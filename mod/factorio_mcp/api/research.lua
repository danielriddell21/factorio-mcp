-- Research handlers. These only choose what to research; completion still
-- requires labs to consume science packs over time.
local util = require("api.util")

local research = {}

local function validate(force, name)
  local tech = force.technologies[name]
  if not tech then util.err("unknown technology: " .. tostring(name)) end
  if tech.researched then util.err(name .. " is already researched") end
  return tech
end

local function snapshot(force)
  local queue = {}
  for _, tech in pairs(force.research_queue or {}) do
    queue[#queue + 1] = tech.name
  end
  return {
    current = force.current_research and force.current_research.name or "",
    queue = queue,
  }
end

function research.set_research(args)
  local player = util.get_player(args)
  local force = player.force
  if not args.technology then util.err("'technology' is required") end
  validate(force, args.technology)
  force.research_queue = { args.technology }
  return snapshot(force)
end

function research.queue_research(args)
  local player = util.get_player(args)
  local force = player.force
  if not args.technologies or #args.technologies == 0 then
    util.err("'technologies' must be a non-empty list")
  end
  for _, name in pairs(args.technologies) do
    validate(force, name)
    force.add_research(name)
  end
  return snapshot(force)
end

return research
