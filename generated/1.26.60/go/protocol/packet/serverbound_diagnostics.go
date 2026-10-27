// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerBoundDiagnostics is sent by the client to tell the server about the performance diagnostics of the
// client. It is sent by the client roughly every 500ms or 10 in-game ticks when the "Creator > Enable Client
// Diagnostics" setting is enabled.
type ServerboundDiagnostics struct {
	// AverageFramesPerSecond is the average amount of frames per second that the client has been running at.
	AvgFps float32
	// AverageServerSimTickTime is the average time that the server spends simulating a single tick in
	// milliseconds.
	AvgServerSimTickTimeMS float32
	// AverageClientSimTickTime is the average time that the client spends simulating a single tick in
	// milliseconds.
	AvgClientSimTickTimeMS float32
	// AverageBeginFrameTime is the average time that the client spends beginning a frame in milliseconds.
	AvgBeginFrameTimeMS float32
	// AverageInputTime is the average time that the client spends processing input in milliseconds.
	AvgInputTimeMS float32
	// AverageRenderTime is the average time that the client spends rendering in milliseconds.
	AvgRenderTimeMS float32
	// AverageEndFrameTime is the average time that the client spends ending a frame in milliseconds.
	AvgEndFrameTimeMS float32
	// AverageRemainderTimePercent is the average percentage of time that the client spends on tasks that are not
	// accounted for.
	AvgRemainderTimePercent float32
	// AverageUnaccountedTimePercent is the average percentage of time that the client spends on unaccounted
	// tasks.
	AvgUnaccountedTimePercent float32
	// MemoryCategoryValues is a list of memory category counters sent by the client.
	MemoryCategoryValues []protocol.MemoryCategoryCounter
	// EntityDiagnostics is a list of entity timing entries sent by the client.
	EntityDiagnostics []protocol.ECSProfilingDiagnosticsEntityDiagnosticTimingInfo
	// SystemDiagnostics is a list of system timing entries sent by the client.
	SystemDiagnostics []protocol.ECSProfilingDiagnosticsSystemDiagnosticTimingInfo
	// SystemCategories maps diagnostics category names to system indices.
	SystemCategories protocol.Optional[[]protocol.ECSProfilingDiagnosticsSystemCategory]
	// WhiskerScopes is a list of whisker profiler scope diagnostic summaries sent by the client.
	WhiskerScopes []protocol.BedrockProfileWhiskerDiagnosticsScopeDataSummary
}

// ID ...
func (*ServerboundDiagnostics) ID() uint32 {
	return IDServerboundDiagnostics
}

func (pk *ServerboundDiagnostics) Marshal(io protocol.IO) {
	io.Float32(&pk.AvgFps)
	io.Float32(&pk.AvgServerSimTickTimeMS)
	io.Float32(&pk.AvgClientSimTickTimeMS)
	io.Float32(&pk.AvgBeginFrameTimeMS)
	io.Float32(&pk.AvgInputTimeMS)
	io.Float32(&pk.AvgRenderTimeMS)
	io.Float32(&pk.AvgEndFrameTimeMS)
	io.Float32(&pk.AvgRemainderTimePercent)
	io.Float32(&pk.AvgUnaccountedTimePercent)
	protocol.Slice(io, &pk.MemoryCategoryValues)
	protocol.Slice(io, &pk.EntityDiagnostics)
	protocol.Slice(io, &pk.SystemDiagnostics)
	protocol.OptionalFunc(io, &pk.SystemCategories, func(value *[]protocol.ECSProfilingDiagnosticsSystemCategory) {
		protocol.Slice(io, value)
	})
	protocol.Slice(io, &pk.WhiskerScopes)
}
