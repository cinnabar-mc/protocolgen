// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// WhiskerScopeDataSummary represents a whisker profiler scope diagnostic summary.
type BedrockProfileWhiskerDiagnosticsScopeDataSummary struct {
	// Label is the label of the whisker scope.
	Label string
	// Indentation is the indentation string of the whisker scope within the profiler hierarchy.
	Indentation string
	// TotalHighCostNS is the total time, in nanoseconds, spent in the high-cost portion of the scope.
	TotalHighCostNS uint64
	// TotalMidCostNS is the total time, in nanoseconds, spent in the mid-cost portion of the scope.
	TotalMidCostNS uint64
	// TotalLowCostNS is the total time, in nanoseconds, spent in the low-cost portion of the scope.
	TotalLowCostNS uint64
}

// Marshal reads or writes BedrockProfileWhiskerDiagnosticsScopeDataSummary using its canonical wire layout.
func (x *BedrockProfileWhiskerDiagnosticsScopeDataSummary) Marshal(io IO) {
	io.String(&x.Label)
	io.String(&x.Indentation)
	io.Uint64(&x.TotalHighCostNS)
	io.Uint64(&x.TotalMidCostNS)
	io.Uint64(&x.TotalLowCostNS)
}
