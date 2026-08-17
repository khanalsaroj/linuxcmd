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

// awkCommand supports a small, documented subset of AWK: an optional
// /regex/ pattern, and a "{ print EXPR, EXPR, ... }" action whose
// expressions are limited to $N field references (including $0 and
// $NF) and quoted string literals. Variables, arithmetic, user
// functions, and BEGIN/END blocks are not implemented.
type awkCommand struct{}

func (awkCommand) Name() string    { return "awk" }
func (awkCommand) Summary() string { return "field-split and print lines (small AWK subset)" }

var awkSpec = parser.Spec{
	{Short: 'F', HasArg: true},
}

var awkDefaultFieldSplit = regexp.MustCompile(`\s+`)

type awkProgram struct {
	pattern *regexp.Regexp
	fields  []string // expressions to print; nil means "print $0"
}

func parseAwkScript(s string) (awkProgram, error) {
	var prog awkProgram
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "/") {
		end := strings.IndexByte(s[1:], '/')
		if end < 0 {
			return prog, fmt.Errorf("unterminated pattern")
		}
		re, err := regexp.Compile(s[1 : 1+end])
		if err != nil {
			return prog, err
		}
		prog.pattern = re
		s = strings.TrimSpace(s[1+end+1:])
	}

	if s == "" {
		return prog, nil
	}
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return prog, fmt.Errorf("unsupported action (only '{ print ... }' is supported)")
	}
	body := strings.TrimSpace(s[1 : len(s)-1])
	body = strings.TrimSpace(strings.TrimPrefix(body, "print"))
	if body == "" {
		return prog, nil
	}
	for _, p := range strings.Split(body, ",") {
		prog.fields = append(prog.fields, strings.TrimSpace(p))
	}
	return prog, nil
}

func evalAwkExpr(token, line string, fields []string) (string, error) {
	if len(token) >= 2 && strings.HasPrefix(token, `"`) && strings.HasSuffix(token, `"`) {
		return token[1 : len(token)-1], nil
	}
	if strings.HasPrefix(token, "$") {
		idxStr := token[1:]
		if idxStr == "NF" {
			if len(fields) == 0 {
				return "", nil
			}
			return fields[len(fields)-1], nil
		}
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return "", fmt.Errorf("invalid field '%s'", token)
		}
		if idx == 0 {
			return line, nil
		}
		if idx < 1 || idx > len(fields) {
			return "", nil
		}
		return fields[idx-1], nil
	}
	return "", fmt.Errorf("unsupported expression '%s'", token)
}

func (p awkProgram) run(ctx *command.Context, splitRe *regexp.Regexp, line string) error {
	if p.pattern != nil && !p.pattern.MatchString(line) {
		return nil
	}
	fields := splitRe.Split(strings.TrimSpace(line), -1)
	if len(fields) == 1 && fields[0] == "" {
		fields = nil
	}

	if p.fields == nil {
		fmt.Fprintln(ctx.Stdout, line)
		return nil
	}
	parts := make([]string, len(p.fields))
	for i, tok := range p.fields {
		v, err := evalAwkExpr(tok, line, fields)
		if err != nil {
			return err
		}
		parts[i] = v
	}
	fmt.Fprintln(ctx.Stdout, strings.Join(parts, " "))
	return nil
}

func (awkCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, awkSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "awk: %s\n", err)
		return command.ExitUsage
	}
	if len(res.Positional) == 0 {
		fmt.Fprintln(ctx.Stderr, "usage: awk [-F SEP] 'PROGRAM' [FILE...]")
		return command.ExitUsage
	}
	prog, err := parseAwkScript(res.Positional[0])
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "awk: %s\n", err)
		return command.ExitUsage
	}
	splitRe := awkDefaultFieldSplit
	if v, ok := res.Value('F', ""); ok {
		re, err := regexp.Compile(regexp.QuoteMeta(v))
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "awk: invalid field separator '%s'\n", v)
			return command.ExitUsage
		}
		splitRe = re
	}

	exit := command.ExitSuccess
	process := func(r *bufio.Scanner) {
		for r.Scan() {
			if err := prog.run(ctx, splitRe, r.Text()); err != nil {
				fmt.Fprintf(ctx.Stderr, "awk: %s\n", err)
				exit = command.ExitFailure
				return
			}
		}
	}

	files := paths.ExpandGlobs(res.Positional[1:])
	if len(files) == 0 {
		scanner := bufio.NewScanner(ctx.Stdin)
		scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
		process(scanner)
		return exit
	}

	for _, arg := range files {
		resolved, err := paths.Resolve(arg)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "awk", arg, err)
			exit = command.ExitFailure
			continue
		}
		f, err := os.Open(resolved)
		if err != nil {
			output.SimpleErrorf(ctx.Stderr, "awk", arg, err)
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

func init() { command.Register(awkCommand{}) }
