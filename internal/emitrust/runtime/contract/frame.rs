// ---------------------------------------------------------------------------
// Packet frames
// ---------------------------------------------------------------------------

/// Leading byte of an uncompressed game batch.
pub const BATCH_HEADER: u8 = 0xfe;

/// Which peer may send a packet.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum Direction {
    Clientbound,
    Serverbound,
    Bidirectional,
}

impl Direction {
    /// Reports whether a packet with this direction may be sent by the peer.
    pub const fn permits(self, sender: Peer) -> bool {
        matches!(
            (self, sender),
            (Self::Bidirectional, _)
                | (Self::Clientbound, Peer::Server)
                | (Self::Serverbound, Peer::Client)
        )
    }
}

/// The peer that produced a packet.
#[derive(Clone, Copy, Debug, Eq, Hash, PartialEq)]
pub enum Peer {
    Client,
    Server,
}

/// The header varint of a batch entry: a ten-bit packet id and two-bit sender
/// and target subclient ids. Higher bits are ignored on decode.
#[derive(Clone, Copy, Debug, Default, Eq, Hash, PartialEq)]
pub struct PacketHeader {
    pub id: u32,
    pub sender_subclient: u8,
    pub target_subclient: u8,
}

impl PacketHeader {
    pub const ID_MASK: u32 = 0x3ff;

    pub const fn new(id: u32) -> Self {
        Self { id, sender_subclient: 0, target_subclient: 0 }
    }

    pub const fn from_raw(raw: u32) -> Self {
        Self {
            id: raw & Self::ID_MASK,
            sender_subclient: ((raw >> 10) & 0x3) as u8,
            target_subclient: ((raw >> 12) & 0x3) as u8,
        }
    }

    pub const fn to_raw(self) -> u32 {
        (self.id & Self::ID_MASK)
            | ((self.sender_subclient as u32 & 0x3) << 10)
            | ((self.target_subclient as u32 & 0x3) << 12)
    }

    fn validate(self) -> EncodeResult<()> {
        if self.id > Self::ID_MASK || self.sender_subclient > 3 || self.target_subclient > 3 {
            return Err(EncodeError::InvalidSubclient {
                sender: self.sender_subclient,
                target: self.target_subclient,
            });
        }
        Ok(())
    }
}

/// One `[length][header][body]` batch entry with its body left undecoded.
///
/// The entry's length prefix fixes its boundary, so a body that fails to
/// decode, or belongs to an unknown packet, never desynchronizes the batch.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct RawFrame {
    header: PacketHeader,
    frame: bytes::Bytes,
    body_start: usize,
}

impl RawFrame {
    pub fn header(&self) -> PacketHeader {
        self.header
    }

    pub fn id(&self) -> u32 {
        self.header.id
    }

    /// Every byte after the header varint, up to the declared length.
    pub fn body(&self) -> bytes::Bytes {
        self.frame.slice(self.body_start..)
    }

    /// Borrows the body while sharing the frame only when a decoded field needs ownership.
    pub(crate) fn body_reader(&self) -> Reader<'_> {
        Reader { buf: &self.frame[self.body_start..], ..Reader::from_shared(&self.frame) }
    }

    /// The complete entry, including its length prefix, for forwarding as is.
    pub fn frame(&self) -> &bytes::Bytes {
        &self.frame
    }

    /// Copies the entry into its own allocation so that retaining it does not
    /// keep the rest of the batch alive.
    pub fn compact(self) -> Self {
        Self { frame: bytes::Bytes::copy_from_slice(&self.frame), ..self }
    }
}

/// Iterates the entries of an uncompressed batch payload without decoding
/// packet bodies. A framing error is yielded once and ends iteration.
#[derive(Clone, Debug)]
pub struct Frames {
    rest: bytes::Bytes,
    limit: usize,
    yielded: usize,
    done: bool,
}

impl Frames {
    /// Reads entries from a payload without the batch header byte, failing
    /// after `max_frames` entries.
    pub fn new(payload: bytes::Bytes, max_frames: usize) -> Self {
        Self { rest: payload, limit: max_frames, yielded: 0, done: false }
    }

    /// Checks and strips the leading [`BATCH_HEADER`] byte.
    pub fn from_batch(mut batch: bytes::Bytes, max_frames: usize) -> DecodeResult<Self> {
        match batch.first().copied() {
            Some(BATCH_HEADER) => Ok(Self::new(batch.split_off(1), max_frames)),
            actual => Err(DecodeError::InvalidBatchHeader(actual)),
        }
    }

    fn next_frame(&mut self) -> DecodeResult<RawFrame> {
        let mut reader = Reader::new(&self.rest);
        let declared = reader.read_var_u32()? as usize;
        let prefix = reader.position();
        if declared > reader.remaining() {
            return Err(DecodeError::UnexpectedEof { needed: declared, remaining: reader.remaining() });
        }
        let frame = self.rest.split_to(prefix + declared);
        let mut header = Reader::new(&frame[prefix..]);
        let raw = header.read_var_u32()?;
        let body_start = prefix + header.position();
        Ok(RawFrame { header: PacketHeader::from_raw(raw), frame, body_start })
    }
}

impl Iterator for Frames {
    type Item = DecodeResult<RawFrame>;

    fn next(&mut self) -> Option<Self::Item> {
        if self.done || self.rest.is_empty() {
            return None;
        }
        if self.yielded == self.limit {
            self.done = true;
            return Some(Err(DecodeError::TooManyFrames { limit: self.limit }));
        }
        let result = self.next_frame();
        match result {
            Ok(_) => self.yielded += 1,
            Err(_) => self.done = true,
        }
        Some(result)
    }
}

/// Appends one `[length][header][body]` entry whose body `body` writes,
/// failing if header and body exceed `limit` bytes. On error `out` is
/// restored to its prior length.
pub fn encode_frame_with(
    out: &mut Vec<u8>,
    header: PacketHeader,
    limit: usize,
    body: impl FnOnce(&mut Writer<'_>) -> EncodeResult<()>,
) -> EncodeResult<()> {
    let start = out.len();
    let result = write_frame(out, header, limit, body);
    if result.is_err() {
        out.truncate(start);
    }
    result
}

fn write_frame(
    out: &mut Vec<u8>,
    header: PacketHeader,
    limit: usize,
    body: impl FnOnce(&mut Writer<'_>) -> EncodeResult<()>,
) -> EncodeResult<()> {
    header.validate()?;
    let start = out.len();
    {
        // The length is unknown until the body is written, so reserve the
        // widest prefix and close the gap afterwards.
        let mut writer = Writer::with_limit(out, limit.saturating_add(MAX_VAR_U32_LEN));
        writer.write_all(&[0; MAX_VAR_U32_LEN])?;
        writer.write_var_u32(header.to_raw())?;
        body(&mut writer)?;
    }
    let declared = out.len() - start - MAX_VAR_U32_LEN;
    let length = u32::try_from(declared).map_err(|_| EncodeError::LengthOverflow(declared))?;
    let mut prefix = [0u8; MAX_VAR_U32_LEN];
    let prefix_len = put_var_uint(&mut prefix, u64::from(length));
    out.copy_within(start + MAX_VAR_U32_LEN.., start + prefix_len);
    out.truncate(start + prefix_len + declared);
    out[start..start + prefix_len].copy_from_slice(&prefix[..prefix_len]);
    Ok(())
}
