package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/term"
)

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

	cmd := exec.Command(claudePath, os.Args[1:]...)

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

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	io.Copy(os.Stdout, ptmx)

	return cmd.Wait()
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
