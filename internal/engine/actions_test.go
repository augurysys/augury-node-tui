package engine

import (
	"testing"
)

func TestActionContracts(t *testing.T) {
	cases := []struct {
		name   string
		action ActionRequest
		id     string
	}{
		// build-unit
		{"build-unit build", BuildUnitBuild, "build-unit:build"},
		{"build-unit pull", BuildUnitPull, "build-unit:pull"},
		{"build-unit delete", BuildUnitDelete, "build-unit:delete"},
		// platform-cache
		{"platform-cache pull", PlatformCachePull, "platform-cache:pull"},
		{"platform-cache push", PlatformCachePush, "platform-cache:push"},
		{"platform-cache clean", PlatformCacheClean, "platform-cache:clean"},
		// hydration
		{"hydration dry-run", HydrationDryRun, "hydration:dry-run"},
		{"hydration run", HydrationRun, "hydration:run"},
		// validations
		{"validations all", ValidationsAll, "validations:all"},
		{"validations shellcheck", ValidationsShellcheck, "validations:shellcheck"},
		{"validations bats", ValidationsBats, "validations:bats"},
		{"validations parse", ValidationsParse, "validations:parse"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.action.ID()
			if got != c.id {
				t.Errorf("ID() = %q, want %q", got, c.id)
			}
			meta := c.action.Metadata()
			if meta.DisplayName == "" {
				t.Errorf("Metadata().DisplayName must be non-empty")
			}
		})
	}
}
