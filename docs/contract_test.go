package docs

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDocsContract_READMEIncludesStartupSplashBehavior(t *testing.T) {
	readme := readFile(t, "README.md")
	required := []string{"splash", "dismiss"}
	for _, r := range required {
		if !strings.Contains(strings.ToLower(readme), r) {
			t.Errorf("README must document startup splash behavior (missing %q)", r)
		}
	}
}

func TestDocsContract_READMEIncludesAuguryNodePathContract(t *testing.T) {
	readme := readFile(t, "README.md")
	required := []string{"scripts/devices", "scripts/lib", "pkg"}
	for _, r := range required {
		if !strings.Contains(readme, r) {
			t.Errorf("README must document required augury-node path contract (missing %q)", r)
		}
	}
}

func TestDocsContract_READMEIncludesKeybindingTable(t *testing.T) {
	readme := readFile(t, "README.md")
	if !strings.Contains(readme, "keybinding") && !strings.Contains(readme, "keybindings") {
		t.Error("README must include keybinding table or reference to docs/keybindings.md")
	}
}

func TestDocsContract_READMEIncludesLogFilePathContract(t *testing.T) {
	readme := readFile(t, "README.md")
	if !strings.Contains(readme, "tmp/augury-node-tui") {
		t.Errorf("README must document log file path contract (tmp/augury-node-tui)")
	}
}

func TestDocsContract_DocsMandatoryNixPolicy(t *testing.T) {
	phase23 := readFile(t, "phase2-phase3.md")
	required := []string{"nix", "block", "ready"}
	for _, r := range required {
		if !strings.Contains(strings.ToLower(phase23), r) {
			t.Errorf("docs must document mandatory Nix policy (missing %q in phase2-phase3.md)", r)
		}
	}
}

func TestDocsContract_DocsCacheActionKeys(t *testing.T) {
	phase23 := readFile(t, "phase2-phase3.md")
	buildUnit := []string{"B", "R", "D"}
	platform := []string{"P", "U", "X"}
	for _, k := range buildUnit {
		if !strings.Contains(phase23, k) {
			t.Errorf("docs must document build-unit cache action key %q", k)
		}
	}
	for _, k := range platform {
		if !strings.Contains(phase23, k) {
			t.Errorf("docs must document platform-cache action key %q", k)
		}
	}
}

func TestDocsContract_DocsLogTabErrorNavigationKeys(t *testing.T) {
	phase23 := readFile(t, "phase2-phase3.md")
	required := []string{"tab", "e", "full", "error", "j", "k"}
	for _, r := range required {
		if !strings.Contains(strings.ToLower(phase23), r) {
			t.Errorf("docs must document log tab/error navigation (missing %q in phase2-phase3.md)", r)
		}
	}
}

func TestDocsContract_DocsDeveloperDownloadsSourceStates(t *testing.T) {
	phase23 := readFile(t, "phase2-phase3.md")
	required := []string{"built", "hydrated", "missing", "unavailable"}
	for _, r := range required {
		if !strings.Contains(strings.ToLower(phase23), r) {
			t.Errorf("docs must document developer-downloads source state %q", r)
		}
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(1)
	dir := filepath.Dir(filename)
	for dir != "/" {
		p := filepath.Join(dir, name)
		if data, err := os.ReadFile(p); err == nil {
			return string(data)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not find %s", name)
	return ""
}
