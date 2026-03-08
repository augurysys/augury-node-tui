# Onboarding Wizard Design

**Date:** 2026-03-08  
**Status:** Approved

## Problem Statement

Current onboarding requires manual steps that are error-prone:
- Nix first-time download has no progress explanation
- Permission errors (daemon socket, group membership) require sudo troubleshooting
- Binary path must be hardcoded or added to PATH manually
- Config/paths not persisted between runs
- Root resolution fails even when in correct directory

Users experience frustration during first-run and can't easily recover from setup issues.

## Solution

Interactive setup wizard (`augury-node-tui setup`) that guides users through one-time configuration with:
- Auto-detection and path configuration
- Safe auto-fixes (Nix config)
- Interactive guidance for sudo actions (group membership, binary install)
- Live progress for Nix environment build with explanations
- Resumable if interrupted
- Persistent config file

## Architecture

### Command Structure

**Normal mode:**
```bash
augury-node-tui          # Reads from config, launches main TUI
```

**Setup mode:**
```bash
augury-node-tui setup              # Interactive wizard (runnable from anywhere)
augury-node-tui setup --reconfigure # Re-run setup, update config
```

### Config File

**Location:** `~/.config/augury-node-tui/config.toml`

**Schema:**
```toml
augury_node_root = "/home/user/Repos/augury-node"
binary_installed = true
nix_verified = true
setup_completed_at = "2026-03-08T19:45:00Z"

[setup_state]
completed_steps = ["root", "nix", "groups", "install", "nixbuild"]
skipped_steps = []
```

### Package Structure

```
internal/setup/
  wizard.go          - Main wizard orchestration (step sequencing)
  step_root.go       - Root detection/selection step
  step_nix.go        - Nix config check/fix step
  step_groups.go     - Group membership guide step
  step_install.go    - Binary installation step
  step_nixbuild.go   - Nix environment build step (with progress)
  step_success.go    - Success/completion step
  config.go          - Config file read/write/validation
  health.go          - System health check functions
  
cmd/augury-node-tui/
  main.go            - Route to setup vs. normal mode based on args and config
```

## Wizard Flow

### Step 1: Augury-Node Root Detection

**What it does:**
1. Try to auto-detect: Search ancestors from CWD for `scripts/devices`, `scripts/lib`, `pkg`
2. If found: Show path, ask user to confirm
3. If not found: Prompt user to enter path manually (with validation)
4. Validate entered path
5. Save to config

**UI:**
```
╭─────────────────────────────────────╮
│  📁 Augury-Node Repository          │
│                                     │
│  Status: ✓ Found at:                │
│  /home/user/Repos/augury-node       │
│                                     │
│  This will be saved to config.      │
│                                     │
│  [Continue] [Change Path] [Cancel]  │
╰─────────────────────────────────────╯
```

**Actions:**
- `enter` on Continue → next step
- `enter` on Change Path → text input for manual entry
- `q` or Cancel → exit setup

### Step 2: Nix Configuration Check

**What it checks:**
1. Is `nix` command available?
2. Are experimental features enabled?
3. Can it access daemon socket?

**Auto-fix (no sudo):**
- Create `~/.config/nix/nix.conf`
- Add `experimental-features = nix-command flakes`

**UI (before fix):**
```
╭──────────────────────────────────────╮
│  ⚙️  Nix Configuration               │
│                                      │
│  ✓ Nix installed                     │
│  ✗ Experimental features disabled    │
│                                      │
│  Auto-fix available:                 │
│  Add to ~/.config/nix/nix.conf       │
│                                      │
│  [Auto-Fix] [Skip] [Cancel]          │
╰──────────────────────────────────────╯
```

**UI (after fix):**
```
╭──────────────────────────────────────╮
│  ⚙️  Nix Configuration               │
│                                      │
│  ✓ Nix installed                     │
│  ✓ Experimental features enabled     │
│                                      │
│  Configuration updated successfully. │
│                                      │
│  [Continue]                          │
╰──────────────────────────────────────╯
```

### Step 3: Group Membership (Interactive Guide)

**What it checks:**
- Is user in `nix-users` group?

**Guided action (requires sudo):**
- Display exact commands to run
- Provide copy-to-clipboard option
- Allow retry after user runs commands

**UI:**
```
╭─────────────────────────────────────────╮
│  👥 Permissions Setup                   │
│                                         │
│  ✗ Not in nix-users group               │
│                                         │
│  Run this command in a new terminal:    │
│                                         │
│  ┌─────────────────────────────────┐   │
│  │ sudo usermod -aG nix-users $USER │   │
│  │ newgrp nix-users                 │   │
│  └─────────────────────────────────┘   │
│                                         │
│  c copy command | r re-check | s skip  │
╰─────────────────────────────────────────╯
```

**Re-check:**
- Runs `groups | grep nix-users`
- Updates status in real-time
- Continues when verified or skipped

### Step 4: Binary Installation

**What it does:**
1. Check if `/usr/local/bin/augury-node-tui` already exists
2. If not, offer to install (symlink from current binary location)
3. Requires sudo for `/usr/local/bin` write access

**UI (guide mode):**
```
╭─────────────────────────────────────╮
│  📦 Binary Installation              │
│                                     │
│  Install to: /usr/local/bin/        │
│                                     │
│  This requires sudo permission.     │
│                                     │
│  Run in a new terminal:             │
│  ┌────────────────────────────┐    │
│  │ sudo ln -sf <src> <dst>    │    │
│  └────────────────────────────┘    │
│                                     │
│  c copy | r re-check | s skip       │
╰─────────────────────────────────────╯
```

**Skip handling:**
- User can skip if they prefer different install location
- Config notes binary not installed
- Success screen shows PATH setup instructions

### Step 5: Nix Environment Build (Live Progress)

**What it does:**
1. Run `nix develop <root>/.#dev-env --command true` to trigger first build
2. Stream output and parse for progress indicators
3. Show estimated size, current package being built
4. Allow skip (user can build manually later)

**UI (in progress):**
```
╭──────────────────────────────────────────╮
│  🔨 Building Nix Development Environment │
│                                          │
│  This prepares tools for building        │
│  augury-node platforms. First run takes  │
│  5-10 minutes as packages download.      │
│                                          │
│  [▓▓▓▓▓▓░░░░] 12/20 packages             │
│                                          │
│  Current: building gcc-wrapper-13.2.0    │
│  Downloaded: 1.2 GB / ~2.5 GB            │
│  Elapsed: 3m 24s                         │
│                                          │
│  s skip (build manually later)           │
╰──────────────────────────────────────────╯
```

**Progress parsing:**
- Parse output: `[built/total, copied, downloaded]`
- Extract current package name from `building <package>` lines
- Show human-readable progress and time estimates

**Skip option:**
- Saves config anyway
- Success screen notes Nix needs manual setup
- Main TUI will show "Nix: ✗ not ready" until user builds

### Step 6: Success Screen

**UI:**
```
╭────────────────────────────────────╮
│  ✅ Setup Complete!                │
│                                    │
│  augury-node-tui is ready to use.  │
│                                    │
│  Configuration saved to:           │
│  ~/.config/augury-node-tui/        │
│                                    │
│  Quick start:                      │
│  $ augury-node-tui                 │
│                                    │
│  [Launch Now] [Exit]               │
╰────────────────────────────────────╯
```

**Launch Now:**
- Exits wizard
- Immediately launches normal TUI using saved config
- Seamless transition

Does this complete wizard flow look good?