package emitrust

import (
	"fmt"
	"strings"

	"protocolgen/internal/manifest"
)

// emitRoundtripTest emits a conformance test asserting that every packet
// decodes what it encodes. It is the cheapest check that the two directions of
// a generated codec did not drift apart.
func emitRoundtripTest(m manifest.Manifest, infos []packetInfo) string {
	crate := strings.ReplaceAll(rustPackageName(m, 0), "-", "_")
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "use %s::packets::*;\n", crate)
	fmt.Fprintf(&b, "use %s::wire::{self, Decode, Encode};\n\n", crate)
	b.WriteString(`fn roundtrip<T>(name: &str)
where
    T: Encode + Decode + Default + PartialEq + std::fmt::Debug,
{
    let value = T::default();
    let bytes = match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| value.encode_to_vec())) {
        Ok(bytes) => bytes,
        Err(payload) if schema_constraint_panic(payload.as_ref()) => return,
        Err(payload) => std::panic::resume_unwind(payload),
    };
    match T::decode_exact(&bytes) {
        Ok(decoded) => assert_eq!(decoded, value, "{name} did not survive a round trip"),
        Err(error) => panic!("{name} could not decode its own encoding: {error}"),
    }
}

fn schema_constraint_panic(payload: &(dyn std::any::Any + Send)) -> bool {
    let message = payload
        .downcast_ref::<&str>()
        .copied()
        .or_else(|| payload.downcast_ref::<String>().map(String::as_str));
    matches!(
        message,
        Some(
            "length outside schema limits"
                | "number outside schema limits"
                | "string does not match schema pattern"
        )
    )
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
            Direction::Bidirectional => continue,
            Direction::Clientbound => Peer::Client,
            Direction::Serverbound => Peer::Server,
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
`)
	return b.String()
}
