# Configuration

augury-node-tui uses the augury-node repo root as its workspace.

## First-Time Setup

The recommended way to configure augury-node-tui is the setup wizard:

```bash
augury-node-tui setup
```

The wizard walks through: augury-node root path, Nix configuration, nix-users group, binary installation, and a verification build. Config is saved to `~/.config/augury-node-tui/config.toml`. See [setup-wizard.md](setup-wizard.md) for a detailed guide.

To re-run the wizard (e.g. to change the root path):

```bash
augury-node-tui setup --reconfigure
```

## Root Resolution Order

1. Flag (if added)
2. Config file path (if added)
3. Ancestor search from CWD for a directory containing `scripts/devices`, `scripts/lib`, and `pkg`

## Mandatory Nix

Executable actions require Nix. Run from `nix develop .#dev-env` or ensure `nix develop .#dev-env --command sh -c 'echo ready'` succeeds. When Nix is not ready, actions are blocked. See [phase2-phase3.md](phase2-phase3.md).

### Enabling Nix Experimental Features

If you see an error about `nix-command` being disabled, enable Nix experimental features:

```bash
# Option 1: User config (recommended)
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

# Option 2: System-wide (requires root)
echo "experimental-features = nix-command flakes" | sudo tee -a /etc/nix/nix.conf
```

After enabling, restart your shell and press `r` in the TUI to refresh Nix readiness.
