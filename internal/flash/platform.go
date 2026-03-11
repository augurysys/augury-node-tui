package flash

import (
	"strings"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

// Platform type identifiers returned by DetectPlatformType
const (
	PlatformTypeMP255    = "mp255"
	PlatformTypeSWUpdate = "swupdate"
	PlatformTypeUnknown  = "unknown"
)

// DetectPlatformType returns the flash adapter type for a platform.
// Empty IDs yield PlatformTypeUnknown.
func DetectPlatformType(p platform.Platform) string {
	// MP255 platforms
	if strings.HasPrefix(p.ID, PlatformTypeMP255) {
		return PlatformTypeMP255
	}

	// SWUpdate platforms
	if strings.Contains(p.ID, "cassia") || strings.Contains(p.ID, "moxa") {
		return PlatformTypeSWUpdate
	}

	return PlatformTypeUnknown
}
