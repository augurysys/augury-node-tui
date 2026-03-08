package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/augurysys/augury-node-tui/internal/run"
)

func TestExecution_SelectedPlatformsExecuteSequentially(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	orderFile := filepath.Join(absRoot, "order.log")
	specs := []run.RunSpec{
		{Name: "first", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "echo first >> " + orderFile}},
		{Name: "second", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "echo second >> " + orderFile}},
		{Name: "third", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "echo third >> " + orderFile}},
	}

	ctx := context.Background()
	summary := ExecuteSequential(ctx, specs)

	if len(summary.Rows) != 3 {
		t.Fatalf("summary rows = %d, want 3", len(summary.Rows))
	}
	data, err := os.ReadFile(orderFile)
	if err != nil {
		t.Fatalf("order file: %v", err)
	}
	content := string(data)
	if content != "first\nsecond\nthird\n" {
		t.Errorf("execution order wrong: got %q, want first\\nsecond\\nthird\\n", content)
	}
}

func TestExecution_FailuresRecordedAndNextPlatformRuns(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	specs := []run.RunSpec{
		{Name: "fail", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 1"}},
		{Name: "succeed", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
	}

	ctx := context.Background()
	summary := ExecuteSequential(ctx, specs)

	if len(summary.Rows) != 2 {
		t.Fatalf("summary rows = %d, want 2", len(summary.Rows))
	}
	if summary.Rows[0].Status != RowStatusFailure {
		t.Errorf("first row status = %q, want failure", summary.Rows[0].Status)
	}
	if summary.Rows[1].Status != RowStatusSuccess {
		t.Errorf("second row status = %q, want success", summary.Rows[1].Status)
	}
}

func TestExecution_CancellationMarksCurrentAndRemainingAppropriately(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	specs := []run.RunSpec{
		{Name: "fast", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
		{Name: "slow", Root: absRoot, Mode: run.ModeSmart, Command: "sleep", Args: []string{"10"}},
		{Name: "after", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	summary := ExecuteSequential(ctx, specs)

	if len(summary.Rows) != 3 {
		t.Fatalf("summary rows = %d, want 3", len(summary.Rows))
	}
	if summary.Rows[0].Status != RowStatusSuccess {
		t.Errorf("first row (completed) status = %q, want success", summary.Rows[0].Status)
	}
	if summary.Rows[1].Status != RowStatusCancelled {
		t.Errorf("second row (current when cancelled) status = %q, want cancelled", summary.Rows[1].Status)
	}
	if summary.Rows[2].Status != RowStatusSkipped {
		t.Errorf("third row (remaining) status = %q, want skipped", summary.Rows[2].Status)
	}
}

func TestExecution_NextToRunSpecClassifiedFromExecuteNotSkipped(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	specs := []run.RunSpec{
		{Name: "first", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
		{Name: "second", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
		{Name: "third", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	summary := ExecuteSequential(ctx, specs)

	if len(summary.Rows) != 3 {
		t.Fatalf("summary rows = %d, want 3", len(summary.Rows))
	}
	if summary.Rows[0].Status != RowStatusCancelled {
		t.Errorf("first row (next-to-run when ctx already cancelled) must be classified from run.Execute as cancelled, not skipped; got %q", summary.Rows[0].Status)
	}
	if summary.Rows[1].Status != RowStatusSkipped {
		t.Errorf("second row (remaining after first cancelled) status = %q, want skipped", summary.Rows[1].Status)
	}
	if summary.Rows[2].Status != RowStatusSkipped {
		t.Errorf("third row (remaining after first cancelled) status = %q, want skipped", summary.Rows[2].Status)
	}
}

func TestExecution_SummaryRowStatusesIncludeSuccessFailureSkippedCancelled(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "repo")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	specs := []run.RunSpec{
		{Name: "ok", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
		{Name: "fail", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 1"}},
	}

	ctx := context.Background()
	summary := ExecuteSequential(ctx, specs)

	seen := make(map[RowStatus]bool)
	for _, r := range summary.Rows {
		seen[r.Status] = true
	}
	if !seen[RowStatusSuccess] {
		t.Error("summary must include success status")
	}
	if !seen[RowStatusFailure] {
		t.Error("summary must include failure status")
	}

	ctx2, cancel2 := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel2()
	}()
	specs2 := []run.RunSpec{
		{Name: "c", Root: absRoot, Mode: run.ModeSmart, Command: "sleep", Args: []string{"5"}},
		{Name: "s", Root: absRoot, Mode: run.ModeSmart, Command: "sh", Args: []string{"-c", "exit 0"}},
	}
	summary2 := ExecuteSequential(ctx2, specs2)
	seen2 := make(map[RowStatus]bool)
	for _, r := range summary2.Rows {
		seen2[r.Status] = true
	}
	if !seen2[RowStatusCancelled] {
		t.Error("summary must include cancelled status when cancelled")
	}
	if !seen2[RowStatusSkipped] {
		t.Error("summary must include skipped status for remaining after cancel")
	}
}
