package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNixGate_ProbeSucceeds(t *testing.T) {
	tmp := t.TempDir()
	fakeNix := filepath.Join(tmp, "nix")
	if err := os.WriteFile(fakeNix, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmp+string(filepath.ListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	root := t.TempDir()
	state := ProbeNix(root)

	if !state.Ready {
		t.Errorf("ProbeNix: Ready = false, want true when probe succeeds")
	}
	if state.Reason != "" {
		t.Errorf("ProbeNix: Reason = %q, want empty when probe succeeds", state.Reason)
	}
}

func TestNixGate_ProbeFails(t *testing.T) {
	root := t.TempDir()
	state := ProbeNix(root)

	if state.Ready {
		t.Errorf("ProbeNix: Ready = true, want false when probe fails (no flake)")
	}
	if state.Reason == "" {
		t.Errorf("ProbeNix: Reason empty, want non-empty when probe fails")
	}
}

func TestNixGate_NoBypass_ExecutableActionsBlockedWhenNixNotReady(t *testing.T) {
	nix := NixState{Ready: false, Reason: "nix not available"}

	executableActions := []ActionRequest{
		BuildUnitBuild, BuildUnitPull, BuildUnitDelete,
		PlatformCachePull, PlatformCachePush, PlatformCacheClean,
		HydrationDryRun, HydrationRun,
		ValidationsAll, ValidationsShellcheck, ValidationsBats, ValidationsParse,
	}

	for _, req := range executableActions {
		blocked, reason := IsActionBlockedByNix(req, nix)
		if !blocked {
			t.Errorf("IsActionBlockedByNix(%q, nix not ready): blocked = false, want true", req.ID())
		}
		if reason != nix.Reason {
			t.Errorf("IsActionBlockedByNix(%q, nix not ready): reason = %q, want %q", req.ID(), reason, nix.Reason)
		}
	}
}

func TestNixGate_ExecutableActionsAllowedWhenNixReady(t *testing.T) {
	nix := NixState{Ready: true, Reason: ""}

	executableActions := []ActionRequest{
		BuildUnitBuild, HydrationRun, ValidationsAll,
	}

	for _, req := range executableActions {
		blocked, reason := IsActionBlockedByNix(req, nix)
		if blocked {
			t.Errorf("IsActionBlockedByNix(%q, nix ready): blocked = true, want false", req.ID())
		}
		if reason != "" {
			t.Errorf("IsActionBlockedByNix(%q, nix ready): reason = %q, want empty", req.ID(), reason)
		}
	}
}
