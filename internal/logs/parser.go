package logs

import (
	"regexp"
	"strings"
)

// ErrorLevel indicates severity
type ErrorLevel int

const (
	ErrorLevelCritical ErrorLevel = iota
	ErrorLevelError
	ErrorLevelWarning
)

// ErrorLocation represents a detected error in logs
type ErrorLocation struct {
	LineNumber int
	LineText   string
	Level      ErrorLevel
	Context    []string // Lines before/after
	Suggestion string
}

var parseErrorPatterns = []struct {
	regex      *regexp.Regexp
	level      ErrorLevel
	suggestion string
}{
	{
		regex:      regexp.MustCompile(`error: experimental Nix feature.*disabled`),
		level:      ErrorLevelError,
		suggestion: "Enable nix-command and flakes in ~/.config/nix/nix.conf",
	},
	{
		regex:      regexp.MustCompile(`ERROR: Task.*failed`),
		level:      ErrorLevelError,
		suggestion: "Check tmp/work/<package>/temp/log.do_* for details",
	},
	{
		regex:      regexp.MustCompile(`undefined reference to`),
		level:      ErrorLevelError,
		suggestion: "Missing library in DEPENDS or LDFLAGS",
	},
	{
		regex:      regexp.MustCompile(`error:|ERROR:`),
		level:      ErrorLevelError,
		suggestion: "",
	},
	{
		regex:      regexp.MustCompile(`warning:|WARNING:`),
		level:      ErrorLevelWarning,
		suggestion: "",
	},
}

// ParseErrors scans content for error patterns
func ParseErrors(content string) []ErrorLocation {
	lines := strings.Split(content, "\n")
	var errors []ErrorLocation

	for i, line := range lines {
		for _, pattern := range parseErrorPatterns {
			if pattern.regex.MatchString(line) {
				// Extract context (5 lines before/after)
				contextStart := i - 5
				if contextStart < 0 {
					contextStart = 0
				}
				contextEnd := i + 6
				if contextEnd > len(lines) {
					contextEnd = len(lines)
				}

				errors = append(errors, ErrorLocation{
					LineNumber: i + 1, // 1-indexed
					LineText:   line,
					Level:      pattern.level,
					Context:    lines[contextStart:contextEnd],
					Suggestion: pattern.suggestion,
				})
				break // Only match first pattern per line
			}
		}
	}

	return errors
}

var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^[^:]*:\s*error\s*:`),
	regexp.MustCompile(`(?i)\berror:\s`),
	regexp.MustCompile(`(?i)\bfatal:\s`),
	regexp.MustCompile(`\bFAIL\b`),
	regexp.MustCompile(`(?i)\b(?:build|compilation|command|test|execution)\s+failed\b`),
	regexp.MustCompile(`(?i)\berror\b.*\bfailed\b`),
}

func FindFirstErrorLine(log string) (lineIndex int, ok bool) {
	lines := strings.Split(log, "\n")
	for i, line := range lines {
		for _, pat := range errorPatterns {
			if pat.MatchString(line) {
				return i, true
			}
		}
	}
	return -1, false
}

func ExtractContextAround(log string, lineIndex int, before, after int) string {
	if before < 0 {
		before = 0
	}
	if after < 0 {
		after = 0
	}
	lines := strings.Split(log, "\n")
	if lineIndex < 0 || lineIndex >= len(lines) {
		return ""
	}
	start := lineIndex - before
	if start < 0 {
		start = 0
	}
	end := lineIndex + after + 1
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}
