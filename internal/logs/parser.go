package logs

import (
	"regexp"
	"strings"
)

var errorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^[^:]*:\s*error\s*:`),
	regexp.MustCompile(`(?i)\berror:\s`),
	regexp.MustCompile(`(?i)\bfatal:\s`),
	regexp.MustCompile(`\bFAIL\b`),
	regexp.MustCompile(`(?i)\bfailed\b`),
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
