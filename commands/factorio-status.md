---
description: Summarize the current Factorio game state (player, nearby entities, resources, research, production).
---

Report the current state of the running Factorio game using the `factorio_*` MCP
tools. Call them as needed and present a concise summary:

1. `factorio_get_player_state` — position, surface, health.
2. `factorio_scan_resources` — nearby ore/oil patches.
3. `factorio_scan_entities` — what's built nearby (group by name/type, note any
   bad `status` like no_power / no_ingredients).
4. `factorio_get_research_state` — current research, progress, queue.
5. `factorio_get_production_stats` (precision "1m") — top produced/consumed items.

Then give a short assessment: what's working, the current bottleneck, and the
single most useful next action toward launching a rocket.
