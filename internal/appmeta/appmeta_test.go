package appmeta

import "testing"

func TestAppName(t *testing.T) {
	if AppName() != "augury-node-tui" {
		t.Fatalf("unexpected app name: %q", AppName())
	}
}
