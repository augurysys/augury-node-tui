package nav

type NavigateMsg struct {
	Route string
}

type ReplaySplashMsg struct{}

type NavigateBackMsg struct{}

// GoToFlash navigates to flash screen
type GoToFlash struct{}
