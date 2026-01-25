package inject

import (
	"log"
	"time"
)

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

func Transform(state TransformState, chunk []byte, mode Mode, dt time.Duration) TransformResult {
	if DebugLog != nil {
		modeStr := "Passthrough"
		if mode == GameActive {
			modeStr = "GameActive"
		}
		DebugLog.Printf("[TRANSFORM] mode=%s chunk_len=%d buffer_len=%d", modeStr, len(chunk), len(state.Buffer))
	}

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
