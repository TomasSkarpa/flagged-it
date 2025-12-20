package components

import (
	"image/color"
	"runtime"

	"flagged-it/pkg/assets"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// PromoCard represents a promotional game card with proper sizing
type PromoCard struct {
	widget.BaseWidget
	bg        *canvas.Rectangle
	icon      *canvas.Image
	title     *canvas.Text
	desc      *canvas.Text
	badgeBg   *canvas.Rectangle
	badgeText *canvas.Text
	onTap     func()
	hasBadge  bool
	isMobile  bool
}

// PromoCardConfig holds the configuration for a promo card
type PromoCardConfig struct {
	Title       string
	Description string
	IconPath    string
	Badge       string
	BadgeColor  color.Color
	OnTap       func()
	IsMobile    bool
}

// NewPromoCard creates a new promotional card
func NewPromoCard(config PromoCardConfig) *PromoCard {
	card := &PromoCard{
		onTap:    config.OnTap,
		isMobile: config.IsMobile,
	}
	card.ExtendBaseWidget(card)

	// Background - solid dark blue
	card.bg = canvas.NewRectangle(color.RGBA{30, 41, 59, 255}) // slate-800
	card.bg.CornerRadius = 12

	// Load icon
	if config.IconPath != "" {
		var iconResource fyne.Resource
		var err error
		if runtime.GOOS == "js" {
			iconResource, err = fyne.LoadResourceFromURLString(config.IconPath)
		} else {
			iconResource, err = assets.LoadResourceFromPath(config.IconPath)
		}
		if err == nil {
			card.icon = canvas.NewImageFromResource(iconResource)
			card.icon.FillMode = canvas.ImageFillContain
		}
	}

	// Title - white, bold, centered (responsive sizes)
	card.title = canvas.NewText(config.Title, color.White)
	card.title.TextStyle = fyne.TextStyle{Bold: true}
	if config.IsMobile {
		card.title.TextSize = 13
	} else {
		card.title.TextSize = 18
	}
	card.title.Alignment = fyne.TextAlignCenter

	// Description - light gray, smaller (responsive sizes)
	card.desc = canvas.NewText(config.Description, color.RGBA{148, 163, 184, 255}) // slate-400
	if config.IsMobile {
		card.desc.TextSize = 11
	} else {
		card.desc.TextSize = 14
	}
	card.desc.Alignment = fyne.TextAlignCenter

	// Badge (responsive sizes)
	if config.Badge != "" {
		card.hasBadge = true
		card.badgeBg = canvas.NewRectangle(config.BadgeColor)
		card.badgeBg.CornerRadius = 4
		card.badgeText = canvas.NewText(config.Badge, color.White)
		if config.IsMobile {
			card.badgeText.TextSize = 9
		} else {
			card.badgeText.TextSize = 12
		}
		card.badgeText.TextStyle = fyne.TextStyle{Bold: true}
	}

	return card
}

func (c *PromoCard) CreateRenderer() fyne.WidgetRenderer {
	return &promoCardRenderer{card: c}
}

func (c *PromoCard) Tapped(_ *fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *PromoCard) TappedSecondary(_ *fyne.PointEvent) {}

func (c *PromoCard) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

// Custom renderer with proper flexbox-like centering
type promoCardRenderer struct {
	card *PromoCard
}

func (r *promoCardRenderer) MinSize() fyne.Size {
	// Mobile: compact cards, Desktop: 3x taller
	if r.card.isMobile {
		return fyne.NewSize(120, 130)
	}
	return fyne.NewSize(140, 390) // 130 * 3 = 390
}

func (r *promoCardRenderer) Layout(size fyne.Size) {
	// Responsive sizing: larger on desktop
	var cardPadding, iconSize, badgePadding, badgeMargin, elementGap float32
	if r.card.isMobile {
		cardPadding = 12
		iconSize = 36
		badgePadding = 4
		badgeMargin = 8
		elementGap = 6
	} else {
		// Desktop: larger sizes
		cardPadding = 20
		iconSize = 100 // Much larger icon
		badgePadding = 6
		badgeMargin = 12
		elementGap = 16
	}

	// Background fills entire card
	r.card.bg.Resize(size)
	r.card.bg.Move(fyne.NewPos(0, 0))

	// Calculate content area
	contentWidth := size.Width - cardPadding*2

	// Calculate total content height for vertical centering
	titleHeight := r.card.title.MinSize().Height
	descHeight := r.card.desc.MinSize().Height

	var totalContentHeight float32
	if r.card.icon != nil {
		totalContentHeight = iconSize + elementGap + titleHeight + elementGap + descHeight
	} else {
		totalContentHeight = titleHeight + elementGap + descHeight
	}

	// Start Y position to center content vertically
	startY := (size.Height - totalContentHeight) / 2
	yPos := startY

	// Icon - centered horizontally, standardized size
	if r.card.icon != nil {
		r.card.icon.Resize(fyne.NewSize(iconSize, iconSize))
		r.card.icon.Move(fyne.NewPos((size.Width-iconSize)/2, yPos))
		yPos += iconSize + elementGap
	}

	// Title - centered
	titleWidth := r.card.title.MinSize().Width
	if titleWidth > contentWidth {
		titleWidth = contentWidth
	}
	r.card.title.Resize(fyne.NewSize(titleWidth, titleHeight))
	r.card.title.Move(fyne.NewPos((size.Width-titleWidth)/2, yPos))
	yPos += titleHeight + elementGap

	// Description - centered
	descWidth := r.card.desc.MinSize().Width
	if descWidth > contentWidth {
		descWidth = contentWidth
	}
	r.card.desc.Resize(fyne.NewSize(descWidth, descHeight))
	r.card.desc.Move(fyne.NewPos((size.Width-descWidth)/2, yPos))

	// Badge - positioned inside card with proper margin
	if r.card.hasBadge && r.card.badgeBg != nil {
		textSize := r.card.badgeText.MinSize()
		badgeWidth := textSize.Width + badgePadding*2
		badgeHeight := textSize.Height + badgePadding*2

		// Position badge inside the card with margin from edges
		badgeX := size.Width - badgeWidth - badgeMargin
		badgeY := badgeMargin

		r.card.badgeBg.Resize(fyne.NewSize(badgeWidth, badgeHeight))
		r.card.badgeBg.Move(fyne.NewPos(badgeX, badgeY))

		// Center text in badge
		r.card.badgeText.Resize(textSize)
		r.card.badgeText.Move(fyne.NewPos(badgeX+badgePadding, badgeY+badgePadding))
	}
}

func (r *promoCardRenderer) Refresh() {
	r.card.bg.Refresh()
	if r.card.icon != nil {
		r.card.icon.Refresh()
	}
	r.card.title.Refresh()
	r.card.desc.Refresh()
	if r.card.badgeBg != nil {
		r.card.badgeBg.Refresh()
	}
	if r.card.badgeText != nil {
		r.card.badgeText.Refresh()
	}
}

func (r *promoCardRenderer) Objects() []fyne.CanvasObject {
	objects := []fyne.CanvasObject{r.card.bg}
	if r.card.icon != nil {
		objects = append(objects, r.card.icon)
	}
	objects = append(objects, r.card.title, r.card.desc)
	if r.card.badgeBg != nil {
		objects = append(objects, r.card.badgeBg)
	}
	if r.card.badgeText != nil {
		objects = append(objects, r.card.badgeText)
	}
	return objects
}

func (r *promoCardRenderer) Destroy() {}

// CreatePromoCardsGrid creates a responsive grid of promotional cards
func CreatePromoCardsGrid(cards []*PromoCard, isMobile bool) *fyne.Container {
	columns := 4
	if isMobile {
		columns = 2
	}

	cardObjects := make([]fyne.CanvasObject, len(cards))
	for i, card := range cards {
		cardObjects[i] = card
	}

	return container.NewGridWithColumns(columns, cardObjects...)
}

// GetBadgeColor returns the color for a badge type
func GetBadgeColor(badgeType string) color.Color {
	switch badgeType {
	case "Popular":
		return color.RGBA{239, 68, 68, 255} // Red
	case "New":
		return color.RGBA{34, 197, 94, 255} // Green
	default:
		return color.RGBA{59, 130, 246, 255} // Blue
	}
}
