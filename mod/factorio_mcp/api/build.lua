-- World-mutating handlers that respect the real economy.
local util = require("api.util")

local build = {}

-- placing_item returns the item name/count that places the given entity, using
-- the prototype's items_to_place_this (falls back to the entity name).
local function placing_item(name)
  local proto = prototypes.entity[name]
  if proto and proto.items_to_place_this and proto.items_to_place_this[1] then
    local it = proto.items_to_place_this[1]
    return it.name, it.count or 1
  end
  return name, 1
end

function build.place_entity(args)
  local player = util.get_player(args)
  if not args.name then util.err("'name' is required") end
  local position = util.pos(args.position)
  local direction = args.direction or defines.direction.north
  local surface = player.surface

  if not prototypes.entity[args.name] then
    util.err("unknown entity: " .. tostring(args.name))
  end

  -- Legit economy: the placing item must already be in the inventory.
  local item_name, item_count = placing_item(args.name)
  local have = player.get_item_count(item_name)
  if have < item_count then
    util.err(string.format("missing item to place: need %d %s, have %d", item_count, item_name, have))
  end

  if not surface.can_place_entity({
    name = args.name,
    position = position,
    direction = direction,
    force = player.force,
    build_check_type = defines.build_check_type.manual,
  }) then
    util.err("cannot place " .. args.name .. " here (blocked or invalid position)")
  end

  local ent = surface.create_entity({
    name = args.name,
    position = position,
    direction = direction,
    force = player.force,
    player = player,
    raise_built = true,
  })
  if not ent then util.err("placement failed") end

  player.remove_item({ name = item_name, count = item_count })

  if args.recipe then
    pcall(function() ent.set_recipe(args.recipe) end)
  end

  return {
    unit_number = ent.unit_number,
    position = { x = ent.position.x, y = ent.position.y },
  }
end

function build.remove_entity(args)
  local player = util.get_player(args)
  local ent = util.find_one(player, args)
  if not ent then util.err("entity not found") end

  local before = util.main_counts(player)
  local ok = player.mine_entity(ent, true)
  local after = util.main_counts(player)

  return { removed = ok == true, returned = util.diff_counts(before, after) }
end

function build.set_recipe(args)
  local player = util.get_player(args)
  if not args.recipe then util.err("'recipe' is required") end
  local ent = util.find_one(player, args)
  if not ent then util.err("entity not found") end
  local ok, errmsg = pcall(function() ent.set_recipe(args.recipe) end)
  if not ok then util.err("could not set recipe: " .. tostring(errmsg)) end
  return { ok = true }
end

function build.teleport(args)
  local player = util.get_player(args)
  local position = util.pos(args.position)
  local surface = args.surface and game.surfaces[args.surface] or player.surface
  if not surface then util.err("unknown surface: " .. tostring(args.surface)) end
  local ok = player.teleport(position, surface)
  return { ok = ok == true, position = { x = player.position.x, y = player.position.y } }
end

function build.mine_resource(args)
  local player = util.get_player(args)
  local count = args.count or 1
  local before = util.main_counts(player)

  for _ = 1, count do
    local ent = util.find_one(player, args)
    if not ent or not ent.valid then break end
    if not player.mine_entity(ent) then break end
  end

  local after = util.main_counts(player)
  return { mined = util.diff_counts(before, after) }
end

return build
