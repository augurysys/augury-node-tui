package app

import (
	"time"

	"github.com/augurysys/augury-node-tui/internal/build"
	"github.com/augurysys/augury-node-tui/internal/caches"
	"github.com/augurysys/augury-node-tui/internal/ci"
	"github.com/augurysys/augury-node-tui/internal/config"
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
	ci          *ci.Model
}

func newModel(st status.RepoStatus, platforms []platform.Platform, splashTimeout time.Duration, nix engine.NixState) *Model {
	hm := home.NewModel(st, platforms)
	hm.SetNixState(nix)
	bm := build.NewModel(st, platforms, hm.Selected)
	hyd := hydration.NewModel(st, platforms, hm.Selected)
	c := caches.NewModel(st, platforms)
	v := validations.NewModel(st)
	h := hints.NewModel(st, platforms)
	h.SetNixState(nix)
	bm.SetNixState(nix)
	c.SetNixState(nix)
	hyd.SetNixState(nix)
	v.SetNixState(nix)
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
		ci:          ciModel,
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
			m.home.SetNixState(m.nixState)
			m.build.SetNixState(m.nixState)
			m.caches.SetNixState(m.nixState)
			m.hydrate.SetNixState(m.nixState)
			m.validations.SetNixState(m.nixState)
			m.hints.SetNixState(m.nixState)
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
		if hyd, _ := m.hydrate.Update(msg); hyd != nil {
			m.hydrate = hyd.(*hydration.Model)
		}
		if bm, _ := m.build.Update(msg); bm != nil {
			m.build = bm.(*build.Model)
		}
		if cim, _ := m.ci.Update(msg); cim != nil {
			m.ci = cim.(*ci.Model)
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
	case "ci":
		child, cmd := m.ci.Update(msg)
		m.ci = child.(*ci.Model)
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
	case "ci":
		return m.ci.View()
	default:
		return ""
	}
}
