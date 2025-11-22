package shape

import (
	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

type Game struct {
	content         *fyne.Container
	backFunc        func()
	selectionView   *fyne.Container
	gameView        *fyne.Container
	mainContent     *fyne.Container
	currentCountry  models.Feature
	countries       []models.Feature
	regionCountries []models.Feature
	currentIndex    int
	shapeCanvas     *fyne.Container
	islandContainer *fyne.Container
	guessEntry      *widget.Entry
	resultLabel     *widget.Label
	scoreLabel      *widget.Label
	progressLabel   *widget.Label
	score           int
	total           int
	selectedRegion  string
	currentCoords   [][][][]float64
	coordCache      map[int][][][][]float64
	cacheMutex      sync.RWMutex
}

func NewGame(backFunc func()) *Game {
	g := &Game{
		backFunc:  backFunc,
		countries: data.LoadGeoData().Features,
	}
	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	topBar := components.NewTopBar("Country Shape Game", g.backFunc, g.Reset)

	g.setupSelectionView()
	g.setupGameView()

	headerSection := container.NewVBox(
		topBar.GetContainer(),
		widget.NewSeparator(),
	)

	g.mainContent = container.NewMax(g.selectionView)

	g.content = container.NewBorder(
		headerSection, nil, nil, nil,
		g.mainContent,
	)
}

func (g *Game) setupSelectionView() {
	regionSelector := components.NewRegionSelector(
		"Select Region",
		"Choose a region and guess all country shapes!",
		g.startRegionGame,
	)
	g.selectionView = regionSelector.GetContainer()
}

func (g *Game) setupGameView() {
	g.scoreLabel = widget.NewLabel("Score: 0/0")
	g.progressLabel = widget.NewLabel("")

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder("Enter country name...")
	g.guessEntry.OnSubmitted = g.checkGuess

	guessBtn := widget.NewButton("Guess", func() { g.checkGuess(g.guessEntry.Text) })
	g.resultLabel = widget.NewLabel("")
	guessContainer := container.NewGridWithColumns(2, g.guessEntry, guessBtn)

	// Create main shape canvas and island container
	g.shapeCanvas = container.NewWithoutLayout()
	g.islandContainer = container.NewVBox()

	// Create main shape window
	mainShapeWindow := container.NewMax(g.shapeCanvas)

	// Layout with islands on left, main shape taking full width
	shapeWindow := container.NewBorder(nil, nil, g.islandContainer, nil, mainShapeWindow)

	topSection := container.NewVBox(
		g.scoreLabel,
		g.progressLabel,
		guessContainer,
		g.resultLabel,
		widget.NewSeparator(),
	)

	g.gameView = container.NewBorder(
		topSection, nil, nil, nil,
		shapeWindow,
	)

}

func (g *Game) drawShape(coords [][][][]float64) {
	g.currentCoords = coords
	g.shapeCanvas.RemoveAll()
	g.islandContainer.RemoveAll()

	if len(coords) == 0 {
		return
	}

	// Find largest polygon for main display
	mainPolygon, otherPolygons := g.separatePolygons(coords)

	// Draw main shape (largest polygon)
	if len(mainPolygon) > 0 {
		g.drawMainShape([][][][]float64{mainPolygon})
	}

	// Draw individual island windows if there are detached islands
	if len(otherPolygons) > 0 {
		for _, island := range otherPolygons {
			g.drawIslandWindow([][][][]float64{island})
		}
	}
}

func (g *Game) separatePolygons(coords [][][][]float64) ([][][]float64, [][][][]float64) {
	if len(coords) <= 1 {
		if len(coords) == 1 {
			return coords[0], nil
		}
		return nil, nil
	}

	// Find largest polygon by area
	largestIdx := 0
	largestArea := 0.0

	for i, polygon := range coords {
		if len(polygon) > 0 {
			area := g.calculatePolygonArea(polygon[0])
			if area > largestArea {
				largestArea = area
				largestIdx = i
			}
		}
	}

	// Separate main from others
	mainPolygon := coords[largestIdx]
	otherPolygons := make([][][][]float64, 0)
	for i, polygon := range coords {
		if i != largestIdx {
			otherPolygons = append(otherPolygons, polygon)
		}
	}

	return mainPolygon, otherPolygons
}

func (g *Game) calculatePolygonArea(ring [][]float64) float64 {
	if len(ring) < 3 {
		return 0
	}
	area := 0.0
	for i := 0; i < len(ring); i++ {
		j := (i + 1) % len(ring)
		area += ring[i][0] * ring[j][1]
		area -= ring[j][0] * ring[i][1]
	}
	return math.Abs(area) / 2.0
}

func (g *Game) drawMainShape(coords [][][][]float64) {
	minX, maxX, minY, maxY := g.calculateBounds(coords)
	if minX == maxX || minY == maxY {
		return
	}

	// Use dynamic canvas size for 100% width
	canvasSize := g.shapeCanvas.Size()
	width, height := float64(canvasSize.Width), float64(canvasSize.Height)
	if width == 0 || height == 0 {
		width, height = 600.0, 400.0
	}

	padding := 30.0

	scaleX := (width - 2*padding) / (maxX - minX)
	scaleY := (height - 2*padding) / (maxY - minY)
	scale := math.Min(scaleX, scaleY)

	shapeWidth := (maxX - minX) * scale
	shapeHeight := (maxY - minY) * scale
	offsetX := (width - shapeWidth) / 2
	offsetY := (height - shapeHeight) / 2

	raster := canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}

		for _, polygon := range coords {
			if len(polygon) > 0 {
				g.fillPolygon(img, polygon[0], minX, minY, scale, offsetX, offsetY, float64(h), color.RGBA{0, 0, 0, 255})
			}
		}

		return img
	})

	raster.Resize(fyne.NewSize(float32(width), float32(height)))
	g.shapeCanvas.Add(raster)
}

func (g *Game) drawIslandWindow(coords [][][][]float64) {
	minX, maxX, minY, maxY := g.calculateBounds(coords)
	if minX == maxX || minY == maxY {
		return
	}

	width, height := 150.0, 100.0
	padding := 8.0

	scaleX := (width - 2*padding) / (maxX - minX)
	scaleY := (height - 2*padding) / (maxY - minY)
	scale := math.Min(scaleX, scaleY)

	shapeWidth := (maxX - minX) * scale
	shapeHeight := (maxY - minY) * scale
	offsetX := (width - shapeWidth) / 2
	offsetY := (height - shapeHeight) / 2

	raster := canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))

		// White background
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}

		// Black border (2px)
		for i := 0; i < 2; i++ {
			for x := 0; x < w; x++ {
				img.Set(x, i, color.RGBA{0, 0, 0, 255})
				img.Set(x, h-1-i, color.RGBA{0, 0, 0, 255})
			}
			for y := 0; y < h; y++ {
				img.Set(i, y, color.RGBA{0, 0, 0, 255})
				img.Set(w-1-i, y, color.RGBA{0, 0, 0, 255})
			}
		}

		for _, polygon := range coords {
			if len(polygon) > 0 {
				g.fillPolygon(img, polygon[0], minX, minY, scale, offsetX, offsetY, float64(h), color.RGBA{0, 0, 0, 255})
			}
		}

		return img
	})

	raster.Resize(fyne.NewSize(float32(width), float32(height)))

	// Create container for this island and add to stack
	islandCanvas := container.NewWithoutLayout()
	islandCanvas.Add(raster)
	islandWindow := container.NewBorder(nil, nil, nil, nil, islandCanvas)
	islandWindow.Resize(fyne.NewSize(float32(width), float32(height)))

	g.islandContainer.Add(islandWindow)
}

func (g *Game) calculateBounds(coords [][][][]float64) (minX, maxX, minY, maxY float64) {
	first := true
	for _, polygon := range coords {
		for _, ring := range polygon {
			for _, point := range ring {
				if len(point) < 2 {
					continue
				}
				lon, lat := point[0], point[1]
				if first {
					minX, maxX, minY, maxY = lon, lon, lat, lat
					first = false
				} else {
					if lon < minX {
						minX = lon
					}
					if lon > maxX {
						maxX = lon
					}
					if lat < minY {
						minY = lat
					}
					if lat > maxY {
						maxY = lat
					}
				}
			}
		}
	}
	return
}

func (g *Game) fillPolygon(img *image.RGBA, ring [][]float64, minX, minY, scale, offsetX, offsetY, height float64, fillColor color.RGBA) {
	if len(ring) < 3 {
		return
	}

	// Convert coordinates to screen space
	points := make([][2]int, len(ring))
	for i, point := range ring {
		if len(point) < 2 {
			continue
		}
		x := int((point[0]-minX)*scale + offsetX)
		y := int(height - (point[1]-minY)*scale - offsetY)
		points[i] = [2]int{x, y}
	}

	// Simple polygon fill using scanline algorithm
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		intersections := []int{}

		// Find intersections with polygon edges
		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			p1, p2 := points[i], points[j]

			if (p1[1] <= y && y < p2[1]) || (p2[1] <= y && y < p1[1]) {
				x := p1[0] + (y-p1[1])*(p2[0]-p1[0])/(p2[1]-p1[1])
				intersections = append(intersections, x)
			}
		}

		// Sort intersections and fill between pairs
		sort.Ints(intersections)
		for i := 0; i < len(intersections); i += 2 {
			if i+1 < len(intersections) {
				for x := intersections[i]; x <= intersections[i+1]; x++ {
					if x >= bounds.Min.X && x < bounds.Max.X {
						img.Set(x, y, fillColor)
					}
				}
			}
		}
	}
}

func (g *Game) nextCountry() {
	for g.currentIndex < len(g.regionCountries) {
		idx := g.currentIndex
		g.currentCountry = g.regionCountries[idx]
		g.currentIndex++

		g.cacheMutex.RLock()
		coords, exists := g.coordCache[idx]
		g.cacheMutex.RUnlock()

		if !exists {
			coords = g.parseCoordinates(g.currentCountry.Geometry)
			g.cacheMutex.Lock()
			g.coordCache[idx] = coords
			g.cacheMutex.Unlock()
		}

		if len(coords) > 0 {
			g.drawShape(coords)
			g.guessEntry.SetText("")
			g.resultLabel.SetText("")
			return
		}
	}
}

func (g *Game) parseCoordinates(geom models.Geometry) [][][][]float64 {
	var coords [][][][]float64

	switch geom.Type {
	case "Polygon":
		if polygonCoords := g.parsePolygon(geom.Coordinates); len(polygonCoords) > 0 {
			coords = [][][][]float64{polygonCoords}
		}
	case "MultiPolygon":
		if multiCoords, ok := geom.Coordinates.([]interface{}); ok {
			for _, poly := range multiCoords {
				if polygonCoords := g.parsePolygon(poly); len(polygonCoords) > 0 {
					coords = append(coords, polygonCoords)
				}
			}
		}
	}

	return coords
}

func (g *Game) parsePolygon(coordsInterface interface{}) [][][]float64 {
	var polygonCoords [][][]float64

	if coordsArray, ok := coordsInterface.([]interface{}); ok {
		for _, ringInterface := range coordsArray {
			if ring := g.parseRing(ringInterface); len(ring) > 0 {
				polygonCoords = append(polygonCoords, ring)
			}
		}
	}

	return polygonCoords
}

func (g *Game) parseRing(ringInterface interface{}) [][]float64 {
	var ring [][]float64

	if ringArray, ok := ringInterface.([]interface{}); ok {
		for _, pointInterface := range ringArray {
			if pointArray, ok := pointInterface.([]interface{}); ok && len(pointArray) >= 2 {
				if lon, ok1 := pointArray[0].(float64); ok1 {
					if lat, ok2 := pointArray[1].(float64); ok2 {
						ring = append(ring, []float64{lon, lat})
					}
				}
			}
		}
	}

	return ring
}

func (g *Game) startRegionGame(region string) {
	g.selectedRegion = region
	g.regionCountries = []models.Feature{}
	g.coordCache = make(map[int][][][][]float64)

	// Filter countries by region
	for _, country := range g.countries {
		if region == "World" || country.Properties.Continent == region {
			g.regionCountries = append(g.regionCountries, country)
		}
	}

	g.score = 0
	g.total = 0
	g.currentIndex = 0
	g.updateProgress()

	g.mainContent.RemoveAll()
	g.mainContent.Add(g.gameView)
	g.mainContent.Refresh()

	// Preprocess coordinates in background
	go g.preprocessCoordinates()
	g.nextCountry()
}

func (g *Game) preprocessCoordinates() {
	for i := range g.regionCountries {
		g.cacheMutex.RLock()
		_, exists := g.coordCache[i]
		g.cacheMutex.RUnlock()

		if !exists {
			coords := g.parseCoordinates(g.regionCountries[i].Geometry)
			g.cacheMutex.Lock()
			g.coordCache[i] = coords
			g.cacheMutex.Unlock()
		}
	}
}

func (g *Game) checkGuess(guess string) {
	guess = strings.TrimSpace(guess)
	if guess == "" {
		return
	}

	g.total++

	if utils.MatchesCountryByName(guess, g.currentCountry.Properties.Name) {
		g.score++
		g.resultLabel.SetText("Correct! It's " + g.currentCountry.Properties.Name)
	} else {
		g.resultLabel.SetText("Wrong! It's " + g.currentCountry.Properties.Name)
	}

	g.updateProgress()
	g.guessEntry.Disable()

	// Check if all countries are done
	if g.currentIndex >= len(g.regionCountries)-1 {
		g.resultLabel.SetText(fmt.Sprintf("Game Complete! Final Score: %d/%d (%.1f%%)", g.score, g.total, float64(g.score)/float64(g.total)*100))
		return
	}

	// Auto-advance to next country after 2 seconds
	time.AfterFunc(2*time.Second, func() {
		g.guessEntry.Enable()
		g.nextCountry()
	})
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) updateProgress() {
	g.scoreLabel.SetText(fmt.Sprintf("Score: %d/%d", g.score, g.total))
	g.progressLabel.SetText(fmt.Sprintf("%s: Country %d/%d", g.selectedRegion, g.currentIndex+1, len(g.regionCountries)))
}

func (g *Game) showSelection() {
	g.mainContent.RemoveAll()
	g.mainContent.Add(g.selectionView)
	g.mainContent.Refresh()
}

func (g *Game) Start() {
	g.showSelection()
}

func (g *Game) Reset() {
	g.showSelection()
}
