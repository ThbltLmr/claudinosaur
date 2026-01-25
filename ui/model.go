package ui

import (
	"io"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thibault/claudinosaur/inject"
	"github.com/thibault/claudinosaur/state"
	"golang.org/x/term"
)

const quietThreshold = 16 * time.Millisecond

type Model struct {
	mode           inject.Mode
	transformState inject.TransformState
	ptyOutput      <-chan []byte
	outputWriter   io.Writer
	lastTick       time.Time
	lastOutputTime time.Time
	cursorTracker  *inject.CursorTracker
	lastSpinnerRow int
	lastOverlayRow int
	done           bool
}

type StateChangeMsg struct {
	NewState state.State
}

type ptyOutputMsg []byte

type ptyClosedMsg struct{}

type tickMsg time.Time

func NewModel(ptyOutput <-chan []byte, output io.Writer) Model {
	now := time.Now()
	return Model{
		mode:           inject.Passthrough,
		ptyOutput:      ptyOutput,
		outputWriter:   output,
		lastTick:       now,
		lastOutputTime: now,
		cursorTracker:  inject.NewCursorTracker(),
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
			m.clearOverlay()
			m.flushBuffer()
		}
		return m, nil

	case ptyOutputMsg:
		now := time.Now()
		dt := now.Sub(m.lastTick)
		m.lastTick = now
		m.lastOutputTime = now

		chunk := []byte(msg)
		if row, found := m.cursorTracker.Process(chunk); found {
			m.lastSpinnerRow = row
		}

		m.processTransform(chunk, dt)
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

		if m.mode == inject.GameActive && now.Sub(m.lastOutputTime) >= quietThreshold {
			m.renderOverlay()
		}

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

func (m *Model) clearOverlay() {
	if m.lastOverlayRow != 0 {
		clearSeq := inject.RenderOverlayAtRow("", m.lastOverlayRow)
		if clearSeq != nil {
			m.outputWriter.Write(clearSeq)
		}
		m.lastOverlayRow = 0
	}
}

func (m *Model) processTransform(chunk []byte, dt time.Duration) {
	result := inject.Transform(m.transformState, chunk, m.mode, dt)
	if len(result.Output) > 0 {
		m.outputWriter.Write(result.Output)
	}
	m.transformState = result.NewState
}

func (m *Model) renderOverlay() {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}

	targetRow := m.lastSpinnerRow + 1
	if targetRow < 1 {
		return
	}

	if m.lastOverlayRow != 0 && m.lastOverlayRow != targetRow {
		clearSeq := inject.RenderOverlayAtRow("", m.lastOverlayRow)
		if clearSeq != nil {
			m.outputWriter.Write(clearSeq)
		}
	}

	if inject.DebugLog != nil {
		inject.DebugLog.Printf("[OVERLAY] rendering at row %d (spinner at %d)", targetRow, m.lastSpinnerRow)
	}

	gameLine := generateGameLine(width)
	overlay := inject.RenderOverlayAtRow(gameLine, targetRow)
	if overlay != nil {
		m.outputWriter.Write(overlay)
		m.lastOverlayRow = targetRow
	}
}

func generateGameLine(width int) string {
	if width < 20 {
		return "🦖"
	}
	return "🦖                🌵                                           Score: 00000"
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
