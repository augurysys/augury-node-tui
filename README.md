# augury-node-tui

TUI for augury-node builds.

## Startup splash behavior

On launch, a splash screen is shown. It auto-dismisses after a timeout or on any key press. From the home screen, press `a` to replay the splash.

## augury-node path contract

Run from an augury-node repo root. The root must contain:

- `scripts/devices`
- `scripts/lib`
- `pkg`

## Keybindings

See [docs/keybindings.md](docs/keybindings.md).

## Log file path contract

Build logs are written to `tmp/augury-node-tui/<platform>.log` under the augury-node root.
