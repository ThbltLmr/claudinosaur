package inject

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRenderOverlay_BasicOutput(t *testing.T) {
	content := "=== GAME LINE ==="
	result := RenderOverlay(content, 3, 24)

	if result == nil {
		t.Fatal("expected output, got nil")
	}

	if !bytes.Contains(result, []byte(content)) {
		t.Errorf("output should contain content: %q", result)
	}

	if !bytes.Contains(result, []byte("\x1b[s")) {
		t.Error("should contain save cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[u")) {
		t.Error("should contain restore cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[21;1H")) {
		t.Errorf("should position at row 21 (24-3), got: %q", result)
	}
}

func TestRenderOverlay_RowCalculation(t *testing.T) {
	tests := []struct {
		rowFromBottom int
		termHeight    int
		expectedRow   int
	}{
		{1, 24, 23},
		{3, 24, 21},
		{5, 30, 25},
		{0, 10, 10},
	}

	for _, tt := range tests {
		result := RenderOverlay("test", tt.rowFromBottom, tt.termHeight)
		expected := []byte(fmt.Sprintf("\x1b[%d;1H", tt.expectedRow))
		if !bytes.Contains(result, expected) {
			t.Errorf("rowFromBottom=%d, termHeight=%d: expected row %d, got %q",
				tt.rowFromBottom, tt.termHeight, tt.expectedRow, result)
		}
	}
}

func TestRenderOverlay_InvalidHeight(t *testing.T) {
	result := RenderOverlay("test", 10, 5)
	if result != nil {
		t.Errorf("should return nil when termHeight <= rowFromBottom, got %q", result)
	}
}

func TestRenderOverlay_ClearsLine(t *testing.T) {
	result := RenderOverlay("content", 1, 10)
	if !bytes.Contains(result, []byte("\x1b[2K")) {
		t.Error("should contain clear line sequence")
	}
}
