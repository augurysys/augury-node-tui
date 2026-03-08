package developerdownloads

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseIndex_ParsesValidJSON(t *testing.T) {
	dir := t.TempDir()
	dd := filepath.Join(dir, "developer-downloads")
	if err := os.MkdirAll(dd, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dd, "index.json")
	content := `{"platforms":[{"name":"node2","enabled":true,"source":"built","source_sha":"abc123","source_branch":"main"},{"name":"moxa-uc3100","enabled":true,"source":"hydrated","source_sha":"def456","source_branch":"main"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if idx == nil {
		t.Fatal("index must not be nil")
	}
	if got := idx.SourceState("node2"); got != SourceBuilt {
		t.Errorf("node2 source: want built, got %q", got)
	}
	if got := idx.SourceState("moxa-uc3100"); got != SourceHydrated {
		t.Errorf("moxa-uc3100 source: want hydrated, got %q", got)
	}
}

func TestParseIndex_MapsPerPlatformSourceStates(t *testing.T) {
	dir := t.TempDir()
	dd := filepath.Join(dir, "developer-downloads")
	if err := os.MkdirAll(dd, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dd, "index.json")
	content := `{"platforms":[{"name":"node2","enabled":true,"source":"built"},{"name":"cassia-x2000","enabled":true,"source":"hydrated"},{"name":"mp255-ulrpm","enabled":true,"source":"missing"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	tests := []struct {
		platform string
		want     SourceState
	}{
		{"node2", SourceBuilt},
		{"cassia-x2000", SourceHydrated},
		{"mp255-ulrpm", SourceMissing},
	}
	for _, tt := range tests {
		if got := idx.SourceState(tt.platform); got != tt.want {
			t.Errorf("SourceState(%q): want %q, got %q", tt.platform, tt.want, got)
		}
	}
}

func TestReadAt_ReturnsUnavailableWhenIndexAbsent(t *testing.T) {
	dir := t.TempDir()
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: want nil error (unavailable), got %v", err)
	}
	if idx != nil {
		t.Errorf("index must be nil when file absent; got %v", idx)
	}
}

func TestParseIndex_MoxaUc3100UlrpmResolvesDirectly(t *testing.T) {
	dir := t.TempDir()
	dd := filepath.Join(dir, "developer-downloads")
	if err := os.MkdirAll(dd, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"platforms":[{"name":"moxa-uc3100-ulrpm","enabled":true,"source":"hydrated"}]}`
	if err := os.WriteFile(filepath.Join(dd, "index.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if idx == nil {
		t.Fatal("index must not be nil")
	}
	if got := idx.SourceState("moxa-uc3100-ulrpm"); got != SourceHydrated {
		t.Errorf("moxa-uc3100-ulrpm: want hydrated, got %q", got)
	}
}

func TestParseIndex_EmptyNameEntriesIgnored(t *testing.T) {
	dir := t.TempDir()
	dd := filepath.Join(dir, "developer-downloads")
	if err := os.MkdirAll(dd, 0755); err != nil {
		t.Fatal(err)
	}
	content := `{"platforms":[{"name":"","enabled":true,"source":"built"},{"name":"node2","enabled":true,"source":"hydrated"}]}`
	if err := os.WriteFile(filepath.Join(dd, "index.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if idx == nil {
		t.Fatal("index must not be nil")
	}
	if got := idx.SourceState("node2"); got != SourceHydrated {
		t.Errorf("node2 should be hydrated; got %q", got)
	}
	if got := idx.SourceState(""); got != SourceMissing {
		t.Errorf("empty name should not be stored; want missing, got %q", got)
	}
}

func TestReadAt_ResolvesPathUnderDeveloperDownloads(t *testing.T) {
	dir := t.TempDir()
	dd := filepath.Join(dir, "developer-downloads")
	if err := os.MkdirAll(dd, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dd, "index.json")
	content := `{"platforms":[{"name":"node2","enabled":true,"source":"built"}]}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	idx, err := ReadAt(dir)
	if err != nil {
		t.Fatalf("ReadAt: %v", err)
	}
	if idx == nil {
		t.Fatal("index must not be nil")
	}
	if got := idx.SourceState("node2"); got != SourceBuilt {
		t.Errorf("node2 source: want built, got %q", got)
	}
}
