package run

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var errReadFailed = errors.New("read failed")
var errWriteFailed = errors.New("write failed")

type errReader struct {
	data []byte
	err  error
}

func (r *errReader) Read(p []byte) (n int, err error) {
	if len(r.data) > 0 {
		n = copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	return 0, r.err
}

type errWriter struct {
	err error
}

func (w *errWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestModeConstants(t *testing.T) {
	if ModeSmart != "Smart" {
		t.Errorf("ModeSmart = %q, want Smart", ModeSmart)
	}
	if ModeClean != "Clean" {
		t.Errorf("ModeClean = %q, want Clean", ModeClean)
	}
	if ModeValidationOnly != "ValidationOnly" {
		t.Errorf("ModeValidationOnly = %q, want ValidationOnly", ModeValidationOnly)
	}
}

func TestExecute_CommandRunsInRepoRoot(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	spec := RunSpec{
		Name:    "cwd-check",
		Root:    absRoot,
		Mode:    ModeSmart,
		Command: "pwd",
		Args:    nil,
	}
	ctx := context.Background()
	result := Execute(ctx, spec)

	if result.Status != "success" {
		t.Errorf("Execute: status = %q, want success", result.Status)
	}
	got := strings.TrimSpace(result.Stdout)
	if got != absRoot {
		t.Errorf("Execute: cwd = %q, want %q", got, absRoot)
	}
}

func TestExecute_LogsPersistedUnderTmpAuguryNodeTui(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	spec := RunSpec{
		Name:    "log-test",
		Root:    absRoot,
		Mode:    ModeSmart,
		Command: "echo",
		Args:    []string{"logged-output"},
	}
	ctx := context.Background()
	_ = Execute(ctx, spec)

	logPath := filepath.Join(absRoot, "tmp", "augury-node-tui", "log-test.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file not found or unreadable: %v", err)
	}
	if !strings.Contains(string(data), "logged-output") {
		t.Errorf("log file content %q does not contain logged-output", string(data))
	}
}

func TestExecute_CleanModeInjectsCLEAN1(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	spec := RunSpec{
		Name:    "clean-env",
		Root:    absRoot,
		Mode:    ModeClean,
		Command: "sh",
		Args:    []string{"-c", "echo $CLEAN"},
	}
	ctx := context.Background()
	result := Execute(ctx, spec)

	if result.Status != "success" {
		t.Errorf("Execute: status = %q, want success", result.Status)
	}
	got := strings.TrimSpace(result.Stdout)
	if got != "1" {
		t.Errorf("Execute: CLEAN = %q, want 1", got)
	}
}

func TestExecute_SmartModeDoesNotInjectCLEAN(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	spec := RunSpec{
		Name:    "smart-env",
		Root:    absRoot,
		Mode:    ModeSmart,
		Command: "sh",
		Args:    []string{"-c", "echo ${CLEAN:-unset}"},
	}
	ctx := context.Background()
	result := Execute(ctx, spec)

	if result.Status != "success" {
		t.Errorf("Execute: status = %q, want success", result.Status)
	}
	got := strings.TrimSpace(result.Stdout)
	if got != "unset" {
		t.Errorf("Execute: CLEAN = %q, want unset", got)
	}
}

func TestExecute_CancellationReturnsCancelledStatus(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	spec := RunSpec{
		Name:    "cancel-test",
		Root:    absRoot,
		Mode:    ModeSmart,
		Command: "sleep",
		Args:    []string{"10"},
	}
	result := Execute(ctx, spec)

	if result.Status != "cancelled" {
		t.Errorf("Execute: status = %q, want cancelled", result.Status)
	}
}

func TestExecute_StreamsOutputToMemoryAndLog(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	spec := RunSpec{
		Name:    "stream-test",
		Root:    absRoot,
		Mode:    ModeSmart,
		Command: "sh",
		Args:    []string{"-c", "echo out; echo err >&2"},
	}
	ctx := context.Background()
	result := Execute(ctx, spec)

	if result.Status != "success" {
		t.Errorf("Execute: status = %q, want success", result.Status)
	}
	if !strings.Contains(result.Stdout, "out") {
		t.Errorf("Execute: stdout %q does not contain out", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "err") {
		t.Errorf("Execute: stderr %q does not contain err", result.Stderr)
	}

	logPath := filepath.Join(absRoot, "tmp", "augury-node-tui", "stream-test.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log file: %v", err)
	}
	if !strings.Contains(string(data), "out") || !strings.Contains(string(data), "err") {
		t.Errorf("log file does not contain both out and err: %q", string(data))
	}
}

func TestStreamTo_PropagatesScannerErr(t *testing.T) {
	r := &errReader{data: []byte("line\n"), err: errReadFailed}
	var buf strings.Builder
	log := &strings.Builder{}

	err := streamTo(r, &buf, log)
	if err == nil {
		t.Error("streamTo: want error from scanner.Err(), got nil")
	}
	if !errors.Is(err, errReadFailed) {
		t.Errorf("streamTo: err = %v, want errReadFailed", err)
	}
}

func TestStreamTo_PropagatesWriteError(t *testing.T) {
	r := &errReader{data: []byte("line\n"), err: nil}
	var buf strings.Builder
	log := &errWriter{err: errWriteFailed}

	err := streamTo(r, &buf, log)
	if err == nil {
		t.Error("streamTo: want error from mw.Write, got nil")
	}
	if !errors.Is(err, errWriteFailed) {
		t.Errorf("streamTo: err = %v, want errWriteFailed", err)
	}
}
