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
		{&packet.PlayerVideoCapture{Action: &protocol.StopVideoCapture{}}, &packet.PlayerVideoCapture{}, "01"},
		{&packet.PlayerVideoCapture{Action: &protocol.StartVideoCapture{FrameRate: 30, FilePrefix: "a"}}, &packet.PlayerVideoCapture{}, "001e0000000161"},
		{&packet.LevelEventGeneric{EventID: 1, CTD: []byte{3, 1, 'x', 2, 0}}, &packet.LevelEventGeneric{}, "020301780200"},
		{&packet.StructureTemplateDataResponse{StructureName: "a", ResponseType: 1}, &packet.StructureTemplateDataResponse{}, "01610001"},
		{&packet.StructureTemplateDataResponse{StructureName: "a", StructureSNBT: protocol.Option([]byte{10, 0, 0}), ResponseType: 1}, &packet.StructureTemplateDataResponse{}, "0161010a000001"},
		{&packet.EducationSettings{AgentCapabilities: protocol.Option(protocol.AgentCapabilities{CanModifyBlocks: protocol.Option(true)})}, &packet.EducationSettings{}, "000000000000010101000000"},
		{&packet.EducationSettings{AgentCapabilities: protocol.Option(protocol.AgentCapabilities{})}, &packet.EducationSettings{}, "0000000000000100000000"},
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

// Test12660SplineEasingBytes checks that both key-frame encodings carry easing as an optional string.
func Test12660SplineEasingBytes(t *testing.T) {
	for _, tc := range []struct {
		value interface{ Marshal(protocol.IO) }
		wire  string
	}{
		{&protocol.CameraSplineProgressKeyFrame{}, "000000000000000000"},
		{&protocol.CameraSplineProgressKeyFrame{Easing: protocol.Option("linear")}, "000000000000000001066c696e656172"},
		{&protocol.CameraSplineRotationKeyFrame{}, "0000000000000000000000000000000000"},
		{&protocol.CameraSplineRotationKeyFrame{Easing: protocol.Option("linear")}, "0000000000000000000000000000000001066c696e656172"},
	} {
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
		tc.value.Marshal(reader)
		if reader.Err() != nil || reader.Remaining() != 0 {
			t.Fatalf("easing decode: %v, remaining %d", reader.Err(), reader.Remaining())
		}
	}
}

// Test12660HeightMapRows preserves each row prefix and rejects rows with the wrong length.
func Test12660HeightMapRows(t *testing.T) {
	var rows [16][]int8
	want := []byte{1, 1}
	for index := range rows {
		rows[index] = make([]int8, 16)
		want = append(want, 16)
		want = append(want, make([]byte, 16)...)
	}
	want = append(want, 0, 0)
	value := protocol.HeightmapData{HeightMapType: 1, SubchunkHeightMap: protocol.Option(rows)}
	writer := protocol.NewWriter()
	value.Marshal(writer)
	if writer.Err() != nil || !bytes.Equal(writer.Data(), want) {
		t.Fatalf("height map encoded %x, want %x: %v", writer.Data(), want, writer.Err())
	}
	var decoded protocol.HeightmapData
	reader := protocol.NewReader(want)
	decoded.Marshal(reader)
	if reader.Err() != nil || reader.Remaining() != 0 || !reflect.DeepEqual(value, decoded) {
		t.Fatalf("height map decode: %v, remaining %d", reader.Err(), reader.Remaining())
	}
	malformed := append([]byte(nil), want...)
	malformed[2] = 15
	reader = protocol.NewReader(malformed)
	decoded.Marshal(reader)
	if reader.Err() == nil {
		t.Fatal("accepted a row with 15 heights")
	}
	rows[0] = rows[0][:15]
	value.SubchunkHeightMap = protocol.Option(rows)
	writer = protocol.NewWriter()
	value.Marshal(writer)
	if writer.Err() == nil {
		t.Fatal("encoded a row with 15 heights")
	}
}

// Test12660SkinImageByteBuffer keeps ordinary RGBA images outside collection limits.
func Test12660SkinImageByteBuffer(t *testing.T) {
	image := protocol.SerializedSkinImage{Width: 64, Height: 64, ImageBytes: bytes.Repeat([]byte{0xff}, 64*64*4)}
	writer := protocol.NewWriter()
	image.Marshal(writer)
	if writer.Err() != nil {
		t.Fatal(writer.Err())
	}
	var decoded protocol.SerializedSkinImage
	reader := protocol.NewReader(writer.Data())
	decoded.Marshal(reader)
	if reader.Err() != nil || reader.Remaining() != 0 || !reflect.DeepEqual(image, decoded) {
		t.Fatalf("ordinary skin image did not round-trip: %v, remaining %d", reader.Err(), reader.Remaining())
	}
}

// Test12660SoundUpdatesHaveOneSelector checks fixed packet bytes for stop and volume updates.
func Test12660SoundUpdatesHaveOneSelector(t *testing.T) {
	for _, tc := range []struct {
		event protocol.SoundDataEvent
		wire  string
	}{
		{&protocol.SoundDataEventStop{}, "080706050403020100"},
		{&protocol.SoundDataEventSetVolume{Volume: 0.5}, "0807060504030201010000003f"},
	} {
		value := &packet.ClientboundUpdateSoundData{ServerSoundHandle: protocol.ServerSoundHandle{ServerSoundHandle: 0x0102030405060708}, Event: tc.event}
		want, err := hex.DecodeString(tc.wire)
		if err != nil {
			t.Fatal(err)
		}
		writer := protocol.NewWriter()
		value.Marshal(writer)
		if writer.Err() != nil || !bytes.Equal(writer.Data(), want) {
			t.Fatalf("encoded %x, want %x: %v", writer.Data(), want, writer.Err())
		}
		decoded := &packet.ClientboundUpdateSoundData{}
		reader := protocol.NewReader(want)
		decoded.Marshal(reader)
		if reader.Err() != nil || reader.Remaining() != 0 || !reflect.DeepEqual(value, decoded) {
			t.Fatalf("sound update did not decode: %v, remaining %d", reader.Err(), reader.Remaining())
		}
	}
}

// Test12660StructureNameBytes checks that a structure name carries its filtered name as an optional string.
func Test12660StructureNameBytes(t *testing.T) {
	for _, tc := range []struct {
		value protocol.StructureRedactableName
		wire  string
	}{
		{protocol.StructureRedactableName{Unredacted: "a"}, "016100"},
		{protocol.StructureRedactableName{Unredacted: "a", Redacted: protocol.Option("b")}, "0161010162"},
	} {
		want, err := hex.DecodeString(tc.wire)
		if err != nil {
			t.Fatal(err)
		}
		writer := protocol.NewWriter()
		tc.value.Marshal(writer)
		if writer.Err() != nil || !bytes.Equal(writer.Data(), want) {
			t.Fatalf("encoded %x, want %x: %v", writer.Data(), want, writer.Err())
		}
		var decoded protocol.StructureRedactableName
		reader := protocol.NewReader(want)
		decoded.Marshal(reader)
		if reader.Err() != nil || reader.Remaining() != 0 || !reflect.DeepEqual(decoded, tc.value) {
			t.Fatalf("decoded %#v, want %#v: %v", decoded, tc.value, reader.Err())
		}
	}
}
