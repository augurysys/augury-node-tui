package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
)

func TestPlanBuilder_FromSelectedPlatformsAndRoot(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	platforms := []platform.Platform{platform.Registry()[0]}
	plan := BuildPlan(absRoot, platforms, run.ModeSmart, nil)

	if len(plan.Entries) == 0 {
		t.Fatal("plan must have at least one entry for selected platform")
	}
	e := plan.Entries[0]
	if e.ScriptPath == "" {
		t.Error("pre-flight plan must include script path")
	}
	if e.OutputPath == "" {
		t.Error("pre-flight plan must include output path")
	}
	if e.LocalArtifactPresent == nil {
		t.Error("pre-flight plan must include local artifact presence")
	}
}

func TestPlanBuilder_ValidationOnlySkipsPlatformBuildScripts(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	platforms := platform.Registry()
	plan := BuildPlan(absRoot, platforms, run.ModeValidationOnly, nil)
	if len(plan.Entries) != 0 {
		t.Errorf("validation-only must skip platform build scripts; got %d entries", len(plan.Entries))
	}
}

func TestPlanBuilder_ArtifactPresentTreatsFileAsPresent(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "artifact.tar")
	if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	got := artifactPresent(f)
	if got == nil {
		t.Fatal("artifactPresent must return non-nil")
	}
	if !*got {
		t.Error("artifactPresent must treat existing file as present; got false")
	}
}

func TestPlanBuilder_ArtifactPresentTreatsDirAsPresent(t *testing.T) {
	tmp := t.TempDir()
	if err := os.MkdirAll(tmp, 0755); err != nil {
		t.Fatal(err)
	}
	got := artifactPresent(tmp)
	if got == nil {
		t.Fatal("artifactPresent must return non-nil")
	}
	if !*got {
		t.Error("artifactPresent must treat existing directory as present; got false")
	}
}

func TestPlanBuilder_PerPlatformForceRebuildToggleExists(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	platforms := []platform.Platform{platform.Registry()[0]}
	plan := BuildPlan(absRoot, platforms, run.ModeSmart, nil)

	if plan.ForceRebuild == nil {
		t.Fatal("plan state must have per-platform force rebuild toggle")
	}
	if _, ok := plan.ForceRebuild[platforms[0].ID]; !ok {
		t.Error("force rebuild map must include entry for selected platform")
	}
}
