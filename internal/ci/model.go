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
	if slug == "" {
		m.state = stateError
		m.errMsg = "Cannot determine CircleCI project slug from git remote URL"
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
		logPath := filepath.Join(logDir, filepath.Base(job.Name)+".log")
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
			m.jobsTable.SetHeight(msg.Height - components.ReservedLayoutHeight)
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
		m.jobsTable.SetHeight(m.Height - components.ReservedLayoutHeight)
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

func (m *Model) buildActionKeys() []components.KeyBinding {
	var keys []components.KeyBinding

	switch m.state {
	case stateReady:
		keys = append(keys, components.KeyBinding{Key: "enter", Label: "view log"})
		keys = append(keys, components.KeyBinding{Key: "r", Label: "refresh"})
	case stateViewing:
		// LogViewer has its own keybindings
	case stateError:
		keys = append(keys, components.KeyBinding{Key: "r", Label: "retry"})
	}

	return keys
}

func (m *Model) buildContext() string {
	parts := []string{m.branch}

	if m.pipeline != nil {
		parts = append(parts, "workflow running")
	}

	return strings.Join(parts, "  •  ")
}

func (m *Model) View() string {
	// Keep full-screen LogViewer as-is
	if m.state == stateViewing && m.logViewer != nil {
		return m.logViewer.View()
	}

	layout := components.ScreenLayout{
		Breadcrumb: []string{"🚀 Home", "Pipeline"},
		Context:    m.buildContext(),
		Content:    m.renderContent(),
		ActionKeys: m.buildActionKeys(),
		NavKeys: []components.KeyBinding{
			{Key: "j/k", Label: "navigate"},
			{Key: "esc", Label: "back"},
			{Key: "q", Label: "quit"},
		},
		Width:  m.Width,
		Height: m.Height,
	}
	return layout.Render()
}

func (m *Model) renderContent() string {
	switch m.state {
	case stateNoToken:
		return m.viewNoToken()
	case stateLoading:
		return m.viewLoading()
	case stateReady:
		return m.viewReady()
	case stateDownloading:
		return m.viewDownloading()
	case stateError:
		return m.viewError()
	default:
		return ""
	}
}

func (m *Model) viewNoToken() string {
	return styles.Warning.Render(
		"No CircleCI token configured.\n\n" +
			"Set CIRCLE_TOKEN environment variable or run setup wizard.")
}

func (m *Model) viewLoading() string {
	return styles.Dim.Render("Loading pipeline for " + m.branch + "...")
}

func (m *Model) viewReady() string {
	var sections []string

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

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) viewDownloading() string {
	return styles.Dim.Render("Downloading log...")
}

func (m *Model) viewError() string {
	return styles.Error.Render("Error: " + m.errMsg)
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
