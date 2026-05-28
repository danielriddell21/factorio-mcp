-- Read-only game-state handlers.
local util = require("api.util")

local introspect = {}

function introspect.get_player_state(args)
  local player = util.get_player(args)
  local char = player.character
  return {
    position = { x = player.position.x, y = player.position.y },
    surface = player.surface.name,
    character_alive = char ~= nil and char.valid,
    health = char and char.health or 0,
    walking = player.walking_state and player.walking_state.walking or false,
  }
end

function introspect.get_inventory(args)
  local player = util.get_player(args)
  local which = args.which or "character"
  if which == "character" then
    return { items = util.contents_to_map(player.get_main_inventory()) }
  end
  if which == "entity" then
    local ent = util.find_one(player, args)
    if not ent then util.err("entity not found") end
    local inv
    if args.inventory then
      inv = ent.get_inventory(util.inventory_index(args.inventory))
    else
      inv = ent.get_output_inventory() or ent.get_main_inventory()
    end
    if not inv then util.err("entity has no such inventory") end
    return { items = util.contents_to_map(inv) }
  end
  util.err("unknown 'which': " .. tostring(which))
end

function introspect.scan_entities(args)
  local player = util.get_player(args)
  local filter = {
    position = args.center and util.pos(args.center) or player.position,
    radius = args.radius or 32,
  }
  if args.names and #args.names > 0 then filter.name = args.names end
  if args.types and #args.types > 0 then filter.type = args.types end

  local all = player.surface.find_entities_filtered(filter)
  table.sort(all, function(a, b)
    return (a.unit_number or 0) < (b.unit_number or 0)
  end)

  local page = args.page or 0
  local page_size = args.page_size or 50
  local start = page * page_size

  local entities = {}
  for i = start + 1, math.min(start + page_size, #all) do
    local e = all[i]
    local rec = e.type == "assembling-machine" and e.get_recipe() or nil
    entities[#entities + 1] = {
      unit_number = e.unit_number,
      name = e.name,
      type = e.type,
      position = { x = e.position.x, y = e.position.y },
      direction = e.direction,
      status = util.status_name(e.status),
      recipe = rec and rec.name or nil,
    }
  end

  local result = { entities = entities, page = page, total = #all }
  if start + page_size < #all then result.next_page = page + 1 end
  return result
end

function introspect.scan_resources(args)
  local player = util.get_player(args)
  local found = player.surface.find_entities_filtered({
    position = args.center and util.pos(args.center) or player.position,
    radius = args.radius or 64,
    type = "resource",
  })

  local agg = {}
  for _, e in pairs(found) do
    local a = agg[e.name]
    if not a then
      a = { amount = 0, tiles = 0, sx = 0, sy = 0 }
      agg[e.name] = a
    end
    a.amount = a.amount + (e.amount or 0)
    a.tiles = a.tiles + 1
    a.sx = a.sx + e.position.x
    a.sy = a.sy + e.position.y
  end

  local patches = {}
  for name, a in pairs(agg) do
    patches[#patches + 1] = {
      resource = name,
      amount = a.amount,
      tiles = a.tiles,
      center = { x = a.sx / a.tiles, y = a.sy / a.tiles },
    }
  end
  return { patches = patches }
end

function introspect.get_research_state(args)
  local player = util.get_player(args)
  local force = player.force

  local queue = {}
  for _, tech in pairs(force.research_queue or {}) do
    queue[#queue + 1] = tech.name
  end

  local researched = {}
  for name, tech in pairs(force.technologies) do
    if tech.researched then researched[#researched + 1] = name end
  end

  return {
    current = force.current_research and force.current_research.name or "",
    progress = force.research_progress or 0,
    queue = queue,
    researched = researched,
  }
end

function introspect.get_tech_tree(args)
  local player = util.get_player(args)
  local force = player.force
  local available_only = args.available_only

  local techs = {}
  for name, tech in pairs(force.technologies) do
    local prereqs = {}
    local all_met = true
    for pname, ptech in pairs(tech.prerequisites) do
      prereqs[#prereqs + 1] = pname
      if not ptech.researched then all_met = false end
    end

    local include = true
    if available_only then
      include = (not tech.researched) and all_met and tech.enabled
    end

    if include then
      local ingredients = {}
      for _, ing in pairs(tech.research_unit_ingredients) do
        ingredients[#ingredients + 1] = ing.name
      end
      techs[#techs + 1] = {
        name = name,
        researched = tech.researched,
        enabled = tech.enabled,
        prerequisites = prereqs,
        ingredients = ingredients,
      }
    end
  end
  return { technologies = techs }
end

local PRECISION = nil
local function precision_index(p)
  if not PRECISION then
    PRECISION = {
      ["1m"] = defines.flow_precision_index.one_minute,
      ["10m"] = defines.flow_precision_index.ten_minutes,
      ["1h"] = defines.flow_precision_index.one_hour,
      ["10h"] = defines.flow_precision_index.ten_hours,
    }
  end
  return PRECISION[p or "1m"]
end

function introspect.get_production_stats(args)
  local player = util.get_player(args)
  local force = player.force
  local stats = force.get_item_production_statistics(player.surface)
  local precision = args.precision or "1m"

  local function value(name, category)
    if precision == "all" then
      if category == "input" then
        return stats.get_input_count(name)
      end
      return stats.get_output_count(name)
    end
    local idx = precision_index(precision)
    if not idx then util.err("unknown precision: " .. tostring(precision)) end
    return stats.get_flow_count({ name = name, category = category, precision_index = idx, count = true })
  end

  local names = args.items
  if not names or #names == 0 then
    names = {}
    for name in pairs(prototypes.item) do names[#names + 1] = name end
  end

  local produced, consumed = {}, {}
  for _, name in pairs(names) do
    local p = value(name, "input")
    local c = value(name, "output")
    if p and p ~= 0 then produced[name] = p end
    if c and c ~= 0 then consumed[name] = c end
  end
  return { produced = produced, consumed = consumed }
end

function introspect.get_recipe_info(args)
  local player = util.get_player(args)
  if not args.name then util.err("recipe 'name' is required") end
  local recipe = player.force.recipes[args.name]
  if not recipe then util.err("unknown recipe: " .. tostring(args.name)) end

  local function stacks(list)
    local out = {}
    for _, s in pairs(list) do
      out[#out + 1] = { name = s.name, amount = s.amount, type = s.type }
    end
    return out
  end

  return {
    name = recipe.name,
    category = recipe.category,
    energy = recipe.energy,
    enabled = recipe.enabled,
    ingredients = stacks(recipe.ingredients),
    products = stacks(recipe.products),
  }
end

return introspect
