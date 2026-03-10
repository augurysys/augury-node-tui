package ci

import "os"

// ResolveToken returns a CircleCI API token.
// Priority: CIRCLE_TOKEN env var > config file value > empty string.
func ResolveToken(configToken string) string {
	if env := os.Getenv("CIRCLE_TOKEN"); env != "" {
		return env
	}
	return configToken
}
