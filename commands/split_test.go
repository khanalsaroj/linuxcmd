package commands

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSplitByLines(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "big.txt")
	var b strings.Builder
	for i := 1; i <= 25; i++ {
		b.WriteString("line")
		b.WriteString(strconv.Itoa(i))
		b.WriteByte('\n')
	}
	mustWriteFile(t, src, b.String())

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := runIn(t, outDir, "split", "-l", "10", src, "part-")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}

	for _, suffix := range []string{"aa", "ab", "ac"} {
		if _, err := os.Stat(filepath.Join(outDir, "part-"+suffix)); err != nil {
			t.Errorf("expected part-%s to exist: %v", suffix, err)
		}
	}
	content, err := os.ReadFile(filepath.Join(outDir, "part-aa"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(content), "\n") != 10 {
		t.Errorf("expected 10 lines in part-aa, got %d", strings.Count(string(content), "\n"))
	}
}
