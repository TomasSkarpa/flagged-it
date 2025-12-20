package screens

import (
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fmt"
	"image/color"
	"runtime"
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

	// Language selector button - shows "🇬🇧 EN" format
	langBtn := components.NewLanguageSelectorButton(d.window, func() {
		// Refresh dashboard when language changes
		d.window.SetContent(NewDashboard(d.navigateFunc, d.debugFunc, d.window, d.scoreManager).GetContent())
	})

	// Header with language selector, title and optional settings button
	var header *fyne.Container
	if d.debugManager.IsDebugEnabled() {
		settingsBtn := components.NewButtonWithIcon("", theme.SettingsIcon(), d.debugFunc)
		header = container.NewBorder(
			nil, nil,
			langBtn,
			settingsBtn,
			container.NewCenter(title),
		)
	} else {
		header = container.NewBorder(
			nil, nil,
			langBtn,
			nil,
			container.NewCenter(title),
		)
	}

	// Game buttons
	flagBtn := components.NewButtonWithIcon(lang.X("game.flag.title", "Guess by Flag"), theme.MailForwardIcon(), func() {
		d.navigateFunc("flag")
	})
	countryListBtn := components.NewButtonWithIcon(lang.X("game.list.title", "List All Countries"), theme.ListIcon(), func() {
		d.navigateFunc("list")
	})
	countryGuessBtn := components.NewButtonWithIcon(lang.X("game.shape.title", "Guess by Shape"), theme.MediaRecordIcon(), func() {
		d.navigateFunc("shape")
	})
	hangmanBtn := components.NewButtonWithIcon(lang.X("game.hangman.title", "Hangman"), theme.AccountIcon(), func() {
		d.navigateFunc("hangman")
	})
	factGuessBtn := components.NewButtonWithIcon(lang.X("game.facts.title", "Guess by Facts"), theme.InfoIcon(), func() {
		d.navigateFunc("facts")
	})
	higher_lowerBtn := components.NewButtonWithIcon(lang.X("game.higher_lower.title", "Higher or Lower"), theme.UploadIcon(), func() {
		d.navigateFunc("higher_lower")
	})

	guessingBtn := components.NewButtonWithIcon(lang.X("game.guessing.title", "What Country is This"), theme.GridIcon(), func() {
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

	// Promotional cards section
	promoCards := d.createPromoCards()

	// Main content (header, game buttons, scores)
	mainContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		gameButtons,
		widget.NewSeparator(),
		d.scoresGrid,
	)

	// Use Border layout to pin promo cards at bottom
	d.content = container.NewBorder(
		nil,                              // top
		promoCards,                       // bottom - promo cards pinned here
		nil,                              // left
		nil,                              // right
		container.NewScroll(mainContent), // center - scrollable main content
	)
}

func (d *Dashboard) createPromoCards() *fyne.Container {
	// Define asset paths based on runtime
	var europeFlagsPath, asiaMapPath, hangmanPath, higherLowerPath string
	if runtime.GOOS == "js" {
		europeFlagsPath = "assets/world_map_silhouette.svg"
		asiaMapPath = "assets/world_map_silhouette.svg"
		hangmanPath = "assets/hangman.svg"
		higherLowerPath = "assets/higher_lower.svg"
	} else {
		europeFlagsPath = "assets/world_map_silhouette.svg"
		asiaMapPath = "assets/world_map_silhouette.svg"
		hangmanPath = "assets/hangman.svg"
		higherLowerPath = "assets/higher_lower.svg"
	}

	isMobile := utils.IsMobile()

	// Card 1: Flags - Europe
	card1 := components.NewPromoCard(components.PromoCardConfig{
		Title:       lang.X("promo.europe_flags.title", "European Flags"),
		Description: lang.X("promo.europe_flags.desc", "Master the flags of Europe"),
		IconPath:    europeFlagsPath,
		Badge:       lang.X("promo.badge.popular", "Popular"),
		BadgeColor:  components.GetBadgeColor("Popular"),
		IsMobile:    isMobile,
		OnTap: func() {
			d.navigateFunc("flag_europe")
		},
	})

	// Card 2: Shapes - Asia
	card2 := components.NewPromoCard(components.PromoCardConfig{
		Title:       lang.X("promo.asia_shapes.title", "Asian Shapes"),
		Description: lang.X("promo.asia_shapes.desc", "Guess countries by shape"),
		IconPath:    asiaMapPath,
		Badge:       lang.X("promo.badge.new", "New"),
		BadgeColor:  components.GetBadgeColor("New"),
		IsMobile:    isMobile,
		OnTap: func() {
			d.navigateFunc("shape_asia")
		},
	})

	// Card 3: Hangman
	card3 := components.NewPromoCard(components.PromoCardConfig{
		Title:       lang.X("game.hangman.title", "Hangman"),
		Description: lang.X("promo.hangman.desc", "Classic word guessing game"),
		IconPath:    hangmanPath,
		IsMobile:    isMobile,
		OnTap: func() {
			d.navigateFunc("hangman")
		},
	})

	// Card 4: Higher or Lower
	card4 := components.NewPromoCard(components.PromoCardConfig{
		Title:       lang.X("game.higher_lower.title", "Higher or Lower"),
		Description: lang.X("promo.higher_lower.desc", "Compare country stats"),
		IconPath:    higherLowerPath,
		Badge:       lang.X("promo.badge.popular", "Popular"),
		BadgeColor:  components.GetBadgeColor("Popular"),
		IsMobile:    isMobile,
		OnTap: func() {
			d.navigateFunc("higher_lower")
		},
	})

	cards := []*components.PromoCard{card1, card2, card3, card4}
	return components.CreatePromoCardsGrid(cards, isMobile)
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
