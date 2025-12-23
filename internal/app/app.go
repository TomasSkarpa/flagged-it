package app

import (
	"flagged-it/internal/games/facts"
	"flagged-it/internal/games/flag"
	"flagged-it/internal/games/guessing"
	"flagged-it/internal/games/hangman"
	"flagged-it/internal/games/higher_lower"
	"flagged-it/internal/games/list"
	"flagged-it/internal/games/shape"
	"flagged-it/internal/ui/screens"

	"fyne.io/fyne/v2"
)

type App struct {
	window    fyne.Window
	app       fyne.App
	dashboard *screens.Dashboard
}

func NewApp(window fyne.Window, app fyne.App) *App {
	return &App{
		window: window,
		app:    app,
	}
}

func (a *App) GetDashboard() *fyne.Container {
	a.dashboard = screens.NewDashboard(a.navigateToGame, a.navigateToDebug, a.window, a.app)
	return a.dashboard.GetContent()
}

func (a *App) navigateToGame(gameType string) {
	switch gameType {
	case "shape":
		game := shape.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "shape_asia":
		game := shape.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
		game.StartWithRegion("Asia")
	case "list":
		game := list.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "hangman":
		game := hangman.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
		a.window.Canvas().SetOnTypedKey(game.TypedKey)
	case "facts":
		game := facts.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "higher_lower":
		game := higher_lower.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "flag":
		game := flag.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "flag_europe":
		game := flag.NewGame(a.backToDashboard)
		game.SetRegion("Europe")
		a.window.SetContent(game.GetContent())
	case "guessing":
		game := guessing.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	}
}

func (a *App) backToDashboard() {
	a.window.SetContent(a.GetDashboard())
}

func (a *App) navigateToDebug() {
	debugScreen := screens.NewDebugScreen(a.backToDashboard, a.window)
	a.window.SetContent(debugScreen.GetContent())
}
