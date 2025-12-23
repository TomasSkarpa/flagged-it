package screens

import (
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"image/color"
	"runtime"

	"fyne.io/fyne/v2"
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
	app          fyne.App
	debugManager *utils.DebugManager
}

func NewDashboard(navigateFunc func(string), debugFunc func(), window fyne.Window, app fyne.App) *Dashboard {
	d := &Dashboard{
		navigateFunc: navigateFunc,
		debugFunc:    debugFunc,
		window:       window,
		app:          app,
		debugManager: utils.NewDebugManager(),
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
		d.window.SetContent(NewDashboard(d.navigateFunc, d.debugFunc, d.window, d.app).GetContent())
	})
	
	// Theme selector button - shows "💻 System", "🌙 Dark", or "☀️ Light"
	themeBtn := components.NewThemeSelectorButton(d.window, d.app)

	// Header with language selector, theme selector, title and optional settings button
	var header *fyne.Container
	leftButtons := container.NewHBox(langBtn, themeBtn)
	
	if d.debugManager.IsDebugEnabled() {
		settingsBtn := components.NewButtonWithIcon("", theme.SettingsIcon(), d.debugFunc)
		header = container.NewBorder(
			nil, nil,
			leftButtons,
			settingsBtn,
			container.NewCenter(title),
		)
	} else {
		header = container.NewBorder(
			nil, nil,
			leftButtons,
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

	// Promotional cards section
	promoCards := d.createPromoCards()

	// Main content (header, game buttons)
	mainContent := container.NewVBox(
		header,
		components.NewDashedSeparator(color.RGBA{200, 200, 200, 255}, 5), // Dashed separator 3px
		gameButtons,
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
		europeFlagsPath = "assets/twemoji_flags_cca2/EU.svg"
		asiaMapPath = "assets/iconography/asia_map.png"
		hangmanPath = "assets/iconography/hangman.png"
		higherLowerPath = "assets/iconography/higher_lower.png"
	} else {
		europeFlagsPath = "assets/twemoji_flags_cca2/EU.svg"
		asiaMapPath = "assets/iconography/asia_map.png"
		hangmanPath = "assets/iconography/hangman.png"
		higherLowerPath = "assets/iconography/higher_lower.png"
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

func (d *Dashboard) GetContent() *fyne.Container {
	return d.content
}
