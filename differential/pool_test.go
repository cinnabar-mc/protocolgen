package differential

import (
	"fmt"
	"sort"
	"testing"

	gophertunnelpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	generatedpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

func TestPacketPoolsHaveMatchingIDs(t *testing.T) {
	for _, tc := range []struct {
		name      string
		generated generatedpacket.Pool
		oracle    gophertunnelpacket.Pool
	}{
		{name: "client", generated: generatedpacket.NewClientPool(), oracle: gophertunnelpacket.NewClientPool()},
		{name: "server", generated: generatedpacket.NewServerPool(), oracle: gophertunnelpacket.NewServerPool()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var missing, extra []uint32
			for id := range tc.generated {
				if _, ok := tc.oracle[id]; !ok {
					missing = append(missing, id)
				}
			}
			for id := range tc.oracle {
				if _, ok := tc.generated[id]; !ok {
					extra = append(extra, id)
				}
			}
			for _, id := range extra {
				if !deprecatedOraclePacketIDs[id] {
					missing = append(missing, id)
				}
			}
			extra = filterUnexpectedIDs(extra, deprecatedOraclePacketIDs)
			sort.Slice(missing, func(i, j int) bool { return missing[i] < missing[j] })
			sort.Slice(extra, func(i, j int) bool { return extra[i] < extra[j] })
			if len(missing) != 0 || len(extra) != 0 {
				t.Fatalf("pool mismatch: generated-only=%s oracle-only=%s", formatIDs(missing), formatIDs(extra))
			}
		})
	}
}

var deprecatedOraclePacketIDs = map[uint32]bool{
	55:  true,
	117: true,
	163: true,
	173: true,
	197: true,
}

func filterUnexpectedIDs(ids []uint32, accepted map[uint32]bool) []uint32 {
	result := ids[:0]
	for _, id := range ids {
		if !accepted[id] {
			result = append(result, id)
		}
	}
	return result
}

func formatIDs(ids []uint32) string {
	if len(ids) == 0 {
		return "[]"
	}
	return fmt.Sprint(ids)
}
