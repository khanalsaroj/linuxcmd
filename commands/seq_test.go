package commands

import "testing"

func TestSeqSingleArg(t *testing.T) {
	code, out, errOut := run(t, "seq", "3")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "1\n2\n3\n" {
		t.Errorf("seq 3 output = %q, want %q", out, "1\n2\n3\n")
	}
}

func TestSeqStartEnd(t *testing.T) {
	code, out, errOut := run(t, "seq", "2", "5")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "2\n3\n4\n5\n" {
		t.Errorf("seq 2 5 output = %q, want %q", out, "2\n3\n4\n5\n")
	}
}

func TestSeqStartStepEnd(t *testing.T) {
	code, out, errOut := run(t, "seq", "1", "2", "10")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "1\n3\n5\n7\n9\n" {
		t.Errorf("seq 1 2 10 output = %q, want %q", out, "1\n3\n5\n7\n9\n")
	}
}

func TestSeqInvalidNumber(t *testing.T) {
	code, _, errOut := run(t, "seq", "notanumber")
	if code == 0 {
		t.Error("expected nonzero exit for invalid number")
	}
	if errOut == "" {
		t.Error("expected an error message")
	}
}
