package flash

import "fmt"

// PromptType defines the kind of user interaction required
type PromptType int

const (
	PromptConfirm PromptType = iota // Press Enter to continue
	PromptChoice                     // Select 1/2/3...
	PromptManual                     // Follow instructions, then confirm
)

func (p PromptType) String() string {
	switch p {
	case PromptConfirm:
		return "PromptConfirm"
	case PromptChoice:
		return "PromptChoice"
	case PromptManual:
		return "PromptManual"
	default:
		return fmt.Sprintf("PromptType(%d)", p)
	}
}

// FlashMethod represents a flashing method (e.g., UUU, Manual/Rescue)
type FlashMethod struct {
	ID          string
	Name        string
	Description string
}

// FlashStep represents one interactive step in the flash process
type FlashStep struct {
	ID          string
	Description string
	PromptType  PromptType
	Choices     []string // For PromptChoice
}
