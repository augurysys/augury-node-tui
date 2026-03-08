package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func injectFakeNix(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	fakeNix := filepath.Join(tmp, "nix")
	if err := os.WriteFile(fakeNix, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", tmp+string(filepath.ListSeparator)+os.Getenv("PATH"))
}

func TestJobs_JobLifecycleStates(t *testing.T) {
	injectFakeNix(t)
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	// success: script exists and exits 0
	scriptPath := filepath.Join(absRoot, "scripts", "validate-all.sh")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	job := ExecuteAction(ctx, absRoot, ValidationsAll)
	if job.State != JobStateSuccess {
		t.Errorf("ExecuteAction success: State = %q, want %q", job.State, JobStateSuccess)
	}

	// failed: script exits non-zero
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}
	job = ExecuteAction(ctx, absRoot, ValidationsAll)
	if job.State != JobStateFailed {
		t.Errorf("ExecuteAction failed: State = %q, want %q", job.State, JobStateFailed)
	}

	// cancelled: ctx cancelled before/during run
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()
	job = ExecuteAction(ctxCancel, absRoot, ValidationsAll)
	if job.State != JobStateCancelled {
		t.Errorf("ExecuteAction cancelled: State = %q, want %q", job.State, JobStateCancelled)
	}

	// blocked: capability not available (no script)
	rootEmpty := t.TempDir()
	job = ExecuteAction(ctx, rootEmpty, ValidationsAll)
	if job.State != JobStateBlocked {
		t.Errorf("ExecuteAction blocked by capability: State = %q, want %q", job.State, JobStateBlocked)
	}
	if job.Reason == "" {
		t.Errorf("ExecuteAction blocked: want non-empty Reason")
	}
}

func TestJobs_JobLogPathPropagation(t *testing.T) {
	injectFakeNix(t)
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(absRoot, "scripts", "validate-all.sh")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho logged-output\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	job := ExecuteAction(ctx, absRoot, ValidationsAll)
	if job.State != JobStateSuccess {
		t.Fatalf("ExecuteAction: State = %q, want success", job.State)
	}
	if job.LogPath == "" {
		t.Fatal("ExecuteAction: want non-empty LogPath")
	}
	expectedSuffix := filepath.Join("tmp", "augury-node-tui")
	if !strings.Contains(job.LogPath, expectedSuffix) {
		t.Errorf("LogPath = %q, want to contain %q", job.LogPath, expectedSuffix)
	}
	if !strings.HasSuffix(job.LogPath, ".log") {
		t.Errorf("LogPath = %q, want to end with .log", job.LogPath)
	}
	data, err := os.ReadFile(job.LogPath)
	if err != nil {
		t.Fatalf("log file not readable: %v", err)
	}
	if !strings.Contains(string(data), "logged-output") {
		t.Errorf("log file content %q does not contain logged-output", string(data))
	}
}

func TestJobs_BlockedWhenCapabilityFails(t *testing.T) {
	root := t.TempDir()
	// No scripts - capability unavailable
	ctx := context.Background()
	job := ExecuteAction(ctx, root, ValidationsAll)
	if job.State != JobStateBlocked {
		t.Errorf("ExecuteAction: State = %q, want blocked when capability fails", job.State)
	}
	if job.Reason == "" {
		t.Errorf("ExecuteAction blocked: want non-empty Reason")
	}
	if job.LogPath != "" {
		t.Errorf("ExecuteAction blocked: want empty LogPath, got %q", job.LogPath)
	}
}

func TestJobs_BlockedWhenNixFails(t *testing.T) {
	// Create root with valid script so capability passes, but nix will fail
	// (nix develop in random temp dir will not find flake)
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(absRoot, "scripts", "validate-all.sh")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	job := ExecuteAction(ctx, absRoot, ValidationsAll)
	// Nix probe will fail (no flake.nix in temp dir), so action should be blocked
	if job.State != JobStateBlocked {
		t.Errorf("ExecuteAction: State = %q, want blocked when nix fails", job.State)
	}
	if job.Reason == "" {
		t.Errorf("ExecuteAction blocked by nix: want non-empty Reason")
	}
}

func TestJobs_HydrationPassesPlatformArg(t *testing.T) {
	injectFakeNix(t)
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(absRoot, "scripts", "hydrate")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte(`#!/bin/sh
[ "$1" = "--platform" ] && [ "$2" = "node2" ] || exit 1
exit 0
`), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	req := ActionRequest{Kind: KindHydration, Target: TargetRun, PlatformID: "node2"}
	job := ExecuteAction(ctx, absRoot, req)
	if job.State != JobStateSuccess {
		t.Errorf("ExecuteAction hydration with PlatformID: State = %q, want success (script must receive --platform)", job.State)
	}
}

func TestJobs_HydrationLogPathIncludesPlatformID(t *testing.T) {
	injectFakeNix(t)
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(absRoot, "scripts", "hydrate")
	if err := mkdirAndTouch(scriptPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	req := ActionRequest{Kind: KindHydration, Target: TargetRun, PlatformID: "node2"}
	job := ExecuteAction(ctx, absRoot, req)
	if job.State != JobStateSuccess {
		t.Fatalf("ExecuteAction: State = %q, want success", job.State)
	}
	if !strings.Contains(job.LogPath, "node2") {
		t.Errorf("LogPath = %q, want to contain platform ID node2 for per-platform log disambiguation", job.LogPath)
	}
	if !strings.Contains(job.LogPath, "hydration-run") {
		t.Errorf("LogPath = %q, want to contain action name", job.LogPath)
	}
}
