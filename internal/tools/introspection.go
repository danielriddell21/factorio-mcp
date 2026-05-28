package tools

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/danielriddell21/factorio-mcp/internal/factorio"
)

// --- get_player_state ---

type playerStateIn struct {
	PlayerIndex int `json:"player_index,omitempty"`
}

type playerStateOut struct {
	Position       factorio.Position `json:"position"`
	Surface        string            `json:"surface"`
	CharacterAlive bool              `json:"character_alive"`
	Health         float64           `json:"health"`
	Walking        bool              `json:"walking"`
}

// --- get_inventory ---

type inventoryIn struct {
	// Which inventory to read: "character" (default) or "entity".
	Which string `json:"which,omitempty"`
	// UnitNumber identifies the entity when Which is "entity".
	UnitNumber int `json:"unit_number,omitempty"`
	// Inventory selects a sub-inventory for entities (e.g. "fuel", "input").
	Inventory   string `json:"inventory,omitempty"`
	PlayerIndex int    `json:"player_index,omitempty"`
}

type inventoryOut struct {
	Items map[string]int `json:"items"`
}

// --- scan_entities ---

type scanEntitiesIn struct {
	Center   *factorio.Position `json:"center,omitempty"`
	Radius   float64            `json:"radius,omitempty"`
	Names    []string           `json:"names,omitempty"`
	Types    []string           `json:"types,omitempty"`
	Page     int                `json:"page,omitempty"`
	PageSize int                `json:"page_size,omitempty"`
}

type scannedEntity struct {
	UnitNumber int               `json:"unit_number"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Position   factorio.Position `json:"position"`
	Direction  int               `json:"direction"`
	Status     string            `json:"status,omitempty"`
	Recipe     string            `json:"recipe,omitempty"`
}

type scanEntitiesOut struct {
	Entities []scannedEntity `json:"entities"`
	Page     int             `json:"page"`
	NextPage *int            `json:"next_page,omitempty"`
	Total    int             `json:"total"`
}

// --- scan_resources ---

type scanResourcesIn struct {
	Center *factorio.Position `json:"center,omitempty"`
	Radius float64            `json:"radius,omitempty"`
}

type resourcePatch struct {
	Resource string            `json:"resource"`
	Center   factorio.Position `json:"center"`
	Amount   int64             `json:"amount"`
	Tiles    int               `json:"tiles"`
}

type scanResourcesOut struct {
	Patches []resourcePatch `json:"patches"`
}

// --- get_research_state ---

type researchStateIn struct{}

type researchStateOut struct {
	Current    string   `json:"current"`
	Progress   float64  `json:"progress"`
	Queue      []string `json:"queue"`
	Researched []string `json:"researched"`
}

// --- get_tech_tree ---

type techTreeIn struct {
	// AvailableOnly limits the result to technologies whose prerequisites are
	// already researched.
	AvailableOnly bool `json:"available_only,omitempty"`
}

type technology struct {
	Name          string   `json:"name"`
	Researched    bool     `json:"researched"`
	Enabled       bool     `json:"enabled"`
	Prerequisites []string `json:"prerequisites"`
	Ingredients   []string `json:"ingredients"`
}

type techTreeOut struct {
	Technologies []technology `json:"technologies"`
}

// --- get_production_stats ---

type productionStatsIn struct {
	// Precision window: "1m", "10m", "1h", "10h", or "all".
	Precision string   `json:"precision,omitempty"`
	Items     []string `json:"items,omitempty"`
}

type productionStatsOut struct {
	Produced map[string]float64 `json:"produced"`
	Consumed map[string]float64 `json:"consumed"`
}

// --- get_recipe_info ---

type recipeInfoIn struct {
	Name string `json:"name"`
}

type ingredient struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

type recipeInfoOut struct {
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Energy      float64      `json:"energy"`
	Enabled     bool         `json:"enabled"`
	Ingredients []ingredient `json:"ingredients"`
	Products    []ingredient `json:"products"`
}

func registerIntrospection(s *mcp.Server, d factorio.Dispatcher) {
	addTool[playerStateIn, playerStateOut](s, d,
		"factorio_get_player_state",
		"Return the player character's current position, surface, health, and whether it is walking. The coordinate system is +x east, +y south, in tiles.",
		"get_player_state")

	addTool[inventoryIn, inventoryOut](s, d,
		"factorio_get_inventory",
		"Read an inventory as item-name -> count. Set 'which' to \"character\" (default) for the player, or \"entity\" with a 'unit_number' (and optional 'inventory' such as \"fuel\" or \"input\") to read a machine.",
		"get_inventory")

	addTool[scanEntitiesIn, scanEntitiesOut](s, d,
		"factorio_scan_entities",
		"List entities near a point. Filter by 'names' and/or 'types'; 'center' defaults to the player and 'radius' to a small area. Results are paginated via 'page'/'page_size'; use the returned 'next_page' to continue. Use this to verify that an action succeeded.",
		"scan_entities")

	addTool[scanResourcesIn, scanResourcesOut](s, d,
		"factorio_scan_resources",
		"Find ore/resource patches near a point, aggregated per resource with a center, total amount, and tile count. 'center' defaults to the player. Use this to locate iron, copper, coal, stone, oil, etc.",
		"scan_resources")

	addTool[researchStateIn, researchStateOut](s, d,
		"factorio_get_research_state",
		"Return the current research, its progress (0-1), the queued technologies, and the list of already-researched technologies. Research only completes as labs consume science packs over time.",
		"get_research_state")

	addTool[techTreeIn, techTreeOut](s, d,
		"factorio_get_tech_tree",
		"List technologies with their researched/enabled status, prerequisites, and science-pack ingredients. Set 'available_only' to true to show only technologies whose prerequisites are met.",
		"get_tech_tree")

	addTool[productionStatsIn, productionStatsOut](s, d,
		"factorio_get_production_stats",
		"Return produced and consumed item rates for a time window ('precision': \"1m\",\"10m\",\"1h\",\"10h\",\"all\"; default \"1m\"). Optionally restrict to specific 'items'. Use this to monitor throughput and find bottlenecks.",
		"get_production_stats")

	addTool[recipeInfoIn, recipeInfoOut](s, d,
		"factorio_get_recipe_info",
		"Return a recipe's category, crafting energy/time, ingredients, and products. Use this to plan production ratios and what a machine needs.",
		"get_recipe_info")
}
