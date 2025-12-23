package screens

import (
	"flagged-it/internal/ui/components"
	"flagged-it/internal/utils"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

type Scoreboard struct {
	content  *fyne.Container
	backFunc func()
	window   fyne.Window
}

func NewScoreboard(backFunc func(), window fyne.Window) *Scoreboard {
	s := &Scoreboard{
		backFunc: backFunc,
		window:   window,
	}
	s.setupUI()
	return s
}

func (s *Scoreboard) setupUI() {
	topBar := components.NewTopBar(lang.X("scoreboard.title", "My Best Scores"), s.backFunc, nil)

	// Get all scores
	allScores := utils.GetScoreboard()

	// Group by game mode
	scoresByGame := make(map[string][]utils.ScoreEntry)
	for _, score := range allScores {
		scoresByGame[score.GameMode] = append(scoresByGame[score.GameMode], score)
	}

	// Create sections for each game mode
	var sections []fyne.CanvasObject

	if len(allScores) == 0 {
		// Empty state
		emptyLabel := widget.NewLabel(lang.X("scoreboard.empty", "No scores yet! Play some games to see your progress here."))
		emptyLabel.Alignment = fyne.TextAlignCenter
		emptyLabel.Wrapping = fyne.TextWrapWord

		sections = append(sections, container.NewPadded(emptyLabel))
	} else {
		// Add section for each game mode
		gameModes := []string{"flag", "shape", "hangman", "facts", "list", "higher_lower", "guessing"}
		gameNames := map[string]string{
			"flag":         lang.X("game.flag.title", "Guess by Flag"),
			"shape":        lang.X("game.shape.title", "Guess by Shape"),
			"hangman":      lang.X("game.hangman.title", "Hangman"),
			"facts":        lang.X("game.facts.title", "Guess by Facts"),
			"list":         lang.X("game.list.title", "List All Countries"),
			"higher_lower": lang.X("game.higher_lower.title", "Higher or Lower"),
			"guessing":     lang.X("game.guessing.title", "What Country is This"),
		}

		for _, gameMode := range gameModes {
			scores, exists := scoresByGame[gameMode]
			if !exists || len(scores) == 0 {
				continue
			}

			// Game title
			gameTitle := widget.NewLabel(gameNames[gameMode])
			gameTitle.TextStyle = fyne.TextStyle{Bold: true}

			// Create table for this game's scores
			scoreTable := s.createScoreTable(scores)

			// Separator
			separator := components.NewDashedSeparator(color.RGBA{100, 100, 100, 255}, 2)

			sections = append(sections,
				gameTitle,
				scoreTable,
				separator,
			)
		}
	}

	// Scrollable content
	scrollContent := container.NewVBox(sections...)
	scroll := container.NewScroll(scrollContent)

	s.content = container.NewBorder(
		topBar.GetContainer(),
		nil, nil, nil,
		scroll,
	)
}

func (s *Scoreboard) createScoreTable(scores []utils.ScoreEntry) *fyne.Container {
	// Limit to top 10 per game
	if len(scores) > 10 {
		scores = scores[:10]
	}

	rows := []fyne.CanvasObject{}

	// Header row
	headerRank := widget.NewLabel("#")
	headerRank.TextStyle = fyne.TextStyle{Bold: true}

	headerScore := widget.NewLabel(lang.X("scoreboard.score", "Score"))
	headerScore.TextStyle = fyne.TextStyle{Bold: true}

	headerPercent := widget.NewLabel(lang.X("scoreboard.percent", "Percent"))
	headerPercent.TextStyle = fyne.TextStyle{Bold: true}

	headerDate := widget.NewLabel(lang.X("scoreboard.date", "Date"))
	headerDate.TextStyle = fyne.TextStyle{Bold: true}

	headerRow := container.NewGridWithColumns(4,
		headerRank,
		headerScore,
		headerPercent,
		headerDate,
	)
	rows = append(rows, headerRow)

	// Data rows
	for i, score := range scores {
		rank := widget.NewLabel(fmt.Sprintf("%d", i+1))

		scoreText := widget.NewLabel(fmt.Sprintf("%d/%d", score.Score, score.Total))

		percentText := widget.NewLabel(fmt.Sprintf("%.1f%%", score.Percent))

		dateText := widget.NewLabel(score.Date.Format("Jan 2, 15:04"))

		row := container.NewGridWithColumns(4,
			rank,
			scoreText,
			percentText,
			dateText,
		)
		rows = append(rows, row)
	}

	return container.NewVBox(rows...)
}

func (s *Scoreboard) GetContent() *fyne.Container {
	return s.content
}
