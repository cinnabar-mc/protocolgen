package retail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	gtprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SyntheticDir is the committed synthetic corpus, relative to this module.
const SyntheticDir = "../../testdata/bench"

const syntheticManifest = "manifest.json"

// Bench sets a synthetic file belongs to.
const (
	SetLarge = "large" // a large packet the brief names
	SetHot   = "hot"   // a small packet sent many times a tick
	SetBatch = "batch" // a realistic multi-packet batch
)

// SyntheticFile is one committed synthetic batch: packets built as a server
// or client builds them and encoded by the pinned gophertunnel.
type SyntheticFile struct {
	Name    string   `json:"name"` // bench case name, shared by the Go and Rust benches
	File    string   `json:"file"`
	Sender  string   `json:"sender"`
	Set     string   `json:"set"`
	Packets []uint32 `json:"packets"` // entry ids, in batch order
	SHA256  string   `json:"sha256"`
	Batch   []byte   `json:"-"`
}

// SyntheticCorpus builds every synthetic file. It is deterministic: the
// committed corpus must equal its output byte for byte.
func SyntheticCorpus() ([]SyntheticFile, error) {
	rng := rand.New(rand.NewPCG(2193, 1))
	type spec struct {
		name, file, sender, set string
		packets                 []gtpacket.Packet
	}
	one := func(name, file, sender, set string, pk gtpacket.Packet) spec {
		return spec{name, file, sender, set, []gtpacket.Packet{pk}}
	}
	chunks := make([]gtpacket.Packet, 0, 10)
	chunks = append(chunks, &gtpacket.NetworkChunkPublisherUpdate{Position: gtprotocol.BlockPos{8, 64, 8}, Radius: 48})
	for x := int32(-1); x <= 1; x++ {
		for z := int32(-1); z <= 1; z++ {
			chunks = append(chunks, levelChunk(rng, x, z))
		}
	}
	specs := []spec{
		one("StartGame", "synthetic_start_game.bin", SenderServer, SetLarge, startGame(rng)),
		one("LevelChunk", "synthetic_level_chunk.bin", SenderServer, SetLarge, levelChunk(rng, 3, -7)),
		one("SubChunk", "synthetic_sub_chunk.bin", SenderServer, SetLarge, subChunk(rng)),
		one("CraftingData", "synthetic_crafting_data.bin", SenderServer, SetLarge, craftingData(rng)),
		one("AvailableCommands", "synthetic_available_commands.bin", SenderServer, SetLarge, availableCommands(rng)),
		one("BiomeDefinitionList", "synthetic_biome_definition_list.bin", SenderServer, SetLarge, biomeDefinitionList(rng)),
		one("ItemRegistry", "synthetic_item_registry.bin", SenderServer, SetLarge, itemRegistry(rng)),
		one("CreativeContent", "synthetic_creative_content.bin", SenderServer, SetLarge, creativeContent(rng)),
		one("PlayerAuthInput", "synthetic_player_auth_input.bin", SenderClient, SetHot, playerAuthInput(rng, false)),
		one("PlayerAuthInputInteract", "synthetic_player_auth_input_interact.bin", SenderClient, SetHot, playerAuthInput(rng, true)),
		one("InventoryTransactionUseItem", "synthetic_inventory_transaction_use_item.bin", SenderClient, SetHot, useItem(rng)),
		one("InventoryTransactionUseItemOnEntity", "synthetic_inventory_transaction_use_item_on_entity.bin", SenderClient, SetHot, useItemOnEntity(rng)),
		one("InventoryTransactionNormal", "synthetic_inventory_transaction_normal.bin", SenderClient, SetHot, normalTransaction(rng)),
		one("ItemStackRequest", "synthetic_item_stack_request.bin", SenderClient, SetHot, itemStackRequest(rng)),
		one("MovePlayer", "synthetic_move_player.bin", SenderServer, SetHot, movePlayer(rng)),
		one("MoveActorDelta", "synthetic_move_actor_delta.bin", SenderServer, SetHot, moveActorDelta(rng)),
		one("SetActorMotion", "synthetic_set_actor_motion.bin", SenderServer, SetHot, setActorMotion(rng)),
		one("SetActorData", "synthetic_set_actor_data.bin", SenderServer, SetHot, setActorData(rng)),
		one("AddActor", "synthetic_add_actor.bin", SenderServer, SetHot, addActor(rng)),
		one("UpdateBlock", "synthetic_update_block.bin", SenderServer, SetHot, updateBlock(rng)),
		one("UpdateAttributes", "synthetic_update_attributes.bin", SenderServer, SetHot, updateAttributes(rng)),
		one("LevelSoundEvent", "synthetic_level_sound_event.bin", SenderServer, SetHot, levelSoundEvent(rng)),
		one("Text", "synthetic_text.bin", SenderServer, SetHot, chat(rng)),
		one("Animate", "synthetic_animate.bin", SenderClient, SetHot, &gtpacket.Animate{ActionType: gtpacket.AnimateActionSwingArm, EntityRuntimeID: 1}),
		one("NetworkStackLatency", "synthetic_network_stack_latency.bin", SenderServer, SetHot, &gtpacket.NetworkStackLatency{Timestamp: 1_760_000_000_123, NeedsResponse: true}),
		one("MobEquipment", "synthetic_mob_equipment.bin", SenderServer, SetHot, mobEquipment(rng)),
		{"ServerTick", "synthetic_batch_server_tick.bin", SenderServer, SetBatch, serverTick(rng)},
		{"ChunkRadius", "synthetic_batch_chunk_radius.bin", SenderServer, SetBatch, chunks},
		{"Spawn", "synthetic_batch_spawn.bin", SenderServer, SetBatch, spawn(rng)},
		{"ClientTick", "synthetic_batch_client_tick.bin", SenderClient, SetBatch, clientTick(rng)},
	}
	files := make([]SyntheticFile, 0, len(specs))
	for _, s := range specs {
		entries := make([]Entry, 0, len(s.packets))
		ids := make([]uint32, 0, len(s.packets))
		for _, pk := range s.packets {
			payload, err := encodeGophertunnel(gtpacket.Header{PacketID: pk.ID()}, pk)
			if err != nil {
				return nil, fmt.Errorf("encode %s: %w", s.name, err)
			}
			entries = append(entries, Entry{Payload: payload})
			ids = append(ids, pk.ID())
		}
		batch := JoinBatch(entries)
		sum := sha256.Sum256(batch)
		files = append(files, SyntheticFile{
			Name: s.name, File: s.file, Sender: s.sender, Set: s.set, Packets: ids,
			SHA256: hex.EncodeToString(sum[:]), Batch: batch,
		})
	}
	return files, nil
}

// WriteSyntheticCorpus writes files and their manifest into dir.
func WriteSyntheticCorpus(dir string, files []SyntheticFile) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f.File), f.Batch, 0o644); err != nil {
			return err
		}
	}
	manifest, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, syntheticManifest), append(manifest, '\n'), 0o644)
}

// ReadSyntheticCorpus reads the committed synthetic files, rejecting any
// whose SHA-256 differs from the manifest.
func ReadSyntheticCorpus(dir string) ([]SyntheticFile, error) {
	data, err := os.ReadFile(filepath.Join(dir, syntheticManifest))
	if err != nil {
		return nil, err
	}
	var files []SyntheticFile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&files); err != nil {
		return nil, fmt.Errorf("%s: %w", syntheticManifest, err)
	}
	if err := CheckSyntheticDir(dir, files); err != nil {
		return nil, err
	}
	for i, f := range files {
		batch, err := os.ReadFile(filepath.Join(dir, f.File))
		if err != nil {
			return nil, err
		}
		if sum := sha256.Sum256(batch); hex.EncodeToString(sum[:]) != f.SHA256 {
			return nil, fmt.Errorf("%s does not match its pinned SHA-256", f.File)
		}
		files[i].Batch = batch
	}
	return files, nil
}

// AddSynthetic appends the committed synthetic files in dir to corpus as
// retail fixtures, so every retail case and probe runs over them too.
func AddSynthetic(corpus *Corpus, batches map[string][]byte, dir string) error {
	files, err := ReadSyntheticCorpus(dir)
	if err != nil {
		return err
	}
	directions, err := syntheticDirections(filepath.Join("..", "..", "generated", corpus.MinecraftVersion, "directions.json"))
	if err != nil {
		return err
	}
	for _, f := range files {
		if _, repeated := batches[f.File]; repeated {
			return fmt.Errorf("synthetic file %s repeats a retail fixture name", f.File)
		}
		if f.Sender != SenderServer && f.Sender != SenderClient || len(f.Packets) == 0 {
			return fmt.Errorf("synthetic file %s has sender %q and %d packets", f.File, f.Sender, len(f.Packets))
		}
		direction := ""
		for _, id := range f.Packets {
			d, ok := directions[id]
			if !ok {
				return fmt.Errorf("%s: packet %d has no manifest direction", f.File, id)
			}
			if direction != "" && d != direction {
				d = "bidirectional"
			}
			direction = d
		}
		corpus.Fixtures = append(corpus.Fixtures, Fixture{Name: f.Name, File: f.File, Packet: "synthetic", ID: f.Packets[0], Direction: direction, Sender: f.Sender, SHA256: f.SHA256})
		batches[f.File] = f.Batch
	}
	return nil
}

func startGame(rng *rand.Rand) *gtpacket.StartGame {
	rules := []gtprotocol.GameRule{}
	for _, name := range []string{"commandblockoutput", "dodaylightcycle", "doentitydrops", "dofiretick", "domobloot", "domobspawning",
		"dotiledrops", "doweathercycle", "drowningdamage", "falldamage", "firedamage", "keepinventory", "mobgriefing", "naturalregeneration",
		"pvp", "sendcommandfeedback", "showcoordinates", "showdaysplayed", "tntexplodes", "doimmediaterespawn", "doinsomnia",
		"freezedamage", "respawnblocksexplode", "showbordereffect", "showdeathmessages", "showtags", "projectilescanbreakblocks"} {
		rules = append(rules, gtprotocol.GameRule{Name: name, CanBeModifiedByPlayer: true, Value: rng.IntN(2) == 0})
	}
	rules = append(rules,
		gtprotocol.GameRule{Name: "functioncommandlimit", CanBeModifiedByPlayer: true, Value: uint32(10000)},
		gtprotocol.GameRule{Name: "maxcommandchainlength", CanBeModifiedByPlayer: true, Value: uint32(65535)},
		gtprotocol.GameRule{Name: "randomtickspeed", CanBeModifiedByPlayer: true, Value: uint32(1)},
		gtprotocol.GameRule{Name: "spawnradius", CanBeModifiedByPlayer: true, Value: uint32(5)},
		gtprotocol.GameRule{Name: "playerssleepingpercentage", CanBeModifiedByPlayer: true, Value: uint32(100)},
	)
	blocks := make([]gtprotocol.BlockEntry, 0, 48)
	for i := range 48 {
		blocks = append(blocks, gtprotocol.BlockEntry{
			Name:       fmt.Sprintf("synthetic:block_%02d", i),
			Properties: map[string]any{"properties": customBlock(rng, i)},
		})
	}
	return &gtpacket.StartGame{
		EntityUniqueID: -4294967291, EntityRuntimeID: 1, PlayerGameMode: 1,
		PlayerPosition: mgl32.Vec3{0.5, 65.62, 0.5}, Pitch: 12.5, Yaw: -90,
		WorldSeed: rng.Int64(), SpawnBiomeType: 0, UserDefinedBiomeName: "plains", Dimension: 0, Generator: 2, WorldGameMode: 1,
		Difficulty: 2, WorldSpawn: gtprotocol.BlockPos{0, 64, 0}, AchievementsDisabled: true, DayCycleLockTime: -1,
		RainLevel: 0, LightningLevel: 0, MultiPlayerGame: true, LANBroadcastEnabled: true, XBLBroadcastMode: 4, PlatformBroadcastMode: 4,
		CommandsEnabled: true, GameRules: rules,
		Experiments:       []gtprotocol.ExperimentData{{Name: "data_driven_items", Enabled: true}, {Name: "upcoming_creator_features", Enabled: true}},
		PlayerPermissions: 1, ServerChunkTickRadius: 4, BaseGameVersion: "*", LimitedWorldWidth: 16, LimitedWorldDepth: 16, NewNether: true,
		ForceExperimentalGameplay: gtprotocol.Option(true), LevelID: "c3ludGhldGljd29ybGQ=", WorldName: "§bSynthetic §fworld",
		PlayerMovementSettings: gtprotocol.PlayerMovementSettings{RewindHistorySize: 40, ServerAuthoritativeBlockBreaking: true},
		Time:                   6000, EnchantmentSeed: rng.Int32(), Blocks: blocks, MultiPlayerCorrelationID: uuid.UUID{0: 0x5e, 6: 0x40, 8: 0x80, 15: 1}.String(),
		ServerAuthoritativeInventory: true, GameVersion: "1.26.51",
		PropertyData:             map[string]any{"properties": []any{map[string]any{"version": int32(3)}, map[string]any{"seed": rng.Int64()}, map[string]any{"name": "synthetic"}}},
		ServerBlockStateChecksum: rng.Uint64(), WorldTemplateID: uuid.UUID{0: 1, 6: 0x40, 8: 0x80, 15: 2}, ChatRestrictionLevel: 0,
		UseBlockNetworkIDHashes: true, ServerAuthoritativeSound: true, ServerID: "synthetic-server", ScenarioID: "", WorldID: "world-1", OwnerID: "",
	}
}

// nbtList returns a list of single-key compounds. gophertunnel decodes
// compounds into Go maps and re-encodes them in map order, so a compound with
// two keys would not re-encode deterministically.
func nbtList(pairs ...any) []any {
	out := make([]any, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, map[string]any{pairs[i].(string): pairs[i+1]})
	}
	return out
}

func customBlock(rng *rand.Rand, i int) []any {
	return nbtList(
		"menu_category", nbtList("category", "construction", "group", fmt.Sprintf("itemGroup.name.synthetic_%d", i%6)),
		"molangVersion", int32(12),
		"vanilla_block_data", byte(rng.IntN(2)),
		"properties", nbtList("name", "synthetic:facing", "enum", []any{"north", "south", "east", "west"}),
	)
}

// levelChunk serializes a flat overworld column as Dragonfly's network
// encoding does: paletted v9 sub-chunks, one biome palette per sub-chunk, no
// border blocks and no block entities.
func levelChunk(rng *rand.Rand, x, z int32) *gtpacket.LevelChunk {
	const subChunks = 8
	var payload []byte
	for y := range subChunks {
		payload = append(payload, subChunkPayload(rng, int8(y-4))...)
	}
	for range 24 {
		payload = append(payload, 1)                               // single-value biome palette
		payload = binary.AppendVarint(payload, int64(rng.IntN(3))) // biome id
	}
	payload = append(payload, 0) // border blocks
	return &gtpacket.LevelChunk{Position: gtprotocol.ChunkPos{x, z}, SubChunkCount: subChunks, RawPayload: payload}
}

// subChunkPayload returns one paletted sub-chunk with stone, ores, dirt,
// grass and air, at 4 bits per block.
func subChunkPayload(rng *rand.Rand, y int8) []byte {
	palette := []int32{-604749536, 1104582342, -1287233473, 2052370311, 981234511, 125471203, -98234511, 714421334, 0x7fff_1234}
	out := []byte{9, 1, byte(y), 4<<1 | 1}
	var word uint32
	for i := range 4096 {
		layer := (i >> 4) & 0xf
		var index uint32
		switch {
		case y >= 0 && layer > 4:
			index = 8 // air
		case y >= 0 && layer == 4:
			index = 6 // grass
		case y >= 0:
			index = 5 // dirt
		default:
			index = uint32(rng.IntN(5)) // stone and ores
		}
		word |= index << (4 * (i % 8))
		if i%8 == 7 {
			out = binary.LittleEndian.AppendUint32(out, word)
			word = 0
		}
	}
	out = binary.AppendVarint(out, int64(len(palette)))
	for _, id := range palette {
		out = binary.AppendVarint(out, int64(id))
	}
	return out
}

func subChunk(rng *rand.Rand) *gtpacket.SubChunk {
	entries := make([]gtprotocol.SubChunkEntry, 0, 27)
	for dy := int8(-1); dy <= 1; dy++ {
		for dx := int8(-1); dx <= 1; dx++ {
			for dz := int8(-1); dz <= 1; dz++ {
				var heights gtprotocol.HeightMap
				for i := range heights {
					for j := range heights[i] {
						heights[i][j] = int8(rng.IntN(16))
					}
				}
				entries = append(entries, gtprotocol.SubChunkEntry{
					Offset: gtprotocol.SubChunkOffset{dx, dy, dz}, Result: gtprotocol.SubChunkResultSuccess,
					RawPayload:    gtprotocol.Option(subChunkPayload(rng, dy)),
					HeightMapType: gtprotocol.HeightMapDataHasData, HeightMapData: gtprotocol.Option(heights),
					RenderHeightMapType: gtprotocol.HeightMapDataAllCopied,
				})
			}
		}
	}
	return &gtpacket.SubChunk{Dimension: 0, Position: gtprotocol.SubChunkPos{4, 3, -2}, SubChunkEntries: entries}
}

var itemNames = []string{"stone", "granite", "diorite", "andesite", "grass_block", "dirt", "cobblestone", "oak_planks", "spruce_planks",
	"birch_planks", "oak_log", "glass", "sandstone", "white_wool", "gold_block", "iron_block", "brick_block", "bookshelf", "obsidian",
	"torch", "chest", "diamond_ore", "crafting_table", "furnace", "ladder", "rail", "lever", "redstone_torch", "snow", "ice", "cactus",
	"clay", "jukebox", "oak_fence", "pumpkin", "netherrack", "glowstone", "stick", "diamond_sword", "iron_pickaxe", "bread", "apple",
	"bow", "arrow", "coal", "iron_ingot", "gold_ingot", "diamond", "bucket", "water_bucket", "string", "feather", "gunpowder"}

func itemName(i int) string {
	return fmt.Sprintf("minecraft:%s_%d", itemNames[i%len(itemNames)], i/len(itemNames))
}

func craftingData(rng *rand.Rand) *gtpacket.CraftingData {
	pk := &gtpacket.CraftingData{ClearRecipes: true}
	descriptor := func() gtprotocol.ItemDescriptorCount {
		switch rng.IntN(8) {
		case 0:
			return gtprotocol.ItemDescriptorCount{Descriptor: &gtprotocol.ItemTagItemDescriptor{Tag: "minecraft:planks"}, Count: 1}
		case 1:
			return gtprotocol.ItemDescriptorCount{Descriptor: &gtprotocol.MoLangItemDescriptor{Expression: "q.any_tag('minecraft:logs')", Version: 12}, Count: 1}
		default:
			return gtprotocol.ItemDescriptorCount{Descriptor: &gtprotocol.DefaultItemDescriptor{Name: itemName(rng.IntN(500)), MetadataValue: 32767}, Count: 1}
		}
	}
	output := func() []gtprotocol.ItemStack {
		return []gtprotocol.ItemStack{{ItemType: gtprotocol.ItemType{NetworkID: int32(1 + rng.IntN(1200))}, Count: uint16(1 + rng.IntN(8))}}
	}
	recipeUUID := func() uuid.UUID {
		var id uuid.UUID
		for i := range id {
			id[i] = byte(rng.Uint32())
		}
		return id
	}
	network := uint32(1)
	for i := range 360 {
		width, height := int32(1+rng.IntN(3)), int32(1+rng.IntN(3))
		inputs := make([]gtprotocol.ItemDescriptorCount, width*height)
		for j := range inputs {
			inputs[j] = descriptor()
		}
		pk.ShapedRecipes = append(pk.ShapedRecipes, gtprotocol.ShapedRecipe{
			RecipeID: fmt.Sprintf("minecraft:shaped_%d", i), Width: width, Height: height, Input: inputs, Output: output(),
			UUID: recipeUUID(), Block: "crafting_table", AssumeSymmetry: true, RecipeNetworkID: network,
		})
		network++
	}
	for i := range 280 {
		inputs := make([]gtprotocol.ItemDescriptorCount, 1+rng.IntN(4))
		for j := range inputs {
			inputs[j] = descriptor()
		}
		block := "crafting_table"
		if i%5 == 0 {
			block = "stonecutter"
		}
		pk.ShapelessRecipes = append(pk.ShapelessRecipes, gtprotocol.ShapelessRecipe{
			RecipeID: fmt.Sprintf("minecraft:shapeless_%d", i), Input: inputs, Output: output(), UUID: recipeUUID(),
			Block: block, Priority: int32(i % 3), RecipeNetworkID: network,
		})
		network++
	}
	for i := range 20 {
		pk.SmithingTransformRecipes = append(pk.SmithingTransformRecipes, gtprotocol.SmithingTransformRecipe{
			RecipeNetworkID: network, RecipeID: fmt.Sprintf("minecraft:smithing_%d", i), Template: descriptor(), Base: descriptor(),
			Addition: descriptor(), Result: output()[0], Block: "smithing_table",
		})
		network++
	}
	for i := range 48 {
		pk.PotionRecipes = append(pk.PotionRecipes, gtprotocol.PotionRecipe{
			InputPotionID: int32(400 + i%12), ReagentItemID: int32(300 + i), OutputPotionID: int32(400 + (i+1)%12), OutputPotionMetadata: int32(i % 3),
		})
	}
	for i := range 3 {
		pk.PotionContainerChangeRecipes = append(pk.PotionContainerChangeRecipes, gtprotocol.PotionContainerChangeRecipe{
			InputItemID: int32(430 + i), ReagentItemID: int32(500 + i), OutputItemID: int32(431 + i),
		})
	}
	for i := range 8 {
		pk.MaterialReducers = append(pk.MaterialReducers, gtprotocol.MaterialReducer{
			InputItem: gtprotocol.ItemType{NetworkID: int32(-700 - i), MetadataValue: uint32(i)},
			Outputs:   []gtprotocol.MaterialReducerOutput{{NetworkID: int32(600 + i), Count: int32(1 + i)}, {NetworkID: 601, Count: 2}},
		})
	}
	return pk
}

func availableCommands(rng *rand.Rand) *gtpacket.AvailableCommands {
	pk := &gtpacket.AvailableCommands{Suffixes: []string{"L"}}
	for i := range 900 {
		pk.EnumValues = append(pk.EnumValues, itemName(i))
	}
	for i := range 32 {
		count := 2 + rng.IntN(40)
		if i == 0 {
			count = 600
		}
		indices := make([]uint32, count)
		for j := range indices {
			indices[j] = uint32(rng.IntN(len(pk.EnumValues)))
		}
		pk.Enums = append(pk.Enums, gtprotocol.CommandEnum{Type: fmt.Sprintf("Enum%d", i), ValueIndices: indices})
	}
	argTypes := []uint32{gtprotocol.CommandArgTypeInt, gtprotocol.CommandArgTypeFloat, gtprotocol.CommandArgTypeTarget,
		gtprotocol.CommandArgTypeString, gtprotocol.CommandArgTypePosition, gtprotocol.CommandArgTypeMessage, gtprotocol.CommandArgTypeJSON}
	for i := range 80 {
		command := gtprotocol.Command{
			Name: fmt.Sprintf("cmd%02d", i), Description: fmt.Sprintf("commands.synthetic_%d.description", i),
			PermissionLevel: byte(i % 4), AliasesOffset: ^uint32(0),
		}
		for range 1 + rng.IntN(4) {
			var overload gtprotocol.CommandOverload
			for p := range 1 + rng.IntN(5) {
				typ := gtprotocol.CommandArgValid | argTypes[rng.IntN(len(argTypes))]
				if rng.IntN(3) == 0 {
					typ = gtprotocol.CommandArgValid | gtprotocol.CommandArgEnum | uint32(rng.IntN(len(pk.Enums)))
				}
				overload.Parameters = append(overload.Parameters, gtprotocol.CommandParameter{
					Name: fmt.Sprintf("arg%d", p), Type: typ, Optional: p > 0 && rng.IntN(2) == 0,
				})
			}
			command.Overloads = append(command.Overloads, overload)
		}
		pk.Commands = append(pk.Commands, command)
	}
	for i := range 6 {
		values := make([]string, 4+rng.IntN(30))
		for j := range values {
			values[j] = fmt.Sprintf("soft_%d_%d", i, j)
		}
		pk.DynamicEnums = append(pk.DynamicEnums, gtprotocol.DynamicEnum{Type: fmt.Sprintf("SoftEnum%d", i), Values: values})
	}
	pk.Constraints = []gtprotocol.CommandEnumConstraint{{EnumValueIndex: 3, EnumIndex: 0, Constraints: []byte{1}}}
	return pk
}

func biomeDefinitionList(rng *rand.Rand) *gtpacket.BiomeDefinitionList {
	pk := &gtpacket.BiomeDefinitionList{}
	for _, tag := range []string{"overworld", "monster", "animal", "forest", "plains", "ocean", "deep", "cold", "frozen", "warm",
		"mountains", "hills", "river", "beach", "mesa", "jungle", "swamp", "nether", "the_end", "spawn_few_zombified_piglins"} {
		pk.StringList = append(pk.StringList, tag)
	}
	for i := range 88 {
		pk.StringList = append(pk.StringList, fmt.Sprintf("synthetic_biome_%d", i))
		tags := make([]uint16, 2+rng.IntN(6))
		for j := range tags {
			tags[j] = uint16(rng.IntN(20))
		}
		pk.BiomeDefinitions = append(pk.BiomeDefinitions, gtprotocol.BiomeDefinition{
			NameIndex: int16(20 + i), BiomeID: int16(i), Temperature: float32(rng.IntN(20)) / 10, Downfall: float32(rng.IntN(10)) / 10,
			FoliageSnow: 0, Depth: 0.1, Scale: 0.2, MapWaterColour: int32(rng.Uint32()), Rain: rng.IntN(2) == 0,
			Tags: gtprotocol.Option(tags),
		})
	}
	return pk
}

func itemRegistry(rng *rand.Rand) *gtpacket.ItemRegistry {
	pk := &gtpacket.ItemRegistry{}
	for i := range 1300 {
		entry := gtprotocol.ItemEntry{Name: itemName(i), RuntimeID: int16(i - 300), Version: 1, Data: map[string]any{}}
		if i%13 == 0 {
			entry.ComponentBased, entry.Version = true, 2
			entry.Data = map[string]any{"components": nbtList(
				"item_properties", nbtList("max_stack_size", int32(1+rng.IntN(64)), "hand_equipped", byte(rng.IntN(2)), "creative_category", "equipment"),
				"minecraft:icon", entry.Name,
			)}
		}
		pk.Items = append(pk.Items, entry)
	}
	return pk
}

func itemStack(rng *rand.Rand, id int32) gtprotocol.ItemStack {
	stack := gtprotocol.ItemStack{ItemType: gtprotocol.ItemType{NetworkID: id}, Count: 1, NBTData: map[string]any{}}
	switch rng.IntN(4) {
	case 0:
		stack.BlockRuntimeID = int32(rng.Uint32())
	case 1:
		stack.NBTData = map[string]any{"tag": map[string]any{"Damage": int32(rng.IntN(100))}}
	}
	return stack
}

func creativeContent(rng *rand.Rand) *gtpacket.CreativeContent {
	pk := &gtpacket.CreativeContent{}
	for i := range 40 {
		pk.Groups = append(pk.Groups, gtprotocol.CreativeGroup{
			Category: byte(1 + i%4), Name: fmt.Sprintf("itemGroup.name.synthetic_%d", i), Icon: itemStack(rng, int32(1+i)),
		})
	}
	for i := range 1100 {
		pk.Items = append(pk.Items, gtprotocol.CreativeItem{
			CreativeItemNetworkID: uint32(i + 1), Item: itemStack(rng, int32(1+rng.IntN(1300))), GroupIndex: uint32(rng.IntN(40)),
		})
	}
	return pk
}

func vec3(rng *rand.Rand, scale float32) mgl32.Vec3 {
	return mgl32.Vec3{(rng.Float32() - 0.5) * scale, (rng.Float32() - 0.5) * scale, (rng.Float32() - 0.5) * scale}
}

func heldItem(rng *rand.Rand) gtprotocol.ItemInstance {
	return gtprotocol.ItemInstance{StackNetworkID: int32(1 + rng.IntN(500)), Stack: itemStack(rng, int32(1+rng.IntN(1300)))}
}

func playerAuthInput(rng *rand.Rand, interact bool) *gtpacket.PlayerAuthInput {
	flags := gtprotocol.NewInputFlags(gtpacket.InputFlagCount)
	for _, flag := range []int{gtpacket.InputFlagUp, gtpacket.InputFlagSprinting, gtpacket.InputFlagSprintDown, gtpacket.InputFlagVerticalCollision} {
		flags.Set(flag)
	}
	position := mgl32.Vec3{120.5 + rng.Float32(), 65.62, -340.25 + rng.Float32()}
	pk := &gtpacket.PlayerAuthInput{
		Pitch: 12.5, Yaw: 171.25, Position: position, MoveVector: mgl32.Vec2{0, 1}, HeadYaw: 171.25, InputData: flags,
		InputMode: gtpacket.InputModeMouse, PlayMode: gtpacket.PlayModeNormal, InteractionModel: gtpacket.InteractionModelCrosshair,
		InteractPitch: 12.5, InteractYaw: 171.25, Tick: 482_113, Delta: mgl32.Vec3{0.021, -0.078, 0.281},
		AnalogueMoveVector: mgl32.Vec2{0, 1}, CameraOrientation: mgl32.Vec3{-0.15, -0.21, -0.96}, RawMoveVector: mgl32.Vec2{0, 1},
	}
	if interact {
		pk.InputData.Set(gtpacket.InputFlagPerformItemInteraction)
		pk.InputData.Set(gtpacket.InputFlagPerformBlockActions)
		pk.ItemInteractionData = gtprotocol.Option(useItemData(rng))
		pk.BlockActions = gtprotocol.Option([]gtprotocol.PlayerBlockAction{
			{Action: gtprotocol.PlayerActionContinueDestroyBlock, BlockPos: gtprotocol.BlockPos{121, 64, -341}, Face: 1},
			{Action: gtprotocol.PlayerActionPredictDestroyBlock, BlockPos: gtprotocol.BlockPos{121, 64, -341}, Face: 1},
		})
	}
	return pk
}

func useItemData(rng *rand.Rand) gtprotocol.UseItemTransactionData {
	return gtprotocol.UseItemTransactionData{
		ActionType: gtprotocol.UseItemActionClickBlock, TriggerType: gtprotocol.TriggerTypePlayerInput,
		BlockPosition: gtprotocol.BlockPos{121, 64, -341}, BlockFace: 1, HotBarSlot: 3, HeldItem: heldItem(rng),
		Position: mgl32.Vec3{120.8, 65.62, -340.1}, ClickedPosition: mgl32.Vec3{0.5, 1, 0.25}, BlockRuntimeID: rng.Uint32(),
		ClientPrediction: gtprotocol.ClientPredictionSuccess,
	}
}

func useItem(rng *rand.Rand) *gtpacket.InventoryTransaction {
	data := useItemData(rng)
	data.Actions = []gtprotocol.InventoryAction{{SourceType: gtprotocol.InventoryActionSourceWorld, SourceFlags: gtprotocol.Option(uint32(0)),
		InventorySlot: 0, OldItem: gtprotocol.ItemInstance{}, NewItem: heldItem(rng)}}
	return &gtpacket.InventoryTransaction{TransactionData: &data}
}

func useItemOnEntity(rng *rand.Rand) *gtpacket.InventoryTransaction {
	return &gtpacket.InventoryTransaction{TransactionData: &gtprotocol.UseItemOnEntityTransactionData{
		TargetEntityRuntimeID: 4021, ActionType: gtprotocol.UseItemOnEntityActionAttack, HotBarSlot: 0, HeldItem: heldItem(rng),
		Position: mgl32.Vec3{12.5, 70, -3.25}, ClickedPosition: mgl32.Vec3{0.1, 1.2, 0.05},
	}}
}

func normalTransaction(rng *rand.Rand) *gtpacket.InventoryTransaction {
	actions := make([]gtprotocol.InventoryAction, 0, 4)
	for i := range 4 {
		actions = append(actions, gtprotocol.InventoryAction{
			SourceType: gtprotocol.InventoryActionSourceContainer, WindowID: gtprotocol.Option(int8(0)),
			InventorySlot: uint32(i), OldItem: heldItem(rng), NewItem: heldItem(rng),
		})
	}
	return &gtpacket.InventoryTransaction{
		LegacyRequestID: -5, LegacySetItemSlots: []gtprotocol.LegacySetItemSlot{{ContainerID: 28, Slots: []byte{0, 1}}},
		Actions: actions, TransactionData: &gtprotocol.NormalTransactionData{},
	}
}

func slot(container byte, index byte, stack int32) gtprotocol.StackRequestSlotInfo {
	return gtprotocol.StackRequestSlotInfo{Container: gtprotocol.FullContainerName{ContainerID: container}, Slot: index, StackNetworkID: stack}
}

func itemStackRequest(rng *rand.Rand) *gtpacket.ItemStackRequest {
	place := &gtprotocol.PlaceStackRequestAction{}
	place.Count, place.Source, place.Destination = 16, slot(28, 3, 41), slot(7, 0, 0)
	swap := &gtprotocol.SwapStackRequestAction{Source: slot(28, 1, 12), Destination: slot(28, 8, 77)}
	consume := &gtprotocol.ConsumeStackRequestAction{}
	consume.Count, consume.Source = 1, slot(13, 28, 101)
	take := &gtprotocol.TakeStackRequestAction{}
	take.Count, take.Source, take.Destination = 4, slot(60, 50, -3), slot(28, 0, 0)
	return &gtpacket.ItemStackRequest{Requests: []gtprotocol.ItemStackRequest{
		{RequestID: -7, Actions: []gtprotocol.StackRequestAction{place, swap}, FilterCause: -1},
		{RequestID: -9, Actions: []gtprotocol.StackRequestAction{
			&gtprotocol.CraftRecipeStackRequestAction{RecipeNetworkID: uint32(1 + rng.IntN(600)), NumberOfCrafts: 1},
			&gtprotocol.CraftResultsDeprecatedStackRequestAction{ResultItems: []gtprotocol.StackRequestItem{{Identifier: "minecraft:oak_planks", Count: 4}}, TimesCrafted: 1},
			consume, take,
		}, FilterCause: -1},
		{RequestID: -11, Actions: []gtprotocol.StackRequestAction{&gtprotocol.DropStackRequestAction{Count: 1, Source: slot(28, 2, 33)}},
			FilterStrings: []string{"renamed"}, FilterCause: 0},
	}}
}

func movePlayer(rng *rand.Rand) *gtpacket.MovePlayer {
	return &gtpacket.MovePlayer{EntityRuntimeID: 17, Position: vec3(rng, 2000), Pitch: 10, Yaw: 92.5, HeadYaw: 92.5, OnGround: true, Tick: 482_120}
}

func moveActorDelta(rng *rand.Rand) *gtpacket.MoveActorDelta {
	return &gtpacket.MoveActorDelta{
		EntityRuntimeID: uint64(100 + rng.IntN(400)), PositionX: gtprotocol.Option(rng.Float32() * 300), PositionZ: gtprotocol.Option(rng.Float32() * -300),
		RotationY: gtprotocol.Option(rng.Float32() * 360), RotationYHead: gtprotocol.Option(rng.Float32() * 360), OnGround: true, Ticks: 482_121,
	}
}

func setActorMotion(rng *rand.Rand) *gtpacket.SetActorMotion {
	return &gtpacket.SetActorMotion{EntityRuntimeID: uint64(100 + rng.IntN(400)), Velocity: vec3(rng, 1), Tick: 482_121}
}

func metadata(rng *rand.Rand) gtprotocol.EntityMetadata {
	md := gtprotocol.NewEntityMetadata()
	md[gtprotocol.EntityDataKeyFlags] = int64(rng.Uint64() & 0x0000_00ff_ffff_ffff)
	md[gtprotocol.EntityDataKeyName] = "§eSynthetic §fZombie"
	md[gtprotocol.EntityDataKeyScale] = float32(1)
	md[gtprotocol.EntityDataKeyWidth] = float32(0.6)
	md[gtprotocol.EntityDataKeyHeight] = float32(1.9)
	md[gtprotocol.EntityDataKeyAirSupply] = int16(300)
	md[gtprotocol.EntityDataKeyVariant] = int32(rng.IntN(4))
	return md
}

func setActorData(rng *rand.Rand) *gtpacket.SetActorData {
	return &gtpacket.SetActorData{
		EntityRuntimeID: uint64(100 + rng.IntN(400)), EntityMetadata: metadata(rng),
		EntityProperties: gtprotocol.EntityProperties{IntegerProperties: []gtprotocol.IntegerEntityProperty{{Index: 0, Value: 2}}},
		Tick:             482_121,
	}
}

func addActor(rng *rand.Rand) *gtpacket.AddActor {
	return &gtpacket.AddActor{
		EntityUniqueID: -401, EntityRuntimeID: 401, EntityType: "minecraft:zombie", Position: vec3(rng, 200), Velocity: vec3(rng, 0.2),
		Pitch: 0, Yaw: 45, HeadYaw: 45, BodyYaw: 45,
		Attributes:     []gtprotocol.AttributeValue{{Name: "minecraft:health", Value: 20, Max: 20}, {Name: "minecraft:movement", Value: 0.23, Max: 3.4e38}},
		EntityMetadata: metadata(rng),
	}
}

func updateBlock(rng *rand.Rand) *gtpacket.UpdateBlock {
	return &gtpacket.UpdateBlock{Position: gtprotocol.BlockPos{121, 64, -341}, NewBlockRuntimeID: rng.Uint32(), Flags: 3}
}

func updateAttributes(rng *rand.Rand) *gtpacket.UpdateAttributes {
	attribute := func(name string, value, max float32) gtprotocol.Attribute {
		return gtprotocol.Attribute{AttributeValue: gtprotocol.AttributeValue{Name: name, Value: value, Max: max}, DefaultMax: max, Default: max}
	}
	movement := attribute("minecraft:movement", 0.13, 3.4e38)
	movement.Modifiers = []gtprotocol.AttributeModifier{{ID: "d208fc00-42aa-4aad-9276-d5446530de43", Name: "Sprinting speed boost", Amount: 0.3, Operation: 2, Operand: 2}}
	return &gtpacket.UpdateAttributes{EntityRuntimeID: 1, Tick: 482_121, Attributes: []gtprotocol.Attribute{
		attribute("minecraft:health", float32(1+rng.IntN(20)), 20), attribute("minecraft:player.hunger", 17, 20),
		attribute("minecraft:player.saturation", 3.2, 20), attribute("minecraft:player.exhaustion", 1.7, 5), movement,
	}}
}

func levelSoundEvent(rng *rand.Rand) *gtpacket.LevelSoundEvent {
	return &gtpacket.LevelSoundEvent{SoundType: "step", Position: vec3(rng, 500), ExtraData: int32(rng.Uint32() >> 8), EntityType: "minecraft:player", EntityUniqueID: -1}
}

func chat(rng *rand.Rand) *gtpacket.Text {
	return &gtpacket.Text{TextType: gtpacket.TextTypeChat, SourceName: "Synthetic_Player", Message: fmt.Sprintf("gg, %d kills this round", rng.IntN(30)), XUID: "2535411111111111"}
}

func mobEquipment(rng *rand.Rand) *gtpacket.MobEquipment {
	return &gtpacket.MobEquipment{EntityRuntimeID: 1, NewItem: heldItem(rng), InventorySlot: 3, HotBarSlot: 3}
}

func serverTick(rng *rand.Rand) []gtpacket.Packet {
	var pks []gtpacket.Packet
	for range 40 {
		pks = append(pks, moveActorDelta(rng))
	}
	for range 10 {
		pks = append(pks, setActorMotion(rng))
	}
	for range 4 {
		pks = append(pks, setActorData(rng), updateBlock(rng))
	}
	for range 3 {
		pks = append(pks, levelSoundEvent(rng))
	}
	return append(pks, updateAttributes(rng), movePlayer(rng), chat(rng), &gtpacket.NetworkStackLatency{Timestamp: 1_760_000_000_456, NeedsResponse: true})
}

func spawn(rng *rand.Rand) []gtpacket.Packet {
	content := make([]gtprotocol.ItemInstance, 36)
	for i := range content {
		if i%3 == 0 {
			content[i] = heldItem(rng)
		}
	}
	pks := []gtpacket.Packet{
		&gtpacket.SetTime{Time: 6000}, updateAttributes(rng), setActorData(rng),
		&gtpacket.InventoryContent{WindowID: 0, Content: content}, mobEquipment(rng),
		&gtpacket.PlayerHotBar{SelectedHotBarSlot: 3, WindowID: 0, SelectHotBarSlot: true},
		&gtpacket.UpdateSoftEnum{EnumType: "SoftEnum0", Options: []string{"alpha", "beta", "gamma"}, ActionType: gtpacket.SoftEnumActionSet},
	}
	for range 12 {
		pks = append(pks, addActor(rng))
	}
	return pks
}

func clientTick(rng *rand.Rand) []gtpacket.Packet {
	return []gtpacket.Packet{
		playerAuthInput(rng, false), &gtpacket.Animate{ActionType: gtpacket.AnimateActionSwingArm, EntityRuntimeID: 1},
		useItemOnEntity(rng), itemStackRequest(rng), &gtpacket.NetworkStackLatency{Timestamp: 1_760_000_000_456},
	}
}

// syntheticDirections maps every packet id the synthetic corpus uses to its
// corpus direction, from the manifest's directions.json.
func syntheticDirections(path string) (map[uint32]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file struct {
		Packets []struct {
			ID        uint32 `json:"id"`
			Direction string `json:"direction"`
		} `json:"packets"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	out := map[uint32]string{}
	names := map[string]string{"server": "clientbound", "client": "serverbound", "both": "bidirectional"}
	for _, p := range file.Packets {
		if !slices.Contains([]string{"server", "client", "both"}, p.Direction) {
			return nil, fmt.Errorf("%s: packet %d has direction %q", path, p.ID, p.Direction)
		}
		out[p.ID] = names[p.Direction]
	}
	return out, nil
}
