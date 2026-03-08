package diagram

import (
	"strings"
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

func TestPlatformFlow_Deterministic(t *testing.T) {
	platforms := platform.Registry()
	a := PlatformFlow(platforms)
	b := PlatformFlow(platforms)
	if a != b {
		t.Errorf("PlatformFlow must be deterministic; two calls differ")
	}
}

func TestPlatformFlow_ContainsFlowElements(t *testing.T) {
	platforms := platform.Registry()
	out := PlatformFlow(platforms)
	if !strings.Contains(out, "Platforms") {
		t.Errorf("PlatformFlow must contain Platforms; got %q", out)
	}
	if !strings.Contains(out, "Build") {
		t.Errorf("PlatformFlow must contain Build; got %q", out)
	}
	if !strings.Contains(out, "Hydrate") {
		t.Errorf("PlatformFlow must contain Hydrate; got %q", out)
	}
}

func TestCacheTopology_Deterministic(t *testing.T) {
	a := CacheTopology(0)
	b := CacheTopology(0)
	if a != b {
		t.Errorf("CacheTopology must be deterministic; two calls differ")
	}
}

func TestCacheTopology_PlatformTab_Deterministic(t *testing.T) {
	a := CacheTopology(1)
	b := CacheTopology(1)
	if a != b {
		t.Errorf("CacheTopology(activeTab==1) must be deterministic; two calls differ")
	}
	if !strings.Contains(a, "platform") && !strings.Contains(a, "Platform") {
		t.Errorf("CacheTopology(1) must contain platform; got %q", a)
	}
}

func TestCacheTopology_ContainsTopologyElements(t *testing.T) {
	out := CacheTopology(1)
	if !strings.Contains(out, "build-unit") && !strings.Contains(out, "Build-unit") {
		t.Errorf("CacheTopology must contain build-unit; got %q", out)
	}
	if !strings.Contains(out, "platform") && !strings.Contains(out, "Platform") {
		t.Errorf("CacheTopology must contain platform; got %q", out)
	}
}

func TestValidationPipeline_Deterministic(t *testing.T) {
	a := ValidationPipeline()
	b := ValidationPipeline()
	if a != b {
		t.Errorf("ValidationPipeline must be deterministic; two calls differ")
	}
}

func TestValidationPipeline_ContainsPipelineElements(t *testing.T) {
	out := ValidationPipeline()
	if !strings.Contains(out, "all") {
		t.Errorf("ValidationPipeline must contain all; got %q", out)
	}
	if !strings.Contains(out, "shellcheck") {
		t.Errorf("ValidationPipeline must contain shellcheck; got %q", out)
	}
	if !strings.Contains(out, "bats") {
		t.Errorf("ValidationPipeline must contain bats; got %q", out)
	}
	if !strings.Contains(out, "parse") {
		t.Errorf("ValidationPipeline must contain parse; got %q", out)
	}
}
