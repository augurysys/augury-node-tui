# Configuration

augury-node-tui uses the augury-node repo root as its workspace. Resolution order:

1. Flag (if added)
2. Config file path (if added)
3. Ancestor search from CWD for a directory containing `scripts/devices`, `scripts/lib`, and `pkg`

## Mandatory Nix

Executable actions require Nix. Run from `nix develop .#dev-env` or ensure `nix develop .#dev-env --command sh -c 'echo ready'` succeeds. When Nix is not ready, actions are blocked. See [phase2-phase3.md](phase2-phase3.md).
