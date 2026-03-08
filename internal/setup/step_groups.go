package setup

import (
	"strings"

	"github.com/atotto/clipboard"
	"github.com/augurysys/augury-node-tui/internal/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type GroupsStepModel struct {
	inNixUsers bool
	state      string
	confirmed  bool
}

func NewGroupsStep() *GroupsStepModel {
	return &GroupsStepModel{
		state: "checking",
	}
}

func (m *GroupsStepModel) Init() tea.Cmd {
	return func() tea.Msg {
		result := CheckNixGroup()
		return GroupCheckMsg{InNixUsers: result.Available}
	}
}

func (m *GroupsStepModel) Update(msg tea.Msg) (*GroupsStepModel, tea.Cmd) {
	switch msg := msg.(type) {
	case GroupCheckMsg:
		m.inNixUsers = msg.InNixUsers
		m.state = "ready"

		if m.inNixUsers {
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		}
		return m, nil

	case tea.KeyMsg:
		if m.state != "ready" {
			return m, nil
		}

		switch msg.String() {
		case "r":
			m.state = "rechecking"
			return m, func() tea.Msg {
				result := CheckNixGroup()
				return GroupCheckMsg{InNixUsers: result.Available}
			}
		case "s":
			m.confirmed = true
			return m, func() tea.Msg { return NextStepMsg{} }
		case "c":
			return m, copyToClipboard("sudo usermod -aG nix-users $USER && newgrp nix-users")
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *GroupsStepModel) View() string {
	var lines []string

	header := styles.Header.Render("👥 Permissions Setup")
	lines = append(lines, header)
	lines = append(lines, "")

	if m.state == "checking" || m.state == "rechecking" {
		lines = append(lines, "  "+styles.Dim.Render("Checking group membership..."))
		return styles.Border.Render(strings.Join(lines, "\n"))
	}

	if m.inNixUsers {
		lines = append(lines, "  "+styles.Success.Render("✓ In nix-users group"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("Permissions OK"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("enter", "Continue"))
	} else {
		lines = append(lines, "  "+styles.Error.Render("✗ Not in nix-users group"))
		lines = append(lines, "")
		lines = append(lines, "  "+styles.Dim.Render("Run this command in a new terminal:"))
		lines = append(lines, "")

		cmd := styles.Info.Render("  sudo usermod -aG nix-users $USER")
		lines = append(lines, "  "+cmd)
		cmd2 := styles.Info.Render("  newgrp nix-users")
		lines = append(lines, "  "+cmd2)

		lines = append(lines, "")
		lines = append(lines, "  "+styles.KeyBinding("c", "Copy")+"  "+styles.KeyBinding("r", "Re-check")+"  "+styles.KeyBinding("s", "Skip"))
	}

	content := strings.Join(lines, "\n")
	return styles.Border.Render(content)
}

func (m *GroupsStepModel) Confirmed() bool {
	return m.confirmed
}

type GroupCheckMsg struct {
	InNixUsers bool
}

func copyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		return ClipboardCopiedMsg{Success: err == nil, Text: text}
	}
}

type ClipboardCopiedMsg struct {
	Success bool
	Text    string
}
