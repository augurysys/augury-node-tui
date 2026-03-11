package flash

import "context"

// FlashAdapter defines the interface for platform-specific flash implementations
type FlashAdapter interface {
	// PlatformType returns "mp255" or "swupdate"
	PlatformType() string

	// SupportsMethodSelection returns true if platform has multiple flash methods
	SupportsMethodSelection() bool

	// GetMethods returns available flash methods (empty if no selection needed)
	GetMethods() []FlashMethod

	// GetSteps returns the interactive steps for the given method
	GetSteps(method string) []FlashStep

	// ExecuteStep runs one step and returns output/error
	ExecuteStep(ctx context.Context, step FlashStep) (output string, err error)

	// CanFlash validates the image path and platform prerequisites
	CanFlash(imagePath string) error
}
