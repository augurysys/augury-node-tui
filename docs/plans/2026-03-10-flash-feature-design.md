# Flash Feature Design

## Overview

Add firmware flashing capability to augury-node-tui for development workflow. Support MP255 (via `deploy.sh` wrapper) and SWUpdate platforms (via `augury_update` wrapper) with unified UI.

**Primary Use Case:** Development workflow - flash freshly built images to connected dev boards for testing.

**Future Use Case:** Manufacturing line (phase 2) - auto mode, batch operations, operator-friendly.

## Architecture

### Core Components

**1. Flash Screen (`internal/flash/model.go`)**
- Main Bubbletea model using ScreenLayout
- State machine: `stateIdle → statePlatformSelect → stateMethodSelect (MP255 only) → stateFlashing → stateComplete/stateError`
- Delegates platform-specific logic to adapters
- Handles UI rendering, key bindings, window sizing

**2. FlashAdapter Interface (`internal/flash/adapter.go`)**
```go
type FlashAdapter interface {
    // Platform identification
    PlatformType() string  // "mp255", "swupdate"
    
    // Method selection (MP255 only)
    SupportsMethodSelection() bool
    GetMethods() []FlashMethod  // e.g., ["UUU", "Manual/Rescue"]
    
    // Interactive steps
    GetSteps(method string) []FlashStep
    ExecuteStep(ctx context.Context, step FlashStep) (output string, err error)
    
    // Validation
    CanFlash(imagePath string) error  // Pre-flight checks
}

type FlashStep struct {
    ID          string
    Description string
    PromptType  PromptType  // Confirm, Choice, Manual
    Choices     []string    // For choice prompts
}

type PromptType int
const (
    PromptConfirm PromptType = iota  // Press Enter
    PromptChoice                     // Select 1/2/3
    PromptManual                     // Follow instructions, then confirm
)
```

**3. Concrete Adapters**
- `MP255Adapter` (`internal/flash/mp255_adapter.go`): Wraps `deploy.sh`, parses interactive prompts
- `SWUpdateAdapter` (`internal/flash/swupdate_adapter.go`): Wraps `augury_update`, simple verify → flash → reboot flow

**4. App Router Integration**
- Add `f` key handler in `internal/home/model.go`
- Route to Flash screen with platform context
- Flash screen in `internal/app/model.go` alongside Build, Hydration, etc.

### Data Flow

1. User presses `f` on Home → Flash screen with platform list
2. User selects platform → Detect type, instantiate adapter
3. Adapter provides steps → UI renders them inline
4. User confirms/selects → Adapter executes, streams output
5. Complete → Show summary, offer reboot/exit

## UI Layout & User Flow

### Flash Screen Layout (ScreenLayout pattern)

**Top Bar:** `🚀 Home > Flash  •  [platform name]`

**Content Area (adapts by state):**

**State 1: Platform Selection**
```
┌─ Platform Selection ─────────────────────┐
│  Select platform to flash:               │
│                                           │
│  > mp255-ulrpm        [Yocto/UUU]        │
│    cassia-x2000       [SWUpdate]         │
│    moxa-uc3100        [SWUpdate]         │
│                                           │
│  Image: /path/to/image.swu               │
│  Status: Ready to flash                  │
└───────────────────────────────────────────┘
```

**State 2: Method Selection (MP255 only)**
```
┌─ Choose Flash Method ────────────────────┐
│  Platform: mp255-ulrpm                   │
│                                           │
│  Select method:                           │
│  > 1) Official UUU (automated)           │
│    2) Manual/Rescue (step-by-step)       │
│                                           │
│  UUU method uses install_linux_fw_uuu.sh │
│  and requires serial connection.         │
└───────────────────────────────────────────┘
```

**State 3: Flashing (Interactive Steps)**
```
┌─ Flashing: Step 2/5 ─────────────────────┐
│  Connect USB-C cable to board             │
│                                           │
│  ⚠ Ensure:                                │
│   • USB-C connected to PC                │
│   • Power off board                      │
│   • Serial cable connected                │
│                                           │
│  Press Enter when ready...               │
│                                           │
│  [Script output streaming here]          │
└───────────────────────────────────────────┘
```

**Bottom Help:**
- Platform selection: `enter select  •  esc back  •  q quit`
- Method selection: `1/2 choose  •  esc back`
- Flashing: `enter continue  •  c cancel  •  esc abort`

### User Flow

**Dev Workflow (Phase 1):**
1. Build image → Home screen shows "built" state
2. Press `f` → Flash screen, platform list
3. Select platform → Auto-detect type (MP255 vs SWUpdate)
4. If MP255: Choose method (UUU/Manual)
5. Step through prompts → Stream output inline
6. Complete → Show summary, return to Home

**Manufacturing Additions (Phase 2 - Future):**
- Auto mode: Skip all prompts, log to file
- Batch mode: Flash multiple devices sequentially
- QR/barcode scanning for serial tracking
- Pass/fail LED indicators
- Operator-friendly error messages

## Error Handling

### Pre-flight Validation (before flashing)
- Image file exists and is readable
- Platform type matches image extension (`.swu` for SWUpdate, release dir for MP255)
- Required tools present (`deploy.sh`, `augury_update`, `expect`, `minicom` for MP255)
- USB/serial device detection (MP255 only)
- Disk space check

### During Flash
- Command timeout detection (script hangs)
- Process failure detection (non-zero exit code)
- Serial connection loss (MP255)
- User cancellation (`c` key or `Esc`)
- Streaming output parsing for error patterns

### Error Display
```
┌─ Flash Failed ───────────────────────────┐
│  ❌ Error: Serial connection lost        │
│                                           │
│  Step: Waiting for U-Boot prompt         │
│  Cause: /dev/ttyACM0 disconnected        │
│                                           │
│  Suggestions:                             │
│   • Check USB cable connection           │
│   • Verify board is powered              │
│   • Try different USB port               │
│                                           │
│  Log: /path/to/flash.log                 │
└───────────────────────────────────────────┘

r retry  •  l view log  •  esc back
```

## Testing Strategy

### Unit Tests
- Adapter interface mocking
- State machine transitions
- Platform detection logic
- Prompt parsing (MP255)
- Error categorization

### Integration Tests
- Mock `deploy.sh` with simulated prompts/output
- Mock `augury_update` with verify/flash/reboot sequence
- Test cancellation at each step
- Test error recovery

### Manual Testing Scenarios
- Happy path: MP255 UUU method start to finish
- Happy path: SWUpdate platform verify → flash
- Error: Missing image file
- Error: Wrong platform selected for image
- Error: USB disconnected mid-flash
- Cancellation: User presses `c` during flash
- UI: Window resize during flash, output scrolling

## Implementation Details

### Platform Detection Logic

```go
func DetectPlatformType(platform platform.Platform) string {
    // Check platform ID prefix
    if strings.HasPrefix(platform.ID, "mp255") {
        return "mp255"
    }
    // Check for debian-based platforms
    if strings.Contains(platform.ID, "cassia") || 
       strings.Contains(platform.ID, "moxa") {
        return "swupdate"
    }
    return "unknown"
}
```

### File Structure

```
internal/flash/
├── model.go          # Main Bubbletea model, state machine
├── adapter.go        # FlashAdapter interface
├── mp255_adapter.go  # deploy.sh wrapper
├── swupdate_adapter.go  # augury_update wrapper
├── messages.go       # Bubbletea messages
└── model_test.go     # Unit tests

internal/app/model.go  # Add flash routing
internal/home/model.go # Add 'f' key handler
```

### Key Design Decisions

1. **Interactive mode only** - Defer auto mode until manufacturing requirements are clear
2. **No image picker** - Use built platform's output (matches Build screen UX)
3. **Log to file** - Same pattern as Build logs (`tmp/augury-node-tui/`)
4. **Streaming output** - Parse stdout/stderr line-by-line, display in content area
5. **State-driven UI** - Single View() method, switch on state for different layouts
6. **Cancel = kill process** - Send SIGTERM to running script, clean up

## Scope

### Phase 1 - Dev Workflow (In Scope)

- Flash screen accessible via `f` key from Home
- Platform selection from built platforms
- MP255 adapter: Interactive mode only, wraps `deploy.sh`
  - Method selection (UUU vs Manual/Rescue)
  - Step-by-step prompts with Enter to continue
  - Stream output inline
  - Basic error display
- SWUpdate adapter: Simple verify → flash → reboot flow
  - Wraps `augury_update` commands
  - Progress display
- Unified ScreenLayout UI for both platform types
- Pre-flight validation (file exists, tools present)
- Log file capture (`tmp/augury-node-tui/flash-<platform>.log`)
- Cancel during flash (`c` key)

### Phase 2 - Manufacturing (Out of Scope / Future)

- Auto mode (manufacturing requirement)
- Batch/multi-device flashing
- QR/barcode scanning
- Serial number tracking
- Pass/fail LED indicators
- Advanced error recovery (auto-retry, fallback methods)
- Flash verification after complete
- Disk usage monitoring during flash
- Network-based image fetching

## Success Criteria

**MVP is complete when:**
- Can flash MP255 using UUU method from TUI
- Can flash Cassia/Moxa using SWUpdate from TUI
- Errors display helpful messages with retry option
- Output streams to UI in real-time
- Cancel works and doesn't leave zombie processes
- Log files captured for debugging

## Rationale

**Why adapter pattern:** Clean separation makes adding platforms trivial. State machine keeps UI logic simple.

**Why dev-first:** Defers manufacturing complexity until requirements solidify. Allows iterating on core functionality first.

**Why unified UI:** Consistent user experience across all platforms. Single code path to maintain.

**Why interactive mode only:** Simpler initial implementation. Auto mode requires serial automation (expect/minicom) and more error handling complexity.
