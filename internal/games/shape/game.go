package shape

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"flagged-it/internal/data"
	"flagged-it/internal/data/models"
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	canvasWidth  = 1000.0
	canvasHeight = 700.0
)

type Game struct {
	content         *fyne.Container
	backFunc        func()
	selectionView   *fyne.Container
	gameView        *fyne.Container
	mainContent     *fyne.Container
	currentCountry  models.Country
	countries       []models.Country
	regionCountries []models.Country
	currentIndex    int
	shapeCanvas     *fyne.Container
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
	scoreManager    *utils.ScoreManager
}

func NewGame(backFunc func(), scoreManager *utils.ScoreManager) *Game {
	g := &Game{
		backFunc:     backFunc,
		scoreManager: scoreManager,
		countries:    data.LoadCountries(),
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
	availableRegions := g.getAvailableRegions()
	regionSelector := components.NewRegionSelector(
		lang.X("game.shape.select_region", "Select Region"),
		lang.X("game.shape.choose_region", "Choose a region and guess all country shapes!"),
		availableRegions,
		g.startRegionGame,
	)
	g.selectionView = regionSelector.GetContainer()
}

func (g *Game) getAvailableRegions() []string {
	regionMap := make(map[string]bool)
	regionMap["World"] = true

	for _, country := range g.countries {
		if country.Region != "" {
			regionMap[country.Region] = true
		}
	}

	var regions []string
	for region := range regionMap {
		regions = append(regions, region)
	}

	sort.Slice(regions, func(i, j int) bool {
		if regions[i] == "World" {
			return true
		}
		if regions[j] == "World" {
			return false
		}
		return regions[i] < regions[j]
	})

	return regions
}

func (g *Game) setupGameView() {
	g.scoreLabel = widget.NewLabel(fmt.Sprintf(lang.X("game.shape.score", "Score: %d/%d"), 0, 0))
	g.progressLabel = widget.NewLabel("")

	g.guessEntry = widget.NewEntry()
	g.guessEntry.SetPlaceHolder(lang.X("game.shape.enter_country", "Enter country name..."))
	g.guessEntry.OnSubmitted = g.checkGuess

	guessBtn := widget.NewButton(lang.X("game.shape.guess", "Guess"), func() { g.checkGuess(g.guessEntry.Text) })
	g.resultLabel = widget.NewLabel("")
	guessContainer := container.NewBorder(
		nil, nil,
		guessBtn, nil,
		g.guessEntry,
	)

	g.shapeCanvas = container.NewWithoutLayout()
	shapeWindow := container.NewMax(g.shapeCanvas)

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

	if len(coords) == 0 {
		return
	}

	g.drawMainShape(coords)
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

	raster := canvas.NewRaster(func(w, h int) image.Image {
		img := image.NewRGBA(image.Rect(0, 0, w, h))

		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}

		scaleX := float64(w) / (maxX - minX)
		scaleY := float64(h) / (maxY - minY)
		scale := math.Min(scaleX, scaleY) * 0.9

		shapeWidth := (maxX - minX) * scale
		shapeHeight := (maxY - minY) * scale
		offsetX := (float64(w) - shapeWidth) / 2
		offsetY := (float64(h) - shapeHeight) / 2

		for _, polygon := range coords {
			if len(polygon) > 0 {
				g.fillPolygon(img, polygon[0], minX, minY, scale, offsetX, offsetY, float64(h))
			}
		}

		return img
	})

	canvasSize := g.shapeCanvas.Size()
	if canvasSize.Width > 0 && canvasSize.Height > 0 {
		raster.Resize(canvasSize)
	} else {
		raster.Resize(fyne.NewSize(float32(canvasWidth*0.9), float32(canvasHeight*0.9)))
	}
	g.shapeCanvas.Add(raster)
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

func (g *Game) fillPolygon(img *image.RGBA, ring [][]float64, minX, minY, scale, offsetX, offsetY, height float64) {
	if len(ring) < 3 {
		return
	}

	points := make([][2]int, len(ring))
	for i, point := range ring {
		if len(point) < 2 {
			continue
		}
		x := int((point[0]-minX)*scale + offsetX)
		y := int(height - (point[1]-minY)*scale - offsetY)
		points[i] = [2]int{x, y}
	}

	var fillColor color.RGBA
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		fillColor = color.RGBA{255, 255, 255, 255}
	} else {
		fillColor = color.RGBA{0, 0, 0, 255}
	}

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		intersections := []int{}

		for i := 0; i < len(points); i++ {
			j := (i + 1) % len(points)
			p1, p2 := points[i], points[j]

			if (p1[1] <= y && y < p2[1]) || (p2[1] <= y && y < p1[1]) {
				x := p1[0] + (y-p1[1])*(p2[0]-p1[0])/(p2[1]-p1[1])
				intersections = append(intersections, x)
			}
		}

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

		if g.currentCountry.CCA3 == "" {
			continue
		}

		g.cacheMutex.RLock()
		coords, exists := g.coordCache[idx]
		g.cacheMutex.RUnlock()

		if !exists {
			geoData, err := data.LoadGeoData(g.currentCountry.CCA3)
			if err == nil && len(geoData.Features) > 0 {
				coords = g.parseCoordinates(geoData.Features[0].Geometry)
			}
			g.cacheMutex.Lock()
			g.coordCache[idx] = coords
			g.cacheMutex.Unlock()
		}

		if len(coords) > 0 {
			g.total++
			g.drawShape(coords)
			g.guessEntry.SetText("")
			g.resultLabel.SetText("")
			g.updateProgress()
			return
		}
	}

	if g.total > 0 {
		g.scoreManager.SetTotal("shape", g.total)
		g.scoreManager.UpdateScore("shape", g.score)
		g.resultLabel.SetText(fmt.Sprintf(lang.X("game.shape.complete", "Game Complete! Final Score: %d/%d (%.1f%%)"), g.score, g.total, float64(g.score)/float64(g.total)*100))
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
	g.regionCountries = []models.Country{}
	g.coordCache = make(map[int][][][][]float64)

	for _, country := range g.countries {
		if region == "World" || country.Region == region {
			g.regionCountries = append(g.regionCountries, country)
		}
	}

	rand.Shuffle(len(g.regionCountries), func(i, j int) {
		g.regionCountries[i], g.regionCountries[j] = g.regionCountries[j], g.regionCountries[i]
	})

	g.score = 0
	g.total = 0
	g.currentIndex = 0

	g.mainContent.RemoveAll()
	g.mainContent.Add(g.gameView)
	g.mainContent.Refresh()

	go g.preprocessCoordinates()
	g.nextCountry()
}

func (g *Game) preprocessCoordinates() {
	for i := range g.regionCountries {
		g.cacheMutex.RLock()
		_, exists := g.coordCache[i]
		g.cacheMutex.RUnlock()

		if !exists && g.regionCountries[i].CCA3 != "" {
			geoData, err := data.LoadGeoData(g.regionCountries[i].CCA3)
			if err == nil && len(geoData.Features) > 0 {
				coords := g.parseCoordinates(geoData.Features[0].Geometry)
				g.cacheMutex.Lock()
				g.coordCache[i] = coords
				g.cacheMutex.Unlock()
			}
		}
	}
}

func (g *Game) checkGuess(guess string) {
	guess = strings.TrimSpace(guess)
	if guess == "" {
		return
	}

	if utils.MatchCountry(guess, g.currentCountry, utils.MatchAll) {
		g.score++
		g.resultLabel.SetText(fmt.Sprintf(lang.X("game.shape.correct", "Correct! It's %s"), g.currentCountry.Name.Common))
	} else {
		g.resultLabel.SetText(fmt.Sprintf(lang.X("game.shape.wrong", "Wrong! It's %s"), g.currentCountry.Name.Common))
	}

	g.guessEntry.Disable()

	time.AfterFunc(2*time.Second, func() {
		fyne.Do(func() {
			g.guessEntry.Enable()
			g.nextCountry()
		})
	})
}

func (g *Game) GetContent() *fyne.Container {
	return g.content
}

func (g *Game) updateProgress() {
	g.scoreLabel.SetText(fmt.Sprintf(lang.X("game.shape.score", "Score: %d/%d"), g.score, g.total))
	g.progressLabel.SetText(fmt.Sprintf(lang.X("game.shape.progress", "%s: Country %d/%d"), g.selectedRegion, g.currentIndex, g.total))
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
