package flash

import (
	"testing"

	"github.com/augurysys/augury-node-tui/internal/platform"
)

func TestDetectPlatformType(t *testing.T) {
	tests := []struct {
		name     string
		platform platform.Platform
		want     string
	}{
		{
			name:     "mp255 platform",
			platform: platform.Platform{ID: "mp255-ulrpm"},
			want:     PlatformTypeMP255,
		},
		{
			name:     "cassia platform",
			platform: platform.Platform{ID: "cassia-x2000"},
			want:     PlatformTypeSWUpdate,
		},
		{
			name:     "moxa platform",
			platform: platform.Platform{ID: "moxa-uc3100"},
			want:     PlatformTypeSWUpdate,
		},
		{
			name:     "unknown platform",
			platform: platform.Platform{ID: "unknown-device"},
			want:     PlatformTypeUnknown,
		},
		{
			name:     "empty ID",
			platform: platform.Platform{ID: ""},
			want:     PlatformTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectPlatformType(tt.platform)
			if got != tt.want {
				t.Errorf("DetectPlatformType() = %v, want %v", got, tt.want)
			}
		})
	}
}
