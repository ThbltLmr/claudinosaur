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

func TestRenderMultiLineOverlay_TwoLines(t *testing.T) {
	lines := []string{"sky line", "ground line"}
	result := RenderMultiLineOverlay(lines, 10)

	if result == nil {
		t.Fatal("expected output, got nil")
	}

	if !bytes.Contains(result, []byte("\x1b[s")) {
		t.Error("should contain save cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[u")) {
		t.Error("should contain restore cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[10;1H\x1b[2Ksky line")) {
		t.Errorf("should contain first line at row 10, got: %q", result)
	}

	if !bytes.Contains(result, []byte("\x1b[11;1H\x1b[2Kground line")) {
		t.Errorf("should contain second line at row 11, got: %q", result)
	}
}

func TestRenderMultiLineOverlay_SingleLine(t *testing.T) {
	lines := []string{"only line"}
	result := RenderMultiLineOverlay(lines, 5)

	if result == nil {
		t.Fatal("expected output, got nil")
	}

	if !bytes.Contains(result, []byte("\x1b[5;1H\x1b[2Konly line")) {
		t.Errorf("should contain line at row 5, got: %q", result)
	}
}

func TestRenderMultiLineOverlay_InvalidInputs(t *testing.T) {
	if result := RenderMultiLineOverlay(nil, 5); result != nil {
		t.Errorf("should return nil for nil lines, got: %q", result)
	}

	if result := RenderMultiLineOverlay([]string{}, 5); result != nil {
		t.Errorf("should return nil for empty lines, got: %q", result)
	}

	if result := RenderMultiLineOverlay([]string{"test"}, 0); result != nil {
		t.Errorf("should return nil for row 0, got: %q", result)
	}

	if result := RenderMultiLineOverlay([]string{"test"}, -1); result != nil {
		t.Errorf("should return nil for negative row, got: %q", result)
	}
}

func TestClearMultipleRows_Basic(t *testing.T) {
	result := ClearMultipleRows(10, 2)

	if result == nil {
		t.Fatal("expected output, got nil")
	}

	if !bytes.Contains(result, []byte("\x1b[s")) {
		t.Error("should contain save cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[u")) {
		t.Error("should contain restore cursor sequence")
	}

	if !bytes.Contains(result, []byte("\x1b[10;1H\x1b[2K")) {
		t.Errorf("should clear row 10, got: %q", result)
	}

	if !bytes.Contains(result, []byte("\x1b[11;1H\x1b[2K")) {
		t.Errorf("should clear row 11, got: %q", result)
	}
}

func TestClearMultipleRows_InvalidInputs(t *testing.T) {
	if result := ClearMultipleRows(0, 2); result != nil {
		t.Errorf("should return nil for row 0, got: %q", result)
	}

	if result := ClearMultipleRows(-1, 2); result != nil {
		t.Errorf("should return nil for negative row, got: %q", result)
	}

	if result := ClearMultipleRows(5, 0); result != nil {
		t.Errorf("should return nil for count 0, got: %q", result)
	}

	if result := ClearMultipleRows(5, -1); result != nil {
		t.Errorf("should return nil for negative count, got: %q", result)
	}
}
