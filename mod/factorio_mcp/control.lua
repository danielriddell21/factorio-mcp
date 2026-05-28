-- factorio_mcp companion mod.
--
-- Exposes a single "factorio_mcp" remote interface with two functions:
--   dispatch(payload_json) -> envelope_json   (named, typed API)
--   eval(lua_source)       -> envelope_json   (cheat/debug escape hatch)
--
-- The MCP server calls these over RCON via /silent-command. All game logic
-- lives here; the Go side only ships JSON. The interface is registered at the
-- top level so it is available to RCON immediately.

local introspect = require("api.introspect")
local build = require("api.build")
local inventory = require("api.inventory")
local research = require("api.research")

-- handlers maps a function name to its implementation. Each takes a decoded
-- args table and returns a plain Lua table (serialized as the envelope "data").
local handlers = {
  get_player_state = introspect.get_player_state,
  get_inventory = introspect.get_inventory,
  scan_entities = introspect.scan_entities,
  scan_resources = introspect.scan_resources,
  get_research_state = introspect.get_research_state,
  get_tech_tree = introspect.get_tech_tree,
  get_production_stats = introspect.get_production_stats,
  get_recipe_info = introspect.get_recipe_info,

  place_entity = build.place_entity,
  remove_entity = build.remove_entity,
  set_recipe = build.set_recipe,
  teleport = build.teleport,
  mine_resource = build.mine_resource,

  insert_items = inventory.insert_items,
  remove_items = inventory.remove_items,
  craft = inventory.craft,

  set_research = research.set_research,
  queue_research = research.queue_research,
}

local function encode(tbl)
  local ok, json = pcall(helpers.table_to_json, tbl)
  if ok then return json end
  return '{"ok":false,"error":"encode_error"}'
end

local function dispatch(payload)
  local ok_decode, req = pcall(helpers.json_to_table, payload)
  if not ok_decode or type(req) ~= "table" then
    return encode({ ok = false, error = "bad_payload" })
  end

  local handler = handlers[req.fn]
  if not handler then
    return encode({ ok = false, error = "unknown_fn", detail = tostring(req.fn) })
  end

  local ok, result = pcall(handler, req.args or {})
  if not ok then
    return encode({ ok = false, error = "handler_error", detail = tostring(result) })
  end
  return encode({ ok = true, data = result })
end

local function eval(source)
  local fn, compile_err = load(source)
  if not fn then
    return encode({ ok = false, error = "compile_error", detail = tostring(compile_err) })
  end
  local ok, result = pcall(fn)
  if not ok then
    return encode({ ok = false, error = "runtime_error", detail = tostring(result) })
  end
  local t = type(result)
  if t == "table" or t == "string" or t == "number" or t == "boolean" or t == "nil" then
    return encode({ ok = true, data = result })
  end
  return encode({ ok = true, data = tostring(result) })
end

remote.add_interface("factorio_mcp", {
  dispatch = dispatch,
  eval = eval,
})
