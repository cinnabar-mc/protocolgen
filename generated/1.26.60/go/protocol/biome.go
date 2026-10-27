// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// BiomeCappedSurface specifies the materials to use for the capped surface of a biome, such as in the Nether.
type BiomeCappedSurfaceData struct {
	// FloorBlocks is a list of runtime IDs to use for the floor blocks.
	FloorBlocks []uint32
	// CeilingBlocks is a list of runtime IDs to use for the ceiling blocks.
	CeilingBlocks []uint32
	// SeaBlock is an optional runtime ID to use for the sea block.
	SeaBlock Optional[uint32]
	// FoundationBlock is an optional runtime ID to use for the foundation block.
	FoundationBlock Optional[uint32]
	// BeachBlock is an optional runtime ID to use for the beach block.
	BeachBlock Optional[uint32]
}

// Marshal reads or writes BiomeCappedSurfaceData using its canonical wire layout.
func (x *BiomeCappedSurfaceData) Marshal(io IO) {
	FuncSlice(io, &x.FloorBlocks, io.Varuint32, io.Uint32)
	FuncSlice(io, &x.CeilingBlocks, io.Varuint32, io.Uint32)
	OptionalFunc(io, &x.SeaBlock, io.Uint32)
	OptionalFunc(io, &x.FoundationBlock, io.Uint32)
	OptionalFunc(io, &x.BeachBlock, io.Uint32)
}

// BiomeClimate represents the climate of a biome, mainly for ambience but also defines certain behaviours.
type BiomeClimateData struct {
	// Temperature is the temperature of the biome, used for weather, biome behaviours and sky colour.
	Temperature float32
	// Downfall is the amount that precipitation affects colours and block changes.
	Downfall float32
	// SnowAccumulationMin is the minimum amount of snow that can accumulate in the biome, every 0.125 is another
	// layer of snow.
	SnowAccumulationMin float32
	// SnowAccumulationMax is the maximum amount of snow that can accumulate in the biome, every 0.125 is another
	// layer of snow.
	SnowAccumulationMax float32
}

// Marshal reads or writes BiomeClimateData using its canonical wire layout.
func (x *BiomeClimateData) Marshal(io IO) {
	io.Float32(&x.Temperature)
	io.Float32(&x.Downfall)
	io.Float32(&x.SnowAccumulationMin)
	io.Float32(&x.SnowAccumulationMax)
}

// BiomeConditionalTransformation is the legacy method of transforming biomes.
type BiomeConditionalTransformationData struct {
	// WeightedBiomes is a list of biomes and their weights.
	TransformsInto []BiomeWeightedData
	// ConditionJSON is an index of the condition JSON data in the string list.
	ConditionJSON uint16
	// MinPassingNeighbours is the minimum number of neighbours that must pass the condition for the
	// transformation to be applied.
	MinPassingNeighbors uint32
}

// Marshal reads or writes BiomeConditionalTransformationData using its canonical wire layout.
func (x *BiomeConditionalTransformationData) Marshal(io IO) {
	Slice(io, &x.TransformsInto)
	io.Uint16(&x.ConditionJSON)
	io.Uint32(&x.MinPassingNeighbors)
}

// BiomeConsolidatedFeature represents a feature that is consolidated into a single feature for the biome.
type BiomeConsolidatedFeatureData struct {
	// Scatter defines how the feature is scattered in the biome.
	Scatter BiomeScatterParamData
	// Feature is the index of the feature's name in the string list.
	Feature uint16
	// Identifier is the index of the feature's identifier in the string list.
	Identifier uint16
	// Pass is the index of the feature's pass in the string list.
	Pass uint16
	// CanUseInternal is true if the feature can use internal features.
	CanUseInternalFeature bool
}

// Marshal reads or writes BiomeConsolidatedFeatureData using its canonical wire layout.
func (x *BiomeConsolidatedFeatureData) Marshal(io IO) {
	x.Scatter.Marshal(io)
	io.Uint16(&x.Feature)
	io.Uint16(&x.Identifier)
	io.Uint16(&x.Pass)
	io.Bool(&x.CanUseInternalFeature)
}

type BiomeConsolidatedFeaturesData struct {
	Features []BiomeConsolidatedFeatureData
}

// Marshal reads or writes BiomeConsolidatedFeaturesData using its canonical wire layout.
func (x *BiomeConsolidatedFeaturesData) Marshal(io IO) {
	Slice(io, &x.Features)
}

// BiomeCoordinate specifies coordinate rules for where features can be scattered in the biome.
type BiomeCoordinateData struct {
	// MinValueType is the type of expression operation to use for the minimum value, and is one of the
	// BiomeExpressionOp constants above.
	MinValueType int32
	// MinValue is the index of the minimum value expression in the string list.
	MinValue uint16
	// MaxValueType is the type of expression operation to use for the maximum value, and is one of the
	MaxValueType int32
	// MaxValue is the index of the maximum value expression in the string list.
	MaxValue uint16
	// GridOffset is the offset of the grid, used for fixed grid and jittered grid distributions.
	GridOffset uint32
	// GridStepSize is the step size of the grid, used for fixed grid and jittered grid distributions.
	GridStepSize uint32
	// Distribution is the type of distribution to use for the coordinate, and is one of the
	// BiomeRandomDistributionType constants above.
	Distribution RandomDistributionType
}

// Marshal reads or writes BiomeCoordinateData using its canonical wire layout.
func (x *BiomeCoordinateData) Marshal(io IO) {
	io.Varint32(&x.MinValueType)
	io.Uint16(&x.MinValue)
	io.Varint32(&x.MaxValueType)
	io.Uint16(&x.MaxValue)
	io.Uint32(&x.GridOffset)
	io.Uint32(&x.GridStepSize)
	x.Distribution.Marshal(io)
}

type BiomeDefinitionChunkGenData struct {
	Climate                    Optional[BiomeClimateData]
	ConsolidatedFeatures       Optional[BiomeConsolidatedFeaturesData]
	MountainParams             Optional[BiomeMountainParamsData]
	SurfaceMaterialAdjustments Optional[BiomeSurfaceMaterialAdjustmentData]
	OverworldGenRules          Optional[BiomeOverworldGenRulesData]
	MultinoiseGenRules         Optional[BiomeMultinoiseGenRulesData]
	LegacyWorldGenRules        Optional[BiomeLegacyWorldGenRulesData]
	ReplacementBiomes          Optional[BiomeReplacementsData]
	VillageType                Optional[VillageType]
	SurfaceBuilderData         Optional[BiomeSurfaceBuilderData]
	SubsurfaceBuilderData      Optional[BiomeSurfaceBuilderData]
}

// Marshal reads or writes BiomeDefinitionChunkGenData using its canonical wire layout.
func (x *BiomeDefinitionChunkGenData) Marshal(io IO) {
	OptionalMarshaler(io, &x.Climate)
	OptionalMarshaler(io, &x.ConsolidatedFeatures)
	OptionalMarshaler(io, &x.MountainParams)
	OptionalMarshaler(io, &x.SurfaceMaterialAdjustments)
	OptionalMarshaler(io, &x.OverworldGenRules)
	OptionalMarshaler(io, &x.MultinoiseGenRules)
	OptionalMarshaler(io, &x.LegacyWorldGenRules)
	OptionalMarshaler(io, &x.ReplacementBiomes)
	OptionalMarshaler(io, &x.VillageType)
	OptionalMarshaler(io, &x.SurfaceBuilderData)
	OptionalMarshaler(io, &x.SubsurfaceBuilderData)
}

// BiomeDefinition represents a biome definition in the game. This can be a vanilla biome or a completely
// custom biome.
type BiomeDefinitionData struct {
	ID uint16
	// Temperature is the temperature of the biome, used for weather, biome behaviours and sky colour.
	Temperature float32
	// Downfall is the amount that precipitation affects colours and block changes.
	Downfall float32
	// FoliageSnow is the progression factor for foliage turning white due to snow.
	FoliageSnow float32
	// Depth is the depth of the biome.
	Depth float32
	// Scale is the scale of the biome.
	Scale             float32
	MapWaterColorARGB int32
	// Rain is true if the biome has rain, false if it is a dry biome.
	Rain bool
	// Tags are a list of indices of tags in the string list. These are used to group biomes together for biome
	// generation and other purposes.
	Tags         Optional[BiomeTagsData]
	ChunkGenData Optional[BiomeDefinitionChunkGenData]
}

// Marshal reads or writes BiomeDefinitionData using its canonical wire layout.
func (x *BiomeDefinitionData) Marshal(io IO) {
	io.Uint16(&x.ID)
	io.Float32(&x.Temperature)
	io.Float32(&x.Downfall)
	io.Float32(&x.FoliageSnow)
	io.Float32(&x.Depth)
	io.Float32(&x.Scale)
	io.Int32(&x.MapWaterColorARGB)
	io.Bool(&x.Rain)
	OptionalMarshaler(io, &x.Tags)
	OptionalMarshaler(io, &x.ChunkGenData)
}

// BiomeElementData are set rules to adjust the surface materials of the biome.
type BiomeElementData struct {
	// NoiseFrequencyScale is the frequency scale of the noise used to adjust the surface materials.
	NoiseFreqScale float32
	// NoiseLowerBound is the minimum noise value required to be selected.
	NoiseLowerBound float32
	// NoiseUpperBound is the maximum noise value required to be selected.
	NoiseUpperBound float32
	// HeightMinType is the type of expression operation to use for the minimum height, and is one of the
	// BiomeExpressionOp constants above.
	HeightMinType int32
	// HeightMin is the index of the minimum height expression in the string list.
	HeightMin uint16
	// HeightMaxType is the type of expression operation to use for the maximum height, and is one of the
	// BiomeExpressionOp constants above.
	HeightMaxType int32
	// HeightMax is the index of the maximum height expression in the string list.
	HeightMax uint16
	// AdjustedMaterials is the materials to use for the surface layers of the biome if selected.
	AdjustedMaterials BiomeSurfaceMaterialData
}

// Marshal reads or writes BiomeElementData using its canonical wire layout.
func (x *BiomeElementData) Marshal(io IO) {
	io.Float32(&x.NoiseFreqScale)
	io.Float32(&x.NoiseLowerBound)
	io.Float32(&x.NoiseUpperBound)
	io.Varint32(&x.HeightMinType)
	io.Uint16(&x.HeightMin)
	io.Varint32(&x.HeightMaxType)
	io.Uint16(&x.HeightMax)
	x.AdjustedMaterials.Marshal(io)
}

type BiomeLegacyWorldGenRulesData struct {
	LegacyPreHillsEdge []BiomeConditionalTransformationData
}

// Marshal reads or writes BiomeLegacyWorldGenRulesData using its canonical wire layout.
func (x *BiomeLegacyWorldGenRulesData) Marshal(io IO) {
	Slice(io, &x.LegacyPreHillsEdge)
}

// BiomeMesaSurface specifies the materials to use for the mesa biome.
type BiomeMesaSurfaceData struct {
	// ClayMaterial is the runtime ID of the block to use for clay layers.
	ClayMaterial uint32
	// HardClayMaterial is the runtime ID of the block to use for hard clay layers.
	HardClayMaterial uint32
	// BrycePillars is true if the biome has bryce pillars, which are tall spire-like structures.
	BrycePillars bool
	// HasForest is true if the biome has a forest.
	HasForest bool
}

// Marshal reads or writes BiomeMesaSurfaceData using its canonical wire layout.
func (x *BiomeMesaSurfaceData) Marshal(io IO) {
	io.Uint32(&x.ClayMaterial)
	io.Uint32(&x.HardClayMaterial)
	io.Bool(&x.BrycePillars)
	io.Bool(&x.HasForest)
}

// BiomeMountainParameters specifies the parameters for a mountain biome.
type BiomeMountainParamsData struct {
	// SteepBlock is the runtime ID of the block to use for steep slopes.
	SteepBlock uint32
	// NorthSlopes is true if the biome has north slopes.
	NorthSlopes bool
	// SouthSlopes is true if the biome has south slopes.
	SouthSlopes bool
	// WestSlopes is true if the biome has west slopes.
	WestSlopes bool
	// EastSlopes is true if the biome has east slopes.
	EastSlopes bool
	// TopSlideEnabled is true if the biome has top slide enabled.
	TopSlideEnabled bool
}

// Marshal reads or writes BiomeMountainParamsData using its canonical wire layout.
func (x *BiomeMountainParamsData) Marshal(io IO) {
	io.Uint32(&x.SteepBlock)
	io.Bool(&x.NorthSlopes)
	io.Bool(&x.SouthSlopes)
	io.Bool(&x.WestSlopes)
	io.Bool(&x.EastSlopes)
	io.Bool(&x.TopSlideEnabled)
}

// BiomeMultiNoiseRules specifies the rules for multi-noise biomes, which are biomes that are defined by
// multiple noise parameters instead of just temperature and humidity.
type BiomeMultinoiseGenRulesData struct {
	// Temperature is the temperature level of the biome.
	Temperature float32
	// Humidity is the humidity level of the biome.
	Humidity float32
	// Altitude is the altitude level of the biome.
	Altitude float32
	// Weirdness is the weirdness level of the biome.
	Weirdness float32
	// Weight is the weight of the biome, with a higher weight being more likely to be selected.
	Weight float32
}

// Marshal reads or writes BiomeMultinoiseGenRulesData using its canonical wire layout.
func (x *BiomeMultinoiseGenRulesData) Marshal(io IO) {
	io.Float32(&x.Temperature)
	io.Float32(&x.Humidity)
	io.Float32(&x.Altitude)
	io.Float32(&x.Weirdness)
	io.Float32(&x.Weight)
}

// BiomeNoiseGradientSurface specifies noise-gradient surface block data for a biome.
type BiomeNoiseGradientSurfaceData struct {
	// NonReplaceableBlocks is a list of block runtime IDs that may not be replaced.
	NonReplaceableBlocks []uint32
	// GradientBlocks is a list of noise block specifiers used by the gradient.
	GradientBlocks []SerializedNoiseBlockSpecifier
	// Noise is the noise descriptor used by the gradient.
	Noise NoiseDescriptor
}

// Marshal reads or writes BiomeNoiseGradientSurfaceData using its canonical wire layout.
func (x *BiomeNoiseGradientSurfaceData) Marshal(io IO) {
	FuncSlice(io, &x.NonReplaceableBlocks, io.Varuint32, io.Uint32)
	Slice(io, &x.GradientBlocks)
	x.Noise.Marshal(io)
}

type BiomeOverworldGenRulesData struct {
	HillsTransformations  []BiomeWeightedData
	MutateTransformations []BiomeWeightedData
	RiverTransformations  []BiomeWeightedData
	ShoreTransformations  []BiomeWeightedData
	PreHillsEdge          []BiomeConditionalTransformationData
	PostShoreEdge         []BiomeConditionalTransformationData
	Climate               []BiomeWeightedTemperatureData
}

// Marshal reads or writes BiomeOverworldGenRulesData using its canonical wire layout.
func (x *BiomeOverworldGenRulesData) Marshal(io IO) {
	Slice(io, &x.HillsTransformations)
	Slice(io, &x.MutateTransformations)
	Slice(io, &x.RiverTransformations)
	Slice(io, &x.ShoreTransformations)
	Slice(io, &x.PreHillsEdge)
	Slice(io, &x.PostShoreEdge)
	Slice(io, &x.Climate)
}

// BiomeReplacementData represents data for biome replacements.
type BiomeReplacementData struct {
	// Biome is the biome ID to replace.
	ReplacementBiome uint16
	// Dimension is the dimension ID where the replacement applies.
	Dimension uint16
	// TargetBiomes is a list of target biome IDs for the replacement.
	TargetBiomes []uint16
	// Amount is the amount of replacement to apply.
	Amount float32
	// NoiseFrequencyScale ...
	NoiseFrequencyScale float32
	// ReplacementIndex is the index of the replacement.
	ReplacementIndex uint32
}

// Marshal reads or writes BiomeReplacementData using its canonical wire layout.
func (x *BiomeReplacementData) Marshal(io IO) {
	io.Uint16(&x.ReplacementBiome)
	io.Uint16(&x.Dimension)
	FuncSlice(io, &x.TargetBiomes, io.Varuint32, io.Uint16)
	io.Float32(&x.Amount)
	io.Float32(&x.NoiseFrequencyScale)
	io.Uint32(&x.ReplacementIndex)
}

type BiomeReplacementsData struct {
	BiomeReplacements []BiomeReplacementData
}

// Marshal reads or writes BiomeReplacementsData using its canonical wire layout.
func (x *BiomeReplacementsData) Marshal(io IO) {
	Slice(io, &x.BiomeReplacements)
}

type BiomeScatterParamData struct {
	// Coordinates is a list of coordinate rules to scatter the feature within.
	Coordinates []BiomeCoordinateData
	// EvaluationOrder is the order in which the coordinates are evaluated, and is one of the
	// CoordinateEvaluationOrder constants above.
	EvalOrder CoordinateEvaluationOrder
	// ChancePercentType is the type of expression operation to use for the chance percent, and is one of the
	// BiomeExpressionOp constants above.
	ChancePercentType int32
	// ChangePercent is the index of the chance expression in the string list.
	ChancePercent uint16
	// ChanceNumerator is the numerator of the chance expression.
	ChanceNumerator int32
	// ChanceDenominator is the denominator of the chance expression.
	ChanceDenominator int32
	// IterationsType is the type of expression operation to use for the iterations, and is one of the
	// BiomeExpressionOp constants above.
	IterationsType int32
	// Iterations is the index of the iterations expression in the string list.
	Iterations uint16
}

// Marshal reads or writes BiomeScatterParamData using its canonical wire layout.
func (x *BiomeScatterParamData) Marshal(io IO) {
	Slice(io, &x.Coordinates)
	x.EvalOrder.Marshal(io)
	io.Varint32(&x.ChancePercentType)
	io.Uint16(&x.ChancePercent)
	io.Int32(&x.ChanceNumerator)
	io.Int32(&x.ChanceDenominator)
	io.Varint32(&x.IterationsType)
	io.Uint16(&x.Iterations)
}

type BiomeStringList struct {
	Strings []string
}

// Marshal reads or writes BiomeStringList using its canonical wire layout.
func (x *BiomeStringList) Marshal(io IO) {
	FuncSlice(io, &x.Strings, io.Varuint32, io.String)
}

// BiomeSurfaceBuilder specifies the materials and special surface rules to use for a biome surface.
type BiomeSurfaceBuilderData struct {
	// SurfaceMaterials is a set of materials to use for the surface layers of the biome.
	SurfaceMaterials Optional[BiomeSurfaceMaterialData]
	// HasDefaultOverworldSurface is true if the biome has a default overworld surface.
	HasDefaultOverworldSurface bool
	// HasSwampSurface is true if the biome has a swamp surface.
	HasSwampSurface bool
	// HasFrozenOceanSurface is true if the biome has a frozen ocean surface.
	HasFrozenOceanSurface bool
	// HasTheEndSurface is true if the biome has an end surface.
	HasTheEndSurface bool
	// MesaSurface is optional information to specify the biome's mesa surface.
	MesaSurface Optional[BiomeMesaSurfaceData]
	// CappedSurface is optional information to specify the biome's capped surface, i.e. in the Nether.
	CappedSurface Optional[BiomeCappedSurfaceData]
	// NoiseGradientSurface is optional information to specify noise-gradient surface data.
	NoiseGradientSurface Optional[BiomeNoiseGradientSurfaceData]
}

// Marshal reads or writes BiomeSurfaceBuilderData using its canonical wire layout.
func (x *BiomeSurfaceBuilderData) Marshal(io IO) {
	OptionalMarshaler(io, &x.SurfaceMaterials)
	io.Bool(&x.HasDefaultOverworldSurface)
	io.Bool(&x.HasSwampSurface)
	io.Bool(&x.HasFrozenOceanSurface)
	io.Bool(&x.HasTheEndSurface)
	OptionalMarshaler(io, &x.MesaSurface)
	OptionalMarshaler(io, &x.CappedSurface)
	OptionalMarshaler(io, &x.NoiseGradientSurface)
}

type BiomeSurfaceMaterialAdjustmentData struct {
	Adjustments []BiomeElementData
}

// Marshal reads or writes BiomeSurfaceMaterialAdjustmentData using its canonical wire layout.
func (x *BiomeSurfaceMaterialAdjustmentData) Marshal(io IO) {
	Slice(io, &x.Adjustments)
}

// BiomeSurfaceMaterial specifies the materials to use for the surface layers of the biome.
type BiomeSurfaceMaterialData struct {
	// TopBlock is the runtime ID of the block to use for the top layer.
	TopBlock uint32
	// MidBlock is the runtime ID to use for the middle layers.
	MidBlock uint32
	// SeaFloorBlock is the runtime ID to use for the sea floor.
	SeaFloorBlock uint32
	// FoundationBlock is the runtime ID to use for the foundation layers.
	FoundationBlock uint32
	// SeaBlock is the runtime ID to use for the sea layers.
	SeaBlock uint32
	// SeaFloorDepth is the depth of the sea floor, in blocks.
	SeaFloorDepth int32
}

// Marshal reads or writes BiomeSurfaceMaterialData using its canonical wire layout.
func (x *BiomeSurfaceMaterialData) Marshal(io IO) {
	io.Uint32(&x.TopBlock)
	io.Uint32(&x.MidBlock)
	io.Uint32(&x.SeaFloorBlock)
	io.Uint32(&x.FoundationBlock)
	io.Uint32(&x.SeaBlock)
	io.Int32(&x.SeaFloorDepth)
}

type BiomeTagsData struct {
	Tags []uint16
}

// Marshal reads or writes BiomeTagsData using its canonical wire layout.
func (x *BiomeTagsData) Marshal(io IO) {
	FuncSlice(io, &x.Tags, io.Varuint32, io.Uint16)
}

type BiomeWeightedData struct {
	BiomeIdentifier uint16
	Weight          uint32
}

// Marshal reads or writes BiomeWeightedData using its canonical wire layout.
func (x *BiomeWeightedData) Marshal(io IO) {
	io.Uint16(&x.BiomeIdentifier)
	io.Uint32(&x.Weight)
}

// BiomeTemperatureWeight defines the weight for a temperature, used for weighted randomness.
type BiomeWeightedTemperatureData struct {
	// Temperature is the temperature that can be selected.
	Temperature int32
	// Weight is the weight of the temperature, with a higher weight being more likely to be selected.
	Weight uint32
}

// Marshal reads or writes BiomeWeightedTemperatureData using its canonical wire layout.
func (x *BiomeWeightedTemperatureData) Marshal(io IO) {
	io.Varint32(&x.Temperature)
	io.Uint32(&x.Weight)
}

// FloatRange is an inclusive minimum/maximum pair of float32 values.
type FloatRange struct {
	// Min is the minimum value of the range.
	Min float32
	// Max is the maximum value of the range.
	Max float32
}

// Marshal reads or writes FloatRange using its canonical wire layout.
func (x *FloatRange) Marshal(io IO) {
	io.Float32(&x.Min)
	io.Float32(&x.Max)
}

// NoiseDescriptor describes the gradient noise used by a BiomeNoiseGradientSurface.
type NoiseDescriptor struct {
	// Name is the string used to initialise the noise.
	Name string
	// FirstOctave is the first octave used by the noise.
	FirstOctave int32
	// Amplitudes is a list of amplitude values used by the noise. It must contain between 1 and 100 entries.
	Amplitudes []float32
}

// Marshal reads or writes NoiseDescriptor using its canonical wire layout.
func (x *NoiseDescriptor) Marshal(io IO) {
	io.String(&x.Name)
	io.Int32(&x.FirstOctave)
	FuncSliceLimits(io, &x.Amplitudes, io.Varuint32, 1, 100, io.Float32)
}
