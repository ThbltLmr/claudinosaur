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

const quietThreshold = 10 * time.Millisecond
const overlayHeight = 2

type Model struct {
	mode                inject.Mode
	transformState      inject.TransformState
	ptyOutput           <-chan []byte
	outputWriter        io.Writer
	lastTick            time.Time
	lastOutputTime      time.Time
	cursorTracker       *inject.CursorTracker
	lastSpinnerRow      int
	lastOverlayRowStart int
	done                bool
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
	if m.lastOverlayRowStart != 0 {
		clearSeq := inject.ClearMultipleRows(m.lastOverlayRowStart, overlayHeight)
		if clearSeq != nil {
			m.outputWriter.Write(clearSeq)
		}
		m.lastOverlayRowStart = 0
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
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return
	}

	skyRow := m.lastSpinnerRow + 2
	if skyRow < 1 {
		return
	}

	if skyRow+overlayHeight-1 > height {
		return
	}

	if m.lastOverlayRowStart != 0 && m.lastOverlayRowStart != skyRow {
		clearSeq := inject.ClearMultipleRows(m.lastOverlayRowStart, overlayHeight)
		if clearSeq != nil {
			m.outputWriter.Write(clearSeq)
		}
	}

	if inject.DebugLog != nil {
		inject.DebugLog.Printf("[OVERLAY] rendering at rows %d-%d (spinner at %d)", skyRow, skyRow+overlayHeight-1, m.lastSpinnerRow)
	}

	skyLine, groundLine := generateGameLines(width)
	overlay := inject.RenderMultiLineOverlay([]string{skyLine, groundLine}, skyRow)
	if overlay != nil {
		m.outputWriter.Write(overlay)
		m.lastOverlayRowStart = skyRow
	}
}

func generateGameLines(width int) (string, string) {
	if width < 20 {
		return "☁️", "🦖"
	}
	skyLine := "    ☁️           ☁️                    ☁️                              [SKY]"
	groundLine := "🦖                🌵                                           Score: 00000"
	return skyLine, groundLine
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
