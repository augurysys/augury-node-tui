# CI Dashboard Screen Design

## Problem

CircleCI's v2 API has no endpoint for downloading raw build step output.
The augury-node pipeline now captures all job stdout/stderr as CircleCI
artifacts (via BASH_ENV auto-tee), but there is no way to browse pipeline
status or view those logs without leaving the terminal.

## Goals

1. Show the current branch's latest CI pipeline status in the TUI
2. List all jobs with their status, duration, and log availability
3. Download and display job logs inline using the existing LogViewer
4. Degrade gracefully when no CircleCI token is configured

## Approach

HTTP client for CircleCI v2 API metadata, download log artifacts to disk,
LogViewer reads from local files. This follows the existing pattern where
build logs go to `tmp/augury-node-tui/<platform>.log`.

## Package Structure

```
internal/
  ci/
    client.go       # CircleCI v2 API HTTP client
    model.go        # Bubbletea model for the CI screen
    messages.go     # Tea messages (pipeline loaded, jobs loaded, log downloaded, errors)
    types.go        # API response structs (Pipeline, Workflow, Job, Artifact)
  config/
    config.go       # Add CircleToken field
  setup/
    ...             # Add optional CircleCI token step
```

The `ci` package is self-contained. It uses stdlib `net/http` and
`encoding/json` -- no new dependencies.

## CircleCI API Client

Project slug is derived from the git remote URL already parsed by
`internal/status/` (e.g. `gh/augurysys/augury-node`).

API calls (all GET, read-only):

| Endpoint | Purpose | When |
|---|---|---|
| `GET /v2/project/{slug}/pipeline?branch={branch}` | Latest pipeline | Screen entry |
| `GET /v2/pipeline/{id}/workflow` | Workflows | After pipeline |
| `GET /v2/workflow/{id}/job` | Jobs list | After workflow |
| `GET /v2/project/{slug}/{job_number}/artifacts` | Artifact URLs | User selects job |
| `GET {artifact_url}` | Download log file | User presses Enter |

All calls are async via `tea.Cmd`. The UI shows a spinner while loading.

Errors (network, 401, 404, empty results) are typed and rendered as
user-visible status messages.

## Screen Layout

```
+--[CI Pipeline: branch-name]-------------------------------+
|  Pipeline #1234  |  SHA: abc123  |  Status: success       |
+-----------------------------------------------------------|
|  DataTable: Jobs                                          |
|  NAME                  | STATUS    | DURATION  | LOGS     |
|  validate-certificates | success   | 0:42      | [1 file] |
|  lint-and-test         | success   | 1:15      | [1 file] |
|  build-node2           | failed    | 45:22     | [1 file] |
|  build-mp255-ulrpm     | success   | 112:05    | [1 file] |
+-----------------------------------------------------------+
|  [Enter] View logs  [r] Refresh  [Esc] Back               |
+-----------------------------------------------------------+
```

Components reused:
- Card (Emphasized) for pipeline header
- DataTable for jobs list
- StatusBadge for job status
- LogViewer for log content (sub-view)

## Navigation

- Home screen: `c` keybind navigates to `ci` route
- CI screen: Enter on a job downloads log, opens LogViewer sub-view
- Esc from LogViewer: back to jobs table
- Esc from jobs table: back to home

## Screen States

1. `loading` -- fetching pipeline/workflow/jobs (spinner)
2. `ready` -- jobs table displayed
3. `downloading` -- fetching log artifact (spinner)
4. `viewing` -- LogViewer showing log content
5. `error` -- API error with message
6. `no_token` -- no token configured, shows setup hint

## Authentication

Token resolution order:
1. `CIRCLE_TOKEN` environment variable
2. `config.toml` field `CircleToken`
3. If neither: `no_token` state with message

Config changes:
- Add `CircleToken string` to Config struct
- Stored in plaintext in `~/.config/augury-node-tui/config.toml`

Setup wizard:
- New optional step after existing steps
- Prompt: "Enter your CircleCI personal API token (optional)"
- Saved to config.toml if provided

## Log Download

Download location: `{augury-node-root}/tmp/augury-node-tui/ci-logs/{job-name}.log`

This follows the existing convention for build logs. Files persist for
re-inspection. No automatic cleanup (same as build logs).

## Project Slug Derivation

Parse git remote URL from RepoStatus (already available):
- `git@github.com:augurysys/augury-node.git` -> `gh/augurysys/augury-node`
- `https://github.com/augurysys/augury-node.git` -> `gh/augurysys/augury-node`

## What Does Not Change

- No changes to existing screens
- No changes to existing components (reused as-is)
- No new Go dependencies
- No changes to augury-node CI pipeline
- CI screen is optional (degrades to "no token" message)

## Alternatives Considered

**Native HTTP + in-memory logs:** Cleaner architecture but logs don't
persist. User can't re-inspect without re-downloading.

**Shell out to curl:** Consistent with engine/run pattern but fragile,
depends on curl/jq being installed, harder to test.
