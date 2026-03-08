package caches

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/engine"
)

func TestBuildUnitTabActionKeys(t *testing.T) {
	tests := []struct {
		key        string
		wantKind   engine.ActionKind
		wantTarget engine.ActionTarget
		wantPlat   string
	}{
		{"B", engine.KindBuildUnit, engine.TargetBuild, "node2"},
		{"R", engine.KindBuildUnit, engine.TargetPull, ""},
		{"D", engine.KindBuildUnit, engine.TargetDelete, ""},
	}
	for _, tt := range tests {
		req, ok := ActionForKey(TabBuildUnit, tt.key, tt.wantPlat)
		if !ok {
			t.Errorf("ActionForKey(build-unit, %q) = _, false; want true", tt.key)
			continue
		}
		if req.Kind != tt.wantKind {
			t.Errorf("ActionForKey(build-unit, %q).Kind = %v; want %v", tt.key, req.Kind, tt.wantKind)
		}
		if req.Target != tt.wantTarget {
			t.Errorf("ActionForKey(build-unit, %q).Target = %v; want %v", tt.key, req.Target, tt.wantTarget)
		}
		if tt.wantPlat != "" && req.PlatformID != tt.wantPlat {
			t.Errorf("ActionForKey(build-unit, %q).PlatformID = %q; want %q", tt.key, req.PlatformID, tt.wantPlat)
		}
	}
}

func TestPlatformCacheTabActionKeys(t *testing.T) {
	tests := []struct {
		key        string
		wantKind   engine.ActionKind
		wantTarget engine.ActionTarget
		wantPlat   string
	}{
		{"P", engine.KindPlatformCache, engine.TargetPull, "moxa-uc3100"},
		{"U", engine.KindPlatformCache, engine.TargetPush, "moxa-uc3100"},
		{"X", engine.KindPlatformCache, engine.TargetClean, "moxa-uc3100"},
	}
	for _, tt := range tests {
		req, ok := ActionForKey(TabPlatform, tt.key, tt.wantPlat)
		if !ok {
			t.Errorf("ActionForKey(platform-cache, %q) = _, false; want true", tt.key)
			continue
		}
		if req.Kind != tt.wantKind {
			t.Errorf("ActionForKey(platform-cache, %q).Kind = %v; want %v", tt.key, req.Kind, tt.wantKind)
		}
		if req.Target != tt.wantTarget {
			t.Errorf("ActionForKey(platform-cache, %q).Target = %v; want %v", tt.key, req.Target, tt.wantTarget)
		}
	}
}

func TestDisabledActionBehaviorWithCapabilityReason(t *testing.T) {
	root := t.TempDir()
	req := engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetPull}
	cap := engine.ResolveCapability(root, req)
	if cap.Available {
		t.Error("ResolveCapability on empty root should return Available=false")
	}
	if cap.Reason == "" {
		t.Error("ResolveCapability when unavailable must include Reason")
	}
}

func TestDestructiveActionConfirmationRequired(t *testing.T) {
	deleteReq := engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetDelete}
	cleanReq := engine.ActionRequest{Kind: engine.KindPlatformCache, Target: engine.TargetClean}
	if !IsDestructive(deleteReq) {
		t.Error("build-unit:delete must be destructive")
	}
	if !IsDestructive(cleanReq) {
		t.Error("platform-cache:clean must be destructive")
	}
	pullReq := engine.ActionRequest{Kind: engine.KindBuildUnit, Target: engine.TargetPull}
	if IsDestructive(pullReq) {
		t.Error("build-unit:pull must not be destructive")
	}
}
