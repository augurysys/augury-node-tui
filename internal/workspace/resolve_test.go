package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func mkValidRoot(t *testing.T, base string) string {
	t.Helper()
	dir := filepath.Join(base, "scripts", "devices")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(base, "scripts", "lib")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	dir = filepath.Join(base, "pkg")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	return base
}

func TestResolveRoot_ExplicitFlagWins(t *testing.T) {
	tmp := t.TempDir()
	root := mkValidRoot(t, filepath.Join(tmp, "repo"))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot(absRoot, "", tmp)
	if err != nil {
		t.Fatalf("ResolveRoot: %v", err)
	}
	if got != absRoot {
		t.Errorf("ResolveRoot = %q, want %q", got, absRoot)
	}
}

func TestResolveRoot_ConfigPathWhenFlagAbsent(t *testing.T) {
	tmp := t.TempDir()
	root := mkValidRoot(t, filepath.Join(tmp, "repo"))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot("", absRoot, tmp)
	if err != nil {
		t.Fatalf("ResolveRoot: %v", err)
	}
	if got != absRoot {
		t.Errorf("ResolveRoot = %q, want %q", got, absRoot)
	}
}

func TestResolveRoot_AncestorDiscoveryFallback(t *testing.T) {
	tmp := t.TempDir()
	root := mkValidRoot(t, filepath.Join(tmp, "augury-node"))
	subdir := filepath.Join(root, "some", "nested", "dir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	got, err := ResolveRoot("", "", subdir)
	if err != nil {
		t.Fatalf("ResolveRoot: %v", err)
	}
	if got != absRoot {
		t.Errorf("ResolveRoot = %q, want %q", got, absRoot)
	}
}

func TestResolveRoot_ErrorWhenRequiredRootsMissing(t *testing.T) {
	tmp := t.TempDir()
	emptyDir := filepath.Join(tmp, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := ResolveRoot("", "", emptyDir)
	if err == nil {
		t.Error("ResolveRoot: want error when no valid root found, got nil")
	}
}

func TestValidateRoot_Valid(t *testing.T) {
	tmp := t.TempDir()
	root := mkValidRoot(t, filepath.Join(tmp, "repo"))

	if err := ValidateRoot(root); err != nil {
		t.Errorf("ValidateRoot(%q): %v", root, err)
	}
}

func TestValidateRoot_MissingScriptsDevices(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(root, "scripts", "lib"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0755); err != nil {
		t.Fatal(err)
	}

	err := ValidateRoot(root)
	if err == nil {
		t.Error("ValidateRoot: want error when scripts/devices/ missing, got nil")
	}
}

func TestValidateRoot_MissingScriptsLib(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(root, "scripts", "devices"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0755); err != nil {
		t.Fatal(err)
	}

	err := ValidateRoot(root)
	if err == nil {
		t.Error("ValidateRoot: want error when scripts/lib/ missing, got nil")
	}
}

func TestValidateRoot_MissingPkg(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(filepath.Join(root, "scripts", "devices"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "scripts", "lib"), 0755); err != nil {
		t.Fatal(err)
	}

	err := ValidateRoot(root)
	if err == nil {
		t.Error("ValidateRoot: want error when pkg/ missing, got nil")
	}
}
