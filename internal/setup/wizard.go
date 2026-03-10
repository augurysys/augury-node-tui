package setup

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/augurysys/augury-node-tui/internal/config"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type LaunchMainTUIMsg struct{}

type WizardModel struct {
	currentStep   int
	stepRoot      *RootStep
	stepNix       *NixStepModel
	stepGroups    *GroupsStepModel
	stepInstall   *InstallStepModel
	stepNixBuild  *BuildStepModel
	stepCircleCI  *CircleCIStepModel
	stepSuccess   *SuccessStepModel
	config        config.Config
	launchMain    bool
	reconfiguring bool
	width         int
	height        int
}

func NewWizard(reconfigure bool) *WizardModel {
	cwd, _ := os.Getwd()
	detected, _ := FindAuguryNodeRoot(cwd)

	reconfiguring := reconfigure
	var existingCfg config.Config
	if path, err := config.DefaultPath(); err == nil {
		if cfg, err := config.Read(path); err == nil {
			existingCfg = cfg
			if !reconfiguring && cfg.AuguryNodeRoot != "" {
				reconfiguring = true
			}
		}
	}

	return &WizardModel{
		config:        existingCfg,
		currentStep:   0,
		stepRoot:      NewRootStep(detected),
		stepNix:       NewNixStep(),
		stepGroups:    NewGroupsStep(),
		stepInstall:   nil,
		stepNixBuild:  nil,
		stepCircleCI:  NewCircleCIStep(),
		stepSuccess:   nil,
		reconfiguring: reconfiguring,
		width:         80,
		height:        24,
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return nil
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case NextStepMsg:
		return m.advanceStep()
	case LaunchMainTUIMsg:
		m.launchMain = true
		return m, tea.Quit
	case RootConfirmedMsg:
		if m.currentStep == 0 {
			m.config.AuguryNodeRoot = msg.Path
			m.config.CompletedSteps = append(m.config.CompletedSteps, "root")
			m.saveConfig()
			m.currentStep = 1
			return m, m.stepNix.Init()
		}
	}
	return m.updateCurrentStep(msg)
}

func (m *WizardModel) advanceStep() (tea.Model, tea.Cmd) {
	m.persistConfigAtStep()
	m.currentStep++
	var cmd tea.Cmd
	switch m.currentStep {
	case 1:
		cmd = m.stepNix.Init()
	case 2:
		cmd = m.stepGroups.Init()
	case 3:
		if m.stepInstall == nil && m.config.AuguryNodeRoot != "" {
			m.stepInstall = NewInstallStep(m.config.AuguryNodeRoot)
		}
		if m.stepInstall != nil {
			cmd = m.stepInstall.Init()
		}
	case 4:
		if m.stepNixBuild == nil && m.config.AuguryNodeRoot != "" {
			m.stepNixBuild = NewBuildStep(m.config.AuguryNodeRoot)
		}
		if m.stepNixBuild != nil {
			cmd = m.stepNixBuild.Init()
		}
	case 5:
		cmd = m.stepCircleCI.Init()
	case 6:
		m.stepSuccess = NewSuccessStep(m.config.SkippedSteps)
		cmd = m.stepSuccess.Init()
	}
	return m, cmd
}

func (m *WizardModel) persistConfigAtStep() {
	switch m.currentStep {
	case 0:
		if m.stepRoot.Confirmed() {
			m.config.AuguryNodeRoot = m.stepRoot.GetRootPath()
			m.config.CompletedSteps = append(m.config.CompletedSteps, "root")
		}
	case 1:
		if m.stepNix.Confirmed() {
			m.config.NixVerified = m.stepNix.AllChecksPassed()
			m.config.CompletedSteps = append(m.config.CompletedSteps, "nix")
			if !m.stepNix.AllChecksPassed() {
				m.config.SkippedSteps = append(m.config.SkippedSteps, "nix")
			}
		}
	case 2:
		if m.stepGroups.Confirmed() {
			m.config.CompletedSteps = append(m.config.CompletedSteps, "groups")
			if !m.stepGroups.inNixUsers {
				m.config.SkippedSteps = append(m.config.SkippedSteps, "groups")
			}
		}
	case 3:
		if m.stepInstall.Confirmed() {
			m.config.BinaryInstalled = m.stepInstall.alreadyInstalled
			m.config.CompletedSteps = append(m.config.CompletedSteps, "install")
			if !m.stepInstall.alreadyInstalled {
				m.config.SkippedSteps = append(m.config.SkippedSteps, "install")
			}
			if m.stepNixBuild == nil && m.config.AuguryNodeRoot != "" {
				m.stepNixBuild = NewBuildStep(m.config.AuguryNodeRoot)
			}
		}
	case 4:
		if m.stepNixBuild != nil && m.stepNixBuild.Confirmed() {
			m.config.CompletedSteps = append(m.config.CompletedSteps, "nixbuild")
		}
	case 5:
		if m.stepCircleCI != nil && m.stepCircleCI.Confirmed() {
			m.config.CompletedSteps = append(m.config.CompletedSteps, "circleci")
			if m.stepCircleCI.Skipped() {
				m.config.SkippedSteps = append(m.config.SkippedSteps, "circleci")
			} else {
				m.config.CircleToken = m.stepCircleCI.Token()
			}
			m.config.SetupCompletedAt = time.Now().Format(time.RFC3339)
		}
	}
	m.saveConfig()
}

func (m *WizardModel) saveConfig() {
	path, err := config.DefaultPath()
	if err == nil {
		_ = config.Write(path, m.config)
	}
}

func (m *WizardModel) updateCurrentStep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentStep {
	case 0:
		var s *RootStep
		s, cmd = m.stepRoot.Update(msg)
		m.stepRoot = s
	case 1:
		var s *NixStepModel
		s, cmd = m.stepNix.Update(msg)
		m.stepNix = s
	case 2:
		var s *GroupsStepModel
		s, cmd = m.stepGroups.Update(msg)
		m.stepGroups = s
	case 3:
		if m.stepInstall != nil {
			var s *InstallStepModel
			s, cmd = m.stepInstall.Update(msg)
			m.stepInstall = s
		}
	case 4:
		if m.stepNixBuild != nil {
			var s *BuildStepModel
			s, cmd = m.stepNixBuild.Update(msg)
			m.stepNixBuild = s
		}
	case 5:
		var s *CircleCIStepModel
		s, cmd = m.stepCircleCI.Update(msg)
		m.stepCircleCI = s
	case 6:
		var s *SuccessStepModel
		s, cmd = m.stepSuccess.Update(msg)
		m.stepSuccess = s
	}
	return m, cmd
}

func (m *WizardModel) View() string {
	titleStr := fmt.Sprintf("Step %d/7", m.currentStep+1)
	if m.reconfiguring {
		titleStr = "Reconfiguring... " + titleStr
	}
	title := styles.Title.Render(titleStr)
	var stepView string
	switch m.currentStep {
	case 0:
		stepView = m.stepRoot.View()
	case 1:
		stepView = m.stepNix.View()
	case 2:
		stepView = m.stepGroups.View()
	case 3:
		if m.stepInstall != nil {
			stepView = m.stepInstall.View()
		} else {
			stepView = ""
		}
	case 4:
		if m.stepNixBuild != nil {
			stepView = m.stepNixBuild.View()
		} else {
			stepView = ""
		}
	case 5:
		stepView = m.stepCircleCI.View()
	case 6:
		stepView = m.stepSuccess.View()
	default:
		stepView = ""
	}
	
	helpPanel := m.renderHelpPanel()
	
	return strings.Join([]string{title, stepView, helpPanel}, "\n")
}

func (m *WizardModel) renderHelpPanel() string {
	var keys []string
	
	switch m.currentStep {
	case 0:
		keys = []string{
			styles.KeyBinding("tab", "complete"),
			styles.KeyBinding("enter", "confirm"),
			styles.KeyBinding("q", "quit"),
		}
	case 1:
		if m.stepNix != nil && m.stepNix.state == "unhealthy" {
			keys = []string{
				styles.KeyBinding("f", "fix"),
				styles.KeyBinding("s", "skip"),
				styles.KeyBinding("q", "quit"),
			}
		} else {
			keys = []string{
				styles.KeyBinding("enter", "continue"),
				styles.KeyBinding("q", "quit"),
			}
		}
	case 2:
		if m.stepGroups != nil && !m.stepGroups.inNixUsers {
			keys = []string{
				styles.KeyBinding("c", "copy cmd"),
				styles.KeyBinding("r", "recheck"),
				styles.KeyBinding("s", "skip"),
				styles.KeyBinding("q", "quit"),
			}
		} else {
			keys = []string{
				styles.KeyBinding("enter", "continue"),
				styles.KeyBinding("q", "quit"),
			}
		}
	case 3:
		if m.stepInstall != nil && m.stepInstall.state == "ready" && !m.stepInstall.alreadyInstalled {
			keys = []string{
				styles.KeyBinding("i", "auto-install"),
				styles.KeyBinding("c", "copy cmd"),
				styles.KeyBinding("s", "skip"),
				styles.KeyBinding("q", "quit"),
			}
		} else {
			keys = []string{
				styles.KeyBinding("enter", "continue"),
				styles.KeyBinding("q", "quit"),
			}
		}
	case 4:
		if m.stepNixBuild != nil && m.stepNixBuild.state == "building" {
			keys = []string{
				styles.KeyBinding("q", "cancel"),
			}
		} else if m.stepNixBuild != nil && m.stepNixBuild.state == "failed" {
			keys = []string{
				styles.KeyBinding("r", "retry"),
				styles.KeyBinding("s", "skip"),
				styles.KeyBinding("q", "quit"),
			}
		} else {
			keys = []string{
				styles.KeyBinding("enter", "continue"),
				styles.KeyBinding("q", "quit"),
			}
		}
	case 5:
		keys = []string{
			styles.KeyBinding("enter", "confirm"),
			styles.KeyBinding("s", "skip"),
			styles.KeyBinding("q", "quit"),
		}
	case 6:
		keys = []string{
			styles.KeyBinding("enter", "launch"),
			styles.KeyBinding("q", "quit"),
		}
	default:
		keys = []string{
			styles.KeyBinding("q", "quit"),
		}
	}
	
	helpText := strings.Join(keys, "  ")
	
	separator := strings.Repeat("─", m.width)
	if m.width == 0 {
		separator = strings.Repeat("─", 80)
	}
	
	return "\n" + styles.Dim.Render(separator) + "\n" + styles.KeyHelp.Render(helpText)
}

func (m *WizardModel) LaunchMainRequested() bool {
	return m.launchMain
}

// CurrentStep returns the 0-based index of the current wizard step (for testing).
func (m *WizardModel) CurrentStep() int {
	return m.currentStep
}
