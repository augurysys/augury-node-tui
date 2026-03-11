package flash

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MP255Adapter wraps deploy.sh for MP255 platforms
type MP255Adapter struct {
	root         string
	platform     string
	releasePath  string
}

// NewMP255Adapter creates a new MP255 adapter
func NewMP255Adapter(root, platform, releasePath string) *MP255Adapter {
	return &MP255Adapter{
		root:        root,
		platform:    platform,
		releasePath: releasePath,
	}
}

// PlatformType returns PlatformTypeMP255
func (a *MP255Adapter) PlatformType() string {
	return PlatformTypeMP255
}

// SupportsMethodSelection returns true (UUU vs Manual choice)
func (a *MP255Adapter) SupportsMethodSelection() bool {
	return true
}

// GetMethods returns UUU and Manual/Rescue methods
func (a *MP255Adapter) GetMethods() []FlashMethod {
	return []FlashMethod{
		{
			ID:          "uuu",
			Name:        "Official UUU (automated)",
			Description: "Uses install_linux_fw_uuu.sh with serial automation",
		},
		{
			ID:          "manual",
			Name:        "Manual/Rescue (step-by-step)",
			Description: "DFU + Fastboot + UMS with manual control",
		},
	}
}

// GetSteps returns placeholder steps (will be implemented later)
func (a *MP255Adapter) GetSteps(method string) []FlashStep {
	// TODO: Parse deploy.sh for actual steps
	// Note: method parameter (uuu/manual) is ignored in stub; will differ when implemented
	return []FlashStep{
		{
			ID:          "placeholder",
			Description: "Deploy.sh integration coming soon",
			PromptType:  PromptConfirm,
		},
	}
}

// ExecuteStep runs one step (stub for now)
func (a *MP255Adapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	return "MP255 flashing not yet implemented", fmt.Errorf("not implemented")
}

// CanFlash validates prerequisites. The imagePath parameter is ignored;
// the adapter always validates a.releasePath (the path it was constructed with).
func (a *MP255Adapter) CanFlash(imagePath string) error {
	// Check release directory exists (validate what will actually be used)
	if _, err := os.Stat(a.releasePath); err != nil {
		return fmt.Errorf("release directory not found: %w", err)
	}

	// Check deploy.sh exists
	deployScript := filepath.Join(a.root, "yocto/deploy.sh")
	if _, err := os.Stat(deployScript); err != nil {
		return fmt.Errorf("deploy.sh not found: %w", err)
	}

	return nil
}
