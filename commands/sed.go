package commands

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"linuxcmd/internal/command"
	"linuxcmd/internal/output"
	"linuxcmd/internal/parser"
	"linuxcmd/internal/paths"
)

// sedCommand supports a useful initial subset of sed: a single
// substitution (s/pat/repl/flags) or deletion (d) command, each
// optionally restricted to lines matching an address (/regex/ or a line
// number). Multiple semicolon-chained commands, ranges (addr1,addr2),
// and most other sed commands (a/i/c/y/...) are not implemented.
type sedCommand struct{}

func (sedCommand) Name() string    { return "sed" }
func (sedCommand) Summary() string { return "stream-edit lines (subset: one s/// or d command)" }

var sedSpec = parser.Spec{
	{Short: 'n', Long: "quiet"},
}

type sedOp struct {
	kind        byte // 's' or 'd'
	addr        *regexp.Regexp
	addrLine    int
	pattern     *regexp.Regexp
	replacement string
	global      bool
}

func convertSedReplacement(repl string) string {
	var b strings.Builder
	for i := 0; i < len(repl); i++ {
		c := repl[i]
		switch {
		case c == '\\' && i+1 < len(repl) && repl[i+1] >= '0' && repl[i+1] <= '9':
			b.WriteString("${")
			b.WriteByte(repl[i+1])
			b.WriteString("}")
			i++
		case c == '&':
			b.WriteString("${0}")
		case c == '$':
			b.WriteString("$$")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func parseSedScript(s string) (sedOp, error) {
	var op sedOp
	i := 0

	if i < len(s) && s[i] == '/' {
		end := strings.IndexByte(s[i+1:], '/')
		if end < 0 {
			return op, fmt.Errorf("unterminated address")
		}
		re, err := regexp.Compile(s[i+1 : i+1+end])
		if err != nil {
			return op, err
		}
		op.addr = re
		i = i + 1 + end + 1
	} else {
		start := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		if i > start {
			n, _ := strconv.Atoi(s[start:i])
			op.addrLine = n
		}
	}

	if i >= len(s) {
		return op, fmt.Errorf("empty command")
	}

	switch s[i] {
	case 'd':
		op.kind = 'd'
	case 's':
		if i+1 >= len(s) {
			return op, fmt.Errorf("incomplete substitution")
		}
		delim := s[i+1]
		parts := strings.SplitN(s[i+2:], string(delim), 3)
		if len(parts) < 2 {
			return op, fmt.Errorf("unterminated substitution")
		}
		flags := ""
		if len(parts) == 3 {
			flags = parts[2]
		}
		reSrc := parts[0]
		if strings.Contains(flags, "i") {
			reSrc = "(?i)" + reSrc
		}
		re, err := regexp.Compile(reSrc)
		if err != nil {
			return op, err
		}
		op.kind = 's'
		op.pattern = re
		op.replacement = convertSedReplacement(parts[1])
		op.global = strings.Contains(flags, "g")
	default:
		return op, fmt.Errorf("unsupported command '%c'", s[i])
	}
	return op, nil
}

// apply returns the (possibly rewritten) line and whether it should be
// deleted.
func (op sedOp) apply(line string, lineNo int) (string, bool) {
	matched := op.addr == nil && op.addrLine == 0
	if op.addr != nil && op.addr.MatchString(line) {
		matched = true
	}
	if op.addrLine != 0 && op.addrLine == lineNo {
		matched = true
	}
	if !matched {
		return line, false
	}

	switch op.kind {
	case 'd':
		return "", true
	case 's':
		if op.global {
			return op.pattern.ReplaceAllString(line, op.replacement), false
		}
		loc := op.pattern.FindStringSubmatchIndex(line)
		if loc == nil {
			return line, false
		}
		var buf []byte
		buf = op.pattern.ExpandString(buf, op.replacement, line, loc)
		return line[:loc[0]] + string(buf) + line[loc[1]:], false
	}
	return line, false
}

func (sedCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, sedSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "sed: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: sed [-n] 's/PATTERN/REPLACEMENT/[gi]' [FILE...]")
		return command.ExitUsage
	}
	op, err := parseSedScript(res.Positional[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "sed: %s\n", err)
		return command.ExitUsage
	}
	suppress := res.Bool('n', "quiet")

	process := func(r *bufio.Scanner) {
		lineNo := 0
		for r.Scan() {
			lineNo++
			result, deleted := op.apply(r.Text(), lineNo)
			if deleted || suppress {
				continue
			}
			fmt.Fprintln(ctx.Stdout, result)
		}
	}

	files := paths.ExpandGlobs(res.Positional[1:])
	if len(files) == 0 {
		scanner := bufio.NewScanner(ctx.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		return command.ExitSuccess
	}

	exit := command.ExitSuccess
	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "sed", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "sed", arg, err)
			exit = command.ExitFailure
			continue
		}
		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		f.Close()
	}
	return exit
}

func init() { command.Register(sedCommand{}) }
