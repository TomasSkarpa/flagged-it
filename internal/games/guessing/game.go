package guessing

import (
	"fmt"
	"image/color"
	"math/rand"
	"runtime"
	"strings"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type fixedHeightTile struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	content fyne.CanvasObject
	height  float32
}

func newFixedHeightTile(bg *canvas.Rectangle, content fyne.CanvasObject, height float32) *fixedHeightTile {
	t := &fixedHeightTile{bg: bg, content: content, height: height}
	t.ExtendBaseWidget(t)
	return t
}

func (t *fixedHeightTile) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(t.bg, t.content))
}

func (t *fixedHeightTile) MinSize() fyne.Size {
	return fyne.NewSize(0, t.height)
}

type Game struct {
	content        *fyne.Container
	backFunc       func()
	countries      []models.Country
	currentCountry *models.Country
	guessEntry     *widget.Entry
	statusLabel    *widget.Label
	guessBtn       *widget.Button
	headerGrid     *fyne.Container
	bodyGrid       *fyne.Container
	bodyScroll     *container.Scroll
	guesses        []models.Country
}

func NewGame(backFunc func()) *Game {
	g := &Game{
		backFunc: backFunc,
		guesses:  []models.Country{},
	}
	g.loadCountries()
	g.setupUI()
	g.newGame()
	return g
}

func (g *Game) loadCountries() {
	g.countries = data.LoadCountries()
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar(lang.X("game.guessing.title", "What country is this?"), g.backFunc, g.newGame)

	g.statusLabel = widget.NewLabel(lang.X("game.guessing.make_guess", "Make a guess!"))

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder(lang.X("game.guessing.enter_country", "Enter country name..."))
	g.guessEntry.OnSubmitted = func(text string) { g.makeGuess() }

	g.guessBtn = widget.NewButton(lang.X("game.guessing.guess", "Guess"), g.makeGuess)

	guessContainer := container.NewGridWithColumns(2, g.guessEntry, g.guessBtn)

	g.headerGrid = container.NewGridWithColumns(5)
	g.addHeaderRow()

	g.bodyGrid = container.NewVBox()
	g.bodyScroll = container.NewVScroll(g.bodyGrid)

	topSection := container.NewVBox(
		widget.NewSeparator(),
		g.statusLabel,
		widget.NewSeparator(),
		guessContainer,
		widget.NewSeparator(),
		widget.NewLabel(lang.X("game.guessing.history", "Guess History:")),
		g.headerGrid,
	)

	g.content = container.NewBorder(
		topBar.GetContainer(),
		nil,
		nil,
		nil,
		container.NewBorder(
			topSection,
			nil,
			nil,
			nil,
			g.bodyScroll,
		),
	)
}

func (g *Game) addHeaderRow() {
	g.headerGrid.Add(g.createTile(lang.X("game.guessing.flag", "Flag"), nil, color.RGBA{100, 100, 100, 255}))
	g.headerGrid.Add(g.createTile(lang.X("game.guessing.country", "Country"), nil, color.RGBA{100, 100, 100, 255}))
	g.headerGrid.Add(g.createTile(lang.X("game.guessing.continent", "Continent"), nil, color.RGBA{100, 100, 100, 255}))
	g.headerGrid.Add(g.createTile(lang.X("game.guessing.population", "Population"), nil, color.RGBA{100, 100, 100, 255}))
	g.headerGrid.Add(g.createTile(lang.X("game.guessing.area", "Area"), nil, color.RGBA{100, 100, 100, 255}))
}

func (g *Game) createTile(text string, icon fyne.Resource, bgColor color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(bgColor)
	label := canvas.NewText(text, color.White)
	label.TextSize = 24
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter

	var content fyne.CanvasObject
	if icon != nil {
		iconWidget := widget.NewIcon(icon)
		content = container.NewHBox(label, iconWidget)
	} else {
		content = container.NewVBox(label)
	}

	return newFixedHeightTile(bg, container.NewCenter(content), 50)
}

func (g *Game) getCompareIcon(guessVal, targetVal float64) fyne.Resource {
	if guessVal < targetVal {
		return theme.MoveUpIcon()
	} else if guessVal > targetVal {
		return theme.MoveDownIcon()
	}
	return nil
}

func (g *Game) getCompareColor(guessVal, targetVal float64) color.Color {
	if guessVal == targetVal {
		return color.RGBA{0, 180, 0, 255}
	}
	return color.RGBA{180, 0, 0, 255}
}

func (g *Game) createFlagTile(country *models.Country) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.RGBA{100, 100, 100, 255})
	flagIcon := widget.NewIcon(g.getCountryFlag(country))
	return newFixedHeightTile(bg, container.NewCenter(flagIcon), 50)
}

func (g *Game) getCountryFlag(country *models.Country) fyne.Resource {
	var flagResource fyne.Resource
	var err error
	if runtime.GOOS == "js" {
		flagURL := fmt.Sprintf("assets/twemoji_flags_cca2/%s.svg", country.CCA2)
		flagResource, err = fyne.LoadResourceFromURLString(flagURL)
	} else {
		flagPath := fmt.Sprintf("assets/twemoji_flags_cca2/%s.svg", country.CCA2)
		flagResource, err = fyne.LoadResourceFromPath(flagPath)
	}
	if err != nil {
		return nil
	}
	return flagResource
}

func (g *Game) addGuessRow(country *models.Country) {
	isCorrectContinent := country.Region == g.currentCountry.Region

	continentColor := color.RGBA{0, 180, 0, 255}
	if !isCorrectContinent {
		continentColor = color.RGBA{180, 0, 0, 255}
	}

	popIcon := g.getCompareIcon(float64(country.Population), float64(g.currentCountry.Population))
	popColor := g.getCompareColor(float64(country.Population), float64(g.currentCountry.Population))

	areaIcon := g.getCompareIcon(country.Area, g.currentCountry.Area)
	areaColor := g.getCompareColor(country.Area, g.currentCountry.Area)

	flagTile := g.createFlagTile(country)
	countryTile := g.createTile(country.Name.Common, nil, color.RGBA{100, 100, 100, 255})

	translatedRegion := utils.TranslateRegion(country.Region)
	row := container.NewGridWithColumns(5,
		flagTile,
		countryTile,
		g.createTile(translatedRegion, nil, continentColor),
		g.createTile(fmt.Sprintf("%d", country.Population), popIcon, popColor),
		g.createTile(fmt.Sprintf("%.0f", country.Area), areaIcon, areaColor),
	)
	g.bodyGrid.Add(row)
	g.bodyGrid.Refresh()
}

func (g *Game) newGame() {
	if len(g.countries) == 0 {
		g.statusLabel.SetText(lang.X("error.loading_countries", "Error loading countries data"))
		return
	}

	g.currentCountry = &g.countries[rand.Intn(len(g.countries))]
	g.guesses = []models.Country{}
	g.bodyGrid.RemoveAll()
	g.guessEntry.SetText("")
	g.guessEntry.Enable()
	g.guessBtn.Enable()
	g.statusLabel.SetText(lang.X("game.guessing.make_guess", "Make a guess!"))
}

func (g *Game) getCountry(countries []models.Country, name string) *models.Country {
	for _, country := range countries {
		if utils.MatchCountry(name, country, utils.MatchAll) {
			return &country
		}
	}
	return nil
}

func (g *Game) makeGuess() {
	guess := strings.TrimSpace(g.guessEntry.Text)
	if guess == "" {
		return
	}

	guessedCountry := g.getCountry(g.countries, guess)
	if guessedCountry == nil {
		g.statusLabel.SetText(lang.X("game.guessing.not_found", "Country not found!"))
		return
	}

	g.guesses = append(g.guesses, *guessedCountry)
	g.addGuessRow(guessedCountry)

	if utils.MatchCountry(guess, *g.currentCountry, utils.MatchAll) {
		g.statusLabel.SetText(fmt.Sprintf(lang.X("game.guessing.correct", "Correct! It was %s!"), g.currentCountry.Name.Common))
		g.guessEntry.Disable()
		g.guessBtn.Disable()
		return
	}

	g.guessEntry.SetText("")
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) Start() {
	g.newGame()
}

func (g *Game) Reset() {
	g.newGame()
}
