package interfaces

import "fyne.io/fyne/v2"

type GameModule interface {
	GetContent() *fyne.Container
	Start()
	Reset()
}