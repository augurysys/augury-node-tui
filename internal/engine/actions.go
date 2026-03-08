package engine

import "fmt"

// actionDisplayNames maps action IDs to human-readable display names.
// Package-level to avoid re-allocation on each Metadata() call.
var actionDisplayNames = map[string]string{
	"build-unit:build":       "Build",
	"build-unit:pull":        "Pull",
	"build-unit:delete":      "Delete",
	"platform-cache:pull":    "Pull",
	"platform-cache:push":    "Push",
	"platform-cache:clean":   "Clean",
	"hydration:dry-run":      "Dry run",
	"hydration:run":          "Run",
	"validations:all":        "All",
	"validations:shellcheck": "ShellCheck",
	"validations:bats":       "Bats",
	"validations:parse":      "Parse",
}

// ActionKind identifies the category of an engine action.
type ActionKind string

const (
	KindBuildUnit     ActionKind = "build-unit"
	KindPlatformCache ActionKind = "platform-cache"
	KindHydration     ActionKind = "hydration"
	KindValidations   ActionKind = "validations"
)

// ActionTarget identifies the specific operation within an action kind.
type ActionTarget string

const (
	TargetBuild      ActionTarget = "build"
	TargetPull       ActionTarget = "pull"
	TargetDelete     ActionTarget = "delete"
	TargetPush       ActionTarget = "push"
	TargetClean      ActionTarget = "clean"
	TargetDryRun     ActionTarget = "dry-run"
	TargetRun        ActionTarget = "run"
	TargetAll        ActionTarget = "all"
	TargetShellcheck ActionTarget = "shellcheck"
	TargetBats       ActionTarget = "bats"
	TargetParse      ActionTarget = "parse"
)

// ActionRequest describes a single engine action to execute.
type ActionRequest struct {
	Kind   ActionKind
	Target ActionTarget
}

// ActionMetadata holds display and metadata for an action.
type ActionMetadata struct {
	DisplayName string
}

// ID returns a unique identifier for the action in "kind:target" form.
func (r ActionRequest) ID() string {
	return fmt.Sprintf("%s:%s", r.Kind, r.Target)
}

// Metadata returns display metadata for the action.
// Unknown actions fall back to using the raw ID as DisplayName.
func (r ActionRequest) Metadata() ActionMetadata {
	id := r.ID()
	if n, ok := actionDisplayNames[id]; ok {
		return ActionMetadata{DisplayName: n}
	}
	return ActionMetadata{DisplayName: id}
}

var (
	BuildUnitBuild   = ActionRequest{Kind: KindBuildUnit, Target: TargetBuild}
	BuildUnitPull    = ActionRequest{Kind: KindBuildUnit, Target: TargetPull}
	BuildUnitDelete  = ActionRequest{Kind: KindBuildUnit, Target: TargetDelete}
	PlatformCachePull = ActionRequest{Kind: KindPlatformCache, Target: TargetPull}
	PlatformCachePush = ActionRequest{Kind: KindPlatformCache, Target: TargetPush}
	PlatformCacheClean = ActionRequest{Kind: KindPlatformCache, Target: TargetClean}
	HydrationDryRun  = ActionRequest{Kind: KindHydration, Target: TargetDryRun}
	HydrationRun     = ActionRequest{Kind: KindHydration, Target: TargetRun}
	ValidationsAll   = ActionRequest{Kind: KindValidations, Target: TargetAll}
	ValidationsShellcheck = ActionRequest{Kind: KindValidations, Target: TargetShellcheck}
	ValidationsBats  = ActionRequest{Kind: KindValidations, Target: TargetBats}
	ValidationsParse = ActionRequest{Kind: KindValidations, Target: TargetParse}
)
