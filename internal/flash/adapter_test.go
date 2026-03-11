package flash

import "context"

// Compile-time interface implementation check
var _ FlashAdapter = (*mockAdapter)(nil)

type mockAdapter struct {
	platformType string
}

func (m *mockAdapter) PlatformType() string                              { return m.platformType }
func (m *mockAdapter) SupportsMethodSelection() bool                     { return false }
func (m *mockAdapter) GetMethods() []FlashMethod                         { return nil }
func (m *mockAdapter) GetSteps(method string) []FlashStep                { return nil }
func (m *mockAdapter) ExecuteStep(ctx context.Context, step FlashStep) (string, error) {
	return "", nil
}
func (m *mockAdapter) CanFlash(imagePath string) error { return nil }
