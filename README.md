# factorio-mcp

A Claude Code plugin that lets Claude play **Factorio 2.0** (base game, no Space
Age) and work toward **launching a rocket** — under the real game economy.

Instead of a fixed scripted strategy, the brain is an LLM: Claude perceives live
game state, reasons about it, acts, and recovers from the unexpected.

## How it works

```
Claude Code ──stdio──> Go MCP server ──RCON/TCP──> Factorio + companion mod
```

- **Go MCP server** (`cmd/factorio-mcp`) exposes typed `factorio_*` tools and
  talks to the game over RCON, one `/silent-command` per call.
- **Companion mod** (`mod/factorio_mcp`) registers a `factorio_mcp` remote
  interface with a clean named API (plus a guarded `eval` escape hatch) and
  returns state as JSON.
- **Skills** (`skills/`) encode the strategy: orientation, early bootstrap,
  smelting, red/green/blue science, research progression, mall builds, rocket.

The economy is **legit**: placing a building consumes a crafted item; crafting
and research take real time and ingredients. Only navigation is a concession
(`factorio_teleport`, since real-time pathfinding is out of scope).

## Tools

Introspection: `get_player_state`, `get_inventory`, `scan_entities`,
`scan_resources`, `get_research_state`, `get_tech_tree`, `get_production_stats`,
`get_recipe_info`.

Actions: `place_entity`, `remove_entity`, `set_recipe`, `insert_items`,
`remove_items`, `craft`, `set_research`, `queue_research`, `teleport`,
`mine_resource`. Optional cheat/debug: `eval` (off by default).

All are exposed as MCP tools prefixed `factorio_`.

## Quickstart

See [`docs/QUICKSTART.md`](docs/QUICKSTART.md) — install the mod, enable RCON,
host a game, then:

```sh
just build
export FACTORIO_RCON_PASSWORD=changeme
./bin/factorio-mcp
```

## Configuration

| Env var | Default | Meaning |
| --- | --- | --- |
| `FACTORIO_RCON_HOST` | `127.0.0.1` | RCON host |
| `FACTORIO_RCON_PORT` | `27015` | RCON port |
| `FACTORIO_RCON_PASSWORD` | _(required)_ | RCON password |
| `FACTORIO_MCP_ALLOW_EVAL` | _(unset)_ | set to enable the `eval` cheat tool |

## Development

```sh
just test     # unit + integration tests (no game required)
just fuzz     # fuzz the Lua escaping
just vet      # go vet
just fmt      # gofmt -s -w
just mod-zip  # package the mod
just ci       # fmt-check + vet + test
```

Requires Go 1.25+. The Factorio side can't run in CI; verify it with the
quickstart smoke test.
