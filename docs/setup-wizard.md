# Setup Wizard

The setup wizard (`augury-node-tui setup`) guides first-time configuration. Run it before using the main TUI.

## Steps

1. **Augury Node Root** — Enter or confirm the path to your augury-node repository. The wizard auto-detects if you run it from inside the repo.

2. **Nix Configuration** — Verifies Nix is installed, experimental features (nix-command, flakes) are enabled, and the daemon is accessible. Press `f` to auto-fix experimental features, or `s` to skip.

3. **Permissions** — Checks that your user is in the `nix-users` group. If not, run the shown command in a new terminal, then press `r` to re-check. Press `c` to copy the command, or `s` to skip.

4. **Binary Installation** — Optionally install a symlink to `/usr/local/bin/augury-node-tui` so you can run the binary from anywhere. Press `c` to copy the install command, or `s` to skip.

5. **Nix Build** — Runs a verification build to ensure the augury-node environment works. Press `r` to retry on failure, or `q` to quit.

6. **Success** — Setup complete. Press Enter to exit.

## Reconfiguring

To change settings (e.g. a different augury-node root):

```bash
augury-node-tui setup --reconfigure
```

The wizard shows "Reconfiguring..." in the title when an existing config is present.

## Troubleshooting

### "Nix not found in PATH"

Install Nix: <https://nixos.org/download.html>. Restart your shell after installation.

### "nix-command and flakes not enabled"

Add to `~/.config/nix/nix.conf`:

```
experimental-features = nix-command flakes
```

Or press `f` in the wizard to auto-fix.

### "Not in nix-users group"

Run:

```bash
sudo usermod -aG nix-users $USER
newgrp nix-users
```

Open a new terminal if `newgrp` does not apply to the current session.

### Root not auto-detected

Run the wizard from inside the augury-node repo, or enter the absolute path manually. The root must contain `scripts/devices`, `scripts/lib`, and `pkg`.
