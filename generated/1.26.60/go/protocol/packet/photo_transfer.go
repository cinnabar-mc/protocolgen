// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PhotoTransfer is sent by the server to transfer a photo (image) file to the client. It is typically used to
// transfer photos so that the client can display it in a portfolio in Education Edition. While previously
// usable in the default Bedrock Edition, the displaying of photos in books was disabled and the packet now
// has little use anymore.
type PhotoTransfer struct {
	// PhotoName is the name of the photo to transfer. It is the exact file name that the client will download the
	// photo as, including the extension of the file.
	PhotoName string
	// PhotoData is the raw data of the photo image. The format of this data may vary: Formats such as JPEG or PNG
	// work, as long as PhotoName has the correct extension.
	PhotoData []byte
	// BookID is the ID of the book that the photo is associated with. If the PhotoName in a book with this ID is
	// set to PhotoName, it will display the photo (provided Education Edition is used). The photo image is
	// downloaded to a sub-folder with this book ID.
	BookID string
	// Type is one of the three photo types above.
	Type protocol.PhotoType
	// SourceType is the source photo type. It is one of the three photo types above.
	SourceType protocol.PhotoType
	// OwnerID is the entity unique ID of the photo's owner.
	OwnerID int64
	// NewPhotoName is the new name of the photo.
	NewPhotoName string
}

// ID ...
func (*PhotoTransfer) ID() uint32 {
	return IDPhotoTransfer
}

func (pk *PhotoTransfer) Marshal(io protocol.IO) {
	io.String(&pk.PhotoName)
	protocol.Pattern(io, &pk.PhotoName, "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\\.jpeg$")
	io.ByteSliceLimits(&pk.PhotoData, 0, 20971520)
	io.String(&pk.BookID)
	pk.Type.Marshal(io)
	pk.SourceType.Marshal(io)
	io.Int64(&pk.OwnerID)
	io.String(&pk.NewPhotoName)
}
