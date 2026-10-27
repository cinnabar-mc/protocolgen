package roundtrip

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	protocol "protocolgen/generated/1.26.60/go/protocol"
	packet "protocolgen/generated/1.26.60/go/protocol/packet"
)

// Test12660PacketGoldens checks new packet field order and signed encodings against fixed bytes.
func Test12660PacketGoldens(t *testing.T) {
	cases := []struct {
		value packet.Packet
		empty packet.Packet
		wire  string
	}{
		{&packet.ServerboundStonecutterSetRecipe{ContainerID: 7, RecipeIndex: -2}, &packet.ServerboundStonecutterSetRecipe{}, "0703"},
		{&packet.ClientboundStonecutterSetRecipe{PlayerID: -1, ContainerID: 7, RecipeIndex: 130}, &packet.ClientboundStonecutterSetRecipe{}, "01078402"},
		{&packet.SetPassengerOfBlock{Passenger: -1}, &packet.SetPassengerOfBlock{}, "0100"},
		{&packet.ClientboundPlayAudioContent{
			SharedMetadata:  protocol.AudioSignedAudioContent{CompactJWT: "a"},
			PlaybackContent: protocol.AudioSignedAudioContent{CompactJWT: "c"},
			PlaybackType:    "Sound",
			PlaySound: protocol.AudioSound{
				Name: "b", Position: protocol.BlockPosition{X: -1, Y: 2, Z: 3},
				Volume: 1, Pitch: 0.5, LoopCount: -1,
			},
		}, &packet.ClientboundPlayAudioContent{}, "0161016305536f756e6401620104060000803f0000003f01000000"},
	}
	for _, tc := range cases {
		t.Run(reflect.TypeOf(tc.value).Elem().Name(), func(t *testing.T) {
			want, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			writer := protocol.NewWriter()
			tc.value.Marshal(writer)
			if writer.Err() != nil || !bytes.Equal(writer.Data(), want) {
				t.Fatalf("encoded %x, want %x: %v", writer.Data(), want, writer.Err())
			}
			reader := protocol.NewReader(want)
			tc.empty.Marshal(reader)
			if reader.Err() != nil || reader.Remaining() != 0 || !reflect.DeepEqual(tc.value, tc.empty) {
				t.Fatalf("decoded %#v, want %#v: %v, remaining %d", tc.empty, tc.value, reader.Err(), reader.Remaining())
			}
		})
	}
}
