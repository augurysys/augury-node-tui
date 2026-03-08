package setup

import (
	"os/exec"
	"os/user"
	"strings"
)

type HealthCheckResult struct {
	Available bool
	Message   string
	Error     error
}

func CheckNixInstalled() HealthCheckResult {
	_, err := exec.LookPath("nix")
	if err != nil {
		return HealthCheckResult{
			Available: false,
			Message:   "Nix not found in PATH",
			Error:     err,
		}
	}
	return HealthCheckResult{
		Available: true,
		Message:   "Nix is installed",
	}
}

func CheckNixExperimentalFeatures() HealthCheckResult {
	cmd := exec.Command("nix", "show-config", "experimental-features")
	out, err := cmd.Output()
	if err != nil {
		return HealthCheckResult{
			Available: false,
			Message:   "Cannot check experimental features",
			Error:     err,
		}
	}

	features := string(out)
	hasNixCommand := strings.Contains(features, "nix-command")
	hasFlakes := strings.Contains(features, "flakes")

	if hasNixCommand && hasFlakes {
		return HealthCheckResult{
			Available: true,
			Message:   "Experimental features enabled",
		}
	}

	return HealthCheckResult{
		Available: false,
		Message:   "nix-command and flakes not enabled",
	}
}

func CheckNixGroup() HealthCheckResult {
	currentUser, err := user.Current()
	if err != nil {
		return HealthCheckResult{
			Available: false,
			Message:   "Cannot determine current user",
			Error:     err,
		}
	}

	gids, err := currentUser.GroupIds()
	if err != nil {
		return HealthCheckResult{
			Available: false,
			Message:   "Cannot get group IDs",
			Error:     err,
		}
	}

	for _, gid := range gids {
		grp, err := user.LookupGroupId(gid)
		if err == nil && grp.Name == "nix-users" {
			return HealthCheckResult{
				Available: true,
				Message:   "User is in nix-users group",
			}
		}
	}

	return HealthCheckResult{
		Available: false,
		Message:   "User not in nix-users group",
	}
}

func CheckDaemonSocket() HealthCheckResult {
	socketPath := "/nix/var/nix/daemon-socket/socket"

	cmd := exec.Command("test", "-S", socketPath)
	if err := cmd.Run(); err != nil {
		return HealthCheckResult{
			Available: false,
			Message:   "Nix daemon socket not accessible",
			Error:     err,
		}
	}

	return HealthCheckResult{
		Available: true,
		Message:   "Nix daemon socket is accessible",
	}
}
