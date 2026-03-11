# Flash Module

Firmware flashing for MP255 and SWUpdate platforms.

## Architecture

- `model.go`: Main Bubbletea model with state machine
- `messages.go`: Custom Bubbletea messages for flash events
- `types.go`: Common types (PromptType, FlashMethod, FlashStep)
- `adapter.go`: Platform-agnostic flash interface
- `mp255_adapter.go`: Wraps deploy.sh for MP255; invokes `deploy.sh` to perform actual flashing
- `swupdate_adapter.go`: Wraps augury_update for SWUpdate; when OutputRelPath is a directory, automatically resolves the `.swu` file within it
- `platform.go`: Platform type detection

## States

1. `stateIdle` - Initial state
2. `statePlatformSelect` - Choose platform
3. `stateMethodSelect` - Choose flash method (MP255 only)
4. `stateFlashing` - Execute flash steps
5. `stateComplete` - Success
6. `stateError` - Failure

## Adding New Platforms

1. Implement `FlashAdapter` interface
2. Update `DetectPlatformType()` to recognize platform
3. Add adapter creation in `handlePlatformSelected()`

## Testing

```bash
# Unit tests
go test ./internal/flash/...

# Integration test (manual)
./augury-node-tui
# Press 'f' → select platform → follow prompts
```

## Current Status

**Implemented:**
- Platform selection UI
- Method selection UI (MP255)
- Adapter pattern foundation
- SWUpdate adapter (with .swu resolution from directories)
- MP255 adapter (with deploy.sh integration)

**TODO:**
- Step execution with output streaming
- Error handling and retry
- Log file capture
- Cancel support
