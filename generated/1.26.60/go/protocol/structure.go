// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

type StructureBlockType int32

const (
	StructureBlockTypeData    StructureBlockType = 0
	StructureBlockTypeSave    StructureBlockType = 1
	StructureBlockTypeLoad    StructureBlockType = 2
	StructureBlockTypeCorner  StructureBlockType = 3
	StructureBlockTypeInvalid StructureBlockType = 4
	StructureBlockTypeExport  StructureBlockType = 5
)

// Marshal reads or writes StructureBlockType through its int32 wire encoding.
func (x *StructureBlockType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type StructureEditorData struct {
	StructureName         string
	FilteredStructureName string
	DataField             string
	ShouldIncludePlayers  bool
	ShouldShowBoundingBox bool
	StructureBlockType    StructureBlockType
	StructureSettings     StructureSettings
	RedstoneSaveMode      StructureRedstoneSaveMode
}

// Marshal reads or writes StructureEditorData using its canonical wire layout.
func (x *StructureEditorData) Marshal(io IO) {
	io.String(&x.StructureName)
	io.String(&x.FilteredStructureName)
	io.String(&x.DataField)
	io.Bool(&x.ShouldIncludePlayers)
	io.Bool(&x.ShouldShowBoundingBox)
	x.StructureBlockType.Marshal(io)
	x.StructureSettings.Marshal(io)
	x.RedstoneSaveMode.Marshal(io)
}

type StructurePlacementSettings struct {
	StructurePaletteName                            string
	ShouldIgnoreEntities                            bool
	ShouldIgnoreBlocks                              bool
	ShouldAllowNonTickingPlayerAndTickingAreaChunks bool
	StructureSize                                   BlockPosition
	StructureOffset                                 BlockPosition
	LastEditPlayer                                  ActorUniqueIdentifier
	Rotation                                        StructureRotation
	Mirror                                          StructureMirror
	AnimationMode                                   SkinAnimationMode
	AnimationSeconds                                float32
	IntegrityValue                                  float32
	IntegritySeed                                   uint32
	RotationPivot                                   Vec3
}

// Marshal reads or writes StructurePlacementSettings using its canonical wire layout.
func (x *StructurePlacementSettings) Marshal(io IO) {
	io.StringLimits(&x.StructurePaletteName, 0, 256)
	io.Bool(&x.ShouldIgnoreEntities)
	io.Bool(&x.ShouldIgnoreBlocks)
	io.Bool(&x.ShouldAllowNonTickingPlayerAndTickingAreaChunks)
	x.StructureSize.Marshal(io)
	x.StructureOffset.Marshal(io)
	x.LastEditPlayer.Marshal(io)
	x.Rotation.Marshal(io)
	x.Mirror.Marshal(io)
	x.AnimationMode.Marshal(io)
	io.Float32(&x.AnimationSeconds)
	io.Float32(&x.IntegrityValue)
	io.Uint32(&x.IntegritySeed)
	x.RotationPivot.Marshal(io)
}

type StructureRedstoneSaveMode uint8

const (
	StructureRedstoneSaveModeSavesToMemory StructureRedstoneSaveMode = 0
	StructureRedstoneSaveModeSavesToDisk   StructureRedstoneSaveMode = 1
)

// Marshal reads or writes StructureRedstoneSaveMode through its uint8 wire encoding.
func (x *StructureRedstoneSaveMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// StructureSettings is a struct holding settings of a structure block. Its fields may be changed using the
// in-game UI on the client-side.
type StructureSettings struct {
	// PaletteName is the name of the palette used in the structure. Currently, it seems that this field is always
	// 'default'.
	StructurePaletteName string
	// IgnoreEntities specifies if the structure should ignore entities or include them. If set to false, entities
	// will also show up in the exported structure.
	ShouldIgnoreEntities bool
	// IgnoreBlocks specifies if the structure should ignore blocks or include them. If set to false, blocks will
	// show up in the exported structure.
	ShouldIgnoreBlocks bool
	// AllowNonTickingChunks specifies if the structure should allow non-ticking chunks. If set to false, the
	// structure will export non-ticking chunks.
	ShouldAllowNonTickingPlayerAndTickingAreaChunks bool
	// Size is the size of the area that is about to be exported. The area exported will start at the Position +
	// Offset, and will extend as far as Size specifies.
	StructureSize BlockPos
	// Offset is the offset position that was set in the structure block. The area exported is offset by this
	// position.
	StructureOffset BlockPos
	// LastEditingPlayerUniqueID is the unique ID of the player that last edited the structure block that these
	// settings concern.
	LastEditPlayer int64
	// Rotation is the rotation that the structure block should obtain. See the constants above for available
	// options.
	Rotation Rotation
	// Mirror specifies the way the structure should be mirrored. It is either no mirror at all, mirror on the x/z
	// axis or both.
	Mirror Mirror
	// AnimationMode ...
	AnimationMode AnimationMode
	// AnimationDuration ...
	AnimationSeconds float32
	// Integrity is usually 1, but may be set to a number between 0 and 1 to omit blocks randomly, using the Seed
	// that follows.
	IntegrityValue float32
	// Seed is the seed used to omit blocks if Integrity is not equal to one. If the Seed is 0, a random seed is
	// selected to omit blocks.
	IntegritySeed uint32
	// Pivot is the pivot around which the structure may be rotated.
	RotationPivot mgl32.Vec3
}

// Marshal reads or writes StructureSettings using its canonical wire layout.
func (x *StructureSettings) Marshal(io IO) {
	io.StringLimits(&x.StructurePaletteName, 0, 256)
	io.Bool(&x.ShouldIgnoreEntities)
	io.Bool(&x.ShouldIgnoreBlocks)
	io.Bool(&x.ShouldAllowNonTickingPlayerAndTickingAreaChunks)
	x.StructureSize.Marshal(io)
	x.StructureOffset.Marshal(io)
	io.ActorUniqueID(&x.LastEditPlayer)
	x.Rotation.Marshal(io)
	x.Mirror.Marshal(io)
	x.AnimationMode.Marshal(io)
	io.Float32(&x.AnimationSeconds)
	io.Float32(&x.IntegrityValue)
	io.Uint32(&x.IntegritySeed)
	io.Vec3(&x.RotationPivot)
}

type StructureTemplateRequestOperation uint8

const (
	StructureTemplateRequestOperationNone                StructureTemplateRequestOperation = 0
	StructureTemplateRequestOperationExportFromSaveMode  StructureTemplateRequestOperation = 1
	StructureTemplateRequestOperationExportFromLoadMode  StructureTemplateRequestOperation = 2
	StructureTemplateRequestOperationQuerySavedStructure StructureTemplateRequestOperation = 3
)

// Marshal reads or writes StructureTemplateRequestOperation through its uint8 wire encoding.
func (x *StructureTemplateRequestOperation) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type StructureTemplateResponseType uint8

const (
	StructureTemplateResponseTypeNone   StructureTemplateResponseType = 0
	StructureTemplateResponseTypeExport StructureTemplateResponseType = 1
	StructureTemplateResponseTypeQuery  StructureTemplateResponseType = 2
)

// Marshal reads or writes StructureTemplateResponseType through its uint8 wire encoding.
func (x *StructureTemplateResponseType) Marshal(io IO) { io.Uint8((*uint8)(x)) }
