# augury-node-tui

TUI for augury-node builds.

## Quick Start

```bash
# 1. Build the TUI
go build -o augury-node-tui ./cmd/augury-node-tui

# 2. Enable Nix experimental features (first time only)
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

# 3. Run from augury-node workspace
cd /path/to/augury-node
nix develop .#dev-env
/path/to/augury-node-tui/augury-node-tui
```

The home screen shows Nix readiness status. Press `r` to refresh after enabling features.

## Startup splash behavior

On launch, a splash screen is shown. It auto-dismisses after a timeout or on any key press. From the home screen, press `a` to replay the splash.

## augury-node path contract

Run from an augury-node repo root. The root must contain:

- `scripts/devices`
- `scripts/lib`
- `pkg`

## Keybindings

See [docs/keybindings.md](docs/keybindings.md).

## Phase 2 and Phase 3

See [docs/phase2-phase3.md](docs/phase2-phase3.md) for cache actions, log navigation, developer-downloads states, and mandatory Nix policy.

## Log file path contract

Build logs are written to `tmp/augury-node-tui/<platform>.log` under the augury-node root.
