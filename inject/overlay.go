package inject

import "fmt"

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
