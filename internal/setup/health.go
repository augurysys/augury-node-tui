package setup

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/augurysys/augury-node-tui/internal/workspace"
)

func FindAuguryNodeRoot(cwd string) (string, error) {
	return workspace.ResolveRoot("", "", cwd)
}

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

func AutoFixNixConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "nix")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path := filepath.Join(dir, "nix.conf")

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	content := string(data)
	hasNixCommand := strings.Contains(content, "nix-command")
	hasFlakes := strings.Contains(content, "flakes")

	// If both are present, nothing to do
	if hasNixCommand && hasFlakes {
		return nil
	}

	// Add or update experimental-features line
	if !strings.Contains(content, "experimental-features") {
		content += "\nexperimental-features = nix-command flakes\n"
	} else {
		// Update existing line (basic approach: append to end)
		content += "\n# Updated by augury-node-tui\nexperimental-features = nix-command flakes\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}
