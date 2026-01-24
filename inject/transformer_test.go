package inject

import (
	"bytes"
	"testing"
	"time"
)

func TestTransform_PassthroughMode(t *testing.T) {
	state := TransformState{}
	chunk := []byte("hello world")

	result := Transform(state, chunk, Passthrough, 0)

	if !bytes.Equal(result.Output, chunk) {
		t.Errorf("expected %q, got %q", chunk, result.Output)
	}
	if len(result.NewState.Buffer) != 0 {
		t.Error("buffer should be empty in passthrough mode")
	}
}

func TestTransform_PassthroughFlushesBuffer(t *testing.T) {
	state := TransformState{Buffer: []byte("buffered ")}
	chunk := []byte("new data")

	result := Transform(state, chunk, Passthrough, 0)

	expected := []byte("buffered new data")
	if !bytes.Equal(result.Output, expected) {
		t.Errorf("expected %q, got %q", expected, result.Output)
	}
}

func TestTransform_GameActive_NoPattern(t *testing.T) {
	state := TransformState{}
	chunk := []byte("some text without pattern")

	result := Transform(state, chunk, GameActive, 10*time.Millisecond)

	if result.Output != nil {
		t.Error("should buffer when pattern not found")
	}
	if !bytes.Equal(result.NewState.Buffer, chunk) {
		t.Errorf("buffer should contain chunk: %q", result.NewState.Buffer)
	}
}

func TestTransform_GameActive_FlushOnTimeout(t *testing.T) {
	state := TransformState{
		Buffer:    []byte("buffered content"),
		BufferAge: 40 * time.Millisecond,
	}
	chunk := []byte(" more")

	result := Transform(state, chunk, GameActive, 15*time.Millisecond)

	expected := []byte("buffered content more")
	if !bytes.Equal(result.Output, expected) {
		t.Errorf("expected flush after timeout, got %q", result.Output)
	}
	if len(result.NewState.Buffer) != 0 {
		t.Error("buffer should be cleared after flush")
	}
}

func TestTransform_GameActive_FlushOnMaxBufferSize(t *testing.T) {
	largeBuffer := make([]byte, MaxBufferSize+1)
	for i := range largeBuffer {
		largeBuffer[i] = 'x'
	}
	state := TransformState{Buffer: largeBuffer}

	result := Transform(state, []byte("more"), GameActive, 0)

	if result.Output == nil {
		t.Error("should flush when buffer exceeds max size")
	}
	if len(result.NewState.Buffer) != 0 {
		t.Error("buffer should be cleared after max size flush")
	}
}

func TestTransform_GameActive_InjectsOnSpinnerPattern(t *testing.T) {
	state := TransformState{}
	// Spinner character followed by text and closing paren
	chunk := []byte("✻ Computing… (Esc to interrupt)\nMore content")

	result := Transform(state, chunk, GameActive, 0)

	if !bytes.Contains(result.Output, gameLinePlaceholder) {
		t.Errorf("game lines should be injected, got: %q", result.Output)
	}
	if !bytes.Contains(result.Output, []byte("✻ Computing… (Esc to interrupt)")) {
		t.Error("spinner line should be preserved")
	}
}

func TestTransform_GameActive_PatternAcrossChunks(t *testing.T) {
	state := TransformState{}

	// First chunk: spinner without paren
	result1 := Transform(state, []byte("✻ Computing…"), GameActive, 10*time.Millisecond)
	if result1.Output != nil {
		t.Error("should buffer first chunk")
	}

	// Second chunk: closing paren
	result2 := Transform(result1.NewState, []byte(" (done) rest"), GameActive, 10*time.Millisecond)

	if !bytes.Contains(result2.Output, gameLinePlaceholder) {
		t.Errorf("should inject when pattern completes across chunks, got: %q", result2.Output)
	}
}

func TestTransform_NoSpinnerNoInjection(t *testing.T) {
	state := TransformState{}
	chunk := []byte("Some text (with parens)\nSome other text\n")

	result := Transform(state, chunk, GameActive, 60*time.Millisecond)

	if bytes.Contains(result.Output, gameLinePlaceholder) {
		t.Error("should not inject without spinner")
	}
}

func TestTransform_SpinnerWithoutParenNoInjection(t *testing.T) {
	state := TransformState{}
	chunk := []byte("✻ Computing… no closing paren here\n")

	result := Transform(state, chunk, GameActive, 60*time.Millisecond)

	// Should flush due to timeout without injecting
	if bytes.Contains(result.Output, gameLinePlaceholder) {
		t.Error("should not inject without closing paren")
	}
}

func TestFindInjectionPoint_FindsSpinnerParen(t *testing.T) {
	data := []byte("✻ Computing… (Esc to interrupt)\nrest")

	idx := findInjectionPoint(data)

	if idx < 0 {
		t.Fatal("should find injection point")
	}
	// Should inject after the closing paren
	if !bytes.HasPrefix(data[idx:], []byte("\nrest")) {
		t.Errorf("injection point should be after paren, got: %q", data[idx:])
	}
}

func TestFindInjectionPoint_NoPattern(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{"no spinner", []byte("text (with parens)\n")},
		{"empty", []byte{}},
		{"single line no spinner", []byte("text")},
		{"spinner without paren", []byte("✻ Computing… no paren")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			idx := findInjectionPoint(tc.data)
			if idx >= 0 {
				t.Errorf("should not find injection point in %q", tc.data)
			}
		})
	}
}

func TestTransform_PreservesContentOrder(t *testing.T) {
	state := TransformState{}
	chunk := []byte("✻ Computing… (done)\nafter")

	result := Transform(state, chunk, GameActive, 0)

	spinnerIdx := bytes.Index(result.Output, []byte("✻"))
	gameIdx := bytes.Index(result.Output, gameLinePlaceholder)
	afterIdx := bytes.Index(result.Output, []byte("after"))

	if spinnerIdx < 0 || gameIdx < 0 || afterIdx < 0 {
		t.Fatalf("missing expected content in output: %q", result.Output)
	}
	if !(spinnerIdx < gameIdx && gameIdx < afterIdx) {
		t.Errorf("content order incorrect: spinner=%d, game=%d, after=%d", spinnerIdx, gameIdx, afterIdx)
	}
}

func TestTransform_MultipleInjections(t *testing.T) {
	state := TransformState{}

	chunk1 := []byte("✻ First (done)\n")
	result1 := Transform(state, chunk1, GameActive, 0)

	if !bytes.Contains(result1.Output, gameLinePlaceholder) {
		t.Error("first injection should work")
	}

	chunk2 := []byte("✶ Second (done)\n")
	result2 := Transform(result1.NewState, chunk2, GameActive, 0)

	if !bytes.Contains(result2.Output, gameLinePlaceholder) {
		t.Error("second injection should work")
	}
}

func TestFindInjectionPoint_AllSpinnerRunes(t *testing.T) {
	spinners := []rune{'✢', '✶', '✻', '✸', '✹', '✺', '✷'}

	for _, s := range spinners {
		t.Run(string(s), func(t *testing.T) {
			data := []byte(string(s) + " Working… (done)\nrest")
			idx := findInjectionPoint(data)
			if idx < 0 {
				t.Errorf("should find injection point for spinner %q", s)
			}
		})
	}
}

func TestTransform_AnsiEscapeHandling(t *testing.T) {
	state := TransformState{}
	// Spinner with ANSI escape sequences
	chunk := []byte("\x1b[1m✻\x1b[0m Computing… (done)\nrest")

	result := Transform(state, chunk, GameActive, 0)

	if !bytes.Contains(result.Output, gameLinePlaceholder) {
		t.Errorf("should handle ANSI escapes, got: %q", result.Output)
	}
}
