// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type Cancel struct {
	ResponseType string
}

func (*Cancel) tagResourcePackClientResponseData() uint32 { return 0 }

// Marshal reads or writes Cancel using its canonical wire layout.
func (x *Cancel) Marshal(io IO) {
	io.String(&x.ResponseType)
}

type Downloading struct {
	ResponseType     string
	DownloadingPacks []string
}

func (*Downloading) tagResourcePackClientResponseData() uint32 { return 1 }

// Marshal reads or writes Downloading using its canonical wire layout.
func (x *Downloading) Marshal(io IO) {
	io.String(&x.ResponseType)
	FuncSliceLimits(io, &x.DownloadingPacks, io.Varuint32, 0, 65535, io.String)
}

type DownloadingFinished struct {
	ResponseType string
}

func (*DownloadingFinished) tagResourcePackClientResponseData() uint32 { return 2 }

// Marshal reads or writes DownloadingFinished using its canonical wire layout.
func (x *DownloadingFinished) Marshal(io IO) {
	io.String(&x.ResponseType)
}

type ResourcePackClientResponseData interface {
	Marshaler
	tagResourcePackClientResponseData() uint32
}

// MarshalResourcePackClientResponseData reads or writes the ResourcePackClientResponseData union using its canonical wire layout.
func MarshalResourcePackClientResponseData(io IO, x *ResourcePackClientResponseData) {
	Union(io, x, io.Varuint32, ResourcePackClientResponseData.tagResourcePackClientResponseData, func(tag uint32) ResourcePackClientResponseData {
		switch tag {
		case 0:
			return new(Cancel)
		case 1:
			return new(Downloading)
		case 2:
			return new(DownloadingFinished)
		case 3:
			return new(ResourcePackStackFinished)
		}
		return nil
	})
}

type ResourcePackStackFinished struct {
	ResponseType string
}

func (*ResourcePackStackFinished) tagResourcePackClientResponseData() uint32 { return 3 }

// Marshal reads or writes ResourcePackStackFinished using its canonical wire layout.
func (x *ResourcePackStackFinished) Marshal(io IO) {
	io.String(&x.ResponseType)
}
