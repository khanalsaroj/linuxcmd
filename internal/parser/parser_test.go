package parser

import (
	"reflect"
	"testing"
)

var lsLikeSpec = Spec{
	{Short: 'l'},
	{Short: 'a', Long: "all"},
	{Short: 'h'},
	{Short: 'n', Long: "number", HasArg: true},
}

func TestParseCombinedShortFlags(t *testing.T) {
	res, err := Parse([]string{"-la"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !res.Bool('l', "") || !res.Bool('a', "all") {
		t.Errorf("expected both -l and -a set from -la, got bools=%v", res.bools)
	}
}

func TestParseSeparateShortFlags(t *testing.T) {
	res, err := Parse([]string{"-l", "-a"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !res.Bool('l', "") || !res.Bool('a', "all") {
		t.Errorf("expected -l and -a set")
	}
}

func TestParsePositionalArgs(t *testing.T) {
	res, err := Parse([]string{"-la", "/tmp", "file.txt"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	want := []string{"/tmp", "file.txt"}
	if !reflect.DeepEqual(res.Positional, want) {
		t.Errorf("Positional = %v, want %v", res.Positional, want)
	}
}

func TestParseLongFlags(t *testing.T) {
	res, err := Parse([]string{"--all", "--number=5"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !res.Bool('a', "all") {
		t.Errorf("expected --all to set the 'a' bool")
	}
	v, ok := res.Value('n', "number")
	if !ok || v != "5" {
		t.Errorf("Value(n,number) = %q,%v want 5,true", v, ok)
	}
}

func TestParseValueAttachedToShortFlag(t *testing.T) {
	res, err := Parse([]string{"-n5"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	v, ok := res.Value('n', "number")
	if !ok || v != "5" {
		t.Errorf("Value(n,number) = %q,%v want 5,true", v, ok)
	}
}

func TestParseValueAsNextArg(t *testing.T) {
	res, err := Parse([]string{"-n", "5"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	v, ok := res.Value('n', "number")
	if !ok || v != "5" {
		t.Errorf("Value(n,number) = %q,%v want 5,true", v, ok)
	}
}

func TestParseDoubleDashStopsFlags(t *testing.T) {
	res, err := Parse([]string{"-l", "--", "-a", "file.txt"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if res.Bool('a', "all") {
		t.Errorf("-a after -- should be positional, not a flag")
	}
	want := []string{"-a", "file.txt"}
	if !reflect.DeepEqual(res.Positional, want) {
		t.Errorf("Positional = %v, want %v", res.Positional, want)
	}
}

func TestParseSingleDashIsPositional(t *testing.T) {
	res, err := Parse([]string{"-"}, lsLikeSpec)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(res.Positional) != 1 || res.Positional[0] != "-" {
		t.Errorf("Positional = %v, want [-]", res.Positional)
	}
}

func TestParseUnknownShortFlagErrors(t *testing.T) {
	if _, err := Parse([]string{"-z"}, lsLikeSpec); err == nil {
		t.Error("expected error for unknown short flag -z")
	}
}

func TestParseUnknownLongFlagErrors(t *testing.T) {
	if _, err := Parse([]string{"--bogus"}, lsLikeSpec); err == nil {
		t.Error("expected error for unknown long flag --bogus")
	}
}

func TestParseMissingRequiredValueErrors(t *testing.T) {
	if _, err := Parse([]string{"-n"}, lsLikeSpec); err == nil {
		t.Error("expected error when -n has no following value")
	}
}

func TestParseArgWithSpaces(t *testing.T) {
	// Simulates what the OS already delivers for a quoted CLI argument
	// like `grep "hello world" file.txt` -- Windows argv splitting has
	// already produced one token containing the space by the time it
	// reaches Parse.
	res, err := Parse([]string{"hello world", "file.txt"}, Spec{})
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	want := []string{"hello world", "file.txt"}
	if !reflect.DeepEqual(res.Positional, want) {
		t.Errorf("Positional = %v, want %v", res.Positional, want)
	}
}
