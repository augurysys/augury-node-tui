package status

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mkGitRepo(t *testing.T, base string) string {
	t.Helper()
	root := filepath.Join(base, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@test")
	runGit(t, root, "config", "user.name", "Test")
	writeFile(t, root, "README", "x")
	runGit(t, root, "add", "README")
	runGit(t, root, "commit", "-m", "init")
	return root
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCollect_ReturnsRootBranchSha(t *testing.T) {
	tmp := t.TempDir()
	root := mkGitRepo(t, tmp)
	runGit(t, root, "checkout", "-b", "feature/test")
	runGit(t, root, "commit", "--allow-empty", "-m", "empty")

	got, err := Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if got.Root != root {
		t.Errorf("Root = %q, want %q", got.Root, root)
	}
	if got.Branch != "feature/test" {
		t.Errorf("Branch = %q, want feature/test", got.Branch)
	}
	if len(got.SHA) < 7 {
		t.Errorf("SHA too short: %q", got.SHA)
	}
}

func TestCollect_DirtyIndicatorsForRequiredPaths(t *testing.T) {
	tmp := t.TempDir()
	root := mkGitRepo(t, tmp)
	for _, p := range []string{"common", "submodules/halo-node", "submodules/apus-installation-service", "submodules/check-tinc-start"} {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(full, 0755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, full, "x", "x")
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-m", "add paths")

	got, err := Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if got.Dirty["common/"] {
		t.Error("common/ should be clean after commit")
	}

	writeFile(t, root, "common/foo", "dirty")
	got, err = Collect(root)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if !got.Dirty["common/"] {
		t.Error("common/ should be dirty after uncommitted change")
	}
}

func TestCollect_RequiredPathsDefined(t *testing.T) {
	if len(RequiredPaths) == 0 {
		t.Fatal("RequiredPaths must be non-empty")
	}
	for _, p := range RequiredPaths {
		if !strings.HasSuffix(p, "/") {
			t.Errorf("RequiredPaths entry %q should end with /", p)
		}
	}
}
