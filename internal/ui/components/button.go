package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// Button is a wrapper around widget.Button that shows a pointer cursor on hover
type Button struct {
	widget.Button
}

// NewButton creates a new button with pointer cursor
func NewButton(label string, tapped func()) *Button {
	b := &Button{}
	b.ExtendBaseWidget(b)
	b.Text = label
	b.OnTapped = tapped
	return b
}

// NewButtonWithIcon creates a new button with an icon and pointer cursor
func NewButtonWithIcon(label string, icon fyne.Resource, tapped func()) *Button {
	b := &Button{}
	b.ExtendBaseWidget(b)
	b.Text = label
	b.Icon = icon
	b.OnTapped = tapped
	return b
}

// Cursor returns the pointer cursor for this button
func (b *Button) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}



