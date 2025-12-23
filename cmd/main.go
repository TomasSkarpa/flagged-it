package main

import (
	"flagged-it/internal/app"
	"flagged-it/internal/utils"
	"image/color"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	utils.LoadTranslation(utils.GetCurrentLocale())
	myApp := fyneApp.New()

	myWindow := myApp.NewWindow("Flagged It - Country Guessing Games")
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.SetOnClosed(myApp.Quit)
	myWindow.SetMaster()

	scale := float32(1.7)
	if utils.IsMobile() {
		scale = 1.0
	}
	
	// Get theme from localStorage
	effectiveTheme := utils.GetEffectiveTheme()
	var variant fyne.ThemeVariant
	if effectiveTheme == "light" {
		variant = theme.VariantLight
	} else {
		variant = theme.VariantDark
	}
	
	myApp.Settings().SetTheme(&scaledTheme{
		scale:   scale,
		variant: variant,
	})

	appController := app.NewApp(myWindow, myApp)
	myWindow.SetContent(appController.GetDashboard())

	// Handle tab visibility changes to prevent freezing
	setupVisibilityHandler(myWindow)

	myWindow.ShowAndRun()
}

type scaledTheme struct {
	scale   float32
	variant fyne.ThemeVariant
}

func (s *scaledTheme) SetVariant(variant fyne.ThemeVariant) {
	s.variant = variant
}

func (s *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Use the theme's variant instead of the passed one
	return theme.DefaultTheme().Color(name, s.variant)
}

func (s *scaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (s *scaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (s *scaledTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name) * s.scale
}
