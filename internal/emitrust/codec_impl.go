package emitrust

import (
	"fmt"
	"strings"

	"protocolgen/internal/manifest"
)

func boxedInner(typ string) (string, bool) {
	if strings.HasPrefix(typ, "Box<") && strings.HasSuffix(typ, ">") {
		return typ[len("Box<") : len(typ)-1], true
	}
	return typ, false
}

// emitDefinitionCodec writes the Encode/Decode pair for one named type.
func (e *codecEmitter) emitDefinitionCodec(b *strings.Builder, item definition) error {
	switch item.Kind {
	case manifest.KindStruct:
		return e.emitStructCodec(b, item)
	case manifest.KindUnion:
		return e.emitUnionCodec(b, item)
	case manifest.KindBitset:
		e.emitBitsetCodec(b, item)
		return nil
	case manifest.KindEnum:
		return e.emitEnumCodec(b, item)
	default:
		return nil
	}
}

func (e *codecEmitter) emitStructCodec(b *strings.Builder, item definition) error {
	if item.Tuple {
		e.openEncode(b, item.Name)
		e.stmt(b, "        ", fmt.Sprintf("wire::%s(self.0).encode(writer)", item.WrapperCodec))
		e.closeEncode(b)
		fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", item.Name)
		fmt.Fprintf(b, "        Ok(Self(<wire::%s as wire::Decode>::decode(reader)?.0))\n", item.WrapperCodec)
		b.WriteString("    }\n}\n\n")
		return nil
	}

	e.openEncode(b, item.Name)
	if len(item.Fields) == 0 {
		b.WriteString("        let _ = writer;\n")
	}
	for _, field := range item.Fields {
		if err := e.encode(b, field.Node, "self."+field.Name, "        "); err != nil {
			return fmt.Errorf("%s.%s: %w", item.Name, field.Name, err)
		}
	}
	e.closeEncode(b)

	fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", item.Name)
	if len(item.Fields) == 0 {
		b.WriteString("        let _ = reader;\n")
	}
	for _, field := range item.Fields {
		expr, err := e.decode(field.Node, field.Hint, item.Name+"."+siteField(field.Name), "        ")
		if err != nil {
			return fmt.Errorf("%s.%s: %w", item.Name, field.Name, err)
		}
		if _, boxed := boxedInner(field.Type); boxed {
			expr = "Box::new(" + expr + ")"
		}
		fmt.Fprintf(b, "        let %s = %s;\n", codecBinding(field.Name), expr)
	}
	b.WriteString("        Ok(Self {\n")
	for _, field := range item.Fields {
		fmt.Fprintf(b, "            %s,\n", codecField(field.Name))
	}
	b.WriteString("        })\n    }\n}\n\n")
	return nil
}

func (e *codecEmitter) emitUnionCodec(b *strings.Builder, item definition) error {
	controlCodec := wireCodecType(item.ControlCode)
	if controlCodec == "" {
		return fmt.Errorf("union %s has unsupported control code %q", item.Name, item.ControlCode)
	}

	e.openEncode(b, item.Name)
	e.stmt(b, "        ", fmt.Sprintf("wire::%s(self.discriminant()).encode(writer)", controlCodec))
	b.WriteString("        match self {\n")
	for _, variant := range item.Union {
		switch {
		case len(variant.Fields) != 0:
			names := make([]string, 0, len(variant.Fields))
			for _, field := range variant.Fields {
				names = append(names, codecField(field.Name))
			}
			fmt.Fprintf(b, "            Self::%s { %s } => {\n", variant.Name, strings.Join(names, ", "))
			for _, field := range variant.Fields {
				if err := e.encode(b, field.Node, codecBinding(field.Name), "                "); err != nil {
					return fmt.Errorf("%s::%s.%s: %w", item.Name, variant.Name, field.Name, err)
				}
			}
			b.WriteString("            }\n")
		case variant.Payload != "":
			fmt.Fprintf(b, "            Self::%s(value) => {\n", variant.Name)
			if err := e.encode(b, variant.Node, "value", "                "); err != nil {
				return fmt.Errorf("%s::%s: %w", item.Name, variant.Name, err)
			}
			b.WriteString("            }\n")
		default:
			fmt.Fprintf(b, "            Self::%s => {}\n", variant.Name)
		}
	}
	b.WriteString("        }\n")
	e.closeEncode(b)

	fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", item.Name)
	fmt.Fprintf(b, "        let discriminant = <wire::%s as wire::Decode>::decode(reader)?.0;\n", controlCodec)
	b.WriteString("        Ok(match discriminant {\n")
	for _, variant := range item.Union {
		switch {
		case len(variant.Fields) != 0:
			fmt.Fprintf(b, "            %d => {\n", variant.Discriminant)
			names := make([]string, 0, len(variant.Fields))
			for _, field := range variant.Fields {
				expr, err := e.decode(field.Node, field.Hint, item.Name+"::"+variant.Name+"."+siteField(field.Name), "                ")
				if err != nil {
					return fmt.Errorf("%s::%s.%s: %w", item.Name, variant.Name, field.Name, err)
				}
				if _, boxed := boxedInner(field.Type); boxed {
					expr = "Box::new(" + expr + ")"
				}
				fmt.Fprintf(b, "                let %s = %s;\n", codecBinding(field.Name), expr)
				names = append(names, codecField(field.Name))
			}
			fmt.Fprintf(b, "                Self::%s { %s }\n", variant.Name, strings.Join(names, ", "))
			b.WriteString("            }\n")
		case variant.Payload != "":
			expr, err := e.decode(variant.Node, variant.Hint, item.Name+"::"+variant.Name, "                ")
			if err != nil {
				return fmt.Errorf("%s::%s: %w", item.Name, variant.Name, err)
			}
			if _, boxed := boxedInner(variant.Payload); boxed {
				expr = "Box::new(" + expr + ")"
			}
			fmt.Fprintf(b, "            %d => Self::%s(%s),\n", variant.Discriminant, variant.Name, expr)
		default:
			fmt.Fprintf(b, "            %d => Self::%s,\n", variant.Discriminant, variant.Name)
		}
	}
	fmt.Fprintf(b, "            value => {\n                return Err(wire::DecodeError::UnknownVariant {\n                    type_name: %q,\n                    value: value as i64,\n                })\n            }\n", item.Name)
	b.WriteString("        })\n    }\n}\n\n")
	return nil
}

func (e *codecEmitter) emitBitsetCodec(b *strings.Builder, item definition) {
	e.openEncode(b, item.Name)
	e.stmt(b, "        ", fmt.Sprintf("wire::encode_bitset(writer, &self.0, %d)", item.BitLength))
	e.closeEncode(b)
	fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", item.Name)
	fmt.Fprintf(b, "        Ok(Self(wire::decode_bitset(reader, %d)?))\n", item.BitLength)
	b.WriteString("    }\n}\n\n")
}

// emitEnumCodec relies on the generated From/to_raw pair, which is total: an
// unrecognised discriminant round trips through the Unknown variant rather than
// failing the packet.
func (e *codecEmitter) emitEnumCodec(b *strings.Builder, item definition) error {
	codec := wireCodecType(item.PrimitiveCode)
	if codec == "" {
		return fmt.Errorf("enum %s has unsupported primitive code %q", item.Name, item.PrimitiveCode)
	}
	e.openEncode(b, item.Name)
	e.stmt(b, "        ", fmt.Sprintf("wire::%s(self.to_raw()).encode(writer)", codec))
	e.closeEncode(b)
	fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", item.Name)
	fmt.Fprintf(b, "        Ok(Self::from(<wire::%s as wire::Decode>::decode(reader)?.0))\n", codec)
	b.WriteString("    }\n}\n\n")
	return nil
}

func (e *codecEmitter) emitPacketCodec(b *strings.Builder, info packetInfo) error {
	e.openEncode(b, info.name)
	if len(info.fields) == 0 {
		b.WriteString("        let _ = writer;\n")
	}
	for _, field := range info.fields {
		if err := e.encode(b, field.node, "self."+field.name, "        "); err != nil {
			return fmt.Errorf("%s.%s: %w", info.name, field.name, err)
		}
	}
	e.closeEncode(b)

	fmt.Fprintf(b, "impl wire::Decode for %s {\n    fn decode(reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n", info.name)
	if len(info.fields) == 0 {
		b.WriteString("        let _ = reader;\n")
	}
	for _, field := range info.fields {
		expr, err := e.decode(field.node, field.hint, info.name+"."+siteField(field.name), "        ")
		if err != nil {
			return fmt.Errorf("%s.%s: %w", info.name, field.name, err)
		}
		fmt.Fprintf(b, "        let %s = %s;\n", codecBinding(field.name), expr)
	}
	b.WriteString("        Ok(Self {\n")
	for _, field := range info.fields {
		fmt.Fprintf(b, "            %s,\n", codecField(field.name))
	}
	b.WriteString("        })\n    }\n}\n\n")
	return nil
}

// emitDirectionRegistry emits the packet direction table and the dispatchers
// that use it. Decoding a packet the local role must never receive is rejected
// on the id, before any field is read. Under the codec contract Direction and
// Peer live in the wire runtime.
func emitDirectionRegistry(b *strings.Builder, infos []packetInfo, contract int) {
	if contract > 0 {
		b.WriteString("\n")
	} else {
		b.WriteString(`
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

`)
	}
	b.WriteString("impl PacketId {\n    pub const fn direction(self) -> Direction {\n        match self {\n")
	for _, info := range infos {
		direction := "Bidirectional"
		switch info.packet.Direction {
		case manifest.DirectionClientbound:
			direction = "Clientbound"
		case manifest.DirectionServerbound:
			direction = "Serverbound"
		}
		fmt.Fprintf(b, "            Self::%s => Direction::%s,\n", info.name, direction)
	}
	b.WriteString("        }\n    }\n\n")
	b.WriteString("    /// Every packet id in the canonical manifest, in id order.\n")
	b.WriteString("    pub const ALL: &'static [PacketId] = &[\n")
	for _, info := range infos {
		fmt.Fprintf(b, "        Self::%s,\n", info.name)
	}
	b.WriteString("    ];\n}\n\n")

	b.WriteString("impl Packet {\n    pub const fn id(&self) -> PacketId {\n        match self {\n")
	for _, info := range infos {
		fmt.Fprintf(b, "            Self::%s(..) => PacketId::%s,\n", info.name, info.name)
	}
	b.WriteString("        }\n    }\n\n")

	if contract > 0 {
		b.WriteString("    pub fn encode(&self, writer: &mut wire::Writer<'_>) -> wire::EncodeResult<()> {\n        match self {\n")
	} else {
		b.WriteString("    pub fn encode(&self, writer: &mut wire::Writer) {\n        match self {\n")
	}
	for _, info := range infos {
		receiver := "value"
		if packetNeedsBox(info) {
			receiver = "value.as_ref()"
		}
		fmt.Fprintf(b, "            Self::%s(value) => wire::Encode::encode(%s, writer),\n", info.name, receiver)
	}
	b.WriteString("        }\n    }\n\n")

	b.WriteString(`    /// Decodes a packet body by id, rejecting ids the sender may not use.
    pub fn decode_from(
        id: u32,
        sender: Peer,
        reader: &mut wire::Reader<'_>,
    ) -> wire::DecodeResult<Self> {
        let packet = PacketId::from_raw(id).ok_or(wire::DecodeError::UnknownPacketId(id))?;
        if !packet.direction().permits(sender) {
            return Err(wire::DecodeError::UnexpectedDirection(id));
        }
        Self::decode_body(packet, reader)
    }

    /// Decodes a packet body and rejects bytes left inside the declared entry.
    pub fn decode_exact_from(
        id: u32,
        sender: Peer,
        reader: &mut wire::Reader<'_>,
    ) -> wire::DecodeResult<Self> {
        let packet = Self::decode_from(id, sender, reader)?;
        reader.expect_consumed()?;
        Ok(packet)
    }

`)
	b.WriteString("    pub fn decode_body(id: PacketId, reader: &mut wire::Reader<'_>) -> wire::DecodeResult<Self> {\n        Ok(match id {\n")
	for _, info := range infos {
		decoded := fmt.Sprintf("<%s as wire::Decode>::decode(reader)?", info.name)
		if packetNeedsBox(info) {
			decoded = "Box::new(" + decoded + ")"
		}
		fmt.Fprintf(b, "            PacketId::%s => Self::%s(%s),\n", info.name, info.name, decoded)
	}
	b.WriteString("        })\n    }\n}\n")
}

// codecBinding avoids parameter shadowing with names that fieldName cannot produce.
func codecBinding(name string) string {
	if name == "reader" || name == "writer" {
		return "__" + name
	}
	return name
}

// codecField maps a public field to its local binding in patterns and initializers.
func codecField(name string) string {
	if binding := codecBinding(name); binding != name {
		return name + ": " + binding
	}
	return name
}
