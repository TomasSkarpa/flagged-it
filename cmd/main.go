package main

import (
	"flagged-it/internal/app"
	"runtime"
	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

func main() {
	myApp := fyneApp.New()
	myWindow := myApp.NewWindow("Flagged It - Country Guessing Games")
	myWindow.Resize(fyne.NewSize(1024, 768))
	myWindow.SetOnClosed(myApp.Quit)
	myWindow.SetMaster()

	scale := float32(1.4)
	if runtime.GOOS == "js" {
		scale = 1.0
	} else if myWindow.Canvas().Size().Width < 768 {
		scale = 1.0
	}
	myApp.Settings().SetTheme(&scaledTheme{scale: scale})

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
