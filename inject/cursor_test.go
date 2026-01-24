package inject

import "testing"

func TestCursorTracker_InitialPosition(t *testing.T) {
	tracker := NewCursorTracker()
	if tracker.Row != 1 || tracker.Col != 1 {
		t.Errorf("expected (1,1), got (%d,%d)", tracker.Row, tracker.Col)
	}
}

func TestCursorTracker_AbsolutePosition(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedRow int
		expectedCol int
	}{
		{"home", "\x1b[H", 1, 1},
		{"row only", "\x1b[5H", 5, 1},
		{"row and col", "\x1b[10;20H", 10, 20},
		{"f command", "\x1b[3;4f", 3, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewCursorTracker()
			tracker.Row = 99
			tracker.Col = 99
			tracker.Process([]byte(tt.input))
			if tracker.Row != tt.expectedRow || tracker.Col != tt.expectedCol {
				t.Errorf("expected (%d,%d), got (%d,%d)",
					tt.expectedRow, tt.expectedCol, tracker.Row, tracker.Col)
			}
		})
	}
}

func TestCursorTracker_RelativeMovement(t *testing.T) {
	tests := []struct {
		name        string
		startRow    int
		startCol    int
		input       string
		expectedRow int
		expectedCol int
	}{
		{"up 1", 10, 5, "\x1b[A", 9, 5},
		{"up 3", 10, 5, "\x1b[3A", 7, 5},
		{"down 1", 10, 5, "\x1b[B", 11, 5},
		{"down 5", 10, 5, "\x1b[5B", 15, 5},
		{"right 1", 10, 5, "\x1b[C", 10, 6},
		{"left 1", 10, 5, "\x1b[D", 10, 4},
		{"next line", 10, 5, "\x1b[E", 11, 1},
		{"prev line", 10, 5, "\x1b[F", 9, 1},
		{"col absolute", 10, 5, "\x1b[20G", 10, 20},
		{"row absolute", 10, 5, "\x1b[3d", 3, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewCursorTracker()
			tracker.Row = tt.startRow
			tracker.Col = tt.startCol
			tracker.Process([]byte(tt.input))
			if tracker.Row != tt.expectedRow || tracker.Col != tt.expectedCol {
				t.Errorf("expected (%d,%d), got (%d,%d)",
					tt.expectedRow, tt.expectedCol, tracker.Row, tracker.Col)
			}
		})
	}
}

func TestCursorTracker_Newline(t *testing.T) {
	tracker := NewCursorTracker()
	tracker.Row = 5
	tracker.Col = 10
	tracker.Process([]byte("\n"))
	if tracker.Row != 6 || tracker.Col != 1 {
		t.Errorf("expected (6,1), got (%d,%d)", tracker.Row, tracker.Col)
	}
}

func TestCursorTracker_CarriageReturn(t *testing.T) {
	tracker := NewCursorTracker()
	tracker.Row = 5
	tracker.Col = 10
	tracker.Process([]byte("\r"))
	if tracker.Row != 5 || tracker.Col != 1 {
		t.Errorf("expected (5,1), got (%d,%d)", tracker.Row, tracker.Col)
	}
}

func TestCursorTracker_SpinnerDetection(t *testing.T) {
	spinners := []string{"✢", "✶", "✻", "✸", "✹", "✺", "✷"}

	for _, s := range spinners {
		t.Run(s, func(t *testing.T) {
			tracker := NewCursorTracker()
			tracker.Row = 15
			row, found := tracker.Process([]byte(s))
			if !found {
				t.Error("spinner should be detected")
			}
			if row != 15 {
				t.Errorf("expected row 15, got %d", row)
			}
		})
	}
}

func TestCursorTracker_SpinnerAfterPosition(t *testing.T) {
	tracker := NewCursorTracker()
	row, found := tracker.Process([]byte("\x1b[22;1H✻ Working..."))
	if !found {
		t.Error("spinner should be detected")
	}
	if row != 22 {
		t.Errorf("expected row 22, got %d", row)
	}
}

func TestCursorTracker_NoSpinner(t *testing.T) {
	tracker := NewCursorTracker()
	_, found := tracker.Process([]byte("regular text without spinner"))
	if found {
		t.Error("no spinner should be detected")
	}
}

func TestCursorTracker_MixedContent(t *testing.T) {
	tracker := NewCursorTracker()
	data := []byte("\x1b[H\x1b[2J\x1b[10;5Hsome text\n\x1b[15;1H✶ Thinking...")
	row, found := tracker.Process(data)
	if !found {
		t.Error("spinner should be detected")
	}
	if row != 15 {
		t.Errorf("expected row 15, got %d", row)
	}
}

func TestCursorTracker_BoundsCheck(t *testing.T) {
	tracker := NewCursorTracker()
	tracker.Row = 1
	tracker.Process([]byte("\x1b[10A"))
	if tracker.Row != 1 {
		t.Errorf("row should not go below 1, got %d", tracker.Row)
	}

	tracker.Col = 1
	tracker.Process([]byte("\x1b[10D"))
	if tracker.Col != 1 {
		t.Errorf("col should not go below 1, got %d", tracker.Col)
	}
}

func TestCursorTracker_PrivateMode(t *testing.T) {
	tracker := NewCursorTracker()
	tracker.Row = 5
	tracker.Process([]byte("\x1b[?25h"))
	if tracker.Row != 5 {
		t.Errorf("private mode should not change position, got row %d", tracker.Row)
	}
}
