# factorio-mcp

[![CI](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/factorio-mcp/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/factorio-mcp/graph/badge.svg)](https://codecov.io/gh/danielriddell21/factorio-mcp)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_factorio-mcp&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_factorio-mcp)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A Claude Code plugin that lets Claude play Factorio 2.0 (base game, no Space Age) and work toward launching a rocket. Claude perceives live game state over RCON, reasons about it, and acts — under the real game economy, so placing a building consumes a crafted item and crafting and research take real time and ingredients.

## Usage

Install the companion mod, enable RCON and host a game — see [`docs/QUICKSTART.md`](docs/QUICKSTART.md).

```sh
brew install danielriddell21/tap/factorio-mcp

export FACTORIO_RCON_PASSWORD=changeme
factorio-mcp
```

Or from a clone: `just build && ./bin/factorio-mcp`.

| Env var | Default | Meaning |
| --- | --- | --- |
| `FACTORIO_RCON_HOST` | `127.0.0.1` | RCON host |
| `FACTORIO_RCON_PORT` | `27015` | RCON port |
| `FACTORIO_RCON_PASSWORD` | _(required)_ | RCON password |
| `FACTORIO_MCP_ALLOW_EVAL` | _(unset)_ | Set to enable the `eval` cheat tool |

Requires Go 1.26+. The Factorio side can't run in CI — verify it with the quickstart smoke test.
