package list

import (
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content           *fyne.Container
	backFunc          func()
	selectionView     *fyne.Container
	gameView          *fyne.Container
	mainContent       *fyne.Container
	selectedContinent string
	allCountries      []models.Country
	guessedCountries  map[string]bool
	guessEntry        *widget.Entry
	progressLabel     *widget.Label
	countryList       *widget.List
	statusLabel       *widget.Label
}

func NewGame(backFunc func()) *Game {
	g := &Game{backFunc: backFunc}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("List All Countries", g.backFunc, g.Reset)

	g.setupSelectionView()
	g.setupGameView()

	headerSection := container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
	)

	g.mainContent = container.NewMax(g.selectionView)

	g.content = container.NewBorder(
		headerSection, nil, nil, nil,
		g.mainContent,
	)
}

func (g *Game) setupSelectionView() {
	regionSelector := components.NewRegionSelector(
		"Select Region",
		"Choose a region and try to name all countries in it!",
		g.startGame,
	)
	g.selectionView = regionSelector.GetContainer()
}

func (g *Game) setupGameView() {
	g.progressLabel = widget.NewLabel("")
	g.statusLabel = widget.NewLabel("")

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	guessBtn := widget.NewButton("Guess", g.makeGuess)

	g.countryList = widget.NewList(
		func() int { return len(g.allCountries) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			country := g.allCountries[id]
			if g.guessedCountries[strings.ToLower(country.Name.Common)] {
				label.SetText(fmt.Sprintf("%d. %s", id+1, country.Name.Common))
			} else {
				label.SetText(fmt.Sprintf("%d. ?", id+1))
			}
		},
	)

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, guessBtn)

	topSection := container.NewVBox(
		g.progressLabel,
		g.statusLabel,
		widget.NewSeparator(),
		guessContainer,
		widget.NewSeparator(),
	)

	g.gameView = container.NewBorder(
		topSection, nil, nil, nil,
		g.countryList,
	)
}

func (g *Game) startGame(continent string) {
	g.selectedContinent = continent
	countries := data.LoadCountries()

	g.allCountries = []models.Country{}
	for _, country := range countries {
		if continent == "World" || country.Region == continent {
			g.allCountries = append(g.allCountries, country)
		}
	}

	sort.Slice(g.allCountries, func(i, j int) bool {
		return g.allCountries[i].Name.Common < g.allCountries[j].Name.Common
	})
	g.guessedCountries = make(map[string]bool)

	g.updateProgress()
	g.statusLabel.SetText("Start guessing countries!")
	g.guessEntry.SetText("")

	g.mainContent.RemoveAll()
	g.mainContent.Add(g.gameView)
	g.mainContent.Refresh()
	g.countryList.Refresh()
}

func (g *Game) makeGuess() {
	guess := strings.TrimSpace(g.guessEntry.Text)
	if guess == "" {
		return
	}

	found := false
	var matchedCountry models.Country

	for _, country := range g.allCountries {
		if utils.MatchesCountry(guess, country) && !g.guessedCountries[strings.ToLower(country.Name.Common)] {
			g.guessedCountries[strings.ToLower(country.Name.Common)] = true
			matchedCountry = country
			found = true
			break
		}
	}

	if found {
		g.statusLabel.SetText(fmt.Sprintf("Correct! %s added to the list.", matchedCountry.Name.Common))
		g.updateProgress()
		if len(g.guessedCountries) == len(g.allCountries) {
			g.statusLabel.SetText("Congratulations! You've listed all countries!")
		}
	} else {
		g.statusLabel.SetText("Not found or already guessed. Try again!")
	}

	g.guessEntry.SetText("")
	g.countryList.Refresh()
}

func (g *Game) updateProgress() {
	g.progressLabel.SetText(fmt.Sprintf("%s: %d/%d countries found", g.selectedContinent, len(g.guessedCountries), len(g.allCountries)))
}

func (g *Game) showSelection() {
	g.mainContent.RemoveAll()
	g.mainContent.Add(g.selectionView)
	g.mainContent.Refresh()
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	g.showSelection()
}

func (g *Game) Reset() {
	g.showSelection()
}
