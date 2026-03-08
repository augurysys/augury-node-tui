package app

import (
	"time"

	"github.com/augurysys/augury-node-tui/internal/build"
	"github.com/augurysys/augury-node-tui/internal/caches"
	"github.com/augurysys/augury-node-tui/internal/engine"
	"github.com/augurysys/augury-node-tui/internal/home"
	"github.com/augurysys/augury-node-tui/internal/hints"
	"github.com/augurysys/augury-node-tui/internal/hydration"
	"github.com/augurysys/augury-node-tui/internal/nav"
	"github.com/augurysys/augury-node-tui/internal/platform"
	"github.com/augurysys/augury-node-tui/internal/status"
	"github.com/augurysys/augury-node-tui/internal/ui"
	"github.com/augurysys/augury-node-tui/internal/validations"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	route       string
	nixState    engine.NixState
	splash      *ui.SplashModel
	home        *home.Model
	build       *build.Model
	hydrate     *hydration.Model
	caches      *caches.Model
	validations *validations.Model
	hints       *hints.Model
}

func newModel(st status.RepoStatus, platforms []platform.Platform, splashTimeout time.Duration, nix engine.NixState) *Model {
	hm := home.NewModel(st, platforms)
	bm := build.NewModel(st, platforms, hm.Selected)
	hyd := hydration.NewModel(st, platforms, hm.Selected)
	c := caches.NewModel(st, platforms)
	v := validations.NewModel(st)
	h := hints.NewModel(st, platforms)
	bm.SetNixState(nix)
	c.SetNixState(nix)
	hyd.SetNixState(nix)
	v.SetNixState(nix)
	return &Model{
		route:       "splash",
		nixState:    nix,
		splash:      ui.NewSplashModel(splashTimeout),
		home:        hm,
		build:       bm,
		hydrate:     hyd,
		caches:      c,
		validations: v,
		hints:       h,
	}
}

func NewModel(st status.RepoStatus, platforms []platform.Platform, splashTimeout time.Duration) *Model {
	return newModel(st, platforms, splashTimeout, engine.ProbeNix(st.Root))
}

func NewModelWithNix(st status.RepoStatus, platforms []platform.Platform, splashTimeout time.Duration, nix engine.NixState) *Model {
	return newModel(st, platforms, splashTimeout, nix)
}

func (m *Model) Route() string {
	return m.route
}

func (m *Model) Init() tea.Cmd {
	return m.splash.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		if m.route != "splash" && m.route != "home" && (s == "b" || s == "esc") {
			if m.route != "caches" || !m.caches.ConfirmShown() {
				m.route = "home"
				return m, nil
			}
		}
		if s == "r" && (m.route == "home" || m.route == "build" || m.route == "caches" || m.route == "hydrate" || m.route == "validations") {
			m.nixState = engine.ProbeNix(m.caches.Status.Root)
			m.build.SetNixState(m.nixState)
			m.caches.SetNixState(m.nixState)
			m.hydrate.SetNixState(m.nixState)
			m.validations.SetNixState(m.nixState)
			return m, nil
		}
	case nav.NavigateMsg:
		m.route = msg.Route
		return m, nil
	case nav.ReplaySplashMsg:
		m.route = "splash"
		m.splash = ui.NewSplashModel(m.splash.Timeout)
		return m, m.splash.Init()
	case nav.NavigateBackMsg:
		m.route = "home"
		return m, nil
	case tea.QuitMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		// Propagate to diagram-enabled routes so diagrams render after route switch
		if hm, _ := m.home.Update(msg); hm != nil {
			m.home = hm.(*home.Model)
		}
		if cm, _ := m.caches.Update(msg); cm != nil {
			m.caches = cm.(*caches.Model)
		}
		if vm, _ := m.validations.Update(msg); vm != nil {
			m.validations = vm.(*validations.Model)
		}
	}

	switch m.route {
	case "splash":
		child, cmd := m.splash.Update(msg)
		m.splash = child.(*ui.SplashModel)
		if m.splash.Dismissed {
			m.route = "home"
		}
		return m, cmd
	case "home":
		child, cmd := m.home.Update(msg)
		m.home = child.(*home.Model)
		return m, cmd
	case "build":
		child, cmd := m.build.Update(msg)
		m.build = child.(*build.Model)
		return m, cmd
	case "hydrate":
		child, cmd := m.hydrate.Update(msg)
		m.hydrate = child.(*hydration.Model)
		return m, cmd
	case "caches":
		child, cmd := m.caches.Update(msg)
		m.caches = child.(*caches.Model)
		return m, cmd
	case "validations":
		child, cmd := m.validations.Update(msg)
		m.validations = child.(*validations.Model)
		return m, cmd
	case "hints":
		child, cmd := m.hints.Update(msg)
		m.hints = child.(*hints.Model)
		return m, cmd
	default:
		return m, nil
	}
}

func (m *Model) View() string {
	switch m.route {
	case "splash":
		return m.splash.View()
	case "home":
		return m.home.View()
	case "build":
		return m.build.View()
	case "hydrate":
		return m.hydrate.View()
	case "caches":
		return m.caches.View()
	case "validations":
		return m.validations.View()
	case "hints":
		return m.hints.View()
	default:
		return ""
	}
}
