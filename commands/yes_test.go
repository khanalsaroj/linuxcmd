package commands

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"linuxcmd/internal/command"
)

// limitedWriter accepts up to n bytes total and then returns an error on
// every subsequent write, simulating a reader that closed its end of a
// pipe (how a real "yes | head" stops the loop).
type limitedWriter struct {
	remaining int
	written   []byte
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if w.remaining <= 0 {
		return 0, errors.New("broken pipe")
	}
	n := len(p)
	if n > w.remaining {
		n = w.remaining
	}
	w.written = append(w.written, p[:n]...)
	w.remaining -= n
	if n < len(p) {
		return n, errors.New("broken pipe")
	}
	return n, nil
}

func TestYesStopsOnWriteError(t *testing.T) {
	cmd, ok := command.Lookup("yes")
	if !ok {
		t.Fatal("yes is not registered")
	}
	w := &limitedWriter{remaining: 100}
	ctx := &command.Context{
		CommandName: "yes",
		Args:        []string{"hello"},
		Stdin:       strings.NewReader(""),
		Stdout:      w,
		Stderr:      &bytes.Buffer{},
	}
	code := cmd.Run(ctx)
	if code != 0 {
		t.Errorf("expected yes to exit 0 on broken pipe, got %d", code)
	}
	if !strings.HasPrefix(string(w.written), "hello\nhello\n") {
		t.Errorf("expected repeated 'hello' lines, got %q", string(w.written))
	}
}
