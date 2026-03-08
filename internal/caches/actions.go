package caches

import (
	"github.com/augurysys/augury-node-tui/internal/engine"
)

// KeyBuildUnitBuild is the key for build-unit build action.
const KeyBuildUnitBuild = "B"

// KeyBuildUnitPull is the key for build-unit pull action.
const KeyBuildUnitPull = "R"

// KeyBuildUnitDelete is the key for build-unit delete action.
const KeyBuildUnitDelete = "D"

// KeyPlatformPull is the key for platform-cache pull action.
const KeyPlatformPull = "P"

// KeyPlatformPush is the key for platform-cache push action.
const KeyPlatformPush = "U"

// KeyPlatformClean is the key for platform-cache clean action.
const KeyPlatformClean = "X"

// ActionForKey returns the engine ActionRequest for the given tab and key.
// Returns zero value and false if the key does not map to an action for the tab.
func ActionForKey(tab int, key string, platformID string) (engine.ActionRequest, bool) {
	switch tab {
	case TabBuildUnit:
		return actionForBuildUnitTab(key, platformID)
	case TabPlatform:
		return actionForPlatformCacheTab(key, platformID)
	default:
		return engine.ActionRequest{}, false
	}
}

func actionForBuildUnitTab(key string, platformID string) (engine.ActionRequest, bool) {
	switch key {
	case KeyBuildUnitBuild:
		return engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetBuild, PlatformID: platformID}, true
	case KeyBuildUnitPull:
		return engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetPull}, true
	case KeyBuildUnitDelete:
		return engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetDelete}, true
	default:
		return engine.ActionRequest{}, false
	}
}

func actionForPlatformCacheTab(key string, platformID string) (engine.ActionRequest, bool) {
	switch key {
	case KeyPlatformPull:
		return engine.ActionRequest{Kind: engine.KindPlatformCache, Target: engine.TargetPull, PlatformID: platformID}, true
	case KeyPlatformPush:
		return engine.ActionRequest{Kind: engine.KindPlatformCache, Target: engine.TargetPush, PlatformID: platformID}, true
	case KeyPlatformClean:
		return engine.ActionRequest{Kind: engine.KindPlatformCache, Target: engine.TargetClean, PlatformID: platformID}, true
	default:
		return engine.ActionRequest{}, false
	}
}

// IsDestructive returns true if the action requires confirmation before execution.
func IsDestructive(req engine.ActionRequest) bool {
	return req.Target == engine.TargetDelete || req.Target == engine.TargetClean
}
