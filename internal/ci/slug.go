package ci

import (
	"fmt"
	"strings"
)

// SlugFromRemote derives a CircleCI project slug from a git remote URL.
// Supports: git@github.com:org/repo.git, https://github.com/org/repo.git
func SlugFromRemote(remoteURL string) (string, error) {
	u := strings.TrimSpace(remoteURL)

	if strings.HasPrefix(u, "git@github.com:") {
		path := strings.TrimPrefix(u, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", fmt.Errorf("cannot parse SSH remote: %s", remoteURL)
		}
		return "gh/" + parts[0] + "/" + parts[1], nil
	}

	if strings.HasPrefix(u, "https://github.com/") {
		path := strings.TrimPrefix(u, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", fmt.Errorf("cannot parse HTTPS remote: %s", remoteURL)
		}
		return "gh/" + parts[0] + "/" + parts[1], nil
	}

	return "", fmt.Errorf("unsupported remote URL format: %s", remoteURL)
}
