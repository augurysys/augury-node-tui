package engine

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProbeTimeout is the maximum duration for the nix probe command.
const ProbeTimeout = 30 * time.Second

// NixState represents the result of probing Nix readiness.
type NixState struct {
	Ready  bool
	Reason string
}

// ProbeNix runs the nix develop probe command in root and returns readiness.
// Contract: nix develop .#dev-env --command sh -c 'echo ready'
// Empty or invalid root returns blocked state. Probe is bounded by ProbeTimeout.
func ProbeNix(root string) NixState {
	if root == "" {
		return NixState{Ready: false, Reason: "root path is empty"}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return NixState{Ready: false, Reason: "invalid root path: " + err.Error()}
	}
	info, err := os.Stat(abs)
	if err != nil {
		return NixState{Ready: false, Reason: "root path does not exist: " + err.Error()}
	}
	if !info.IsDir() {
		return NixState{Ready: false, Reason: "root path is not a directory"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nix", "develop", ".#dev-env", "--command", "sh", "-c", "echo ready")
	cmd.Dir = abs
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		reason := strings.TrimSpace(stderr.String())
		if reason == "" {
			reason = err.Error()
		}
		if ctx.Err() == context.DeadlineExceeded {
			reason = "probe timed out after " + ProbeTimeout.String()
		}
		return NixState{Ready: false, Reason: reason}
	}
	return NixState{Ready: true, Reason: ""}
}

// IsActionBlockedByNix returns whether the action is blocked by Nix not being ready.
// When nix is not ready, all executable actions are blocked (no bypass).
func IsActionBlockedByNix(req ActionRequest, nix NixState) (bool, string) {
	if nix.Ready {
		return false, ""
	}
	return true, nix.Reason
}
