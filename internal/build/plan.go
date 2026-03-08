package build

import (
	"os"
	"path/filepath"

	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/run"
)

type PlanEntry struct {
	PlatformID          string
	ScriptPath          string
	OutputPath          string
	LocalArtifactPresent *bool
}

type Plan struct {
	Entries      []PlanEntry
	Mode         run.Mode
	ForceRebuild map[string]bool
}

func BuildPlan(root string, platforms []platform.Platform, mode run.Mode, forceRebuild map[string]bool) *Plan {
	p := &Plan{
		Mode:         mode,
		ForceRebuild: make(map[string]bool),
	}
	if forceRebuild != nil {
		for k, v := range forceRebuild {
			p.ForceRebuild[k] = v
		}
	}
	if mode == run.ModeValidationOnly {
		return p
	}
	for _, pl := range platforms {
		p.ForceRebuild[pl.ID] = p.ForceRebuild[pl.ID]
		scriptPath := filepath.Join(root, pl.ScriptRelPath)
		outputPath := filepath.Join(root, pl.OutputRelPath)
		present := artifactPresent(outputPath)
		p.Entries = append(p.Entries, PlanEntry{
			PlatformID:          pl.ID,
			ScriptPath:          scriptPath,
			OutputPath:          outputPath,
			LocalArtifactPresent: present,
		})
	}
	return p
}

func artifactPresent(outputPath string) *bool {
	_, err := os.Stat(outputPath)
	b := err == nil
	return &b
}
