package ansi

import "strings"

// StripAnsi removes ANSI CSI sequences (\x1b[...m) from the string
func StripAnsi(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && s[j] != 'm' {
				j++
			}
			if j < len(s) {
				i = j + 1
			} else {
				i = len(s)
			}
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}
