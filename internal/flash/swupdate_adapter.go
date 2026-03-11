package flash

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveSWUFile resolves basePath to a single .swu file path.
// - If basePath is a file ending in .swu, returns it as-is.
// - If basePath is a directory, finds *.swu files and returns the latest by ModTime.
// - Otherwise returns an error.
func ResolveSWUFile(basePath string) (string, error) {
	info, err := os.Stat(basePath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	if info.Mode().IsRegular() {
		if strings.HasSuffix(strings.ToLower(basePath), ".swu") {
			return basePath, nil
		}
		return "", fmt.Errorf("invalid path: not a .swu file")
	}

	if !info.IsDir() {
		return "", fmt.Errorf("invalid path: not a file or directory")
	}

	pattern := filepath.Join(basePath, "*.swu")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("no .swu files found in directory")
	}
	if len(matches) == 1 {
		return matches[0], nil
	}

	// Multiple files: return latest by ModTime
	var latest string
	var latestMod int64
	for _, p := range matches {
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		mtime := fi.ModTime().Unix()
		if mtime > latestMod {
			latestMod = mtime
			latest = p
		}
	}
	if latest == "" {
		// Fallback if all Stat failed (e.g. permission)
		return matches[0], nil
	}
	return latest, nil
}

// SWUpdateAdapter wraps augury_update for SWUpdate-based platforms
type SWUpdateAdapter struct {
	root      string
	platform  string
	imagePath string
}

// NewSWUpdateAdapter creates a new SWUpdate adapter. It resolves basePath to a
// single .swu file (file or directory containing .swu) and stores the resolved path.
func NewSWUpdateAdapter(root, platform, basePath string) (*SWUpdateAdapter, error) {
	resolved, err := ResolveSWUFile(basePath)
	if err != nil {
		return nil, err
	}
	return &SWUpdateAdapter{
		root:      root,
		platform:  platform,
		imagePath: resolved,
	}, nil
}

// PlatformType returns PlatformTypeSWUpdate
func (a *SWUpdateAdapter) PlatformType() string {
	return PlatformTypeSWUpdate
}

// SupportsMethodSelection returns false (no method choice for SWUpdate)
func (a *SWUpdateAdapter) SupportsMethodSelection() bool {
	return false
}

// GetMethods returns empty slice (no methods to choose)
func (a *SWUpdateAdapter) GetMethods() []FlashMethod {
	return nil
}

// GetSteps returns the three-step SWUpdate flow
func (a *SWUpdateAdapter) GetSteps(method string) []FlashStep {
	return []FlashStep{
		{
			ID:          "verify",
			Description: "Verify firmware image",
			PromptType:  PromptConfirm,
		},
		{
			ID:          "flash",
			Description: "Flash firmware to device",
			PromptType:  PromptConfirm,
		},
		{
			ID:          "reboot",
			Description: "Reboot device to apply firmware",
			PromptType:  PromptConfirm,
		},
	}
}

// ExecuteStep runs one step of the flash process
func (a *SWUpdateAdapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	augury_update := filepath.Join(a.root, "common/otsn/augury_update")

	var cmd *exec.Cmd
	switch step.ID {
	case "verify":
		cmd = exec.CommandContext(ctx, augury_update, "verify", a.imagePath)
	case "flash":
		cmd = exec.CommandContext(ctx, augury_update, "apply_update", a.imagePath)
	case "reboot":
		return "Reboot required. Please power cycle the device.", nil
	default:
		return "", fmt.Errorf("unknown step: %s", step.ID)
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// CanFlash validates prerequisites. The imagePath parameter is ignored;
// the adapter always validates a.imagePath (the resolved .swu file).
func (a *SWUpdateAdapter) CanFlash(imagePath string) error {
	// Check resolved .swu file exists and is a regular file (not directory)
	info, err := os.Stat(a.imagePath)
	if err != nil {
		return fmt.Errorf("image file not found: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("image path is not a regular file: %s", a.imagePath)
	}

	// Check augury_update exists
	augury_update := filepath.Join(a.root, "common/otsn/augury_update")
	if _, err := os.Stat(augury_update); err != nil {
		return fmt.Errorf("augury_update not found: %w", err)
	}

	return nil
}
