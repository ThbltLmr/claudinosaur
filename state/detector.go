package state

import (
	"sync"
	"time"
	"unicode/utf8"
)

type State int

const (
	Idle State = iota
	Working
)

func (s State) String() string {
	switch s {
	case Idle:
		return "idle"
	case Working:
		return "working"
	default:
		return "unknown"
	}
}

var spinnerRunes = []rune{'✢', '✶', '✻', '✸', '✹', '✺', '✷'}

type Detector struct {
	mu           sync.Mutex
	state        State
	lastSpinner  time.Time
	timeout      time.Duration
	onTransition func(from, to State)
}

func NewDetector(timeout time.Duration, onTransition func(from, to State)) *Detector {
	return &Detector{
		state:        Idle,
		timeout:      timeout,
		onTransition: onTransition,
	}
}

func (d *Detector) Write(p []byte) (int, error) {
	if containsSpinner(p) {
		d.mu.Lock()
		d.lastSpinner = time.Now()
		d.transitionTo(Working)
		d.mu.Unlock()
	}
	return len(p), nil
}

func (d *Detector) Check(now time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.state == Working && !d.lastSpinner.IsZero() && now.Sub(d.lastSpinner) >= d.timeout {
		d.transitionTo(Idle)
	}
}

func (d *Detector) CurrentState() State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

func (d *Detector) transitionTo(newState State) {
	if d.state == newState {
		return
	}
	old := d.state
	d.state = newState
	if d.onTransition != nil {
		d.onTransition(old, newState)
	}
}

func containsSpinner(p []byte) bool {
	for len(p) > 0 {
		r, size := utf8.DecodeRune(p)
		if r == utf8.RuneError && size == 1 {
			p = p[1:]
			continue
		}
		for _, spinner := range spinnerRunes {
			if r == spinner {
				return true
			}
		}
		p = p[size:]
	}
	return false
}
