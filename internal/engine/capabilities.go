package engine

import (
	"os"
	"path/filepath"
)

// Capability describes whether an action can be executed and which script to run.
type Capability struct {
	Available  bool
	Reason     string
	ScriptPath string
}

// actionScriptMap maps action ID to required script path (relative to root).
// Platform-specific actions use "<platform>" placeholder.
var actionScriptMap = map[string]string{
	"build-unit:build":       "scripts/devices/<platform>-build.sh",
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
func ResolveCapability(root string, req ActionRequest) Capability {
	id := req.ID()
	template, ok := actionScriptMap[id]
	if !ok {
		return Capability{
			Available:  false,
			Reason:     "unknown action: " + id,
			ScriptPath: "",
		}
	}

	var relPath string
	if template == "scripts/devices/<platform>-build.sh" {
		if req.PlatformID == "" {
			return Capability{
				Available:  false,
				Reason:     "platform required for build-unit:build",
				ScriptPath: filepath.Join(root, "scripts/devices/<platform>-build.sh"),
			}
		}
		relPath = "scripts/devices/" + req.PlatformID + "-build.sh"
	} else {
		relPath = template
	}

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
