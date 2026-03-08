# Phase 2 and Phase 3 Controls

## Mandatory Nix Policy

All executable actions (build, hydration, cache actions, validations) run through Nix. The TUI checks Nix readiness before allowing actions.

- **Nix ready**: Actions are enabled. Commands run via `nix develop .#dev-env --command <script>`.
- **Nix block**: When Nix is not ready, action keys are disabled. Navigation remains available. The UI shows the blocking reason and remediation.

There is no bypass. If Nix is unavailable, actions are blocked.

## Cache Action Keys

### Build-unit tab

| Key | Action |
|-----|--------|
| B | Build unit (force rebuild) |
| R | Pull unit cache from remote |
| D | Delete local unit cache (confirm required) |

### Platform tab

| Key | Action |
|-----|--------|
| P | Pull caches |
| U | Push caches |
| X | Clean local caches (confirm required) |

## Log Tab and Error Navigation

Build and validation logs support:

- **Full log** tab: full output
- **First error** tab: context around first detected error

Keys:

- `tab`: switch between Full log and First error
- `e`: jump to first error region
- `j`/`k` and page keys: scroll

## Developer-downloads Source States

When `developer-downloads/index.json` is present, the TUI shows per-platform provenance:

- **built**: built locally
- **hydrated**: hydrated from cache
- **missing**: expected but absent
- **unavailable**: index file absent or unreadable
