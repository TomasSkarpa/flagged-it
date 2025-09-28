package main

import (
	"flagged-it/internal/app"
	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
)

func main() {
	myApp := fyneApp.New()
	myWindow := myApp.NewWindow("Flagged It - Country Guessing Games")
	myWindow.Resize(fyne.NewSize(800, 600))

	appController := app.NewApp(myWindow)
	myWindow.SetContent(appController.GetDashboard())

	myWindow.ShowAndRun()
}
