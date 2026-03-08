package status

import (
	"os/exec"
	"path/filepath"
	"strings"
)

var RequiredPaths = []string{
	"common/",
	"submodules/halo-node/",
	"submodules/apus-installation-service/",
	"submodules/check-tinc-start/",
}

type RepoStatus struct {
	Root   string
	Branch string
	SHA    string
	Dirty  map[string]bool
}

func Collect(root string) (RepoStatus, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return RepoStatus{}, err
	}
	st := RepoStatus{Root: abs, Dirty: make(map[string]bool)}
	for _, p := range RequiredPaths {
		st.Dirty[p] = false
	}

	branch, err := gitOutput(abs, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return st, err
	}
	st.Branch = strings.TrimSpace(branch)

	sha, err := gitOutput(abs, "rev-parse", "HEAD")
	if err != nil {
		return st, err
	}
	st.SHA = strings.TrimSpace(sha)
	if len(st.SHA) > 7 {
		st.SHA = st.SHA[:7]
	}

	porcelain, err := gitOutput(abs, "status", "--porcelain")
	if err != nil {
		return st, err
	}
	for _, p := range RequiredPaths {
		prefix := strings.TrimSuffix(p, "/")
		for _, line := range strings.Split(porcelain, "\n") {
			line = strings.TrimSpace(line)
			if len(line) < 4 {
				continue
			}
			path := strings.TrimSpace(line[3:])
			if idx := strings.Index(path, " "); idx > 0 {
				path = path[:idx]
			}
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				st.Dirty[p] = true
				break
			}
		}
	}
	return st, nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
