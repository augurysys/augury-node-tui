package main

import (
	"fmt"
	"os"
	"time"

	"github.com/augurysys/augury-node-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.NewSplashModel(2 * time.Second)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
