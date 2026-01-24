package inject

import (
	"bytes"
	"log"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/thibault/claudinosaur/state"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x07]*\x07|\x1b\[[?\d;]*[hlm]`)

var DebugLog *log.Logger

const (
	FlushTimeout  = 50 * time.Millisecond
	MaxBufferSize = 64 * 1024
)

type Mode int

const (
	Passthrough Mode = iota
	GameActive
)

type TransformState struct {
	Buffer    []byte
	BufferAge time.Duration
}

type TransformResult struct {
	Output   []byte
	NewState TransformState
}

var gameLinePlaceholder = []byte("\n=== GAME LINE 1 ===\n=== GAME LINE 2 ===")

func Transform(state TransformState, chunk []byte, mode Mode, dt time.Duration) TransformResult {
	if DebugLog != nil {
		modeStr := "Passthrough"
		if mode == GameActive {
			modeStr = "GameActive"
		}
		DebugLog.Printf("[TRANSFORM] mode=%s chunk_len=%d buffer_len=%d age=%v", modeStr, len(chunk), len(state.Buffer), state.BufferAge)
	}

	if mode == Passthrough {
		if len(state.Buffer) > 0 {
			output := make([]byte, len(state.Buffer)+len(chunk))
			copy(output, state.Buffer)
			copy(output[len(state.Buffer):], chunk)
			return TransformResult{
				Output:   output,
				NewState: TransformState{},
			}
		}
		return TransformResult{
			Output:   chunk,
			NewState: TransformState{},
		}
	}

	newBuffer := append(state.Buffer, chunk...)
	age := state.BufferAge + dt

	if idx := findInjectionPoint(newBuffer); idx >= 0 {
		if DebugLog != nil {
			DebugLog.Printf("[TRANSFORM] INJECTION POINT FOUND at idx=%d, injecting game lines", idx)
		}
		before := newBuffer[:idx]
		after := newBuffer[idx:]

		output := make([]byte, 0, len(newBuffer)+len(gameLinePlaceholder))
		output = append(output, before...)
		output = append(output, gameLinePlaceholder...)
		output = append(output, after...)

		return TransformResult{
			Output:   output,
			NewState: TransformState{},
		}
	}

	if age >= FlushTimeout || len(newBuffer) > MaxBufferSize {
		if DebugLog != nil {
			DebugLog.Printf("[TRANSFORM] FLUSH due to timeout/size, buffer=%d bytes, age=%v", len(newBuffer), age)
		}
		return TransformResult{
			Output:   newBuffer,
			NewState: TransformState{},
		}
	}

	if DebugLog != nil {
		DebugLog.Printf("[TRANSFORM] BUFFERING, no injection point yet, buffer preview: %q", truncate(newBuffer, 100))
	}

	return TransformResult{
		Output:   nil,
		NewState: TransformState{Buffer: newBuffer, BufferAge: age},
	}
}

func truncate(b []byte, max int) []byte {
	if len(b) <= max {
		return b
	}
	return b[:max]
}

func findInjectionPoint(data []byte) int {
	cleaned := stripAnsi(data)

	if DebugLog != nil {
		DebugLog.Printf("[FIND] looking for spinner+paren pattern in %d bytes, cleaned=%d bytes", len(data), len(cleaned))
		DebugLog.Printf("[FIND] cleaned preview: %q", truncate(cleaned, 100))
	}

	spinnerIdx := -1
	for i := 0; i < len(cleaned); i++ {
		r, size := utf8.DecodeRune(cleaned[i:])
		if r == utf8.RuneError {
			continue
		}
		for _, s := range state.SpinnerRunes {
			if r == s {
				spinnerIdx = i
				break
			}
		}
		if spinnerIdx >= 0 {
			break
		}
		i += size - 1
	}

	if spinnerIdx < 0 {
		if DebugLog != nil {
			DebugLog.Printf("[FIND] no spinner found")
		}
		return -1
	}

	parenIdx := bytes.IndexByte(cleaned[spinnerIdx:], ')')
	if parenIdx < 0 {
		if DebugLog != nil {
			DebugLog.Printf("[FIND] spinner found at %d but no closing paren after", spinnerIdx)
		}
		return -1
	}

	cleanedTargetIdx := spinnerIdx + parenIdx + 1

	rawIdx := mapCleanedIdxToRaw(data, cleaned, cleanedTargetIdx)

	if DebugLog != nil {
		DebugLog.Printf("[FIND] FOUND pattern! spinner at %d, paren at %d, injecting after raw idx %d", spinnerIdx, spinnerIdx+parenIdx, rawIdx)
	}

	return rawIdx
}

func mapCleanedIdxToRaw(raw, cleaned []byte, cleanedIdx int) int {
	cleanedPos := 0
	rawPos := 0

	for rawPos < len(raw) && cleanedPos < cleanedIdx {
		if raw[rawPos] == 0x1b {
			end := rawPos + 1
			for end < len(raw) {
				b := raw[end]
				end++
				if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
					break
				}
				if b == 0x07 {
					break
				}
			}
			rawPos = end
		} else {
			rawPos++
			cleanedPos++
		}
	}

	return rawPos
}

func stripAnsi(data []byte) []byte {
	return ansiEscape.ReplaceAll(data, nil)
}
