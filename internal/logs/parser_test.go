package logs

import (
	"strings"
	"testing"
)

func TestParser_FirstErrorMarkerDetection(t *testing.T) {
	tests := []struct {
		name     string
		log      string
		wantLine int
		wantOK   bool
	}{
		{
			name:     "empty log has no error",
			log:      "",
			wantLine: -1,
			wantOK:   false,
		},
		{
			name: "error colon pattern",
			log:  "line1\nline2\nerror: something failed\nline4",
			wantLine: 2,
			wantOK:   true,
		},
		{
			name: "Error uppercase",
			log:  "build started\nError: compilation failed\nmore output",
			wantLine: 1,
			wantOK:   true,
		},
		{
			name: "fatal pattern",
			log:  "a\nb\nc\nfatal: exit\n",
			wantLine: 3,
			wantOK:   true,
		},
		{
			name: "FAIL pattern",
			log:  "running tests\nFAIL\tpkg/foo\nok",
			wantLine: 1,
			wantOK:   true,
		},
		{
			name:     "no error in log",
			log:      "line1\nline2\nline3",
			wantLine: -1,
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line, ok := FindFirstErrorLine(tt.log)
			if ok != tt.wantOK {
				t.Errorf("FindFirstErrorLine() ok = %v, want %v", ok, tt.wantOK)
			}
			if line != tt.wantLine {
				t.Errorf("FindFirstErrorLine() line = %d, want %d", line, tt.wantLine)
			}
		})
	}
}

func TestParser_ContextWindowExtraction(t *testing.T) {
	lines := strings.Split("L0\nL1\nL2\nL3\nL4\nL5\nL6\nL7\nL8\nL9", "\n")
	log := strings.Join(lines, "\n")

	tests := []struct {
		name     string
		lineIdx  int
		before   int
		after    int
		wantCont string
	}{
		{
			name:     "context around middle",
			lineIdx:  4,
			before:   2,
			after:    2,
			wantCont: "L2\nL3\nL4\nL5\nL6",
		},
		{
			name:     "context at start",
			lineIdx:  0,
			before:   2,
			after:    2,
			wantCont: "L0\nL1\nL2",
		},
		{
			name:     "context at end",
			lineIdx:  9,
			before:   2,
			after:    2,
			wantCont: "L7\nL8\nL9",
		},
		{
			name:     "zero before after",
			lineIdx:  5,
			before:   0,
			after:    0,
			wantCont: "L5",
		},
		{
			name:     "negative before after clamped to zero",
			lineIdx:  3,
			before:   -1,
			after:    -2,
			wantCont: "L3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractContextAround(log, tt.lineIdx, tt.before, tt.after)
			if got != tt.wantCont {
				t.Errorf("ExtractContextAround() = %q, want %q", got, tt.wantCont)
			}
		})
	}
}
