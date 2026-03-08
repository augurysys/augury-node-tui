package ui

import (
	"time"

	"github.com/augurysys/augury-node-tui/assets"
	tea "github.com/charmbracelet/bubbletea"
)

type TimeoutMsg struct{}

type ReplayMsg struct{}

type SplashModel struct {
	Dismissed bool
	Timeout   time.Duration
	Width     int
	Height    int
}

func NewSplashModel(timeout time.Duration) *SplashModel {
	return &SplashModel{Timeout: timeout}
}

func (m *SplashModel) Init() tea.Cmd {
	return tea.Tick(m.Timeout, func(t time.Time) tea.Msg {
		return TimeoutMsg{}
	})
}

func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Dismissed && msg.String() == "r" {
			m.Dismissed = false
			return m, m.Init()
		}
		if !m.Dismissed {
			m.Dismissed = true
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	case TimeoutMsg:
		m.Dismissed = true
		return m, nil
	case ReplayMsg:
		m.Dismissed = false
		return m, m.Init()
	}
	return m, nil
}

func (m *SplashModel) View() string {
	if m.Dismissed {
		return ""
	}
	if m.Width < 40 || m.Height < 8 {
		return "augury-node-tui"
	}
	return assets.SplashTxt + "\n"
}
