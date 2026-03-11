# Flash Feature TODOs Implementation Plan

**Date:** 2026-03-10  
**Status:** Draft  
**Branch:** master → feature branch TBD

## Overview

Complete the two remaining TODOs from the initial flash feature implementation:
1. SWUpdate .swu file resolution (OutputRelPath is directory, need to find .swu file)
2. MP255 adapter ExecuteStep implementation (currently stub)

## Context

The flash feature was implemented with platform adapters for MP255 and SWUpdate. Both adapters have placeholder logic that needs completion:

- **SWUpdate**: `CanFlash()` checks if `a.imagePath` exists, but `OutputRelPath` is actually a directory containing `.swu` files. Need logic to find the correct `.swu` file.
- **MP255**: `ExecuteStep()` returns "not implemented" error. Need to integrate with `deploy.sh` script.

## Requirements

### 1. SWUpdate .swu File Resolution

**Current behavior:**
- `NewSWUpdateAdapter` receives `imagePath` (which is `OutputRelPath` from platform config)
- `OutputRelPath` is typically `buildroot/output/mp255-moxa/images` (a directory)
- `CanFlash()` checks if this directory exists
- `ExecuteStep()` passes this path to `augury_update`, which expects a `.swu` file

**Required behavior:**
- Detect if `imagePath` is a directory
- If directory, search for `.swu` files within it
- Select the appropriate `.swu` file (latest by mtime? only one expected?)
- Use the resolved `.swu` file path in `ExecuteStep()`
- Update `CanFlash()` to validate the resolved file

**Edge cases:**
- No `.swu` files found → error
- Multiple `.swu` files → pick latest by modification time
- Path is already a `.swu` file → use as-is

### 2. MP255 deploy.sh Integration

**Current behavior:**
- `GetSteps()` returns placeholder step
- `ExecuteStep()` returns "not implemented" error

**Required behavior:**
- Call `deploy.sh` with appropriate arguments based on method (uuu/manual)
- For UUU method: `deploy.sh -m uuu -r <releasePath>`
- For Manual method: `deploy.sh -m manual -r <releasePath>`
- Capture and return output from deploy.sh
- Handle interactive prompts from deploy.sh (if any)

**Deploy.sh behavior** (based on existing script at `/home/ngurfinkel/Repos/augury-node/yocto/deploy.sh`):
- Accepts `-m` (method: uuu, manual, rescue)
- Accepts `-r` (release directory)
- May prompt for confirmation
- Outputs progress to stdout/stderr

## Tasks

### Task 1: Add .swu File Resolution to SWUpdateAdapter

**Spec:**
- Add `ResolveSWUFile(basePath string) (string, error)` helper function in `swupdate_adapter.go`
- Logic:
  - Check if `basePath` is a file ending in `.swu` → return as-is
  - Check if `basePath` is a directory:
    - Use `filepath.Glob` to find `*.swu` files
    - If no files found → return error "no .swu files found in directory"
    - If one file found → return that path
    - If multiple files found → return latest by `os.Stat().ModTime()`
  - Otherwise → return error "invalid path"
- Update `NewSWUpdateAdapter` to call `ResolveSWUFile` and store resolved path in `a.imagePath`
- Update `CanFlash` to validate the resolved `.swu` file (not directory)
- Add tests for all cases in `swupdate_adapter_test.go`

**TDD Approach:**
1. Write tests for `ResolveSWUFile` helper (file, directory with 0/1/multiple .swu, invalid path)
2. Implement `ResolveSWUFile`
3. Update `NewSWUpdateAdapter` to use `ResolveSWUFile`
4. Update existing tests to create actual `.swu` files in temp directories
5. Verify all tests pass

**Acceptance Criteria:**
- `ResolveSWUFile` handles all edge cases with tests
- `NewSWUpdateAdapter` resolves `.swu` file at construction time
- `CanFlash` validates the resolved file (not directory)
- All existing `SWUpdateAdapter` tests still pass

### Task 2: Implement MP255Adapter ExecuteStep

**Spec:**
- Implement `ExecuteStep(ctx, step)` to call `deploy.sh`
- For now, treat the entire flash as a single step (not parsing deploy.sh internals)
- Command: `bash <root>/yocto/deploy.sh -m <method> -r <releasePath>`
- Use `exec.CommandContext` with `ctx` for cancellation support
- Capture combined output (stdout + stderr)
- Return output and error
- Update `GetSteps()` to return a single "Flash with deploy.sh" step (placeholder removal)

**TDD Approach:**
1. Write test for `ExecuteStep` that mocks `exec.Command` or uses a test script
2. Implement `ExecuteStep` with real `deploy.sh` call
3. Test with actual deploy.sh (manual/integration test)
4. Verify error handling and output capture

**Acceptance Criteria:**
- `ExecuteStep` calls `deploy.sh` with correct arguments
- Output is captured and returned
- Errors are propagated
- Context cancellation works
- Existing `MP255Adapter` tests still pass

### Task 3: Update Documentation

**Spec:**
- Update `/internal/flash/README.md`:
  - Remove "SWUpdate .swu file resolution" from TODO section
  - Remove "MP255 deploy.sh integration" from TODO section
  - Add notes about .swu resolution logic in architecture section
  - Add notes about deploy.sh integration in architecture section

**Acceptance Criteria:**
- README reflects completed work
- TODO section is updated

## Testing Strategy

**Unit Tests:**
- `TestResolveSWUFile_*` (file, dir with 0/1/many .swu, invalid)
- `TestSWUpdateAdapter_NewWithDirectory` (verify resolution at construction)
- `TestMP255Adapter_ExecuteStep` (with mock/test script)

**Integration Tests:**
- Manual: Build a platform, press `f`, select platform, attempt flash
- Verify `.swu` resolution works with real build output
- Verify `deploy.sh` is called with correct args (use dry-run if available)

**Pre-existing Tests:**
- All existing flash tests must continue to pass
- May need to update test fixtures to create `.swu` files

## Risks

1. **Multiple .swu files**: How to pick the right one? → Use latest by mtime (common pattern)
2. **deploy.sh interactivity**: Script may prompt for input → May need `expect` or similar (defer to later if needed)
3. **deploy.sh failure modes**: Script errors need clear propagation → Capture stderr, return as error
4. **Test isolation**: Real deploy.sh shouldn't run in tests → Use test fixtures or skip integration

## Rollout

1. Implement Task 1 (SWUpdate resolution) with full tests
2. Review and merge Task 1
3. Implement Task 2 (MP255 ExecuteStep) with tests
4. Review and merge Task 2
5. Update documentation (Task 3)

## Success Criteria

- SWUpdate adapter correctly resolves `.swu` files from output directories
- MP255 adapter can invoke `deploy.sh` and capture output
- All tests pass
- Documentation updated
- No regressions in existing flash functionality
