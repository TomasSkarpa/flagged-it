package main

import (
	"flagged-it/internal/app"
	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

func main() {
	myApp := fyneApp.New()
	myApp.Settings().SetTheme(&scaledTheme{scale: 1.4})
	myWindow := myApp.NewWindow("Flagged It - Country Guessing Games")
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.SetOnClosed(myApp.Quit)

	appController := app.NewApp(myWindow)
	myWindow.SetContent(appController.GetDashboard())

	myWindow.ShowAndRun()
}

type scaledTheme struct {
	scale float32
}

func (s *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
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
