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
		t.Error("buffer should be empty")
	}
}

func TestTransform_FlushesExistingBuffer(t *testing.T) {
	state := TransformState{Buffer: []byte("buffered ")}
	chunk := []byte("new data")

	result := Transform(state, chunk, Passthrough, 0)

	expected := []byte("buffered new data")
	if !bytes.Equal(result.Output, expected) {
		t.Errorf("expected %q, got %q", expected, result.Output)
	}
}

func TestTransform_GameActiveMode_Passthrough(t *testing.T) {
	state := TransformState{}
	chunk := []byte("some output from claude")

	result := Transform(state, chunk, GameActive, 10*time.Millisecond)

	if !bytes.Equal(result.Output, chunk) {
		t.Errorf("expected passthrough, got %q", result.Output)
	}
	if len(result.NewState.Buffer) != 0 {
		t.Error("buffer should be empty")
	}
}

func TestTransform_EmptyChunk(t *testing.T) {
	state := TransformState{}
	chunk := []byte{}

	result := Transform(state, chunk, Passthrough, 0)

	if len(result.Output) != 0 {
		t.Errorf("expected empty output, got %q", result.Output)
	}
}

func TestTransform_EmptyChunkWithBuffer(t *testing.T) {
	state := TransformState{Buffer: []byte("buffered")}
	chunk := []byte{}

	result := Transform(state, chunk, Passthrough, 0)

	if !bytes.Equal(result.Output, []byte("buffered")) {
		t.Errorf("expected buffer flush, got %q", result.Output)
	}
}
