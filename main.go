package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
	"github.com/thibault/claudinosaur/inject"
	"github.com/thibault/claudinosaur/state"
	"github.com/thibault/claudinosaur/ui"
	"golang.org/x/term"
)

var debugFlag bool

func init() {
	for _, arg := range os.Args[1:] {
		if arg == "--dino-debug" {
			debugFlag = true
			break
		}
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "claudinosaur: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH")
	}

	var debugLog *log.Logger
	if debugFlag {
		home := os.Getenv("HOME")
		if home == "" {
			home = "."
		}
		debugDir := filepath.Join(home, ".claudinosaur")
		if err := os.MkdirAll(debugDir, 0755); err != nil {
			return fmt.Errorf("failed to create debug dir: %w", err)
		}
		debugFile, err := os.OpenFile(
			filepath.Join(debugDir, "debug.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			return fmt.Errorf("failed to open debug log: %w", err)
		}
		defer debugFile.Close()
		debugLog = log.New(debugFile, "", log.LstdFlags)
		debugLog.Println("=== Claudinosaur started ===")
		inject.DebugLog = debugLog
	}

	claudeArgs := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		if arg != "--dino-debug" {
			claudeArgs = append(claudeArgs, arg)
		}
	}
	cmd := exec.Command(claudePath, claudeArgs...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}
	defer ptmx.Close()

	stopSigwinch, err := setupSigwinchHandler(ptmx)
	if err != nil {
		return err
	}
	defer stopSigwinch()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	ptyOutputChan := make(chan []byte, 100)
	model := ui.NewModel(ptyOutputChan, os.Stdout)
	program := tea.NewProgram(model, tea.WithoutRenderer(), tea.WithInput(nil))

	detector := state.NewDetector(500*time.Millisecond, func(from, to state.State) {
		if debugLog != nil {
			debugLog.Printf("[STATE] %s → %s", from, to)
		}
		program.Send(ui.StateChangeMsg{NewState: to})
	})

	detectorTicker := time.NewTicker(100 * time.Millisecond)
	defer detectorTicker.Stop()

	go func() {
		for range detectorTicker.C {
			detector.Check(time.Now())
		}
	}()

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])
				detector.Write(chunk)
				ptyOutputChan <- chunk
			}
			if err != nil {
				close(ptyOutputChan)
				break
			}
		}
	}()

	go func() {
		program.Run()
	}()

	err = cmd.Wait()
	program.Quit()
	return err
}

func setupSigwinchHandler(ptmx *os.File) (stop func(), err error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)

	go func() {
		for range ch {
			pty.InheritSize(os.Stdin, ptmx)
		}
	}()

	stop = func() {
		signal.Stop(ch)
		close(ch)
	}

	return stop, pty.InheritSize(os.Stdin, ptmx)
}
