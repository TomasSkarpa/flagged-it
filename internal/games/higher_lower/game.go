package higher_lower

import (
	"flagged-it/internal/data"
	"flagged-it/internal/ui/components"
	"fmt"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

type Game struct {
	content  *fyne.Container
	backFunc func()

	firstCountry  int
	secondCountry int
	score         int
	highestStreak int

	countryOneNameLabel *widget.Label
	countryTwoNameLabel *widget.Label
	countryOnePopLabel  *widget.Label
	countryTwoPopLabel  *widget.Label
	nextBtn             *components.Button
	higherBtn           *components.Button
	lowerBtn            *components.Button
	currentStreakLabel  *widget.Label
	highestStreakLabel  *widget.Label
}

func NewGame(backFunc func()) *Game {
	g := &Game{backFunc: backFunc}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar(lang.X("game.higher_lower.title", "Higher or Lower Game"), g.backFunc, g.Reset)

	gameDescription := widget.NewLabel(lang.X("game.higher_lower.description", "Try to guess which country has a higher population!"))

	g.countryOneNameLabel = widget.NewLabel("")
	g.countryOnePopLabel = widget.NewLabel("")
	g.countryTwoNameLabel = widget.NewLabel("")
	g.countryTwoPopLabel = widget.NewLabel("")

	// Current streak label
	g.currentStreakLabel = widget.NewLabel("0")
	g.currentStreakLabel.Alignment = fyne.TextAlignTrailing

	// Highest streak label
	g.highestStreakLabel = widget.NewLabel("0")
	g.highestStreakLabel.Alignment = fyne.TextAlignTrailing

	var startBtn *components.Button

	g.higherBtn = components.NewButton(lang.X("game.higher_lower.higher", "Higher"), func() {
		g.makeGuess(true)
	})
	g.lowerBtn = components.NewButton(lang.X("game.higher_lower.lower", "Lower"), func() {
		g.makeGuess(false)
	})
	g.nextBtn = components.NewButton(lang.X("game.higher_lower.next_round", "Next Round"), func() {
		g.nextRound()
	})

	g.higherBtn.Hide()
	g.lowerBtn.Hide()
	g.nextBtn.Hide()

	startBtn = components.NewButton(lang.X("game.higher_lower.start", "Start Game"), func() {
		g.Start()
		startBtn.Hide()
		g.higherBtn.Show()
		g.lowerBtn.Show()
	})

	// Create responsive button grid
	buttonGrid := container.NewGridWithColumns(2, g.higherBtn, g.lowerBtn)

	// Streak displays with labels on left, values on right
	currentStreakTextLabel := widget.NewLabel(lang.X("game.higher_lower.current_streak_label", "Current Streak:"))
	currentStreakContainer := container.NewBorder(nil, nil, currentStreakTextLabel, g.currentStreakLabel, container.NewMax())

	highestStreakTextLabel := widget.NewLabel(lang.X("game.higher_lower.highest_streak_label", "Highest Streak:"))
	highestStreakContainer := container.NewBorder(nil, nil, highestStreakTextLabel, g.highestStreakLabel, container.NewMax())

	// Header section with natural spacing
	headerSection := container.NewVBox(
		topBar.GetContainer(),
		gameDescription,
		currentStreakContainer,
		highestStreakContainer,
	)

	// Game content
	gameContent := container.NewVBox(
		startBtn,
		container.NewCenter(g.countryOneNameLabel),
		container.NewCenter(g.countryOnePopLabel),
		container.NewCenter(g.countryTwoNameLabel),
		container.NewCenter(g.countryTwoPopLabel),
		buttonGrid,
		g.nextBtn,
	)

	g.content = container.NewVBox(
		headerSection,
		gameContent,
	)
}

func (g *Game) makeGuess(isHigher bool) {
	correct := (isHigher && g.secondCountry > g.firstCountry) || (!isHigher && g.secondCountry < g.firstCountry)
	if correct {
		g.score++
		// Update highest streak if current is higher
		if g.score > g.highestStreak {
			g.highestStreak = g.score
			g.highestStreakLabel.SetText(fmt.Sprintf("%d", g.highestStreak))
		}
	} else {
		g.score = 0
	}
	g.currentStreakLabel.SetText(fmt.Sprintf("%d", g.score))
	g.countryTwoPopLabel.SetText(fmt.Sprintf(lang.X("game.higher_lower.population", "Population: %d"), g.secondCountry))
	g.higherBtn.Hide()
	g.lowerBtn.Hide()
	g.nextBtn.Show()
}

func (g *Game) nextRound() {
	countries := data.LoadCountries()
	newCountry := countries[rand.Intn(len(countries))]

	for newCountry.Name.Common == g.countryTwoNameLabel.Text {
		newCountry = countries[rand.Intn(len(countries))]
	}

	g.firstCountry = g.secondCountry
	g.countryOneNameLabel.SetText(g.countryTwoNameLabel.Text)
	g.countryOnePopLabel.SetText(fmt.Sprintf(lang.X("game.higher_lower.population", "Population: %d"), g.firstCountry))

	g.secondCountry = newCountry.Population
	g.countryTwoNameLabel.SetText(newCountry.Name.Common)
	g.countryTwoPopLabel.SetText(lang.X("game.higher_lower.population_unknown", "Population: ?"))

	g.nextBtn.Hide()
	g.higherBtn.Show()
	g.lowerBtn.Show()
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	countries := data.LoadCountries()

	firstRandomCountry := countries[rand.Intn(len(countries))]
	secondRandomCountry := countries[rand.Intn(len(countries))]

	for firstRandomCountry.Name.Common == secondRandomCountry.Name.Common {
		secondRandomCountry = countries[rand.Intn(len(countries))]
	}

	g.firstCountry = firstRandomCountry.Population
	g.secondCountry = secondRandomCountry.Population

	g.countryOneNameLabel.SetText(firstRandomCountry.Name.Common)
	g.countryOnePopLabel.SetText(fmt.Sprintf(lang.X("game.higher_lower.population", "Population: %d"), firstRandomCountry.Population))
	g.countryTwoNameLabel.SetText(secondRandomCountry.Name.Common)
	g.countryTwoPopLabel.SetText(lang.X("game.higher_lower.population_unknown", "Population: ?"))
}

func (g *Game) Reset() {
	g.score = 0
	g.highestStreak = 0
	g.currentStreakLabel.SetText("0")
	g.highestStreakLabel.SetText("0")
	g.Start()
}
