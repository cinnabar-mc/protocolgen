module protocolgen/differential/retail

go 1.26.1

require (
	github.com/go-gl/mathgl v1.1.0
	github.com/google/uuid v1.6.0
	github.com/sandertv/gophertunnel v1.62.0
	protocolgen v0.0.0
)

require (
	github.com/cinnabar-mc/bedrock-net/tools/bench-compare v0.0.0-20261020105954-88ff04df79bc // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/image v0.21.0 // indirect
)

// The fixtures' encoder, as ../gophertunnel.json pins it: the fork commit
// Cinnabar's tools/fixturegen pins at the corpus commit. prepare checks the
// two still agree.
replace github.com/sandertv/gophertunnel => github.com/hashimthearab/gophertunnel v1.25.3-0.20261019151115-85d4022854e1

replace protocolgen => ../..

tool github.com/cinnabar-mc/bedrock-net/tools/bench-compare
