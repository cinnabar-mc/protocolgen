//! Inputs and verdict lines, in the format `differential/retail/verdict.go`
//! parses.

use std::fmt::Display;
use std::path::Path;

use bytes::Bytes;
use sha2::{Digest, Sha256};

pub const OK: &str = "ok";
pub const UNKNOWN: &str = "unknown";
pub const ERROR: &str = "error";
pub const FRAME_ERROR: &str = "frame-error";
pub const PANIC: &str = "panic";
pub const SAME: &str = "same";
pub const DIFFERS: &str = "differs";

const MAX_DETAIL_CHARS: usize = 240;

/// Cases whose decoded fields `compare` checks against the generated Go.
const FIELD_CASES: [&str; 3] = ["retail", "crafting", "large"];

#[derive(Clone, Copy, PartialEq, Eq)]
pub enum Sender {
    Server,
    Client,
}

/// One batch from `inputs.tsv`, checked against its SHA-256.
pub struct Input {
    pub fixture: String,
    pub case: String,
    pub sender: Sender,
    pub batch: Bytes,
}

pub fn read_inputs(out: &Path) -> Result<Vec<Input>, String> {
    let index = out.join("inputs.tsv");
    let text = std::fs::read_to_string(&index).map_err(|e| format!("{}: {e}", index.display()))?;
    let mut inputs = Vec::new();
    for (number, line) in text.lines().enumerate() {
        let columns: Vec<&str> = line.split('\t').collect();
        let [fixture, case, sender, sha256, path] = columns.as_slice() else {
            return Err(format!(
                "{}:{}: want 5 columns",
                index.display(),
                number + 1
            ));
        };
        let sender = match *sender {
            "server" => Sender::Server,
            "client" => Sender::Client,
            other => {
                return Err(format!(
                    "{}:{}: sender {other:?}",
                    index.display(),
                    number + 1
                ));
            }
        };
        let data = std::fs::read(out.join(path)).map_err(|e| format!("{path}: {e}"))?;
        if hex(&Sha256::digest(&data)) != *sha256 {
            return Err(format!("{path} does not match its digest"));
        }
        inputs.push(Input {
            fixture: fixture.to_string(),
            case: case.to_string(),
            sender,
            batch: Bytes::from(data),
        });
    }
    Ok(inputs)
}

/// One decoder's result for one batch entry; see `Verdict` in `verdict.go`.
#[derive(Default)]
pub struct Verdict {
    pub fixture: String,
    pub case: String,
    pub entry: usize,
    pub header: Option<String>,
    pub outcome: &'static str,
    pub class: Option<&'static str>,
    pub payload: Option<String>,
    pub fields: Option<String>,
    pub roundtrip: Option<&'static str>,
    pub encoded: Option<String>,
    pub detail: Option<String>,
    /// The decoded packet's `Debug` rendering, which `compare` checks field
    /// by field against the generated Go decoding of the same entry.
    pub debug: Option<String>,
}

impl Verdict {
    pub fn new(input: &Input, case: &str, entry: usize) -> Self {
        Self {
            fixture: input.fixture.clone(),
            case: case.to_string(),
            entry,
            ..Self::default()
        }
    }

    pub fn frame_error(input: &Input, case: &str, entry: usize, detail: impl Display) -> Self {
        Self {
            outcome: FRAME_ERROR,
            detail: Some(detail.to_string()),
            ..Self::new(input, case, entry)
        }
    }

    pub fn fail(&mut self, class: &'static str, detail: impl Display) {
        self.outcome = ERROR;
        self.class = Some(class);
        self.detail = Some(detail.to_string());
    }

    fn line(&self) -> String {
        let column =
            |value: Option<&str>| value.filter(|v| !v.is_empty()).unwrap_or("-").to_string();
        [
            self.fixture.clone(),
            self.case.clone(),
            self.entry.to_string(),
            column(self.header.as_deref()),
            self.outcome.to_string(),
            column(self.class),
            column(self.payload.as_deref()),
            column(self.fields.as_deref()),
            column(self.roundtrip),
            column(self.encoded.as_deref()),
            sanitize(self.detail.as_deref().unwrap_or("")),
        ]
        .join("\t")
    }
}

/// Runs one entry's decode, recording a panic as its own outcome so the
/// remaining entries still run.
pub fn guarded(verdict: &mut Verdict, decode: impl FnOnce(&mut Verdict)) {
    let result = std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| decode(verdict)));
    if let Err(payload) = result {
        let text = payload
            .downcast_ref::<&str>()
            .map(|text| text.to_string())
            .or_else(|| payload.downcast_ref::<String>().cloned())
            .unwrap_or_default();
        verdict.outcome = PANIC;
        verdict.class = None;
        verdict.detail = Some(format!("panicked: {text}"));
    }
}

pub fn write(out: &Path, decoder: &str, verdicts: &[Verdict]) -> Result<(), String> {
    let dir = out.join("verdicts");
    std::fs::create_dir_all(&dir).map_err(|e| format!("{}: {e}", dir.display()))?;
    let mut text = String::new();
    for verdict in verdicts {
        text.push_str(&verdict.line());
        text.push('\n');
    }
    let path = dir.join(format!("{decoder}.tsv"));
    std::fs::write(&path, text).map_err(|e| format!("{}: {e}", path.display()))?;
    let mut fields = String::new();
    for verdict in verdicts {
        let Some(debug) = verdict.debug.as_deref() else {
            continue;
        };
        if FIELD_CASES.contains(&verdict.case.as_str()) {
            fields.push_str(&format!(
                "{}\t{}\t{}\t{debug}\n",
                verdict.fixture, verdict.case, verdict.entry
            ));
        }
    }
    let dir = out.join("fields");
    std::fs::create_dir_all(&dir).map_err(|e| format!("{}: {e}", dir.display()))?;
    let path = dir.join(format!("{decoder}.tsv"));
    std::fs::write(&path, fields).map_err(|e| format!("{}: {e}", path.display()))
}

/// The first eight bytes of SHA-256, in hex.
pub fn digest(data: &[u8]) -> String {
    hex(&Sha256::digest(data)[..8])
}

/// The id and subclient bits of a raw header varint.
pub fn header_text(raw: u32) -> String {
    format!("{}:{}:{}", raw & 0x3ff, (raw >> 10) & 3, (raw >> 12) & 3)
}

fn hex(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect()
}

fn sanitize(detail: &str) -> String {
    let flat: String = detail
        .chars()
        .map(|c| if c < ' ' || c == '\x7f' { ' ' } else { c })
        .take(MAX_DETAIL_CHARS)
        .collect();
    match flat.trim() {
        "" => "-".to_string(),
        trimmed => trimmed.to_string(),
    }
}
