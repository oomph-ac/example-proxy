# example-proxy

Example Oomph proxy that uses Spectrum to sit in front of a Bedrock server and run the Oomph processing pipeline.

## What this gives you out-of-the-box
- **Transparent proxy**: Listens on a local address and forwards to your remote/backup server.
- **Oomph pipeline wired in**: Components and detections are registered and process client/server packets.
- **Built-in moderator tooling**: Simple `moderators.list` gate for permissions and `/ac` subcommands for alerts, logs, and debugging.
- **Resource pack support**: Optionally loads server packs and enforces them.
- **Graceful shutdown**: Broadcasts a transfer to a backup address (if configured) or disconnects with a message.
- **Operational niceties**: Per-player log files, optional pprof endpoint.

## Quick start

### Prerequisites
- Go 1.24+
- Git (for the update_deps script)

### Setup
1) Obtain the oomph-specific dependencies by running the `.update_deps.sh` script. This script will also install additional dependencies for Oomph and make sure the proxy binary is ready to be compiled.

2) Build the proxy with `go build -o {proxy binary name} -ldflags='-s -w'` and run it `./{proxy binary name}`. It will generate a dummy configuration file which then you would be able to edit.

3) (Optional) Add moderator names to `moderators.list` (one player name per line):

```
YourModeratorIGN
AnotherMod
TheNextModHere
```

## Moderator commands (granted via `moderators.list`)
- **/ac alerts [true|false|enable|disable|delayMs]**: Enable/disable alerts, or set delay in ms.
- **/ac logs <player>**: Print the player’s current detections and violation multipliers.
- **/ac debug <mode>**: Toggle Oomph debug modes. Special cases:
  - `type_message` or `type_log` switches the debug output sink.
  - `gmc` sets client-side creative mode when `OOMPH_GAMEMODE_TEST_BECAUSE_DEV` is set.

## Configuration reference (brief)
- **Global.LocalAddress**: Address the proxy listens on (e.g., `0.0.0.0:19132`).
- **Global.RemoteAddress**: Upstream Bedrock server address.
- **Global.BackupAddress**: Optional `host:port`. On shutdown, online players are transferred here.
- **Global.ShutdownMessage**: Message shown when disconnecting without a backup.
- **Global.GCPercent**: Go GC target percentage. Values lower than 100 can cause high CPU; 100 is recommended.
- **Global.MemThreshold**: Soft memory limit (MB) for the process.
- **Global.Resource.ResourceFolder**: Folder containing resource packs; `content_keys.json` is supported if present.
- **Global.Resource.RequirePacks**: Whether to enforce pack downloads.

Alert message templates for detections can be defined via config and support placeholders used by the handler: `{prefix}`, `{player}`, `{xuid}`, `{detection_type}`, `{detection_subtype}`, `{violations}`.

## Operational notes
- **Logs**: Per-player logs are written to `./logs/<player>.log`.
- **pprof**: Set `PPROF_ENABLED=1` and `PPROF_ADDRESS=host:port` to expose pprof.

## Advanced/custom integrations
Basic proxying and Oomph processing are implemented here. For advanced requirements—such as integrating databases, custom punishment/storage backends, dashboards, webhooks, multi-proxy coordination, or bespoke detections—please commission a custom implementation via the Oomph Discord.