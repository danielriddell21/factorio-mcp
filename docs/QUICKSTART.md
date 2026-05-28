# Quickstart

Get Claude playing Factorio 2.0 end to end. Factorio can't run in CI, so this is
the manual path on your own machine.

## 1. Install the companion mod

Copy or symlink `mod/factorio_mcp` into your Factorio mods folder:

- **Windows:** `%APPDATA%\Factorio\mods`
- **macOS:** `~/Library/Application Support/factorio/mods`
- **Linux:** `~/.factorio/mods`

```sh
ln -s "$(pwd)/mod/factorio_mcp" "<MODS_DIR>/factorio_mcp"
# or package a zip: just mod-zip  ->  copy dist/factorio_mcp_0.1.0.zip to <MODS_DIR>
```

Start Factorio, open **Mods**, enable **Factorio MCP Bridge**, restart if asked.

## 2. Enable RCON

RCON works in multiplayer, including when you host solo. Add to `config.ini`
(in Factorio's `config/` folder):

```ini
[other]
local-rcon-socket=127.0.0.1:27015
local-rcon-password=changeme
```

## 3. Host a game

**Multiplayer -> Host new game**: freeplay, **peaceful / no enemies**,
**Space Age off**. Hosting (not single-player) is what opens the RCON socket.

## 4. Build and run the server

```sh
just build
export FACTORIO_RCON_PASSWORD=changeme
# optional: FACTORIO_RCON_HOST, FACTORIO_RCON_PORT, FACTORIO_MCP_ALLOW_EVAL=1
./bin/factorio-mcp
```

With the plugin installed, `.mcp.json` launches `bin/factorio-mcp` automatically
as long as `FACTORIO_RCON_PASSWORD` is set in your environment.

## 5. Smoke test

Ask Claude (or call the tools) in order and confirm each:

1. `factorio_get_player_state` -> your position/surface.
2. `factorio_scan_resources` -> nearby ore patches.
3. `factorio_mine_resource` near stone, then `factorio_craft` `stone-furnace` x1.
4. `factorio_place_entity` `stone-furnace` on a clear tile.
5. `factorio_scan_entities` (names: `["stone-furnace"]`) -> the furnace shows up.

If all five work, the full Claude -> MCP -> RCON -> mod chain is live. From here,
the skills guide the bootstrap toward launching a rocket.

## Troubleshooting

- `mod_not_loaded` — the mod isn't enabled in the active save.
- connection errors — RCON off, wrong port/password, or the game isn't hosted.
- `missing item to place` — legit economy: craft or obtain the item first.
