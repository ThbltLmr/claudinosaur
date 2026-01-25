package inject

import (
	"log"
	"unicode/utf8"
)

var spinnerRunes = []rune{'✢', '✶', '✻', '✸', '✹', '✺', '✷'}

type CursorTracker struct {
	Row int
	Col int
}

func NewCursorTracker() *CursorTracker {
	return &CursorTracker{Row: 1, Col: 1}
}

func (t *CursorTracker) Process(data []byte) (spinnerRow int, hasSpinner bool) {
	i := 0
	for i < len(data) {
		if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '[' {
			seqEnd := t.parseCSI(data[i:])
			i += seqEnd
			continue
		}

		if data[i] == '\n' {
			t.Row++
			t.Col = 1
			i++
			continue
		}

		if data[i] == '\r' {
			t.Col = 1
			i++
			continue
		}

		r, size := utf8.DecodeRune(data[i:])
		if r != utf8.RuneError {
			for _, s := range spinnerRunes {
				if r == s {
					hasSpinner = true
					spinnerRow = t.Row
					if DebugLog != nil {
						DebugLog.Printf("[CURSOR] spinner %q found at row %d", string(r), t.Row)
					}
					break
				}
			}
			t.Col++
		}
		i += size
	}
	return spinnerRow, hasSpinner
}

func (t *CursorTracker) parseCSI(data []byte) int {
	if len(data) < 3 || data[0] != 0x1b || data[1] != '[' {
		return 1
	}

	i := 2
	params := make([]int, 0, 2)
	currentParam := 0
	hasParam := false

	for i < len(data) {
		b := data[i]

		if b >= '0' && b <= '9' {
			currentParam = currentParam*10 + int(b-'0')
			hasParam = true
			i++
			continue
		}

		if b == ';' {
			params = append(params, currentParam)
			currentParam = 0
			hasParam = false
			i++
			continue
		}

		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
			if hasParam {
				params = append(params, currentParam)
			}
			t.applyCSI(b, params)
			return i + 1
		}

		if b == '?' {
			i++
			continue
		}

		return i + 1
	}

	return len(data)
}

func (t *CursorTracker) applyCSI(cmd byte, params []int) {
	n := 1
	if len(params) > 0 && params[0] > 0 {
		n = params[0]
	}

	switch cmd {
	case 'H', 'f':
		row := 1
		col := 1
		if len(params) > 0 && params[0] > 0 {
			row = params[0]
		}
		if len(params) > 1 && params[1] > 0 {
			col = params[1]
		}
		t.Row = row
		t.Col = col
		if DebugLog != nil {
			DebugLog.Printf("[CURSOR] move to row=%d col=%d", row, col)
		}

	case 'A':
		t.Row -= n
		if t.Row < 1 {
			t.Row = 1
		}

	case 'B':
		t.Row += n

	case 'C':
		t.Col += n

	case 'D':
		t.Col -= n
		if t.Col < 1 {
			t.Col = 1
		}

	case 'E':
		t.Row += n
		t.Col = 1

	case 'F':
		t.Row -= n
		if t.Row < 1 {
			t.Row = 1
		}
		t.Col = 1

	case 'G':
		t.Col = n

	case 'd':
		t.Row = n
	}
}

func SetDebugLog(logger *log.Logger) {
	DebugLog = logger
}
