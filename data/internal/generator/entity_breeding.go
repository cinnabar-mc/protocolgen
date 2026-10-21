package generator

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/bedrock-mc/protocolgen/data/internal/runtimegen"
	"github.com/bedrock-mc/protocolgen/data/source"
)

// EntityBreedingInputs binds breeding projection to authenticated samples and
// the runtime item registry. It does not replace the independent base snapshot.
type EntityBreedingInputs struct {
	Definitions      map[string][]byte
	Items            map[string]runtimegen.RegistryEntry
	SourceLockSHA256 string
	Repository       string
	Revision         string
	ArchiveSHA256    string
}

// ReadEntityBreedingInputs authenticates the existing sample archive and item
// registry pins. Both full generation and the independent projection use it.
func ReadEntityBreedingInputs(ctx context.Context, sources *source.Sources, cache string) (EntityBreedingInputs, error) {
	var result EntityBreedingInputs
	var err error
	result.Definitions, err = sources.BehaviorEntities(ctx, cache)
	if err != nil {
		return result, err
	}
	registry, err := sources.Read(ctx, "item_registry", cache)
	if err != nil {
		return result, err
	}
	result.Items, err = runtimegen.DecodeRegistry(registry)
	if err != nil {
		return result, err
	}
	result.SourceLockSHA256, err = sources.Lock.SHA256()
	input := sources.Lock.Inputs["resource_pack"]
	upstream := sources.Lock.Upstreams[input.Upstream]
	result.Repository, result.Revision, result.ArchiveSHA256 = upstream.Repository, upstream.Revision, input.SHA256
	return result, err
}

// projectedBreeding is the generator's conservative subset of breedable facts.
type projectedBreeding struct {
	Foods              []string `json:"foods"`
	RequiresAdult      bool     `json:"requires_adult"`
	RequiresTamed      bool     `json:"requires_tamed"`
	RequiresFullHealth bool     `json:"requires_full_health"`
}

// breedingEvidence records the source document and exact component owner.
type breedingEvidence struct {
	File           string             `json:"file,omitempty"`
	SHA256         string             `json:"sha256,omitempty"`
	ComponentGroup string             `json:"component_group,omitempty"`
	Breeding       *projectedBreeding `json:"breeding,omitempty"`
	Unavailable    string             `json:"unavailable,omitempty"`
}

// GenerateEntityBreeding projects into existing Entity literals, so there is
// no second runtime registry. It also removes stale projections on regeneration.
// Base fields and the catalog's entity domain remain owned by generateEntities.
func GenerateEntityBreeding(files map[string][]byte, inputs EntityBreedingInputs) (map[string][]byte, int, error) {
	if len(inputs.Definitions) == 0 || len(inputs.Items) == 0 || inputs.SourceLockSHA256 == "" || inputs.Revision == "" || inputs.ArchiveSHA256 == "" {
		return nil, 0, fmt.Errorf("breeding projection requires authenticated definitions, items and provenance")
	}
	evidence := make(map[string]breedingEvidence)
	for _, path := range sortedKeys(inputs.Definitions) {
		data := inputs.Definitions[path]
		stripped, err := stripJSONComments(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf}))
		if err != nil {
			return nil, 0, fmt.Errorf("%s: %w", path, err)
		}
		var raw rawEntity
		if err := json.Unmarshal(stripped, &raw); err != nil {
			return nil, 0, fmt.Errorf("%s: %w", path, err)
		}
		name := raw.Entity.Description.Identifier
		if name == "" {
			return nil, 0, fmt.Errorf("%s has no entity identifier", path)
		}
		if _, exists := evidence[name]; exists {
			return nil, 0, fmt.Errorf("duplicate sample entity %s", name)
		}
		breeding, owner, reason := projectBreeding(raw, inputs.Items)
		evidence[name] = breedingEvidence{path, fmt.Sprintf("%x", sha256.Sum256(data)), owner, breeding, reason}
	}
	output := make(map[string][]byte, len(files)+1)
	used := make(map[string]breedingEvidence)
	count := 0
	for _, name := range sortedKeys(files) {
		content := files[name]
		if strings.HasPrefix(name, "entity/") && strings.HasSuffix(name, "_generated.go") {
			var entityName string
			var breeding *projectedBreeding
			var err error
			content, entityName, err = projectEntityLiteral(name, content, evidence)
			if err != nil {
				return nil, 0, err
			}
			if entityName != "" {
				entry, ok := evidence[entityName]
				if !ok {
					entry.Unavailable = "base entity has no locked sample definition"
				}
				if entry.Unavailable != "no breedable component" {
					used[entityName] = entry
				}
				breeding = entry.Breeding
			}
			if breeding != nil {
				count++
			}
		}
		output[name] = content
	}
	manifest, err := json.MarshalIndent(struct {
		SchemaVersion    int                         `json:"schema_version"`
		SourceLockSHA256 string                      `json:"source_lock_sha256"`
		Repository       string                      `json:"repository"`
		Revision         string                      `json:"revision"`
		ArchiveInput     string                      `json:"archive_input"`
		ArchiveSHA256    string                      `json:"archive_sha256"`
		Supported        int                         `json:"supported"`
		Entities         map[string]breedingEvidence `json:"entities"`
	}{1, inputs.SourceLockSHA256, inputs.Repository, inputs.Revision, "resource_pack", inputs.ArchiveSHA256, count, used}, "", "  ")
	if err != nil {
		return nil, 0, err
	}
	output["entity/breeding_sources.json"] = append(manifest, '\n')
	return output, count, nil
}

// projectBreeding proves the adult component owner from birth, spawn and growth
// events. Unsupported predicates and ambiguous owners produce no descriptor.
func projectBreeding(raw rawEntity, items map[string]runtimegen.RegistryEntry) (*projectedBreeding, string, string) {
	groups, events := raw.Entity.Groups, raw.Entity.Events
	if _, present := raw.Entity.Components["minecraft:breedable"]; present {
		return nil, "", "base breedable lifecycle is unsupported"
	}
	if _, present := raw.Entity.Components["minecraft:is_baby"]; present {
		return nil, "", "base baby lifecycle is unsupported"
	}
	var owner, baby string
	for _, name := range sortedKeys(groups) {
		components := groups[name]
		if _, present := components["minecraft:breedable"]; present {
			if owner != "" {
				return nil, "", "multiple breedable component owners"
			}
			owner = name
		}
		if _, present := components["minecraft:is_baby"]; present {
			if baby != "" {
				return nil, owner, "multiple baby component owners"
			}
			baby = name
		}
	}
	if owner == "" {
		return nil, "", "no breedable component"
	}
	breeding, reason := decodeBreeding(groups[owner]["minecraft:breedable"], raw.Entity.Description.Identifier, breedingMaySit(raw), items)
	if reason != "" {
		return nil, owner, reason
	}
	if baby == "" || baby == owner {
		return nil, owner, "no distinct baby lifecycle"
	}
	var ageable struct {
		GrowUp struct {
			Event  string `json:"event"`
			Target string `json:"target"`
		} `json:"grow_up"`
	}
	if err := json.Unmarshal(groups[baby]["minecraft:ageable"], &ageable); err != nil || ageable.GrowUp.Event == "" || ageable.GrowUp.Target != "self" {
		return nil, owner, "no self-targeted baby growth event"
	}
	projection := breedingLifecycle{events: events, groups: groups, baby: baby, adult: owner}
	grown, ok := projection.event(ageable.GrowUp.Event, 1, 0)
	if !ok || grown != 1<<2 {
		return nil, owner, "growth does not unconditionally install the adult breedable owner"
	}
	spawned, ok := projection.event("minecraft:entity_spawned", 0, 0)
	if !ok || spawned == 0 || spawned & ^uint8((1<<1)|(1<<2)) != 0 {
		return nil, owner, "spawn does not prove baby or breedable adulthood"
	}
	born, ok := projection.event("minecraft:entity_born", 0, 0)
	if !ok || born != 1<<1 {
		return nil, owner, "birth does not prove the baby owner"
	}
	if _, present := events["minecraft:entity_transformed"]; present {
		transformed, ok := projection.event("minecraft:entity_transformed", 0, 0)
		if !ok || transformed == 0 || transformed & ^uint8((1<<1)|(1<<2)) != 0 {
			return nil, owner, "transformation has an unsupported adult state"
		}
	}
	// Every event is checked, not only the growth path. Removing the breeding
	// owner without marking a baby would leave an unobservable admission state.
	for _, name := range sortedKeys(events) {
		for _, state := range []uint8{1, 2} {
			outcomes, ok := projection.event(name, state, 0)
			if !ok || outcomes&1 != 0 {
				return nil, owner, "an event can leave unsupported breeding eligibility"
			}
		}
	}
	breeding.RequiresAdult = true
	return breeding, owner, ""
}

// breedingMaySit conservatively checks base components and every group, since
// sitting can be introduced outside the adult breedable component owner.
func breedingMaySit(raw rawEntity) bool {
	canSit := func(components map[string]json.RawMessage) bool {
		for name := range components {
			switch name {
			case "minecraft:sittable", "minecraft:behavior.stay_while_sitting",
				"minecraft:behavior.random_sitting", "minecraft:behavior.random_look_around_and_sit",
				"minecraft:behavior.ocelot_sit_on_block":
				return true
			}
		}
		return false
	}
	if canSit(raw.Entity.Components) {
		return true
	}
	for _, components := range raw.Entity.Groups {
		if canSit(components) {
			return true
		}
	}
	return false
}

// decodeBreeding rejects unrepresented constraints and item-object semantics.
// Missing require_tame is true in Bedrock; it must never become false by default.
func decodeBreeding(data []byte, entityName string, maySit bool, items map[string]runtimegen.RegistryEntry) (*projectedBreeding, string) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return nil, "invalid breedable component"
	}
	for _, key := range sortedKeys(fields) {
		switch key {
		case "require_tame", "require_full_health", "allow_sitting", "breed_items", "breeds_with":
		default:
			return nil, "unsupported breedable constraint: " + key
		}
	}
	value := &projectedBreeding{RequiresTamed: true}
	for _, field := range []struct {
		name   string
		target *bool
	}{{"require_tame", &value.RequiresTamed}, {"require_full_health", &value.RequiresFullHealth}} {
		if data, present := fields[field.name]; present {
			var decoded *bool
			if json.Unmarshal(data, &decoded) != nil || decoded == nil {
				return nil, "invalid " + field.name
			}
			*field.target = *decoded
		}
	}
	// Sitting is not part of this API. Tame-required or sitting-capable animals
	// need explicit permission to breed while sitting; allow_sitting defaults false.
	if value.RequiresTamed || maySit {
		var allowed bool
		if json.Unmarshal(fields["allow_sitting"], &allowed) != nil || !allowed {
			return nil, "unrepresented sitting prerequisite"
		}
	}
	var partners map[string]map[string]json.RawMessage
	if json.Unmarshal(fields["breeds_with"], &partners) != nil || len(partners) != 1 || partners[entityName] == nil || len(partners[entityName]) != 0 {
		return nil, "unsupported partner or offspring lifecycle"
	}
	var foods []string
	if json.Unmarshal(fields["breed_items"], &foods) != nil || len(foods) == 0 {
		return nil, "unsupported breeding item representation"
	}
	for index, name := range foods {
		if !strings.Contains(name, ":") {
			name = "minecraft:" + name
		}
		if _, present := items[name]; !present {
			return nil, "breeding item is not a canonical runtime item: " + name
		}
		foods[index] = name
	}
	sort.Strings(foods)
	for index := 1; index < len(foods); index++ {
		if foods[index-1] == foods[index] {
			return nil, "duplicate breeding item"
		}
	}
	value.Foods = foods
	return value, ""
}

// breedingLifecycle projects only baby/adult-owner membership. An outcome set
// uses bit (1 << state), where state bit 0 is baby and bit 1 is breedable owner.
type breedingLifecycle struct {
	events      map[string]json.RawMessage
	groups      map[string]map[string]json.RawMessage
	baby, adult string
}

// event follows a named local event with a bounded trigger depth.
func (p breedingLifecycle) event(name string, state uint8, depth int) (uint8, bool) {
	data, present := p.events[name]
	if !present || depth > 32 {
		return 0, false
	}
	return p.node(data, state, depth+1)
}

// node overapproximates filtered choices: both applying and skipping a filter
// are possible. Unknown actions fail closed instead of inferring eligibility.
func (p breedingLifecycle) node(data []byte, state uint8, depth int) (uint8, bool) {
	if depth > 32 {
		return 0, false
	}
	original := state
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil {
		return 0, false
	}
	for key := range fields {
		switch key {
		case "filters", "weight", "add", "remove", "sequence", "randomize", "first_valid", "trigger", "set_property":
		default:
			return 0, false
		}
	}
	outcomes := uint8(1 << state)
	for _, operation := range []string{"remove", "add"} {
		data, present := fields[operation]
		if !present {
			continue
		}
		var effect struct {
			Groups []string `json:"component_groups"`
		}
		if json.Unmarshal(data, &effect) != nil {
			return 0, false
		}
		for _, name := range effect.Groups {
			if _, present := p.groups[name]; !present {
				return 0, false
			}
			var mask uint8
			if name == p.baby {
				mask = 1
			}
			if name == p.adult {
				mask |= 2
			}
			if operation == "add" {
				state |= mask
			} else {
				state &^= mask
			}
		}
	}
	outcomes = 1 << state
	control := false
	for _, name := range []string{"sequence", "randomize", "first_valid"} {
		data, present := fields[name]
		if !present {
			continue
		}
		if control {
			return 0, false
		}
		control = true
		var children []json.RawMessage
		if json.Unmarshal(data, &children) != nil || len(children) == 0 {
			return 0, false
		}
		if name == "sequence" {
			for _, child := range children {
				var next uint8
				for input := uint8(0); input < 4; input++ {
					if outcomes&(1<<input) == 0 {
						continue
					}
					result, ok := p.node(child, input, depth+1)
					if !ok {
						return 0, false
					}
					next |= result
				}
				outcomes = next
			}
		} else {
			outcomes = 0
			for _, child := range children {
				result, ok := p.node(child, state, depth+1)
				if !ok {
					return 0, false
				}
				outcomes |= result
			}
			if name == "first_valid" {
				outcomes |= 1 << state
			}
		}
	}
	if data, present := fields["trigger"]; present {
		var name string
		if json.Unmarshal(data, &name) != nil {
			return 0, false
		}
		var next uint8
		for input := uint8(0); input < 4; input++ {
			if outcomes&(1<<input) == 0 {
				continue
			}
			result, ok := p.event(name, input, depth+1)
			if !ok {
				return 0, false
			}
			next |= result
		}
		outcomes = next
	}
	if _, filtered := fields["filters"]; filtered {
		outcomes |= 1 << original
	}
	return outcomes, true
}

// projectEntityLiteral adds or replaces the one Entity.Breeding field using
// syntax positions. It never reconstructs or guesses the base-component values.
func projectEntityLiteral(name string, content []byte, evidence map[string]breedingEvidence) ([]byte, string, error) {
	if !bytes.HasPrefix(content, []byte(GeneratedMarker+"\n")) {
		return nil, "", fmt.Errorf("%s is not owned by the entity generator", name)
	}
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, name, content, 0)
	if err != nil {
		return nil, "", err
	}
	var literal *ast.CompositeLit
	ast.Inspect(file, func(node ast.Node) bool {
		value, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		typ, ok := value.Type.(*ast.Ident)
		if ok && typ.Name == "Entity" {
			literal = value
			return false
		}
		return true
	})
	if literal == nil {
		return content, "", nil
	}
	var entityName string
	start, end := set.Position(literal.Rbrace).Offset, set.Position(literal.Rbrace).Offset
	for _, field := range literal.Elts {
		entry, ok := field.(*ast.KeyValueExpr)
		if !ok {
			return nil, "", fmt.Errorf("%s has an unkeyed Entity field", name)
		}
		key, ok := entry.Key.(*ast.Ident)
		if !ok {
			return nil, "", fmt.Errorf("%s has an invalid Entity field", name)
		}
		if key.Name == "Name" {
			value, ok := entry.Value.(*ast.BasicLit)
			if !ok {
				return nil, "", fmt.Errorf("%s has a non-literal Entity name", name)
			}
			entityName, err = strconv.Unquote(value.Value)
			if err != nil {
				return nil, "", err
			}
		}
		if key.Name == "Breeding" {
			start, end = set.Position(entry.Pos()).Offset, set.Position(entry.End()).Offset
			if end < len(content) && content[end] == ',' {
				end++
			}
		}
	}
	if entityName == "" {
		return nil, "", fmt.Errorf("%s has no Entity name", name)
	}
	var replacement string
	if breeding := evidence[entityName].Breeding; breeding != nil {
		replacement = fmt.Sprintf("Breeding: &Breeding{Foods: %s, RequiresAdult: %t, RequiresTamed: %t, RequiresFullHealth: %t},", goStrings(breeding.Foods), breeding.RequiresAdult, breeding.RequiresTamed, breeding.RequiresFullHealth)
		if start == end {
			if len(literal.Elts) > 0 {
				last := set.Position(literal.Elts[len(literal.Elts)-1].End()).Offset
				if !bytes.Contains(content[last:start], []byte(",")) {
					replacement = ",\n" + replacement
				}
			}
			replacement += "\n"
		}
	}
	if start == end && replacement == "" {
		return content, entityName, nil
	}
	output := append([]byte(nil), content[:start]...)
	output = append(output, replacement...)
	output = append(output, content[end:]...)
	formatted, err := format.Source(output)
	return formatted, entityName, err
}
