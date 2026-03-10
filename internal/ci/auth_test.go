package ci

import (
	"os"
	"testing"
)

func TestResolveToken(t *testing.T) {
	t.Run("env var takes priority", func(t *testing.T) {
		os.Setenv("CIRCLE_TOKEN", "env-token")
		defer os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("config-token")
		if got != "env-token" {
			t.Errorf("got %q, want %q", got, "env-token")
		}
	})

	t.Run("falls back to config", func(t *testing.T) {
		os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("config-token")
		if got != "config-token" {
			t.Errorf("got %q, want %q", got, "config-token")
		}
	})

	t.Run("empty when neither set", func(t *testing.T) {
		os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("")
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
