# factorio-mcp

> *factorio-mcp* — Claude plays Factorio, under the real economy.

[![CI](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/factorio-mcp/graph/badge.svg)](https://codecov.io/gh/danielriddell21/factorio-mcp)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_factorio-mcp&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_factorio-mcp)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A Claude Code plugin that lets Claude play **Factorio 2.0** (base game, no Space Age) and work toward **launching a rocket**.

Inspired by "Factorio plays itself" Lua-scripting demos, but instead of a fixed scripted strategy the brain is an LLM: Claude perceives live game state, reasons about it, acts, and recovers from the unexpected. The economy is legit — placing a building consumes a crafted item, and crafting and research take real time and ingredients. Only navigation is a concession, since real-time pathfinding is out of scope.

```
Claude Code ──stdio──> Go MCP server ──RCON/TCP──> Factorio + companion mod
```

- **Go MCP server** (`cmd/factorio-mcp`) exposes typed `factorio_*` tools over RCON, one `/silent-command` per call.
- **Companion mod** (`mod/factorio_mcp`) registers a remote interface with a named API and returns state as JSON.
- **Skills** (`skills/`) encode the strategy: orientation, bootstrap, smelting, science, research, mall builds, rocket.

## Install

```sh
brew install danielriddell21/tap/factorio-mcp
```

## Quickstart

See [`docs/QUICKSTART.md`](docs/QUICKSTART.md) to install the mod, enable RCON and host a game. Then:

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
| `FACTORIO_MCP_ALLOW_EVAL` | _(unset)_ | Set to enable the `eval` cheat tool |

Requires Go 1.26+. The Factorio side can't run in CI — verify it with the quickstart smoke test.
