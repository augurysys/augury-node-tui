package ci

import (
	"testing"
	"time"
)

func TestJobDuration(t *testing.T) {
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	stop := time.Date(2026, 1, 1, 10, 5, 30, 0, time.UTC)

	tests := []struct {
		name     string
		job      Job
		expected time.Duration
	}{
		{
			name:     "normal duration",
			job:      Job{StartedAt: &start, StoppedAt: &stop},
			expected: 5*time.Minute + 30*time.Second,
		},
		{
			name:     "nil started_at",
			job:      Job{StartedAt: nil, StoppedAt: &stop},
			expected: 0,
		},
		{
			name:     "nil stopped_at",
			job:      Job{StartedAt: &start, StoppedAt: nil},
			expected: 0,
		},
		{
			name:     "both nil",
			job:      Job{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.Duration()
			if got != tt.expected {
				t.Errorf("Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}
