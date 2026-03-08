package engine

import (
	"os"
	"path/filepath"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

// Capability describes whether an action can be executed and which script to run.
type Capability struct {
	Available  bool
	Reason     string
	ScriptPath string
}

// actionScriptMap maps action ID to required script path (relative to root).
// build-unit:build uses platform registry (single-source-of-truth).
var actionScriptMap = map[string]string{
	"build-unit:pull":        "scripts/dev/pull-artifacts.sh",
	"build-unit:delete":      "scripts/dev/delete-build-unit-cache.sh",
	"platform-cache:pull":    "scripts/dev/pull-artifacts.sh",
	"platform-cache:push":    "scripts/dev/push-artifacts.sh",
	"platform-cache:clean":   "scripts/dev/clean-platform-cache.sh",
	"hydration:dry-run":      "scripts/hydrate",
	"hydration:run":          "scripts/hydrate",
	"validations:all":        "scripts/validate-all.sh",
	"validations:shellcheck": "scripts/validate-shellcheck.sh",
	"validations:bats":       "scripts/validate-bats.sh",
	"validations:parse":      "scripts/validate-parse-test.sh",
}

// ResolveCapability resolves whether the action is available and the script path.
// Available is true only when the required script exists on disk.
// No direct fallback command path exists; execution is script-only.
// Empty root returns unavailable. Unknown platform IDs return unavailable.
func ResolveCapability(root string, req ActionRequest) Capability {
	if root == "" {
		return Capability{
			Available:  false,
			Reason:     "root path is empty",
			ScriptPath: "",
		}
	}

	id := req.ID()
	if id == "build-unit:build" {
		return resolveBuildUnitBuild(root, req)
	}

	template, ok := actionScriptMap[id]
	if !ok {
		return Capability{
			Available:  false,
			Reason:     "unknown action: " + id,
			ScriptPath: "",
		}
	}

	relPath := template
	scriptPath := filepath.Join(root, relPath)
	_, err := os.Stat(scriptPath)
	if err != nil {
		return Capability{
			Available:  false,
			Reason:     "script not found: " + relPath,
			ScriptPath: scriptPath,
		}
	}
	return Capability{
		Available:  true,
		Reason:     "",
		ScriptPath: scriptPath,
	}
}

func resolveBuildUnitBuild(root string, req ActionRequest) Capability {
	if req.PlatformID == "" {
		return Capability{
			Available:  false,
			Reason:     "platform required for build-unit:build",
			ScriptPath: "",
		}
	}
	pl, ok := platform.ByID(req.PlatformID)
	if !ok {
		return Capability{
			Available:  false,
			Reason:     "unknown platform: " + req.PlatformID,
			ScriptPath: "",
		}
	}
	scriptPath := filepath.Join(root, pl.ScriptRelPath)
	_, err := os.Stat(scriptPath)
	if err != nil {
		return Capability{
			Available:  false,
			Reason:     "script not found: " + pl.ScriptRelPath,
			ScriptPath: scriptPath,
		}
	}
	return Capability{
		Available:  true,
		Reason:     "",
		ScriptPath: scriptPath,
	}
}
