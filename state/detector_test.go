package state

import (
	"testing"
	"time"
)

func TestDetector_InitialState(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)
	if d.CurrentState() != Idle {
		t.Errorf("expected Idle, got %v", d.CurrentState())
	}
}

func TestDetector_TransitionToWorkingOnSpinner(t *testing.T) {
	var transitions []State
	d := NewDetector(500*time.Millisecond, func(from, to State) {
		transitions = append(transitions, to)
	})

	d.Write([]byte("some text ✸ more text"))

	if d.CurrentState() != Working {
		t.Errorf("expected Working, got %v", d.CurrentState())
	}
	if len(transitions) != 1 || transitions[0] != Working {
		t.Errorf("expected one transition to Working, got %v", transitions)
	}
}

func TestDetector_TransitionToIdleAfterTimeout(t *testing.T) {
	var transitions []State
	d := NewDetector(500*time.Millisecond, func(from, to State) {
		transitions = append(transitions, to)
	})

	d.Write([]byte("✸"))

	now := time.Now()
	d.Check(now.Add(100 * time.Millisecond))
	if d.CurrentState() != Working {
		t.Error("should still be Working before timeout")
	}

	d.Check(now.Add(500 * time.Millisecond))
	if d.CurrentState() != Idle {
		t.Errorf("expected Idle after timeout, got %v", d.CurrentState())
	}

	if len(transitions) != 2 {
		t.Errorf("expected 2 transitions, got %d", len(transitions))
	}
	if transitions[0] != Working || transitions[1] != Idle {
		t.Errorf("expected [Working, Idle], got %v", transitions)
	}
}

func TestDetector_NoTransitionIfAlreadyInState(t *testing.T) {
	callCount := 0
	d := NewDetector(500*time.Millisecond, func(from, to State) {
		callCount++
	})

	d.Write([]byte("✸"))
	d.Write([]byte("✶"))

	if callCount != 1 {
		t.Errorf("expected 1 callback, got %d", callCount)
	}
}

func TestDetector_AllSpinnerCharacters(t *testing.T) {
	spinners := []string{"✢", "✶", "✻", "✸", "✹", "✺", "✷"}

	for _, spinner := range spinners {
		d := NewDetector(500*time.Millisecond, nil)
		d.Write([]byte(spinner))

		if d.CurrentState() != Working {
			t.Errorf("spinner %s not detected", spinner)
		}
	}
}

func TestDetector_EmptyChunk(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)
	d.Write([]byte{})

	if d.CurrentState() != Idle {
		t.Error("empty chunk should not change state")
	}
}

func TestDetector_SpinnerResetsTimeout(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)

	d.Write([]byte("✸"))
	time.Sleep(10 * time.Millisecond)
	d.Write([]byte("✸"))

	now := time.Now()
	d.Check(now.Add(400 * time.Millisecond))

	if d.CurrentState() != Working {
		t.Error("spinner should reset timeout timer")
	}
}

func TestDetector_SpinnerInMiddleOfText(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)
	d.Write([]byte("Loading ✻ please wait"))

	if d.CurrentState() != Working {
		t.Error("spinner in middle of text not detected")
	}
}

func TestDetector_NoFalsePositives(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)
	d.Write([]byte("normal text with unicode: émojis 🦖 and symbols ★ ☆ ✓"))

	if d.CurrentState() != Idle {
		t.Error("false positive: non-spinner unicode detected as spinner")
	}
}

func TestDetector_StateString(t *testing.T) {
	if Idle.String() != "idle" {
		t.Errorf("expected 'idle', got %q", Idle.String())
	}
	if Working.String() != "working" {
		t.Errorf("expected 'working', got %q", Working.String())
	}
}

func TestDetector_WriteReturnsCorrectLength(t *testing.T) {
	d := NewDetector(500*time.Millisecond, nil)
	input := []byte("test ✸ data")
	n, err := d.Write(input)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != len(input) {
		t.Errorf("expected %d, got %d", len(input), n)
	}
}

func TestDetector_CallbackReceivesCorrectStates(t *testing.T) {
	var fromStates, toStates []State
	d := NewDetector(500*time.Millisecond, func(from, to State) {
		fromStates = append(fromStates, from)
		toStates = append(toStates, to)
	})

	d.Write([]byte("✸"))
	d.Check(time.Now().Add(600 * time.Millisecond))

	if len(fromStates) != 2 || len(toStates) != 2 {
		t.Fatalf("expected 2 transitions, got %d", len(fromStates))
	}
	if fromStates[0] != Idle || toStates[0] != Working {
		t.Errorf("first transition: expected Idle→Working, got %v→%v", fromStates[0], toStates[0])
	}
	if fromStates[1] != Working || toStates[1] != Idle {
		t.Errorf("second transition: expected Working→Idle, got %v→%v", fromStates[1], toStates[1])
	}
}
