module protocolgen/differential

go 1.26.1

require protocolgen v0.0.0

require (
	github.com/go-gl/mathgl v1.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/sandertv/gophertunnel v1.62.0
	golang.org/x/exp v0.0.0-20260611194520-c48552f49976 // indirect
	golang.org/x/image v0.21.0 // indirect
)

replace protocolgen => ..

replace github.com/sandertv/gophertunnel => github.com/hashimthearab/gophertunnel v1.25.3-0.20261028223324-77149b94f802
