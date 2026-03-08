# Augury Node TUI Phase 2/3 Design

**Date:** 2026-03-08  
**Status:** Approved  
**Delivery:** Single PR containing both Phase 2 and Phase 3

## Problem

`augury-node-tui` has a strong MVP baseline, but cache control and operator-grade visibility are still incomplete:

- cache operations are not controllable from the TUI
- log triage is still too raw for fast failure diagnosis
- developer-downloads provenance is not surfaced in UI
- Nix shell policy is not enforced at action boundaries
- visual hierarchy is still minimal for complex flows

## Scope Decisions (Locked)

- Implement **Phase 2 + Phase 3 in one PR**.
- Keep TUI **script-only** for operational actions:
  - no direct native `gsutil` fallback
  - no direct destructive shell fallback
- Add **full visual polish** (tables, diagrams, progress views), not functional text only.
- Phase 3 required now:
  - log slicing + first-error navigation
  - developer-downloads awareness
  - Nix support
- Nix policy is **mandatory** for action execution:
  - if Nix is not ready, actions are blocked
  - no bypass toggle in this phase

## Architecture: Refactor First

Before adding more feature logic, refactor structure to reduce coupling:

### 1) App shell and route boundaries

- `internal/app` handles route transitions and global messages only.
- screen modules own local state/view rendering.
- route-level interactions use shared message types from `internal/nav`.

### 2) Shared action engine

Add `internal/engine` as orchestration boundary between UI and command execution:

- `actions.go`: typed action requests (cache/unit/platform/hydration/validation/build)
- `capabilities.go`: checks script availability + Nix readiness
- `jobs.go`: queue/status/lifecycle for long-running actions
- `results.go`: normalized result model (status, logs, metadata)

Screen modules call engine APIs, not raw script paths.

### 3) Runner integration

Reuse `internal/run` for command execution and log persistence.
The engine maps action requests into `run.RunSpec` and publishes job updates.

### 4) Visual composition layer

Use Bubble Tea + Bubbles + Lip Gloss with a dedicated visual helper package:

- `table` for cache and platform matrices
- `viewport` for full/error log tabs
- `progress/spinner` for active jobs
- Unicode diagram rendering for flow/topology panes

## Mandatory Nix Contract

All executable actions (build, hydration, cache actions, validations) must run through:

`nix develop .#dev-env --command <script-or-command>`

### Enforcement behavior

- Preflight checks Nix readiness.
- If Nix unavailable:
  - action keys are disabled
  - UI displays blocking reason and remediation
  - navigation remains available
- No fallback direct command execution.

## Phase 2 Feature Design

## 2.1 Build-unit cache management

Enhance build-unit tab with actions:

- `B`: Build unit (force rebuild path, script-driven)
- `R`: Pull unit cache from remote (script-driven)
- `D`: Delete local unit cache (script-driven, confirm required)

Table columns:

- `unit`
- `fingerprint`
- `local`
- `remote`
- `last action`
- `availability`

If capability missing, action is disabled with reason text.

## 2.2 Platform cache management

Enhance platform tab with actions:

- `P`: Pull caches
- `U`: Push caches
- `X`: Clean local caches (confirm required)

Rows include:

- buildroot dl/ccache
- yocto downloads/sstate (when relevant)
- go/cargo cache presence
- remote presence/probe status

## 2.3 Build plan simulation improvements

Preflight plan for selected platforms should include:

- local cache presence
- remote cache/artifact presence (script-driven probe)
- expected path (`reuse`, `hydrate`, `rebuild`)
- force-rebuild state per platform

## Phase 3 Feature Design

## 3.1 Log slicing and first-error navigation

For build/validation logs:

- tabs:
  - `Full log`
  - `First error`
- keys:
  - `tab`: switch tab
  - `e`: jump to first error region
  - `j/k` and page keys for scroll

Parser rules use common error markers and return an anchored context window.

## 3.2 Developer-downloads awareness

Consume `developer-downloads/index.json` when present and show per-platform provenance:

- `built`
- `hydrated`
- `missing`
- `unavailable` (index absent)

Display in Home side panel and post-run summary.

## 3.3 Mandatory Nix UX

Home/status includes hard state:

- `Nix: ready`
- `Nix: blocking`

Add explicit command/key to re-check Nix readiness and refresh capability state.

## 3.4 Visual enhancements

Add high-value diagram panes (not decorative-only):

- platform flow diagram on Home/build detail
- cache topology map on Caches view
- validation pipeline status diagram on Validations view

Use consistent status colors for:

- running
- success
- failed
- cancelled
- blocked
- unavailable

## Interaction and Safety Rules

- destructive actions require explicit confirmation modal
- any action failure preserves logs and returns clear status
- blocked capability states are explicit and non-fatal
- script-only policy is always enforced

## Testing Strategy

### Unit tests

- capability resolution (script presence + Nix ready/not-ready)
- action-to-command translation
- log parser extraction
- developer-downloads parsing

### Model tests

- key routing and per-screen interaction
- disabled action behavior when blocked
- confirmation flows for destructive actions

### Integration tests

- deterministic fixture repo and fake scripts
- fake Nix gate:
  - ready path
  - blocked path
- end-to-end action lifecycle and log outputs

## Acceptance Criteria

Phase 2/3 is complete when:

1. cache actions are available from TUI and run only through scripts
2. Nix gating blocks all executable actions when unavailable
3. logs support full/error tabs with first-error navigation
4. developer-downloads source state appears in UI
5. visual diagrams/progress views are present in primary operational screens
6. full test suite passes in CI

## Risks and Mitigations

- **Refactor churn risk**
  - mitigate with incremental commits and preserved test coverage
- **Action drift from scripts**
  - mitigate with centralized engine capability map and strict script-only contract
- **UI complexity growth**
  - mitigate with module boundaries and shared visual primitives
- **Nix startup friction**
  - mitigate with clear blocked-state remediation and quick readiness recheck
