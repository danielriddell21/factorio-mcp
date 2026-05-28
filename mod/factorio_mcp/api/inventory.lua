-- Item-movement and crafting handlers (real economy: items must exist).
local util = require("api.util")

local inventory = {}

local function target_inventory(player, args)
  local ent = util.find_one(player, args)
  if not ent then util.err("target entity not found") end
  local inv
  if args.inventory then
    inv = ent.get_inventory(util.inventory_index(args.inventory))
  else
    inv = ent.get_main_inventory() or ent.get_output_inventory()
  end
  if not inv then util.err("entity has no usable inventory") end
  return ent, inv
end

-- insert_items moves items the character already holds into a machine.
function inventory.insert_items(args)
  local player = util.get_player(args)
  if not args.item then util.err("'item' is required") end
  local want = args.count or 0
  if want <= 0 then util.err("'count' must be > 0") end

  local available = player.get_item_count(args.item)
  if available <= 0 then util.err("character holds no " .. args.item) end
  local to_move = math.min(want, available)

  local ent = select(1, target_inventory(player, args))
  local inserted
  if args.inventory then
    local inv = ent.get_inventory(util.inventory_index(args.inventory))
    inserted = inv.insert({ name = args.item, count = to_move })
  else
    inserted = ent.insert({ name = args.item, count = to_move })
  end

  if inserted > 0 then
    player.remove_item({ name = args.item, count = inserted })
  end
  return { count = inserted }
end

-- remove_items moves items out of a machine into the character.
function inventory.remove_items(args)
  local player = util.get_player(args)
  if not args.item then util.err("'item' is required") end
  local want = args.count or 0
  if want <= 0 then util.err("'count' must be > 0") end

  local _, inv = target_inventory(player, args)
  local available = inv.get_item_count(args.item)
  if available <= 0 then util.err("target holds no " .. args.item) end
  local to_move = math.min(want, available)

  local removed = inv.remove({ name = args.item, count = to_move })
  if removed > 0 then
    player.insert({ name = args.item, count = removed })
  end
  return { count = removed }
end

-- craft hand-crafts a recipe through the player's crafting queue.
function inventory.craft(args)
  local player = util.get_player(args)
  if not args.recipe then util.err("'recipe' is required") end
  local count = args.count or 1
  local started = player.begin_crafting({ recipe = args.recipe, count = count })
  return { started = started }
end

return inventory
