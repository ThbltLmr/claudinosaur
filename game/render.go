package game

import (
	"fmt"
	"strings"
)

const (
	DinoEmoji     = "🦕"
	DeadEmoji     = "💀"
	ObstacleEmoji = "🌵"
	CloudEmoji    = "☁️"
)

func Render(s State, width int) (skyLine, groundLine string) {
	if width < 20 {
		return CloudEmoji, DinoEmoji
	}

	skyLine = renderSkyLine(s, width)
	groundLine = renderGroundLine(s, width)
	return skyLine, groundLine
}

func renderSkyLine(s State, width int) string {
	emojiCount := 0
	cloudPositions := []int{4, 20, 45}
	for _, pos := range cloudPositions {
		if pos < width-2 {
			emojiCount++
		}
	}
	if s.IsInAir && !s.GameOver {
		emojiCount++
	}

	effectiveWidth := width - emojiCount

	line := make([]rune, effectiveWidth)
	for i := range line {
		line[i] = ' '
	}

	for _, pos := range cloudPositions {
		if pos < effectiveWidth-2 {
			placeEmoji(line, pos, CloudEmoji)
		}
	}

	if s.IsInAir && !s.GameOver {
		placeEmoji(line, 0, DinoEmoji)
	}

	return string(line)
}

func renderGroundLine(s State, width int) string {
	scoreStr := formatScore(s)
	scoreLen := len(scoreStr)

	emojiCount := 0
	if !s.IsInAir {
		emojiCount++
	}
	for _, x := range s.Obstacles {
		if int(x) >= 0 && int(x) < width {
			emojiCount++
		}
	}

	effectiveWidth := width - emojiCount

	line := make([]rune, effectiveWidth)
	for i := range line {
		line[i] = ' '
	}

	if !s.IsInAir {
		emoji := DinoEmoji
		if s.GameOver {
			emoji = DeadEmoji
		}
		placeEmoji(line, 0, emoji)
	}

	for _, x := range s.Obstacles {
		pos := int(x)
		if pos >= 0 && pos < effectiveWidth-2 {
			placeEmoji(line, pos, ObstacleEmoji)
		}
	}

	scoreStart := effectiveWidth - scoreLen
	if scoreStart > 0 {
		copy(line[scoreStart:], []rune(scoreStr))
	}

	return string(line)
}

func formatScore(s State) string {
	base := fmt.Sprintf("Score: %05d HI: %05d", s.Score, s.HighScore)
	if s.GameOver {
		return base + " [R]estart"
	}
	return base
}

func placeEmoji(line []rune, pos int, emoji string) {
	runes := []rune(emoji)
	for i, r := range runes {
		if pos+i < len(line) {
			line[pos+i] = r
		}
	}
}

func FormatCountdown(skyLine string, countdown int, width int) string {
	if countdown <= 0 || width < 10 {
		return skyLine
	}
	countdownStr := fmt.Sprintf("▶ %d ◀", countdown)
	center := (width - len(countdownStr)) / 2
	if center < 0 {
		center = 0
	}

	line := []rune(skyLine)
	for len(line) < width {
		line = append(line, ' ')
	}

	runes := []rune(countdownStr)
	for i, r := range runes {
		if center+i < len(line) {
			line[center+i] = r
		}
	}
	return strings.TrimRight(string(line), " ")
}
