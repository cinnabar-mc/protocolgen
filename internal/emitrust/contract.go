package emitrust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"protocolgen/internal/manifest"
)

// CodecContractVersion is the Rust codec contract (docs/rust-codec-contract.md)
// this emitter implements when Options.CodecContract is set.
const CodecContractVersion = 2

// fieldSites names every length-prefixed decode site so consumers can bound
// one field without bounding the rest. Ids follow emission order.
type fieldSites struct {
	ids    map[string]int
	paths  []string
	consts []string
	used   map[string]bool
}

func newFieldSites() *fieldSites {
	return &fieldSites{ids: map[string]int{}, used: map[string]bool{}}
}

// reference returns the Rust path of the FieldId constant for a site path.
func (s *fieldSites) reference(path string) string {
	index, ok := s.ids[path]
	if !ok {
		index = len(s.paths)
		s.ids[path] = index
		s.paths = append(s.paths, path)
		s.consts = append(s.consts, uniqueField(siteConstName(path), s.used))
	}
	return "crate::fields::" + s.consts[index]
}

// siteConstName upper-snake-cases a site path such as `Text.message[]`.
func siteConstName(path string) string {
	flat := strings.NewReplacer("[]", "_item", "{key}", "_key", "{value}", "_value", "::", "_", ".", "_").Replace(path)
	return strings.ToUpper(strings.TrimSuffix(fieldName(flat), "_"))
}

func emitFields(sites *fieldSites) (string, error) {
	if len(sites.paths) >= 0xffff {
		return "", fmt.Errorf("%d length-prefixed field sites exceed the u16 FieldId space", len(sites.paths))
	}
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\n")
	b.WriteString("//! Length-prefixed field sites, for per-field decode limits.\n\n")
	b.WriteString("use crate::wire::FieldId;\n\n")
	for index, name := range sites.consts {
		fmt.Fprintf(&b, "pub const %s: FieldId = FieldId(%d);\n", name, index)
	}
	b.WriteString("\n/// Site paths, indexed by `FieldId`.\npub const PATHS: &[&str] = &[\n")
	for _, path := range sites.paths {
		fmt.Fprintf(&b, "    %q,\n", path)
	}
	b.WriteString(`];

/// Returns the site path of ` + "`field`" + `, or ` + "`None`" + ` for an id this crate does not define.
pub fn path(field: FieldId) -> Option<&'static str> {
    PATHS.get(usize::from(field.0)).copied()
}

/// Finds the site with this path.
pub fn find(path: &str) -> Option<FieldId> {
    PATHS
        .iter()
        .position(|candidate| *candidate == path)
        .map(|index| FieldId(index as u16))
}
`)
	return b.String(), nil
}

func emitContractLib(m manifest.Manifest) (string, error) {
	stable, err := manifest.MarshalStable(m)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(stable)
	return fmt.Sprintf(`// Code generated from canonical protocol manifest v2. DO NOT EDIT.

// Variant layouts follow the manifest; the emitter boxes by its own size estimate.
#![allow(clippy::large_enum_variant)]

pub const GAME_VERSION: &str = %q;
pub const PROTOCOL_VERSION: i32 = %d;
/// Rust codec contract this crate implements; see protocolgen's
/// `+"`docs/rust-codec-contract.md`"+`.
pub const CODEC_CONTRACT: u32 = %d;
/// SHA-256 of the stable serialization of the manifest this crate was generated from.
pub const MANIFEST_SHA256: &str = %q;

pub mod enums;
pub mod fields;
pub mod types;
pub mod packets;
pub mod wire;
`, m.Target.MinecraftVersion, m.Target.ProtocolVersion, CodecContractVersion, hex.EncodeToString(digest[:])), nil
}

// stmt writes one encode statement, propagating its error under the contract.
func (e *codecEmitter) stmt(b *strings.Builder, indent, call string) {
	if e.g.contract > 0 {
		fmt.Fprintf(b, "%s%s?;\n", indent, call)
		return
	}
	fmt.Fprintf(b, "%s%s;\n", indent, call)
}

// openEncode starts an Encode impl for name.
func (e *codecEmitter) openEncode(b *strings.Builder, name string) {
	if e.g.contract > 0 {
		fmt.Fprintf(b, "impl wire::Encode for %s {\n    fn encode(&self, writer: &mut wire::Writer<'_>) -> wire::EncodeResult<()> {\n", name)
		return
	}
	fmt.Fprintf(b, "impl wire::Encode for %s {\n    fn encode(&self, writer: &mut wire::Writer) {\n", name)
}

// closeEncode ends an Encode impl started by openEncode.
func (e *codecEmitter) closeEncode(b *strings.Builder) {
	if e.g.contract > 0 {
		b.WriteString("        Ok(())\n")
	}
	b.WriteString("    }\n}\n\n")
}

// field returns the FieldId expression for a decode site, or "" outside the
// contract.
func (e *codecEmitter) field(site string) string {
	if e.g.contract == 0 {
		return ""
	}
	return e.g.sites.reference(site)
}

// emitFramePackets emits the frame-level decode and encode entry points: an
// unknown id yields its exact body, and every other failure stays inside its
// frame.
func emitFramePackets(b *strings.Builder) {
	b.WriteString(`
/// A decoded batch entry: a packet this crate defines, or the exact bytes of
/// one it does not.
#[derive(Clone, Debug, PartialEq)]
pub enum Decoded {
    Known(Packet),
    Unknown(UnknownPacket),
}

/// A packet id absent from this manifest, with its entry preserved byte-exact.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct UnknownPacket {
    pub header: wire::PacketHeader,
    pub body: bytes::Bytes,
    /// The complete entry as received. The header keeps only the id and
    /// subclient bits, so forwarding uses these bytes instead.
    pub frame: bytes::Bytes,
}

impl UnknownPacket {
    /// Appends the entry exactly as received; on error ` + "`out`" + ` is unchanged.
    pub fn encode_frame(&self, out: &mut Vec<u8>) -> wire::EncodeResult<()> {
        wire::Writer::new(out).write_all(&self.frame)
    }
}

impl Packet {
    /// Decodes one batch entry sent by ` + "`sender`" + `. Unknown ids yield
    /// [` + "`Decoded::Unknown`" + `]; a known id the sender may not use, a malformed
    /// body, or bytes left inside the entry fail this entry only.
    pub fn decode_frame(
        frame: &wire::RawFrame,
        sender: Peer,
        limits: &wire::Limits,
    ) -> wire::DecodeResult<Decoded> {
        let mut reader = frame.body_reader().with_limits(limits);
        Self::decode_entry(frame.header(), frame.frame(), &mut reader, sender)
    }

    /// Decodes exactly one length-prefixed entry, borrowing the input
    /// until a decoded byte field or unknown packet needs shared ownership.
    pub fn decode_framed(
        frame: &bytes::Bytes,
        sender: Peer,
        limits: &wire::Limits,
    ) -> wire::DecodeResult<Decoded> {
        let mut reader = wire::Reader::from_shared(frame).with_limits(limits);
        let declared = reader.read_var_u32()? as usize;
        if declared > reader.remaining() {
            return Err(wire::DecodeError::UnexpectedEof { needed: declared, remaining: reader.remaining() });
        }
        if declared < reader.remaining() {
            return Err(wire::DecodeError::TrailingBytes(reader.remaining() - declared));
        }
        let header = wire::PacketHeader::from_raw(reader.read_var_u32()?);
        Self::decode_entry(header, frame, &mut reader, sender)
    }

    /// Decodes a bounded entry body and preserves unknown entries exactly.
    fn decode_entry(
        header: wire::PacketHeader,
        frame: &bytes::Bytes,
        reader: &mut wire::Reader<'_>,
        sender: Peer,
    ) -> wire::DecodeResult<Decoded> {
        let Some(id) = PacketId::from_raw(header.id) else {
            let body = reader.take_shared(reader.remaining())?;
            return Ok(Decoded::Unknown(UnknownPacket { header, body, frame: frame.clone() }));
        };
        if !id.direction().permits(sender) {
            return Err(wire::DecodeError::UnexpectedDirection(header.id));
        }
        let packet = Self::decode_body(id, reader)?;
        reader.expect_consumed()?;
        Ok(Decoded::Known(packet))
    }

    /// Appends this packet as one batch entry sent by ` + "`sender`" + `; on error
    /// ` + "`out`" + ` is restored to its prior length.
    pub fn encode_frame(
        &self,
        sender: Peer,
        sender_subclient: u8,
        target_subclient: u8,
        out: &mut Vec<u8>,
    ) -> wire::EncodeResult<()> {
        let id = self.id();
        if !id.direction().permits(sender) {
            return Err(wire::EncodeError::UnexpectedDirection(id as u32));
        }
        let header = wire::PacketHeader { id: id as u32, sender_subclient, target_subclient };
        wire::encode_frame_with(out, header, usize::MAX, |writer| self.encode(writer))
    }
}
`)
}

// emitContractRoundtripTest checks every packet's default through both the
// value and frame entry points, plus direction and unknown-id handling.
func emitContractRoundtripTest(m manifest.Manifest, infos []packetInfo) string {
	crate := strings.ReplaceAll(rustPackageName(m, CodecContractVersion), "-", "_")
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "use %s::packets::*;\n", crate)
	fmt.Fprintf(&b, "use %s::wire::{self, Decode, Encode, Peer};\n\n", crate)
	b.WriteString(`fn roundtrip<T>(name: &str)
where
    T: Encode + Decode + Default + PartialEq + std::fmt::Debug + Into<Packet>,
{
    let value = T::default();
    let bytes = match value.encode_to_vec() {
        Ok(bytes) => bytes,
        Err(wire::EncodeError::SchemaConstraint(_)) => return,
        Err(error) => panic!("{name} could not encode its default: {error}"),
    };
    match T::decode_exact(&bytes) {
        Ok(decoded) => assert_eq!(decoded, value, "{name} did not survive a round trip"),
        Err(error) => panic!("{name} could not decode its own encoding: {error}"),
    }

    let packet: Packet = value.into();
    let sender = if packet.id().direction().permits(Peer::Server) {
        Peer::Server
    } else {
        Peer::Client
    };
    let mut out = Vec::new();
    packet
        .encode_frame(sender, 1, 2, &mut out)
        .unwrap_or_else(|error| panic!("{name} could not encode a frame: {error}"));
    let frames = wire::Frames::new(out.into(), 1)
        .collect::<Result<Vec<_>, _>>()
        .unwrap_or_else(|error| panic!("{name} framing failed: {error}"));
    assert_eq!(frames.len(), 1, "{name} did not encode exactly one frame");
    assert_eq!(frames[0].header().sender_subclient, 1);
    assert_eq!(frames[0].header().target_subclient, 2);
    let decoded = Packet::decode_frame(&frames[0], sender, &wire::Limits::receiving_from(sender));
    let direct = Packet::decode_framed(frames[0].frame(), sender, &wire::Limits::receiving_from(sender));
    assert_eq!(direct, decoded, "{name} borrowed and owned frames differ");
    let wrong_sender = match sender { Peer::Client => Peer::Server, Peer::Server => Peer::Client };
    if !packet.id().direction().permits(wrong_sender) {
        assert_eq!(
            Packet::decode_framed(frames[0].frame(), wrong_sender, &wire::Limits::unbounded()),
            Err(wire::DecodeError::UnexpectedDirection(packet.id() as u32)),
        );
    }
    assert_eq!(decoded, Ok(Decoded::Known(packet)), "{name} did not survive a frame round trip");
}

#[test]
fn every_schema_valid_packet_default_round_trips() {
`)
	for _, info := range infos {
		fmt.Fprintf(&b, "    roundtrip::<%s>(%q);\n", info.name, info.name)
	}
	b.WriteString("}\n")
	b.WriteString(`
/// The direction table must reject a packet from the peer that cannot send it,
/// on the id, before any field is read.
#[test]
fn every_packet_rejects_the_wrong_sender() {
    for &id in PacketId::ALL {
        let raw = id as u32;
        let wrong = match id.direction() {
            wire::Direction::Bidirectional => continue,
            wire::Direction::Clientbound => Peer::Client,
            wire::Direction::Serverbound => Peer::Server,
        };
        let mut reader = wire::Reader::new(&[]);
        assert!(
            matches!(
                Packet::decode_from(raw, wrong, &mut reader),
                Err(wire::DecodeError::UnexpectedDirection(_))
            ),
            "{id:?} accepted a packet from the wrong peer"
        );
    }
}

/// An id this manifest lacks keeps its exact body, and the next entry still
/// decodes from its own boundary.
#[test]
fn unknown_packets_keep_exact_boundaries() {
    let unknown = (0..=wire::PacketHeader::ID_MASK)
        .find(|id| PacketId::from_raw(*id).is_none())
        .expect("a free packet id");
    let mut batch = vec![wire::BATCH_HEADER];
    for body in [&[1u8, 2, 3][..], &[][..]] {
        wire::encode_frame_with(&mut batch, wire::PacketHeader::new(unknown), usize::MAX, |writer| {
            writer.write_all(body)
        })
        .unwrap();
    }
    let limits = wire::Limits::unbounded();
    let decoded = wire::Frames::from_batch(batch.into(), 16)
        .unwrap()
        .map(|frame| Packet::decode_frame(&frame.unwrap(), Peer::Server, &limits).unwrap())
        .collect::<Vec<_>>();
    assert_eq!(decoded.len(), 2);
    match &decoded[0] {
        Decoded::Unknown(packet) => {
            assert_eq!(packet.header.id, unknown);
            assert_eq!(packet.body.as_ref(), &[1, 2, 3]);
        }
        other => panic!("unknown id decoded as {other:?}"),
    }
    assert!(matches!(&decoded[1], Decoded::Unknown(packet) if packet.body.is_empty()));
}

/// Borrowed single-entry decoding validates both prefixes and the exact entry boundary.
#[test]
fn framed_decode_rejects_bad_boundaries() {
    let limits = wire::Limits::unbounded();
    for bytes in [
        &[0x80][..],
        &[0x80, 0],
        &[0xff, 0xff, 0xff, 0xff, 0x10],
        &[0],
        &[2, 1],
        &[1, 1, 0],
        &[2, 0x80, 0],
        &[5, 0xff, 0xff, 0xff, 0xff, 0x10],
    ] {
        assert!(Packet::decode_framed(&bytes::Bytes::copy_from_slice(bytes), Peer::Server, &limits).is_err(), "{bytes:?}");
    }
}

/// Header bits above the target subclient are ignored on decode but must
/// survive forwarding an unknown entry.
#[test]
fn unknown_packets_forward_byte_for_byte() {
    let unknown = (0..=wire::PacketHeader::ID_MASK)
        .find(|id| PacketId::from_raw(*id).is_none())
        .expect("a free packet id");
    let mut header = Vec::new();
    wire::Writer::new(&mut header).write_var_u32(unknown | (1 << 14)).unwrap();
    let mut entry = Vec::new();
    let mut writer = wire::Writer::new(&mut entry);
    writer.write_var_u32(header.len() as u32 + 1).unwrap();
    writer.write_all(&header).unwrap();
    writer.write_u8(0xab).unwrap();
    let frame = wire::Frames::new(entry.clone().into(), 1).next().unwrap().unwrap();
    let decoded = Packet::decode_frame(&frame, Peer::Server, &wire::Limits::unbounded());
    assert_eq!(Packet::decode_framed(frame.frame(), Peer::Server, &wire::Limits::unbounded()), decoded);
    let Ok(Decoded::Unknown(packet)) = &decoded else {
        panic!("unknown id decoded as {decoded:?}");
    };
    assert_eq!(packet.header.id, unknown);
    let mut out = Vec::new();
    packet.encode_frame(&mut out).unwrap();
    assert_eq!(out, entry);
}
`)
	return b.String()
}

// fieldArg returns the leading FieldId argument for a runtime decode helper,
// or "" outside the contract.
func (e *codecEmitter) fieldArg(site string) string {
	if e.g.contract == 0 {
		return ""
	}
	return e.field(site) + ", "
}

// allocVec returns the storage expression for a decoded collection of count
// elements.
func (e *codecEmitter) allocVec() string {
	if e.g.contract > 0 {
		return "wire::alloc_vec(count)?"
	}
	return "Vec::with_capacity(count)"
}

// pushItem returns the statement appending one decoded element to out.
func (e *codecEmitter) pushItem(item string) string {
	if e.g.contract > 0 {
		return "wire::push(&mut out, " + item + ")?"
	}
	return "out.push(" + item + ")"
}

// siteField is the field segment of a site path.
func siteField(name string) string {
	return strings.TrimPrefix(name, "r#")
}
