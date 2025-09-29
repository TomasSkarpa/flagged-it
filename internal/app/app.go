package app

import (
	"flagged-it/internal/games/higher_lower"
	"flagged-it/internal/games/shapes"
	"flagged-it/internal/games/list"
	"flagged-it/internal/games/facts"
	"flagged-it/internal/games/hangman"
	"flagged-it/internal/ui/screens"
	"fyne.io/fyne/v2"
)

type App struct {
	window fyne.Window
}

func NewApp(window fyne.Window) *App {
	return &App{window: window}
}

func (a *App) GetDashboard() *fyne.Container {
	dashboard := screens.NewDashboard(a.navigateToGame)
	return dashboard.GetContent()
}

func (a *App) navigateToGame(gameType string) {
	switch gameType {
	case "country_guess":
		game := country_guess.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "country_list":
		game := country_list.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "hangman":
		game := hangman.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "fact_guess":
		game := fact_guess.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	case "higher_lower":
		game := higher_lower.NewGame(a.backToDashboard)
		a.window.SetContent(game.GetContent())
	}
}

func (a *App) backToDashboard() {
	a.window.SetContent(a.GetDashboard())
}
