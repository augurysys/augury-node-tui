package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCapabilities_EachActionMapsToRequiredScriptPath(t *testing.T) {
	root := t.TempDir()
	cases := []struct {
		name   string
		req    ActionRequest
		expect string
	}{
		{"build-unit build moxa-uc3100", ActionRequest{Kind: KindBuildUnit, Target: TargetBuild, PlatformID: "moxa-uc3100"}, "scripts/devices/moxa-uc3100-build.sh"},
		{"build-unit build mp255-ulrpm", ActionRequest{Kind: KindBuildUnit, Target: TargetBuild, PlatformID: "mp255-ulrpm"}, "scripts/devices/mp255-ulrpm.sh"},
		{"build-unit pull", BuildUnitPull, "scripts/dev/pull-artifacts.sh"},
		{"build-unit delete", BuildUnitDelete, "scripts/dev/delete-build-unit-cache.sh"},
		{"platform-cache pull", PlatformCachePull, "scripts/dev/pull-artifacts.sh"},
		{"platform-cache push", PlatformCachePush, "scripts/dev/push-artifacts.sh"},
		{"platform-cache clean", PlatformCacheClean, "scripts/dev/clean-platform-cache.sh"},
		{"hydration dry-run", HydrationDryRun, "scripts/hydrate"},
		{"hydration run", HydrationRun, "scripts/hydrate"},
		{"validations all", ValidationsAll, "scripts/validate-all.sh"},
		{"validations shellcheck", ValidationsShellcheck, "scripts/validate-shellcheck.sh"},
		{"validations bats", ValidationsBats, "scripts/validate-bats.sh"},
		{"validations parse", ValidationsParse, "scripts/validate-parse-test.sh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cap := ResolveCapability(root, c.req)
			got := filepath.ToSlash(cap.ScriptPath)
			if !strings.HasSuffix(got, c.expect) {
				t.Errorf("ScriptPath = %q, want suffix %q", got, c.expect)
			}
		})
	}
}

func TestCapabilities_AvailableOnlyWhenScriptExists(t *testing.T) {
	root := t.TempDir()
	// No scripts exist - all should be unavailable
	cap := ResolveCapability(root, ValidationsAll)
	if cap.Available {
		t.Errorf("ValidationsAll: want Available=false when script missing, got true")
	}

	// Create script - should become available
	scriptPath := filepath.Join(root, "scripts", "validate-all.sh")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatalf("setup: %v", err)
	}
	cap = ResolveCapability(root, ValidationsAll)
	if !cap.Available {
		t.Errorf("ValidationsAll: want Available=true when script exists, got false (Reason=%q)", cap.Reason)
	}
}

func TestCapabilities_NotAvailableIncludesMissingScriptReason(t *testing.T) {
	root := t.TempDir()
	cap := ResolveCapability(root, HydrationRun)
	if cap.Available {
		t.Errorf("want Available=false when script missing")
	}
	if cap.Reason == "" {
		t.Errorf("want non-empty Reason when not available")
	}
	if !strings.Contains(strings.ToLower(cap.Reason), "script") && !strings.Contains(strings.ToLower(cap.Reason), "missing") && !strings.Contains(strings.ToLower(cap.Reason), "not found") {
		t.Errorf("Reason should mention missing script, got %q", cap.Reason)
	}
}

func TestCapabilities_BuildUnitBuildUnknownPlatformUnavailable(t *testing.T) {
	root := t.TempDir()
	req := ActionRequest{Kind: KindBuildUnit, Target: TargetBuild, PlatformID: "unknown-platform"}
	cap := ResolveCapability(root, req)
	if cap.Available {
		t.Errorf("want Available=false for unknown platform")
	}
	if cap.Reason == "" {
		t.Errorf("want non-empty Reason for unknown platform")
	}
	if !strings.Contains(strings.ToLower(cap.Reason), "unknown") && !strings.Contains(cap.Reason, "unknown-platform") {
		t.Errorf("Reason should mention unknown platform, got %q", cap.Reason)
	}
}

func TestCapabilities_EmptyRootReturnsUnavailable(t *testing.T) {
	cap := ResolveCapability("", ValidationsAll)
	if cap.Available {
		t.Errorf("want Available=false for empty root")
	}
	if cap.Reason == "" {
		t.Errorf("want non-empty Reason for empty root")
	}
	if !strings.Contains(strings.ToLower(cap.Reason), "root") && !strings.Contains(strings.ToLower(cap.Reason), "empty") {
		t.Errorf("Reason should mention empty root, got %q", cap.Reason)
	}
}

func TestCapabilities_NoDirectFallbackCommandPathExists(t *testing.T) {
	// ResolveCapability must only return script paths, never raw command paths (e.g. /usr/bin/nix).
	root := t.TempDir()
	actions := []ActionRequest{
		BuildUnitPull, PlatformCachePush, HydrationRun, ValidationsAll,
	}
	for _, req := range actions {
		cap := ResolveCapability(root, req)
		if cap.ScriptPath != "" {
			// ScriptPath must be under root, not an absolute system path
			if !strings.HasPrefix(cap.ScriptPath, root) && filepath.IsAbs(cap.ScriptPath) {
				t.Errorf("%s: ScriptPath must not be absolute system path, got %q", req.ID(), cap.ScriptPath)
			}
			// Must look like a script path (contains scripts/)
			if !strings.Contains(cap.ScriptPath, "scripts") {
				t.Errorf("%s: ScriptPath must be script path, got %q", req.ID(), cap.ScriptPath)
			}
		}
	}
}

func mkdirAndTouch(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return f.Close()
}
