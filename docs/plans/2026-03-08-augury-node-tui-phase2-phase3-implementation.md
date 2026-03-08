# Augury Node TUI Phase 2/3 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement Phase 2 (cache control) and Phase 3 (polish) in a single PR using a refactor-first architecture with mandatory Nix execution gating.

**Architecture:** Introduce a shared action engine layer that owns capability checks, mandatory Nix gating, and job orchestration so screen models remain thin and consistent. Expand screens through typed actions and shared visual primitives (tables, viewports, diagrams) rather than ad-hoc command execution. Add log slicing, developer-downloads awareness, and visual progress/log panes without violating the script-only execution contract.

**Tech Stack:** Go 1.22, Bubble Tea, Bubbles (`table`, `viewport`, `progress`, `spinner`), Lip Gloss, existing `internal/run` runner, deterministic fixture-based integration tests.

---

### Task 0: Add Engine Action Contracts

**Files:**
- Create: `internal/engine/actions.go`
- Create: `internal/engine/actions_test.go`

**Step 1: Write the failing test**

Create `internal/engine/actions_test.go` with table-driven tests asserting action IDs and required metadata for:

- build-unit: `build`, `pull`, `delete`
- platform-cache: `pull`, `push`, `clean`
- hydration: `dry-run`, `run`
- validations: `all`, `shellcheck`, `bats`, `parse`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/engine -run TestActionContracts -v`  
Expected: FAIL (`internal/engine` package or symbols missing).

**Step 3: Write minimal implementation**

Create `internal/engine/actions.go`:

- `type ActionKind string`
- `type ActionTarget string`
- `type ActionRequest struct`
- typed constants for all required actions

**Step 4: Run test to verify it passes**

Run: `go test ./internal/engine -run TestActionContracts -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/engine/actions.go internal/engine/actions_test.go
git commit -m "refactor: add typed action contracts for phase2 and phase3"
```

---

### Task 1: Implement Mandatory Nix Gate Contracts

**Files:**
- Create: `internal/engine/nix_gate.go`
- Create: `internal/engine/nix_gate_test.go`

**Step 1: Write the failing test**

Add tests for:

- nix-ready state when probe command succeeds
- nix-blocked state when probe command fails
- no bypass behavior (all executable actions blocked when nix not ready)

**Step 2: Run test to verify it fails**

Run: `go test ./internal/engine -run TestNixGate -v`  
Expected: FAIL (missing nix gate APIs).

**Step 3: Write minimal implementation**

Implement:

- `type NixState struct { Ready bool; Reason string }`
- `func ProbeNix(root string) NixState`
- `func IsActionBlockedByNix(req ActionRequest, nix NixState) (bool, string)`

Probe command contract:

- `nix develop .#dev-env --command sh -c 'echo ready'`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/engine -run TestNixGate -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/engine/nix_gate.go internal/engine/nix_gate_test.go
git commit -m "feat: add mandatory nix readiness gate for executable actions"
```

---

### Task 2: Add Script Capability Matrix (Script-Only Policy)

**Files:**
- Create: `internal/engine/capabilities.go`
- Create: `internal/engine/capabilities_test.go`

**Step 1: Write the failing test**

Add tests asserting:

- each action maps to a required script path
- `available` only when script exists
- `not available` includes missing script reason
- no direct fallback command path exists

**Step 2: Run test to verify it fails**

Run: `go test ./internal/engine -run TestCapabilities -v`  
Expected: FAIL (missing capability map implementation).

**Step 3: Write minimal implementation**

Implement:

- `type Capability struct { Available bool; Reason string; ScriptPath string }`
- `func ResolveCapability(root string, req ActionRequest) Capability`
- script mapping table for all phase2/phase3 executable actions

**Step 4: Run test to verify it passes**

Run: `go test ./internal/engine -run TestCapabilities -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/engine/capabilities.go internal/engine/capabilities_test.go
git commit -m "feat: enforce script-only capability resolution for engine actions"
```

---

### Task 3: Add Engine Job Orchestration and Status Model

**Files:**
- Create: `internal/engine/jobs.go`
- Create: `internal/engine/jobs_test.go`

**Step 1: Write the failing test**

Add tests for:

- job lifecycle states (`queued`, `running`, `success`, `failed`, `cancelled`, `blocked`)
- job log path propagation
- action blocked status when capability/nix fails

**Step 2: Run test to verify it fails**

Run: `go test ./internal/engine -run TestJobs -v`  
Expected: FAIL (missing job orchestration).

**Step 3: Write minimal implementation**

Implement:

- `type JobState string`
- `type Job struct`
- `func ExecuteAction(ctx context.Context, root string, req ActionRequest) Job`

Use existing `internal/run.Execute` for command execution when not blocked.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/engine -run TestJobs -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/engine/jobs.go internal/engine/jobs_test.go
git commit -m "feat: add engine job lifecycle and runner integration"
```

---

### Task 4: Refactor Caches Screen for Phase 2 Actions

**Files:**
- Modify: `internal/caches/model.go`
- Modify: `internal/caches/model_test.go`
- Create: `internal/caches/actions.go`
- Create: `internal/caches/actions_test.go`

**Step 1: Write the failing test**

Add tests for:

- build-unit tab action key handling (`B`, `R`, `D`)
- platform-cache tab action key handling (`P`, `U`, `X`)
- disabled action behavior with capability reason
- destructive action confirmation requirement for `D` and `X`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/caches -v`  
Expected: FAIL (missing action behavior/confirm state).

**Step 3: Write minimal implementation**

Implement:

- tab-specific action routing
- confirm modal state for destructive actions
- engine action dispatch hooks
- visual action status in rows

**Step 4: Run test to verify it passes**

Run: `go test ./internal/caches -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/caches/model.go internal/caches/model_test.go internal/caches/actions.go internal/caches/actions_test.go
git commit -m "feat: add phase2 cache action controls with confirmations"
```

---

### Task 5: Refactor Hydration Screen for Interactive Execution

**Files:**
- Modify: `internal/hydration/model.go`
- Modify: `internal/hydration/model_test.go`

**Step 1: Write the failing test**

Add tests for:

- `D` triggers dry-run action preview
- `H` triggers hydration action for selected platforms
- blocked state when nix not ready
- not-available state when script missing

**Step 2: Run test to verify it fails**

Run: `go test ./internal/hydration -v`  
Expected: FAIL (missing key-driven execution flow).

**Step 3: Write minimal implementation**

Implement:

- key handlers for `D` and `H`
- engine action invocation per selected platform
- status rows showing blocked/not-available/success/failure

**Step 4: Run test to verify it passes**

Run: `go test ./internal/hydration -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/hydration/model.go internal/hydration/model_test.go
git commit -m "feat: make hydration screen interactive with nix-gated actions"
```

---

### Task 6: Refactor Validations Screen for Preset Execution

**Files:**
- Modify: `internal/validations/model.go`
- Modify: `internal/validations/model_test.go`

**Step 1: Write the failing test**

Add tests for:

- key mapping for preset execution
- action blocked when nix not ready
- not-available when required scripts missing
- result summary updates after run

**Step 2: Run test to verify it fails**

Run: `go test ./internal/validations -v`  
Expected: FAIL (execution behavior not wired to keys).

**Step 3: Write minimal implementation**

Implement:

- key handlers (`1`, `2`, `3`, `4`)
- engine action dispatch
- status + quick log summary rendering

**Step 4: Run test to verify it passes**

Run: `go test ./internal/validations -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/validations/model.go internal/validations/model_test.go
git commit -m "feat: add nix-gated validation preset execution"
```

---

### Task 7: Add Log Parsing and First-Error Extraction

**Files:**
- Create: `internal/logs/parser.go`
- Create: `internal/logs/parser_test.go`
- Modify: `internal/build/model.go`
- Modify: `internal/build/model_test.go`

**Step 1: Write the failing test**

Add tests for:

- first-error marker detection from log content
- context window extraction around first error
- key behavior for tab switching and jump-to-error (`e`)

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/logs -v
go test ./internal/build -run TestLog -v
```

Expected: FAIL (missing parser and UI hook).

**Step 3: Write minimal implementation**

Implement:

- parser with common error pattern list
- full/error tab state in build model
- navigation keys for log view

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/logs -v
go test ./internal/build -run TestLog -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/logs/parser.go internal/logs/parser_test.go internal/build/model.go internal/build/model_test.go
git commit -m "feat: add log slicing and first-error navigation"
```

---

### Task 8: Add Developer-Downloads Data Provider and UI Integration

**Files:**
- Create: `internal/data/developerdownloads/index.go`
- Create: `internal/data/developerdownloads/index_test.go`
- Modify: `internal/home/model.go`
- Modify: `internal/home/model_test.go`

**Step 1: Write the failing test**

Add tests for:

- parsing `developer-downloads/index.json`
- mapping per-platform source states (`built`, `hydrated`, `missing`)
- fallback `unavailable` when index file absent
- rendering state in home/detail view

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/data/developerdownloads -v
go test ./internal/home -run TestDeveloperDownloads -v
```

Expected: FAIL (provider not implemented).

**Step 3: Write minimal implementation**

Implement:

- index reader/parser
- home model field + render hook for per-platform source state

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/data/developerdownloads -v
go test ./internal/home -run TestDeveloperDownloads -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/data/developerdownloads/index.go internal/data/developerdownloads/index_test.go internal/home/model.go internal/home/model_test.go
git commit -m "feat: surface developer-downloads provenance in home view"
```

---

### Task 9: Add Visual Diagram Primitives and Screen Integration

**Files:**
- Create: `internal/visual/diagram/render.go`
- Create: `internal/visual/diagram/render_test.go`
- Modify: `internal/home/model.go`
- Modify: `internal/caches/model.go`
- Modify: `internal/validations/model.go`

**Step 1: Write the failing test**

Add tests for:

- deterministic diagram rendering for:
  - platform flow
  - cache topology
  - validation pipeline
- screen view includes diagram section when enough width is available

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/visual/diagram -v
go test ./internal/home ./internal/caches ./internal/validations -run TestDiagram -v
```

Expected: FAIL (diagram package not present).

**Step 3: Write minimal implementation**

Implement:

- reusable diagram renderers using Unicode box drawing
- screen-side integration using existing view model state

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/visual/diagram -v
go test ./internal/home ./internal/caches ./internal/validations -run TestDiagram -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/visual/diagram/render.go internal/visual/diagram/render_test.go internal/home/model.go internal/caches/model.go internal/validations/model.go
git commit -m "feat: add visual flow and topology diagrams for operational screens"
```

---

### Task 10: Enforce Nix Blocking in App-Level Action Routing

**Files:**
- Modify: `internal/app/model.go`
- Modify: `internal/app/model_test.go`
- Modify: `cmd/augury-node-tui/main.go`

**Step 1: Write the failing test**

Add tests for:

- action keypresses are blocked when Nix is not ready
- action keypresses are allowed when Nix is ready
- blocked reason is surfaced in UI state

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app -run TestNixGate -v`  
Expected: FAIL (no app-level gate wiring for action routes).

**Step 3: Write minimal implementation**

Implement:

- app-level nix readiness state propagation into child screens
- explicit refresh action to re-probe nix state
- hard block only on executable actions (navigation still allowed)

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app -run TestNixGate -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add internal/app/model.go internal/app/model_test.go cmd/augury-node-tui/main.go
git commit -m "feat: enforce mandatory nix gating for executable actions"
```

---

### Task 11: Update Operator Documentation for Phase 2/3

**Files:**
- Modify: `README.md`
- Modify: `docs/keybindings.md`
- Modify: `docs/configuration.md`
- Create: `docs/phase2-phase3.md`
- Modify: `docs/contract_test.go`

**Step 1: Write the failing docs test**

Add/extend docs contract tests to require documentation of:

- mandatory Nix policy
- cache action keys (`B/R/D`, `P/U/X`)
- log tab/error navigation keys
- developer-downloads source states

**Step 2: Run test to verify it fails**

Run: `go test ./docs/... -run TestDocsContract -v`  
Expected: FAIL (new requirements not yet documented).

**Step 3: Write minimal documentation updates**

Document exact operator behavior and keymaps.

**Step 4: Run test to verify it passes**

Run: `go test ./docs/... -run TestDocsContract -v`  
Expected: PASS.

**Step 5: Commit**

```bash
git add README.md docs/keybindings.md docs/configuration.md docs/phase2-phase3.md docs/contract_test.go
git commit -m "docs: document phase2 and phase3 controls with mandatory nix policy"
```

---

### Task 12: End-to-End Integration Coverage for Nix/Capabilities/Logs

**Files:**
- Modify: `integration/app_integration_test.go`
- Create: `integration/nix_gate_integration_test.go`
- Create: `integration/log_slicing_integration_test.go`
- Modify: `testdata/fake-augury-node/**` (only if needed for deterministic cases)

**Step 1: Write failing integration tests**

Add tests for:

- nix blocked mode (actions disabled/blocked)
- nix ready mode (actions execute)
- log slicing and first-error extraction from deterministic fixture logs

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./integration -run TestNix -v
go test ./integration -run TestLogSlicing -v
```

Expected: FAIL before implementation.

**Step 3: Write minimal implementation glue**

Adjust fixture/setup and integration helper behavior only as needed.

**Step 4: Run tests to verify they pass**

Run:

```bash
go test ./integration -v
go test ./... -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add integration/app_integration_test.go integration/nix_gate_integration_test.go integration/log_slicing_integration_test.go testdata/fake-augury-node
git commit -m "test: add phase2 and phase3 integration coverage for nix and logs"
```

---

### Final Verification Gate

Run in order:

```bash
go test ./... -v
go test ./integration -v
```

Expected:

- full suite passes
- no blocked TODOs for phase2/phase3 acceptance criteria

If anything fails, fix and create a new commit; do not amend unless explicitly requested.
