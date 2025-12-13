package screens

import (
	"flagged-it/internal/translations"
	"flagged-it/internal/utils"
	"fmt"
	"image/color"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Dashboard struct {
	content      *fyne.Container
	navigateFunc func(string)
	debugFunc    func()
	window       fyne.Window
	debugManager *utils.DebugManager
	scoreManager *utils.ScoreManager
	scoresGrid   *fyne.Container
}

func NewDashboard(navigateFunc func(string), debugFunc func(), window fyne.Window, scoreManager *utils.ScoreManager) *Dashboard {
	d := &Dashboard{
		navigateFunc: navigateFunc,
		debugFunc:    debugFunc,
		window:       window,
		debugManager: utils.NewDebugManager(),
		scoreManager: scoreManager,
	}
	d.setupUI()
	return d
}

func (d *Dashboard) setupUI() {
	title := widget.NewLabel(lang.X("dashboard.title", "Choose Your Game Mode"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Language selector
	languageList := make([]string, len(translations.TranslationsInfo))
	var langSelIndex int
	currentLocale := utils.GetCurrentLocale()
	for i, tr := range translations.TranslationsInfo {
		languageList[i] = tr.DisplayName
		if tr.Name == currentLocale {
			langSelIndex = i
		}
	}

	langSelect := widget.NewSelect(languageList, nil)
	langSelect.SetSelectedIndex(langSelIndex)
	langSelect.OnChanged = func(_ string) {
		if i := langSelect.SelectedIndex(); i >= 0 {
			utils.SetCurrentLocale(translations.TranslationsInfo[i].Name)
			utils.LoadTranslation(translations.TranslationsInfo[i].Name)
			d.window.SetContent(NewDashboard(d.navigateFunc, d.debugFunc, d.window, d.scoreManager).GetContent())
		}
	}

	// Header with language selector, title and optional settings button
	var header *fyne.Container
	if d.debugManager.IsDebugEnabled() {
		settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), d.debugFunc)
		header = container.NewBorder(
			nil, nil,
			langSelect,
			settingsBtn,
			container.NewHBox(widget.NewLabel(""), title),
		)
	} else {
		header = container.NewBorder(
			nil, nil,
			langSelect,
			nil,
			container.NewHBox(widget.NewLabel(""), title),
		)
	}

	// Game buttons
	flagBtn := widget.NewButtonWithIcon(lang.X("game.flag.title", "Guess by Flag"), theme.MailForwardIcon(), func() {
		d.navigateFunc("flag")
	})
	countryListBtn := widget.NewButtonWithIcon(lang.X("game.list.title", "List All Countries"), theme.ListIcon(), func() {
		d.navigateFunc("list")
	})
	countryGuessBtn := widget.NewButtonWithIcon(lang.X("game.shape.title", "Guess by Shape"), theme.MediaRecordIcon(), func() {
		d.navigateFunc("shape")
	})
	hangmanBtn := widget.NewButtonWithIcon(lang.X("game.hangman.title", "Hangman"), theme.AccountIcon(), func() {
		d.navigateFunc("hangman")
	})
	factGuessBtn := widget.NewButtonWithIcon(lang.X("game.facts.title", "Guess by Facts"), theme.InfoIcon(), func() {
		d.navigateFunc("facts")
	})
	higher_lowerBtn := widget.NewButtonWithIcon(lang.X("game.higher_lower.title", "Higher or Lower"), theme.UploadIcon(), func() {
		d.navigateFunc("higher_lower")
	})

	guessingBtn := widget.NewButtonWithIcon(lang.X("game.guessing.title", "What Country is This"), theme.GridIcon(), func() {
		d.navigateFunc("guessing")
	})

	// Game buttons in responsive grid
	columns := 2
	if utils.IsMobile() {
		columns = 1
	}
	gameButtons := container.NewGridWithColumns(columns,
		flagBtn,
		countryListBtn,
		countryGuessBtn,
		hangmanBtn,
		factGuessBtn,
		higher_lowerBtn,
		guessingBtn,
	)

	// Score boxes
	d.scoresGrid = d.createScoresBox()

	d.content = container.NewVBox(
		header,
		widget.NewSeparator(),
		gameButtons,
		widget.NewSeparator(),
		d.scoresGrid,
	)
}

func (d *Dashboard) getScoreColor(score, total int) color.Color {
	if total == 0 {
		return color.RGBA{200, 200, 200, 255}
	}

	percentage := float64(score) / float64(total) * 100

	if percentage >= 70 {
		return color.RGBA{144, 238, 144, 255}
	} else if percentage >= 40 {
		return color.RGBA{255, 165, 0, 255}
	}
	return color.RGBA{255, 99, 71, 255}
}

func (d *Dashboard) createScoresBox() *fyne.Container {
	scores := d.scoreManager.GetAllScores()
	totals := d.scoreManager.GetAllTotals()

	gameNames := map[string]string{
		"flag":         "Flag",
		"hangman":      "Hangman",
		"facts":        "Facts",
		"shape":        "Shape",
		"higher_lower": "Higher/Lower",
	}

	// Sort games alphabetically
	var gameKeys []string
	for key := range gameNames {
		gameKeys = append(gameKeys, key)
	}
	sort.Strings(gameKeys)

	boxes := []fyne.CanvasObject{}

	for _, gameKey := range gameKeys {
		score := scores[gameKey]

		if score == 0 {
			continue
		}

		total := totals[gameKey]
		name := gameNames[gameKey]

		var scoreText string
		if total > 0 {
			scoreText = fmt.Sprintf("%d/%d", score, total)
		} else {
			scoreText = fmt.Sprintf("%d", score)
		}

		// Create larger, bolder score label
		scoreLabel := canvas.NewText(scoreText, color.White)
		scoreLabel.TextSize = 24
		scoreLabel.TextStyle = fyne.TextStyle{Bold: true}
		scoreLabel.Alignment = fyne.TextAlignCenter

		// Create game name label
		gameLabel := canvas.NewText(name, color.White)
		gameLabel.TextSize = 14
		gameLabel.TextStyle = fyne.TextStyle{Bold: true}
		gameLabel.Alignment = fyne.TextAlignCenter

		// Get color based on score
		bgColor := d.getScoreColor(score, total)
		bg := canvas.NewRectangle(bgColor)

		// Create border
		border := canvas.NewRectangle(color.RGBA{50, 50, 50, 255})

		// Stack: border, background, content
		content := container.NewVBox(
			gameLabel,
			scoreLabel,
		)

		boxWithBorder := container.NewStack(
			border,
			bg,
			container.NewPadded(content),
		)

		boxes = append(boxes, boxWithBorder)
	}

	if len(boxes) == 0 {
		return container.NewVBox()
	}

	// Create responsive grid - 2 columns on mobile
	columns := 2
	if len(boxes) > 4 {
		columns = 3
	}
	if len(boxes) > 6 {
		columns = 4
	}

	scoresGrid := container.NewGridWithColumns(columns, boxes...)

	return container.NewVBox(
		scoresGrid,
	)
}

func (d *Dashboard) RefreshScores() {
	d.scoresGrid.RemoveAll()
	newScoresBox := d.createScoresBox()
	for _, obj := range newScoresBox.Objects {
		d.scoresGrid.Add(obj)
	}
	d.scoresGrid.Refresh()
}

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
