package inject

import (
	"bytes"
	"fmt"
)

func RenderOverlay(content string, rowFromBottom int, termHeight int) []byte {
	if termHeight <= rowFromBottom {
		return nil
	}
	targetRow := termHeight - rowFromBottom
	return RenderOverlayAtRow(content, targetRow)
}

func RenderOverlayAtRow(content string, row int) []byte {
	if row < 1 {
		return nil
	}
	return []byte(fmt.Sprintf(
		"\x1b[s\x1b[%d;1H\x1b[2K%s\x1b[u",
		row, content,
	))
}

func RenderMultiLineOverlay(lines []string, startRow int) []byte {
	if startRow < 1 || len(lines) == 0 {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString("\x1b[s")
	for i, line := range lines {
		row := startRow + i
		buf.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K%s", row, line))
	}
	buf.WriteString("\x1b[u")
	return buf.Bytes()
}

func ClearMultipleRows(startRow, count int) []byte {
	if startRow < 1 || count <= 0 {
		return nil
	}
	var buf bytes.Buffer
	buf.WriteString("\x1b[s")
	for i := 0; i < count; i++ {
		buf.WriteString(fmt.Sprintf("\x1b[%d;1H\x1b[2K", startRow+i))
	}
	buf.WriteString("\x1b[u")
	return buf.Bytes()
}
