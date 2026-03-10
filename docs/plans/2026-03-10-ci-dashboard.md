# CI Dashboard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a CI status/dashboard screen to augury-node-tui that shows the current branch's latest CircleCI pipeline and lets users view job logs inline.

**Architecture:** HTTP client fetches CircleCI v2 API metadata (pipeline -> workflow -> jobs -> artifacts), downloads log artifacts to disk, and LogViewer displays them. Auth via CIRCLE_TOKEN env var with config.toml fallback. New `internal/ci/` package, new route `"ci"` on keybind `p`.

**Tech Stack:** Go stdlib `net/http`, `encoding/json`, Bubbletea/Charm, existing LogViewer/DataTable/Card components.

---

### Task 1: API Types

**Files:**
- Create: `internal/ci/types.go`
- Test: `internal/ci/types_test.go`

**Step 1: Write the types file**

```go
package ci

import "time"

type Pipeline struct {
	ID     string    `json:"id"`
	Number int       `json:"number"`
	State  string    `json:"state"`
	VCS    VCSInfo   `json:"vcs"`
	CreatedAt time.Time `json:"created_at"`
}

type VCSInfo struct {
	Branch   string `json:"branch"`
	Revision string `json:"revision"`
}

type PipelineResponse struct {
	Items    []Pipeline `json:"items"`
	NextPage string     `json:"next_page_token"`
}

type Workflow struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type WorkflowResponse struct {
	Items    []Workflow `json:"items"`
	NextPage string     `json:"next_page_token"`
}

type Job struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	JobNumber   int    `json:"job_number"`
	StartedAt   *time.Time `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

type JobResponse struct {
	Items    []Job  `json:"items"`
	NextPage string `json:"next_page_token"`
}

type Artifact struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

type ArtifactResponse struct {
	Items    []Artifact `json:"items"`
	NextPage string     `json:"next_page_token"`
}

// Duration returns the job duration, or zero if timestamps are missing.
func (j Job) Duration() time.Duration {
	if j.StartedAt == nil || j.StoppedAt == nil {
		return 0
	}
	return j.StoppedAt.Sub(*j.StartedAt)
}
```

**Step 2: Write test for Duration()**

```go
package ci

import (
	"testing"
	"time"
)

func TestJobDuration(t *testing.T) {
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	stop := time.Date(2026, 1, 1, 10, 5, 30, 0, time.UTC)

	tests := []struct {
		name     string
		job      Job
		expected time.Duration
	}{
		{
			name:     "normal duration",
			job:      Job{StartedAt: &start, StoppedAt: &stop},
			expected: 5*time.Minute + 30*time.Second,
		},
		{
			name:     "nil started_at",
			job:      Job{StartedAt: nil, StoppedAt: &stop},
			expected: 0,
		},
		{
			name:     "nil stopped_at",
			job:      Job{StartedAt: &start, StoppedAt: nil},
			expected: 0,
		},
		{
			name:     "both nil",
			job:      Job{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.Duration()
			if got != tt.expected {
				t.Errorf("Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}
```

**Step 3: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/ci/ -v -run TestJobDuration`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/ci/types.go internal/ci/types_test.go
git commit -m "ci: add CircleCI API response types"
```

---

### Task 2: HTTP Client

**Files:**
- Create: `internal/ci/client.go`
- Test: `internal/ci/client_test.go`

**Step 1: Write the client**

The client wraps CircleCI v2 API. Base URL: `https://circleci.com/api/v2`.
All requests need `Circle-Token` header.

```go
package ci

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: "https://circleci.com/api/v2",
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Circle-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// LatestPipeline returns the most recent pipeline for a branch.
// slug format: "gh/org/repo"
func (c *Client) LatestPipeline(slug, branch string) (*Pipeline, error) {
	path := fmt.Sprintf("/project/%s/pipeline?branch=%s",
		url.PathEscape(slug), url.QueryEscape(branch))
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp PipelineResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("no pipelines found for branch %q", branch)
	}
	return &resp.Items[0], nil
}

// Workflows returns workflows for a pipeline.
func (c *Client) Workflows(pipelineID string) ([]Workflow, error) {
	path := fmt.Sprintf("/pipeline/%s/workflow", pipelineID)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp WorkflowResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// Jobs returns jobs for a workflow.
func (c *Client) Jobs(workflowID string) ([]Job, error) {
	path := fmt.Sprintf("/workflow/%s/job", workflowID)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp JobResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// Artifacts returns artifacts for a job.
// slug format: "gh/org/repo"
func (c *Client) Artifacts(slug string, jobNumber int) ([]Artifact, error) {
	path := fmt.Sprintf("/project/%s/%d/artifacts",
		url.PathEscape(slug), jobNumber)
	data, err := c.get(path)
	if err != nil {
		return nil, err
	}
	var resp ArtifactResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// DownloadArtifact fetches raw artifact content.
func (c *Client) DownloadArtifact(artifactURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", artifactURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Circle-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
```

**Step 2: Write tests using httptest**

```go
package ci

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(server *httptest.Server) *Client {
	c := NewClient("test-token")
	c.baseURL = server.URL
	return c
}

func TestLatestPipeline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Circle-Token") != "test-token" {
			t.Error("missing auth token")
		}
		resp := PipelineResponse{
			Items: []Pipeline{
				{ID: "pipe-1", Number: 42, State: "created"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	p, err := c.LatestPipeline("gh/org/repo", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != "pipe-1" {
		t.Errorf("ID = %q, want %q", p.ID, "pipe-1")
	}
	if p.Number != 42 {
		t.Errorf("Number = %d, want %d", p.Number, 42)
	}
}

func TestLatestPipelineNoPipelines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PipelineResponse{})
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.LatestPipeline("gh/org/repo", "main")
	if err == nil {
		t.Fatal("expected error for empty pipelines")
	}
}

func TestWorkflows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := WorkflowResponse{
			Items: []Workflow{
				{ID: "wf-1", Name: "build-and-release", Status: "success"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	wfs, err := c.Workflows("pipe-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wfs) != 1 || wfs[0].Name != "build-and-release" {
		t.Errorf("unexpected workflows: %+v", wfs)
	}
}

func TestJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := JobResponse{
			Items: []Job{
				{ID: "job-1", Name: "lint-and-test", Status: "success", JobNumber: 100},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := newTestClient(server)
	jobs, err := c.Jobs("wf-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 || jobs[0].Name != "lint-and-test" {
		t.Errorf("unexpected jobs: %+v", jobs)
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer server.Close()

	c := newTestClient(server)
	_, err := c.LatestPipeline("gh/org/repo", "main")
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
```

**Step 3: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/ci/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/ci/client.go internal/ci/client_test.go
git commit -m "ci: add CircleCI v2 API HTTP client"
```

---

### Task 3: Project Slug Derivation

**Files:**
- Create: `internal/ci/slug.go`
- Test: `internal/ci/slug_test.go`

**Step 1: Write the slug parser**

Parses git remote URL to CircleCI project slug format `gh/org/repo`.

```go
package ci

import (
	"fmt"
	"strings"
)

// SlugFromRemote derives a CircleCI project slug from a git remote URL.
// Supports: git@github.com:org/repo.git, https://github.com/org/repo.git
func SlugFromRemote(remoteURL string) (string, error) {
	u := strings.TrimSpace(remoteURL)

	// SSH format: git@github.com:org/repo.git
	if strings.HasPrefix(u, "git@github.com:") {
		path := strings.TrimPrefix(u, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", fmt.Errorf("cannot parse SSH remote: %s", remoteURL)
		}
		return "gh/" + parts[0] + "/" + parts[1], nil
	}

	// HTTPS format: https://github.com/org/repo.git
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
```

**Step 2: Write tests**

```go
package ci

import "testing"

func TestSlugFromRemote(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		want    string
		wantErr bool
	}{
		{
			name:   "SSH format",
			remote: "git@github.com:augurysys/augury-node.git",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "HTTPS format",
			remote: "https://github.com/augurysys/augury-node.git",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "HTTPS without .git",
			remote: "https://github.com/augurysys/augury-node",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:   "SSH without .git",
			remote: "git@github.com:augurysys/augury-node",
			want:   "gh/augurysys/augury-node",
		},
		{
			name:    "unsupported",
			remote:  "https://gitlab.com/org/repo.git",
			wantErr: true,
		},
		{
			name:    "empty",
			remote:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SlugFromRemote(tt.remote)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 3: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/ci/ -v -run TestSlugFromRemote`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/ci/slug.go internal/ci/slug_test.go
git commit -m "ci: add project slug derivation from git remote URL"
```

---

### Task 4: Config + Auth

**Files:**
- Modify: `internal/config/config.go` -- add `CircleToken` field
- Create: `internal/ci/auth.go` -- token resolution logic
- Test: `internal/ci/auth_test.go`

**Step 1: Add CircleToken to Config struct**

In `internal/config/config.go`, add field to the struct:

```go
type Config struct {
	AuguryNodeRoot   string   `toml:"augury_node_root"`
	BinaryInstalled  bool     `toml:"binary_installed"`
	NixVerified      bool     `toml:"nix_verified"`
	SetupCompletedAt string   `toml:"setup_completed_at"`
	CompletedSteps   []string `toml:"completed_steps"`
	SkippedSteps     []string `toml:"skipped_steps"`
	CircleToken      string   `toml:"circle_token,omitempty"`
}
```

**Step 2: Write auth.go**

```go
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
```

**Step 3: Write auth test**

```go
package ci

import (
	"os"
	"testing"
)

func TestResolveToken(t *testing.T) {
	t.Run("env var takes priority", func(t *testing.T) {
		os.Setenv("CIRCLE_TOKEN", "env-token")
		defer os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("config-token")
		if got != "env-token" {
			t.Errorf("got %q, want %q", got, "env-token")
		}
	})

	t.Run("falls back to config", func(t *testing.T) {
		os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("config-token")
		if got != "config-token" {
			t.Errorf("got %q, want %q", got, "config-token")
		}
	})

	t.Run("empty when neither set", func(t *testing.T) {
		os.Unsetenv("CIRCLE_TOKEN")

		got := ResolveToken("")
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
```

**Step 4: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/ci/ -v -run TestResolveToken`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/ci/auth.go internal/ci/auth_test.go
git commit -m "ci: add CircleCI token config field and auth resolution"
```

---

### Task 5: Tea Messages

**Files:**
- Create: `internal/ci/messages.go`

**Step 1: Write messages**

These are Bubbletea messages the CI model sends/receives asynchronously.

```go
package ci

type PipelineLoadedMsg struct {
	Pipeline *Pipeline
	Slug     string
}

type JobsLoadedMsg struct {
	Jobs []Job
}

type LogDownloadedMsg struct {
	JobName string
	Path    string
}

type CIErrorMsg struct {
	Err error
}
```

**Step 2: Commit**

```bash
git add internal/ci/messages.go
git commit -m "ci: add Bubbletea messages for CI screen"
```

---

### Task 6: CI Screen Model

**Files:**
- Create: `internal/ci/model.go`
- Test: `internal/ci/model_test.go`

This is the main Bubbletea model for the CI dashboard screen.

**Step 1: Write model.go**

```go
package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/components"
	"github.com/augurysys/augury-node-tui/internal/components/primitives"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screenState int

const (
	stateNoToken screenState = iota
	stateLoading
	stateReady
	stateDownloading
	stateViewing
	stateError
)

type Model struct {
	client    *Client
	slug      string
	branch    string
	repoRoot  string
	state     screenState
	pipeline  *Pipeline
	jobs      []Job
	errMsg    string
	jobsTable *components.DataTable
	logViewer *components.LogViewer
	viewingJob string
	Width     int
	Height    int
}

func NewModel(token, slug, branch, repoRoot string) *Model {
	m := &Model{
		slug:     slug,
		branch:   branch,
		repoRoot: repoRoot,
	}
	if token == "" {
		m.state = stateNoToken
		return m
	}
	m.client = NewClient(token)
	m.state = stateLoading
	return m
}

func (m *Model) Init() tea.Cmd {
	if m.state == stateNoToken {
		return nil
	}
	return m.fetchPipeline()
}

func (m *Model) fetchPipeline() tea.Cmd {
	return func() tea.Msg {
		p, err := m.client.LatestPipeline(m.slug, m.branch)
		if err != nil {
			return CIErrorMsg{Err: err}
		}
		return PipelineLoadedMsg{Pipeline: p, Slug: m.slug}
	}
}

func (m *Model) fetchJobs() tea.Cmd {
	return func() tea.Msg {
		wfs, err := m.client.Workflows(m.pipeline.ID)
		if err != nil {
			return CIErrorMsg{Err: err}
		}
		var allJobs []Job
		for _, wf := range wfs {
			jobs, err := m.client.Jobs(wf.ID)
			if err != nil {
				return CIErrorMsg{Err: err}
			}
			allJobs = append(allJobs, jobs...)
		}
		return JobsLoadedMsg{Jobs: allJobs}
	}
}

func (m *Model) downloadLog(job Job) tea.Cmd {
	return func() tea.Msg {
		artifacts, err := m.client.Artifacts(m.slug, job.JobNumber)
		if err != nil {
			return CIErrorMsg{Err: fmt.Errorf("list artifacts: %w", err)}
		}

		var logArtifact *Artifact
		for _, a := range artifacts {
			if strings.HasPrefix(a.Path, "logs/") && strings.HasSuffix(a.Path, ".log") {
				logArtifact = &a
				break
			}
		}
		if logArtifact == nil {
			return CIErrorMsg{Err: fmt.Errorf("no log artifact found for %s", job.Name)}
		}

		data, err := m.client.DownloadArtifact(logArtifact.URL)
		if err != nil {
			return CIErrorMsg{Err: fmt.Errorf("download log: %w", err)}
		}

		logDir := filepath.Join(m.repoRoot, "tmp", "augury-node-tui", "ci-logs")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return CIErrorMsg{Err: fmt.Errorf("create log dir: %w", err)}
		}
		logPath := filepath.Join(logDir, job.Name+".log")
		if err := os.WriteFile(logPath, data, 0644); err != nil {
			return CIErrorMsg{Err: fmt.Errorf("write log: %w", err)}
		}

		return LogDownloadedMsg{JobName: job.Name, Path: logPath}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.jobsTable != nil {
			m.jobsTable.SetWidth(msg.Width)
			m.jobsTable.SetHeight(msg.Height - 12)
		}
		if m.logViewer != nil {
			m.logViewer.SetWidth(msg.Width)
			m.logViewer.SetHeight(msg.Height - 4)
		}
		return m, nil

	case PipelineLoadedMsg:
		m.pipeline = msg.Pipeline
		return m, m.fetchJobs()

	case JobsLoadedMsg:
		m.jobs = msg.Jobs
		m.initJobsTable()
		m.state = stateReady
		return m, nil

	case LogDownloadedMsg:
		content, err := os.ReadFile(msg.Path)
		if err != nil {
			m.state = stateError
			m.errMsg = fmt.Sprintf("read log: %v", err)
			return m, nil
		}
		m.logViewer = components.NewLogViewer(string(content))
		m.logViewer.SetWidth(m.Width)
		m.logViewer.SetHeight(m.Height - 4)
		m.viewingJob = msg.JobName
		m.state = stateViewing
		return m, m.logViewer.Init()

	case CIErrorMsg:
		m.state = stateError
		m.errMsg = msg.Err.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.state == stateViewing && m.logViewer != nil {
		cmd := m.logViewer.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	s := msg.String()

	switch m.state {
	case stateViewing:
		if s == "esc" || s == "b" {
			m.state = stateReady
			m.logViewer = nil
			m.viewingJob = ""
			return m, nil
		}
		if m.logViewer != nil {
			cmd := m.logViewer.Update(msg)
			return m, cmd
		}

	case stateReady:
		switch s {
		case "r":
			m.state = stateLoading
			return m, m.fetchPipeline()
		case "enter":
			if m.jobsTable != nil {
				row := m.jobsTable.SelectedRow()
				if row != nil {
					entry := row.(jobEntry)
					for _, j := range m.jobs {
						if j.Name == entry.Name {
							m.state = stateDownloading
							return m, m.downloadLog(j)
						}
					}
				}
			}
		case "j", "down", "k", "up":
			if m.jobsTable != nil {
				m.jobsTable.Update(msg)
			}
			return m, nil
		}

	case stateError:
		if s == "r" {
			m.state = stateLoading
			return m, m.fetchPipeline()
		}
	}

	return m, nil
}

type jobEntry struct {
	Name     string
	Status   string
	Duration string
	HasLogs  string
}

func (m *Model) initJobsTable() {
	columns := []components.Column{
		{Header: "Job", Width: 30, Sortable: true, Renderer: func(row interface{}) string {
			return row.(jobEntry).Name
		}},
		{Header: "Status", Width: 12, Sortable: true, Renderer: func(row interface{}) string {
			e := row.(jobEntry)
			st := primitives.StatusSuccess
			switch e.Status {
			case "success":
				st = primitives.StatusSuccess
			case "failed":
				st = primitives.StatusError
			case "running":
				st = primitives.StatusRunning
			case "not_run":
				st = primitives.StatusUnavailable
			default:
				st = primitives.StatusWarning
			}
			return primitives.StatusBadge{Label: e.Status, Status: st}.Render()
		}},
		{Header: "Duration", Width: 12, Sortable: true, Renderer: func(row interface{}) string {
			return row.(jobEntry).Duration
		}},
		{Header: "Logs", Width: 8, Sortable: false, Renderer: func(row interface{}) string {
			return row.(jobEntry).HasLogs
		}},
	}

	m.jobsTable = components.NewDataTable(columns)
	if m.Width > 0 {
		m.jobsTable.SetWidth(m.Width)
	}
	if m.Height > 0 {
		m.jobsTable.SetHeight(m.Height - 12)
	}

	rows := make([]interface{}, 0, len(m.jobs))
	for _, j := range m.jobs {
		rows = append(rows, jobEntry{
			Name:     j.Name,
			Status:   j.Status,
			Duration: formatDuration(j.Duration()),
			HasLogs:  "yes",
		})
	}
	m.jobsTable.SetRows(rows)
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func (m *Model) View() string {
	switch m.state {
	case stateNoToken:
		return m.viewNoToken()
	case stateLoading:
		return m.viewLoading()
	case stateReady:
		return m.viewReady()
	case stateDownloading:
		return m.viewDownloading()
	case stateViewing:
		return m.viewLogs()
	case stateError:
		return m.viewError()
	default:
		return ""
	}
}

func (m *Model) viewNoToken() string {
	title := styles.Title.Render("CI Pipeline")
	msg := styles.Warning.Render(
		"No CircleCI token configured.\n\n" +
			"Set CIRCLE_TOKEN environment variable or run setup wizard.")
	keys := styles.KeyHelp.Render(styles.KeyBinding("esc", "back"))
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg, "", keys)
}

func (m *Model) viewLoading() string {
	title := styles.Title.Render("CI Pipeline")
	msg := styles.Dim.Render("Loading pipeline for " + m.branch + "...")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg)
}

func (m *Model) viewReady() string {
	var sections []string

	title := styles.Title.Render("CI Pipeline")
	sections = append(sections, title)

	if m.pipeline != nil {
		header := primitives.Card{
			Title: fmt.Sprintf("Pipeline #%d", m.pipeline.Number),
			Content: fmt.Sprintf("Branch: %s  SHA: %s  Status: %s",
				m.branch,
				shortSHA(m.pipeline.VCS.Revision),
				m.pipeline.State),
			Style: primitives.CardEmphasized,
		}
		w := m.Width
		if w <= 0 {
			w = 80
		}
		sections = append(sections, header.Render(w))
	}

	if m.jobsTable != nil {
		sections = append(sections, m.jobsTable.View())
	}

	keys := []string{
		styles.KeyBinding("enter", "view logs"),
		styles.KeyBinding("r", "refresh"),
		styles.KeyBinding("esc", "back"),
	}
	sections = append(sections, styles.KeyHelp.Render(strings.Join(keys, "  ")))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) viewDownloading() string {
	title := styles.Title.Render("CI Pipeline")
	msg := styles.Dim.Render("Downloading log...")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg)
}

func (m *Model) viewLogs() string {
	title := styles.Title.Render(fmt.Sprintf("Log: %s", m.viewingJob))
	var content string
	if m.logViewer != nil {
		content = m.logViewer.View()
	}
	keys := styles.KeyHelp.Render(
		styles.KeyBinding("esc", "back") + "  " +
			styles.KeyBinding("e", "first error") + "  " +
			styles.KeyBinding("n/N", "next/prev error"))
	return lipgloss.JoinVertical(lipgloss.Left, title, content, keys)
}

func (m *Model) viewError() string {
	title := styles.Title.Render("CI Pipeline")
	msg := styles.Error.Render("Error: " + m.errMsg)
	keys := styles.KeyHelp.Render(
		styles.KeyBinding("r", "retry") + "  " +
			styles.KeyBinding("esc", "back"))
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg, "", keys)
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
```

**Step 2: Write model test**

```go
package ci

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelNoToken(t *testing.T) {
	m := NewModel("", "gh/org/repo", "main", "/tmp/root")
	if m.state != stateNoToken {
		t.Errorf("state = %v, want stateNoToken", m.state)
	}
}

func TestNewModelWithToken(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	if m.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", m.state)
	}
	if m.client == nil {
		t.Error("client should not be nil")
	}
}

func TestUpdatePipelineLoaded(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	p := &Pipeline{ID: "p1", Number: 10, State: "success"}

	result, cmd := m.Update(PipelineLoadedMsg{Pipeline: p, Slug: "gh/org/repo"})
	model := result.(*Model)
	if model.pipeline == nil || model.pipeline.ID != "p1" {
		t.Error("pipeline should be set")
	}
	if cmd == nil {
		t.Error("should return cmd to fetch jobs")
	}
}

func TestUpdateJobsLoaded(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.Width = 120
	m.Height = 40

	jobs := []Job{
		{ID: "j1", Name: "lint", Status: "success", JobNumber: 1},
		{ID: "j2", Name: "build", Status: "failed", JobNumber: 2},
	}

	result, _ := m.Update(JobsLoadedMsg{Jobs: jobs})
	model := result.(*Model)
	if model.state != stateReady {
		t.Errorf("state = %v, want stateReady", model.state)
	}
	if len(model.jobs) != 2 {
		t.Errorf("jobs count = %d, want 2", len(model.jobs))
	}
	if model.jobsTable == nil {
		t.Error("jobsTable should be initialized")
	}
}

func TestUpdateCIError(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")

	result, _ := m.Update(CIErrorMsg{Err: fmt.Errorf("test error")})
	model := result.(*Model)
	if model.state != stateError {
		t.Errorf("state = %v, want stateError", model.state)
	}
	if model.errMsg != "test error" {
		t.Errorf("errMsg = %q, want %q", model.errMsg, "test error")
	}
}

func TestKeyRefreshInReady(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.state = stateReady

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := result.(*Model)
	if model.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", model.state)
	}
	if cmd == nil {
		t.Error("should return fetchPipeline cmd")
	}
}

func TestKeyRetryInError(t *testing.T) {
	m := NewModel("tok", "gh/org/repo", "main", "/tmp/root")
	m.state = stateError

	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := result.(*Model)
	if model.state != stateLoading {
		t.Errorf("state = %v, want stateLoading", model.state)
	}
	if cmd == nil {
		t.Error("should return fetchPipeline cmd")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "-"},
		{90 * time.Second, "1:30"},
		{5*time.Minute + 22*time.Second, "5:22"},
		{65 * time.Minute, "65:00"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.input)
		if got != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
```

**Step 3: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/ci/ -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/ci/model.go internal/ci/model_test.go
git commit -m "ci: add CI dashboard Bubbletea model"
```

---

### Task 7: Wire into App Router

**Files:**
- Modify: `internal/app/model.go` -- add `ci` route
- Modify: `internal/home/model.go` -- add `p` keybind

**Step 1: Add git remote lookup to status package**

Add to `internal/status/repo_status.go`:

```go
// RemoteURL returns the git remote URL for the given remote name.
func RemoteURL(root, remote string) string {
	out, err := gitOutput(root, "remote", "get-url", remote)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}
```

**Step 2: Modify `internal/app/model.go`**

Add import for `ci` package. Add `ci *ci.Model` field to Model struct.
Add CI model initialization in `newModel()`. Add `"ci"` route to Update() and View().

In the import block, add:
```go
"github.com/augurysys/augury-node-tui/internal/ci"
"github.com/augurysys/augury-node-tui/internal/config"
```

In the Model struct, add:
```go
ci *ci.Model
```

In `newModel()`, after the existing model creation, add:
```go
var circleToken string
if cfgPath, err := config.DefaultPath(); err == nil {
    if cfg, err := config.Read(cfgPath); err == nil {
        circleToken = cfg.CircleToken
    }
}
token := ci.ResolveToken(circleToken)
remoteURL := status.RemoteURL(st.Root, "origin")
slug, _ := ci.SlugFromRemote(remoteURL)
ciModel := ci.NewModel(token, slug, st.Branch, st.Root)
```

Add `ci: ciModel` to the return struct.

In `Update()`, add `"ci"` to the escape-back condition on line 78:
```go
if m.route != "splash" && m.route != "home" && (s == "b" || s == "esc") {
```

In the route switch at line 125, add:
```go
case "ci":
    child, cmd := m.ci.Update(msg)
    m.ci = child.(*ci.Model)
    return m, cmd
```

In `View()`, add:
```go
case "ci":
    return m.ci.View()
```

Add `"ci"` to the `r` refresh route list on line 84.

In the `WindowSizeMsg` handler, propagate to CI model:
```go
if cm, _ := m.ci.Update(msg); cm != nil {
    m.ci = cm.(*ci.Model)
}
```

**Step 3: Modify `internal/home/model.go`**

Add `p` keybind in the `Update()` KeyMsg switch (after the `o` case):
```go
case "p":
    return m, func() tea.Msg { return nav.NavigateMsg{Route: "ci"} }
```

Add `p` to the key help in `renderKeyHelp()`:
```go
styles.KeyBinding("p", "pipeline"),
```

**Step 4: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./... -v`
Expected: PASS (all existing + new tests)

**Step 5: Commit**

```bash
git add internal/status/repo_status.go internal/app/model.go internal/home/model.go
git commit -m "ci: wire CI dashboard screen into app router"
```

---

### Task 8: Setup Wizard Step (Optional)

**Files:**
- Create: `internal/setup/step_circleci.go`
- Create: `internal/setup/step_circleci_test.go`
- Modify: `internal/setup/wizard.go` -- add step

This is the optional CircleCI token step in the setup wizard.

**Step 1: Write step_circleci.go**

Follow the existing step pattern (see `step_root.go`, `step_nix.go`). The step shows a text input for the token. User can Enter to submit or press Tab/Enter with empty input to skip.

**Step 2: Write step_circleci_test.go**

Test that:
- Empty input results in skipped step
- Non-empty input is stored
- Step view renders correctly

**Step 3: Wire into wizard.go**

Add `stepCircleCI` field to WizardModel. Add step 5 (shift success to step 6). Update `View()` title to show "Step N/7". Update `advanceStep()`, `persistConfigAtStep()`, `updateCurrentStep()`.

**Step 4: Run tests**

Run: `cd /home/ngurfinkel/Repos/augury-node-tui && go test ./internal/setup/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/setup/step_circleci.go internal/setup/step_circleci_test.go internal/setup/wizard.go
git commit -m "setup: add optional CircleCI token wizard step"
```

---

## Summary

| Task | What | Files | Dependencies |
|------|------|-------|-------------|
| 1 | API types | types.go, types_test.go | none |
| 2 | HTTP client | client.go, client_test.go | Task 1 |
| 3 | Slug derivation | slug.go, slug_test.go | none |
| 4 | Config + auth | config.go, auth.go, auth_test.go | none |
| 5 | Tea messages | messages.go | Task 1 |
| 6 | CI screen model | model.go, model_test.go | Tasks 1-5 |
| 7 | App router wiring | app/model.go, home/model.go, repo_status.go | Task 6 |
| 8 | Setup wizard step | step_circleci.go, wizard.go | Task 4 |

Tasks 1, 3, 4 can run in parallel. Task 2 depends on 1. Task 5 depends on 1.
Task 6 depends on 1-5. Task 7 depends on 6. Task 8 depends on 4.
