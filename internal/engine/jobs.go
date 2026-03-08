package engine

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/run"
)

// JobState represents the lifecycle state of a job.
type JobState string

const (
	JobStateQueued    JobState = "queued"
	JobStateRunning   JobState = "running"
	JobStateSuccess   JobState = "success"
	JobStateFailed    JobState = "failed"
	JobStateCancelled JobState = "cancelled"
	JobStateBlocked   JobState = "blocked"
)

// Job holds the result of an action execution attempt.
type Job struct {
	State   JobState
	LogPath string
	Reason  string
}

// ExecuteAction runs the action when not blocked by capability or nix.
// Returns a Job with State blocked when capability or nix fails.
// Uses internal/run.Execute for command execution when not blocked.
func ExecuteAction(ctx context.Context, root string, req ActionRequest) Job {
	cap := ResolveCapability(root, req)
	if !cap.Available {
		return Job{
			State:  JobStateBlocked,
			Reason: cap.Reason,
		}
	}

	nix := ProbeNix(root)
	if blocked, reason := IsActionBlockedByNix(req, nix); blocked {
		return Job{
			State:  JobStateBlocked,
			Reason: reason,
		}
	}

	args := buildScriptArgs(cap.ScriptPath, req)
	spec := run.RunSpec{
		Name:    jobNameFromRequest(req),
		Root:    root,
		Mode:    run.ModeSmart,
		Command: "sh",
		Args:    args,
	}
	result := run.Execute(ctx, spec)

	logPath := filepath.Join(root, "tmp", "augury-node-tui", spec.Name+".log")

	switch result.Status {
	case "success":
		return Job{State: JobStateSuccess, LogPath: logPath}
	case "cancelled":
		return Job{State: JobStateCancelled, LogPath: logPath, Reason: "cancelled"}
	case "error":
		reason := result.Stderr
		if reason == "" {
			reason = result.Stdout
		}
		if reason == "" {
			reason = "command failed"
		}
		return Job{State: JobStateFailed, LogPath: logPath, Reason: strings.TrimSpace(reason)}
	default:
		return Job{State: JobStateFailed, LogPath: logPath, Reason: result.Status}
	}
}

func jobNameFromRequest(req ActionRequest) string {
	name := strings.ReplaceAll(req.ID(), ":", "-")
	if req.PlatformID != "" {
		name += "-" + req.PlatformID
	}
	return name
}

func buildScriptArgs(scriptPath string, req ActionRequest) []string {
	args := []string{scriptPath}
	if req.Kind == KindHydration && req.PlatformID != "" {
		args = append(args, "--platform", req.PlatformID)
	}
	return args
}
