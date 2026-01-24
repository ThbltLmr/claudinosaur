package ui

import (
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thibault/claudinosaur/inject"
	"github.com/thibault/claudinosaur/state"
)

type Model struct {
	mode           inject.Mode
	transformState inject.TransformState
	ptyOutput      <-chan []byte
	outputWriter   io.Writer
	lastTick       time.Time
	done           bool
}

type StateChangeMsg struct {
	NewState state.State
}

type ptyOutputMsg []byte

type ptyClosedMsg struct{}

type tickMsg time.Time

func NewModel(ptyOutput <-chan []byte, output io.Writer) Model {
	return Model{
		mode:         inject.Passthrough,
		ptyOutput:    ptyOutput,
		outputWriter: output,
		lastTick:     time.Now(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		waitForPtyOutput(m.ptyOutput),
		tick(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case StateChangeMsg:
		if msg.NewState == state.Working {
			m.mode = inject.GameActive
		} else {
			m.mode = inject.Passthrough
			m.flushBuffer()
		}
		return m, nil

	case ptyOutputMsg:
		now := time.Now()
		dt := now.Sub(m.lastTick)
		m.lastTick = now
		m.processTransform([]byte(msg), dt)
		return m, waitForPtyOutput(m.ptyOutput)

	case ptyClosedMsg:
		m.flushBuffer()
		m.done = true
		return m, tea.Quit

	case tickMsg:
		if m.done {
			return m, nil
		}
		now := time.Time(msg)
		dt := now.Sub(m.lastTick)
		m.lastTick = now
		m.processTransform(nil, dt)
		return m, tick()
	}

	return m, nil
}

func (m Model) View() string {
	return ""
}

func (m *Model) flushBuffer() {
	if len(m.transformState.Buffer) > 0 {
		m.outputWriter.Write(m.transformState.Buffer)
		m.transformState = inject.TransformState{}
	}
}

func (m *Model) processTransform(chunk []byte, dt time.Duration) {
	result := inject.Transform(m.transformState, chunk, m.mode, dt)
	if len(result.Output) > 0 {
		m.outputWriter.Write(result.Output)
	}
	m.transformState = result.NewState
}

func waitForPtyOutput(ch <-chan []byte) tea.Cmd {
	return func() tea.Msg {
		data, ok := <-ch
		if !ok {
			return ptyClosedMsg{}
		}
		return ptyOutputMsg(data)
	}
}

func tick() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
