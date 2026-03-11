package flash

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SWUpdateAdapter wraps augury_update for SWUpdate-based platforms
type SWUpdateAdapter struct {
	root      string
	platform  string
	imagePath string
}

// NewSWUpdateAdapter creates a new SWUpdate adapter
func NewSWUpdateAdapter(root, platform, imagePath string) *SWUpdateAdapter {
	return &SWUpdateAdapter{
		root:      root,
		platform:  platform,
		imagePath: imagePath,
	}
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
// the adapter always validates a.imagePath (the path it was constructed with).
func (a *SWUpdateAdapter) CanFlash(imagePath string) error {
	// Check image file exists (validate what will actually be used)
	if _, err := os.Stat(a.imagePath); err != nil {
		return fmt.Errorf("image file not found: %w", err)
	}

	// Check augury_update exists
	augury_update := filepath.Join(a.root, "common/otsn/augury_update")
	if _, err := os.Stat(augury_update); err != nil {
		return fmt.Errorf("augury_update not found: %w", err)
	}

	return nil
}
