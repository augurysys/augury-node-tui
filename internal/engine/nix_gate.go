package engine

import (
	"bytes"
	"os/exec"
	"strings"
)

// NixState represents the result of probing Nix readiness.
type NixState struct {
	Ready  bool
	Reason string
}

// ProbeNix runs the nix develop probe command in root and returns readiness.
// Contract: nix develop .#dev-env --command sh -c 'echo ready'
func ProbeNix(root string) NixState {
	cmd := exec.Command("nix", "develop", ".#dev-env", "--command", "sh", "-c", "echo ready")
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		reason := strings.TrimSpace(stderr.String())
		if reason == "" {
			reason = err.Error()
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
