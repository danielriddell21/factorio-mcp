package tools

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/danielriddell21/factorio-mcp/internal/factorio"
)

// --- place_entity ---

type placeEntityIn struct {
	Name      string            `json:"name"`
	Position  factorio.Position `json:"position"`
	Direction int               `json:"direction,omitempty"`
	Recipe    string            `json:"recipe,omitempty"`
}

type placeEntityOut struct {
	UnitNumber int               `json:"unit_number"`
	Position   factorio.Position `json:"position"`
}

// --- remove_entity ---

type removeEntityIn struct {
	UnitNumber int                `json:"unit_number,omitempty"`
	Position   *factorio.Position `json:"position,omitempty"`
	Name       string             `json:"name,omitempty"`
}

type removeEntityOut struct {
	Removed  bool           `json:"removed"`
	Returned map[string]int `json:"returned"`
}

// --- set_recipe ---

type setRecipeIn struct {
	UnitNumber int    `json:"unit_number"`
	Recipe     string `json:"recipe"`
}

type setRecipeOut struct {
	OK bool `json:"ok"`
}

// --- insert_items / remove_items ---

type moveItemsIn struct {
	// UnitNumber identifies the machine to move items into/out of.
	UnitNumber int    `json:"unit_number"`
	Inventory  string `json:"inventory,omitempty"`
	Item       string `json:"item"`
	Count      int    `json:"count"`
}

type moveItemsOut struct {
	Count int `json:"count"`
}

// --- craft ---

type craftIn struct {
	Recipe string `json:"recipe"`
	Count  int    `json:"count"`
}

type craftOut struct {
	Started int `json:"started"`
}

// --- set_research / queue_research ---

type setResearchIn struct {
	Technology string `json:"technology"`
}

type queueResearchIn struct {
	Technologies []string `json:"technologies"`
}

type researchOut struct {
	Current string   `json:"current"`
	Queue   []string `json:"queue"`
}

// --- teleport ---

type teleportIn struct {
	Position factorio.Position `json:"position"`
	Surface  string            `json:"surface,omitempty"`
}

type teleportOut struct {
	OK       bool              `json:"ok"`
	Position factorio.Position `json:"position"`
}

// --- mine_resource ---

type mineResourceIn struct {
	Position   *factorio.Position `json:"position,omitempty"`
	UnitNumber int                `json:"unit_number,omitempty"`
	Count      int                `json:"count,omitempty"`
}

type mineResourceOut struct {
	Mined map[string]int `json:"mined"`
}

func registerActions(s *mcp.Server, d factorio.Dispatcher) {
	addTool[placeEntityIn, placeEntityOut](s, d,
		"factorio_place_entity",
		"Place an entity at a tile position, consuming one matching item from the character's inventory (legit economy: the item must be crafted/held first; placement is rejected if it is not, or if the spot is blocked). 'direction' uses Factorio's 16-step encoding (0=N,4=E,8=S,12=W). Optionally set a 'recipe' for assembling machines. Returns the new unit_number.",
		"place_entity")

	addTool[removeEntityIn, removeEntityOut](s, d,
		"factorio_remove_entity",
		"Mine/remove an entity (by 'unit_number', or by 'position' with optional 'name'), returning its item(s) to the character's inventory as in hand-mining.",
		"remove_entity")

	addTool[setRecipeIn, setRecipeOut](s, d,
		"factorio_set_recipe",
		"Set the recipe of an assembling machine / chemical plant / refinery by 'unit_number'.",
		"set_recipe")

	addTool[moveItemsIn, moveItemsOut](s, d,
		"factorio_insert_items",
		"Move existing items from the character into a machine ('unit_number', optional 'inventory' like \"fuel\" or \"input\"). Does not create items; the character must already hold them. Returns the count actually inserted.",
		"insert_items")

	addTool[moveItemsIn, moveItemsOut](s, d,
		"factorio_remove_items",
		"Move items from a machine ('unit_number', optional 'inventory') into the character's inventory. Returns the count actually removed.",
		"remove_items")

	addTool[craftIn, craftOut](s, d,
		"factorio_craft",
		"Hand-craft a recipe 'count' times via the character's crafting queue. Consumes real ingredients and takes in-game time; poll factorio_get_inventory to see results. Returns how many crafts were started.",
		"craft")

	addTool[setResearchIn, researchOut](s, d,
		"factorio_set_research",
		"Set the current research to a technology. Completion still requires labs to consume science packs over time; poll factorio_get_research_state for progress.",
		"set_research")

	addTool[queueResearchIn, researchOut](s, d,
		"factorio_queue_research",
		"Append technologies to the research queue (in order). Completion requires science-pack production over time.",
		"queue_research")

	addTool[teleportIn, teleportOut](s, d,
		"factorio_teleport",
		"Teleport the player character to a position (optionally a named 'surface'). This is the one allowed navigation convenience, since real-time pathfinding is out of scope.",
		"teleport")

	addTool[mineResourceIn, mineResourceOut](s, d,
		"factorio_mine_resource",
		"Hand-mine a resource entity (ore/rock/tree) by 'position' or 'unit_number', adding realistic product amounts to the character's inventory and depleting the patch. 'count' optionally limits mining cycles.",
		"mine_resource")
}

type evalIn struct {
	Lua string `json:"lua"`
}

type evalOut struct {
	Result json.RawMessage `json:"result"`
}

func registerEval(s *mcp.Server, d factorio.Dispatcher) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "factorio_eval",
		Description: "ADVANCED/UNSAFE cheat & debug escape hatch: run arbitrary Lua in the game and return its result as JSON. Bypasses the legit economy. Disabled unless the server was started with eval enabled.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in evalIn) (*mcp.CallToolResult, evalOut, error) {
		data, err := d.Eval(ctx, in.Lua)
		if err != nil {
			return nil, evalOut{}, err
		}
		return nil, evalOut{Result: data}, nil
	})
}
