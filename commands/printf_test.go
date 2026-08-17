package commands

import "testing"

func TestPrintfBasicFormat(t *testing.T) {
	code, out, errOut := run(t, "printf", "%s %d\\n", "name", "3")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "name 3\n" {
		t.Errorf("printf output = %q, want %q", out, "name 3\n")
	}
}

func TestPrintfEscapes(t *testing.T) {
	code, out, errOut := run(t, "printf", "a\\tb\\n")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\tb\n" {
		t.Errorf("printf output = %q, want %q", out, "a\tb\n")
	}
}

func TestPrintfCyclesFormatOverExtraArgs(t *testing.T) {
	code, out, errOut := run(t, "printf", "%s\\n", "a", "b", "c")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "a\nb\nc\n" {
		t.Errorf("printf output = %q, want %q", out, "a\nb\nc\n")
	}
}

func TestPrintfPercentLiteral(t *testing.T) {
	code, out, errOut := run(t, "printf", "100%%\\n")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "100%\n" {
		t.Errorf("printf output = %q, want %q", out, "100%\n")
	}
}

func TestPrintfMissingFormat(t *testing.T) {
	code, _, errOut := run(t, "printf")
	if code == 0 {
		t.Error("expected nonzero exit for missing format")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
