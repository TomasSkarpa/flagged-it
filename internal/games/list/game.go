package list

import (
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fmt"
	"image/color"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
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
	scoreLabel        *widget.Label
	scoreManager      *utils.ScoreManager
}

func NewGame(backFunc func(), scoreManager *utils.ScoreManager) *Game {
	g := &Game{
		backFunc:     backFunc,
		scoreManager: scoreManager,
	}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar(lang.X("game.list.title", "List All Countries"), g.backFunc, g.Reset)

	g.setupSelectionView()
	g.setupGameView()

	headerSection := container.NewVBox(
		topBar.GetContainer(),
		components.NewDashedSeparator(color.RGBA{200, 200, 200, 255}, 5),
	)

	g.mainContent = container.NewMax(g.selectionView)

	g.content = container.NewBorder(
		headerSection, nil, nil, nil,
		g.mainContent,
	)
}

func (g *Game) setupSelectionView() {
	availableRegions := g.getAvailableRegions()
	regionSelector := components.NewRegionSelector(
		lang.X("game.list.select_region", "Select Region"),
		lang.X("game.list.choose_region", "Choose a region and try to name all countries in it!"),
		availableRegions,
		g.startGame,
	)
	g.selectionView = regionSelector.GetContainer()
}

func (g *Game) getAvailableRegions() []string {
	countries := data.LoadCountries()
	regionMap := make(map[string]bool)
	regionMap["World"] = true

	for _, country := range countries {
		if country.Region != "" {
			regionMap[country.Region] = true
		}
	}

	var regions []string
	for region := range regionMap {
		regions = append(regions, region)
	}

	// Sort with World first
	sort.Slice(regions, func(i, j int) bool {
		if regions[i] == "World" {
			return true
		}
		if regions[j] == "World" {
			return false
		}
		return regions[i] < regions[j]
	})

	return regions
}

func (g *Game) setupGameView() {
	g.progressLabel = widget.NewLabel("")
	g.statusLabel = widget.NewLabel("")
	g.scoreLabel = widget.NewLabel("")

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder(lang.X("game.list.enter_country", "Enter country name..."))
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	guessBtn := components.NewButton(lang.X("game.list.guess", "Guess"), g.makeGuess)

	g.countryList = widget.NewList(
		func() int { return len(g.allCountries) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			country := g.allCountries[id]
			if g.guessedCountries[strings.ToLower(country.Name.Common)] {
				label.SetText(fmt.Sprintf(lang.X("game.list.country_item", "%d. %s"), id+1, country.Name.Common))
			} else {
				label.SetText(fmt.Sprintf(lang.X("game.list.country_unknown", "%d. ?"), id+1))
			}
		},
	)

	guessContainer := container.NewBorder(
		nil, nil,
		guessBtn, nil,
		g.guessEntry,
	)

	topSection := container.NewVBox(
		g.scoreLabel,
		g.progressLabel,
		g.statusLabel,
		components.NewDashedSeparator(color.RGBA{200, 200, 200, 255}, 5),
		guessContainer,
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
	g.statusLabel.SetText(lang.X("game.list.start_guessing", "Start guessing countries!"))
	g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.list.completion", "Completion: %.0f%%"), 0.0))
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
		if utils.MatchCountry(guess, country, utils.MatchCommon|utils.MatchOfficial) && !g.guessedCountries[strings.ToLower(country.Name.Common)] {
			g.guessedCountries[strings.ToLower(country.Name.Common)] = true
			matchedCountry = country
			found = true
			break
		}
	}

	if found {
		g.statusLabel.SetText(fmt.Sprintf(lang.X("game.list.correct_added", "Correct! %s added to the list."), matchedCountry.Name.Common))
		g.updateProgress()
		percent := float64(len(g.guessedCountries)) / float64(len(g.allCountries)) * 100
		g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.list.completion", "Completion: %.0f%%"), percent))
		if len(g.guessedCountries) == len(g.allCountries) {
			g.statusLabel.SetText(lang.X("game.list.congratulations", "Congratulations! You've listed all countries!"))
		}
	} else {
		g.statusLabel.SetText(lang.X("game.list.not_found", "Not found or already guessed. Try again!"))
	}

	g.guessEntry.SetText("")
	g.countryList.Refresh()
}

func (g *Game) updateProgress() {
	translatedRegion := utils.TranslateRegion(g.selectedContinent)
	g.progressLabel.SetText(fmt.Sprintf(lang.X("game.list.progress", "%s: %d/%d countries found"), translatedRegion, len(g.guessedCountries), len(g.allCountries)))
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
