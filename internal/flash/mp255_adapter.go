package flash

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// MP255Adapter wraps deploy.sh for MP255 platforms
type MP255Adapter struct {
	root        string
	platform    string
	releasePath string
	method      string // stored when GetSteps is called (uuu/manual)
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

// GetSteps returns a single step for the given method (uuu or manual).
// The method is stored for use by ExecuteStep.
func (a *MP255Adapter) GetSteps(method string) []FlashStep {
	a.method = method
	return []FlashStep{
		{
			ID:          "flash",
			Description: "Flash with deploy.sh",
			PromptType:  PromptConfirm,
		},
	}
}

// methodToDeployArg maps adapter method ID to deploy.sh --method value (1=UUU, 2=Manual)
func (a *MP255Adapter) methodToDeployArg() string {
	switch a.method {
	case "manual":
		return "2"
	case "uuu":
		return "1"
	default:
		return "1"
	}
}

// ExecuteStep runs one step. For the "flash" step, invokes deploy.sh with the stored method.
func (a *MP255Adapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	if step.ID != "flash" {
		return "", fmt.Errorf("unknown step: %s", step.ID)
	}

	deployScript := filepath.Join(a.root, "yocto", "deploy.sh")
	cmd := exec.CommandContext(ctx, "bash", deployScript, "--method", a.methodToDeployArg(), a.releasePath)
	output, err := cmd.CombinedOutput()
	return string(output), err
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
