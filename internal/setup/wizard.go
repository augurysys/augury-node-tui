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
}

func NewWizard(reconfigure bool) *WizardModel {
	cwd, _ := os.Getwd()
	detected, _ := FindAuguryNodeRoot(cwd)
	binaryPath, _ := os.Executable()

	reconfiguring := reconfigure
	if !reconfiguring {
		if path, err := config.DefaultPath(); err == nil {
			if cfg, err := config.Read(path); err == nil && cfg.AuguryNodeRoot != "" {
				reconfiguring = true
			}
		}
	}

	return &WizardModel{
		currentStep:  0,
		stepRoot:     NewRootStep(detected),
		stepNix:      NewNixStep(),
		stepGroups:   NewGroupsStep(),
		stepInstall:  NewInstallStep(binaryPath),
		stepNixBuild: nil,
		stepCircleCI: NewCircleCIStep(),
		stepSuccess:  nil, // created when advancing to step 6 with skipped steps
		reconfiguring: reconfiguring,
	}
}

func (m *WizardModel) Init() tea.Cmd {
	return nil
}

func (m *WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		cmd = m.stepInstall.Init()
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
		var s *InstallStepModel
		s, cmd = m.stepInstall.Update(msg)
		m.stepInstall = s
	case 4:
		if m.stepNixBuild != nil {
			var s *BuildStepModel
			s, cmd = m.stepNixBuild.Update(msg)
			m.stepNixBuild = s
		}
	case 5:
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
		stepView = m.stepInstall.View()
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
	return strings.Join([]string{title, stepView}, "\n")
}

func (m *WizardModel) LaunchMainRequested() bool {
	return m.launchMain
}

// CurrentStep returns the 0-based index of the current wizard step (for testing).
func (m *WizardModel) CurrentStep() int {
	return m.currentStep
}
