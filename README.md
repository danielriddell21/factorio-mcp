# factorio-mcp

[![CI](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/factorio-mcp/graph/badge.svg)](https://codecov.io/gh/danielriddell21/factorio-mcp)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_factorio-mcp&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_factorio-mcp)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A Claude Code plugin that lets Claude play Factorio 2.0 (base game, no Space Age) and work toward launching a rocket. Claude perceives live game state over RCON, reasons about it, and acts — under the real game economy, so placing a building consumes a crafted item and crafting and research take real time and ingredients.

Factorio can't run in CI, so setup is a manual path on your own machine.

## 1. Install the companion mod

Copy or symlink `mod/factorio_mcp` into your Factorio mods folder:

| OS | Mods folder |
| --- | --- |
| Windows | `%APPDATA%\Factorio\mods` |
| macOS | `~/Library/Application Support/factorio/mods` |
| Linux | `~/.factorio/mods` |

```sh
ln -s "$(pwd)/mod/factorio_mcp" "<MODS_DIR>/factorio_mcp"
# or package a zip: just mod-zip
```

Start Factorio, open **Mods**, enable **Factorio MCP Bridge**, restart if asked.

## 2. Enable RCON

Add to `config.ini` in Factorio's `config/` folder:

```ini
[other]
local-rcon-socket=127.0.0.1:27015
local-rcon-password=changeme
```

## 3. Host a game

**Multiplayer → Host new game**: freeplay, peaceful / no enemies, Space Age off. Hosting is what opens the RCON socket — single-player does not.

## 4. Run

```sh
brew install danielriddell21/tap/factorio-mcp

export FACTORIO_RCON_PASSWORD=changeme
factorio-mcp
```

Or from a clone: `just build && ./bin/factorio-mcp`. With the plugin installed, `.mcp.json` launches the binary automatically as long as `FACTORIO_RCON_PASSWORD` is set.

| Env var | Default | Meaning |
| --- | --- | --- |
| `FACTORIO_RCON_HOST` | `127.0.0.1` | RCON host |
| `FACTORIO_RCON_PORT` | `27015` | RCON port |
| `FACTORIO_RCON_PASSWORD` | _(required)_ | RCON password |
| `FACTORIO_MCP_ALLOW_EVAL` | _(unset)_ | Set to enable the `eval` cheat tool |

Ask Claude for `factorio_get_player_state` — if it returns your position, the Claude → MCP → RCON → mod chain is live.

## Troubleshooting

- `mod_not_loaded` — the mod isn't enabled in the active save.
- Connection errors — RCON off, wrong port/password, or the game isn't hosted.
- `missing item to place` — legit economy: craft or obtain the item first.

Requires Go 1.26+.
