-- Shared helpers for the factorio_mcp companion mod.
local util = {}

-- err raises a structured handler error. The dispatcher turns the message into
-- the "detail" field of the JSON error envelope.
function util.err(msg)
  error(msg, 0)
end

-- get_player returns the LuaPlayer for args.player_index (default 1) or errors.
function util.get_player(args)
  local index = (args and args.player_index) or 1
  local player = game.get_player(index)
  if not player then
    util.err("no player at index " .. tostring(index))
  end
  return player
end

-- pos normalizes a {x, y} table into a position, erroring if missing.
function util.pos(p)
  if type(p) ~= "table" or p.x == nil or p.y == nil then
    util.err("expected a position {x=, y=}")
  end
  return { x = p.x, y = p.y }
end

-- contents_to_map converts a 2.0 inventory's contents array (a list of
-- {name, count, quality}) into a flat item-name -> total-count map.
function util.contents_to_map(inv)
  local out = {}
  if not inv then return out end
  for _, stack in pairs(inv.get_contents()) do
    out[stack.name] = (out[stack.name] or 0) + stack.count
  end
  return out
end

-- main_counts snapshots the player's main inventory as a name -> count map.
function util.main_counts(player)
  return util.contents_to_map(player.get_main_inventory())
end

-- diff_counts returns after - before for every item that changed.
function util.diff_counts(before, after)
  local out = {}
  for name, n in pairs(after) do
    local d = n - (before[name] or 0)
    if d ~= 0 then out[name] = d end
  end
  for name, n in pairs(before) do
    if after[name] == nil then out[name] = -n end
  end
  return out
end

-- status_names maps defines.entity_status values to readable strings, built
-- once on first use.
local status_names
function util.status_name(status)
  if status == nil then return nil end
  if not status_names then
    status_names = {}
    for name, value in pairs(defines.entity_status) do
      status_names[value] = name
    end
  end
  return status_names[status] or tostring(status)
end

-- find_one locates a single entity by unit_number or by position. Factorio has
-- no global by-unit-number lookup, so unit_number is resolved with a bounded
-- search around the player (or args.position), which covers a base under
-- construction. Pass args.search_radius to widen it.
function util.find_one(player, args)
  local surface = player.surface

  if args.position then
    local p = util.pos(args.position)
    local found = surface.find_entities_filtered({ position = p, radius = 0.7, name = args.name })
    local best, bestd
    for _, ent in pairs(found) do
      local dx, dy = ent.position.x - p.x, ent.position.y - p.y
      local d = dx * dx + dy * dy
      if not bestd or d < bestd then best, bestd = ent, d end
    end
    return best
  end

  if args.unit_number then
    local center = args.center and util.pos(args.center) or player.position
    local radius = args.search_radius or 512
    for _, ent in pairs(surface.find_entities_filtered({ position = center, radius = radius })) do
      if ent.unit_number == args.unit_number then return ent end
    end
    return nil
  end

  return nil
end

-- inventory_index maps a friendly inventory name to a defines.inventory value.
local INV
function util.inventory_index(name)
  if not name then return nil end
  if not INV then
    INV = {
      fuel = defines.inventory.fuel,
      chest = defines.inventory.chest,
      furnace_source = defines.inventory.furnace_source,
      furnace_result = defines.inventory.furnace_result,
      furnace_modules = defines.inventory.furnace_modules,
      input = defines.inventory.assembling_machine_input,
      assembling_machine_input = defines.inventory.assembling_machine_input,
      output = defines.inventory.assembling_machine_output,
      assembling_machine_output = defines.inventory.assembling_machine_output,
      modules = defines.inventory.assembling_machine_modules,
      lab_input = defines.inventory.lab_input,
    }
  end
  local idx = INV[name]
  if not idx then util.err("unknown inventory name: " .. tostring(name)) end
  return idx
end

return util
