package flash

// PlatformSelectedMsg is sent when user selects a platform
type PlatformSelectedMsg struct {
	PlatformID string
}

// MethodSelectedMsg is sent when user selects a flash method
type MethodSelectedMsg struct {
	MethodID string
}

// StepCompleteMsg is sent when a flash step completes
type StepCompleteMsg struct {
	StepID string
	Output string
}

// FlashErrorMsg is sent when flash fails
type FlashErrorMsg struct {
	Err error
}

// FlashCompleteMsg is sent when flash succeeds
type FlashCompleteMsg struct{}

// CancelFlashMsg is sent when user cancels
type CancelFlashMsg struct{}
