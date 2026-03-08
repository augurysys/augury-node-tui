package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/build"
	"github.com/augurysys/augury-node-tui/internal/home"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/workspace"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	fixture := filepath.Join(dir, "..", "testdata", "fake-augury-node")
	abs, err := filepath.Abs(fixture)
	if err != nil {
		t.Fatalf("abs fixture: %v", err)
	}
	return abs
}

func setupFixtureRoot(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := copyDir(fixtureRoot(t), root); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@test")
	runGit(t, root, "config", "user.name", "Test")
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "init")
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, info.Mode())
	})
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestAppIntegration_FixtureRoot(t *testing.T) {
	root := setupFixtureRoot(t)

	if err := workspace.ValidateRoot(root); err != nil {
		t.Fatalf("fixture must be valid augury-node root: %v", err)
	}

	st, err := status.Collect(root)
	if err != nil {
		dirty := make(map[string]bool)
		for _, p := range status.RequiredPaths {
			dirty[p] = false
		}
		st = status.RepoStatus{Root: root, Branch: "?", SHA: "?", Dirty: dirty}
	}

	platforms := platform.Registry()
	hm := home.NewModel(st, platforms)
	bm := build.NewModel(st, platforms, hm.Selected)

	homeView := hm.View()
	if homeView == "" {
		t.Error("home status must load")
	}
	if !strings.Contains(homeView, root) {
		t.Errorf("home status must contain root; got %q", homeView)
	}

	for _, p := range platforms {
		hm.TogglePlatform(p.ID)
	}

	plan := bm.Plan()
	if plan == nil {
		t.Fatal("build plan must not be nil")
	}
	if len(plan.Entries) == 0 {
		t.Error("build plan must render entries for selected platforms")
	}

	specs := bm.RunSpecs()
	if len(specs) == 0 {
		t.Fatal("RunSpecs must return specs for selected platforms")
	}

	ctx := context.Background()
	summary := build.ExecuteSequential(ctx, specs)

	logDir := filepath.Join(root, "tmp", "augury-node-tui")
	for _, spec := range specs {
		logPath := filepath.Join(logDir, spec.Name+".log")
		if _, err := os.Stat(logPath); err != nil {
			t.Errorf("build execution must produce logs in tmp/augury-node-tui; missing %s: %v", logPath, err)
		}
	}

	for _, row := range summary.Rows {
		if row.Status != build.RowStatusSuccess {
			t.Errorf("summary must contain expected fixture outcomes; row %s = %q, want success", row.PlatformID, row.Status)
		}
	}
}