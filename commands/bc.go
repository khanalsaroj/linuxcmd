package commands

import (
	"bufio"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"linuxcmd/internal/command"
)

// bcCommand evaluates one arithmetic expression per input line using
// float64 (not true arbitrary precision, unlike real bc; documented in
// the README) with +, -, *, /, %, ^, parentheses, and unary minus.
type bcCommand struct{}

func (bcCommand) Name() string    { return "bc" }
func (bcCommand) Summary() string { return "evaluate arithmetic expressions" }

func (bcCommand) Run(ctx *command.Context) int {
	process := func(expr string) bool {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			return true
		}
		val, err := evalBcExpr(expr)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "bc: %s\n", err)
			return false
		}
		fmt.Fprintln(ctx.Stdout, formatBcNumber(val))
		return true
	}

	if len(ctx.Args) > 0 {
		ok := true
		for _, a := range ctx.Args {
			if !process(a) {
				ok = false
			}
		}
		if !ok {
			return command.ExitFailure
		}
		return command.ExitSuccess
	}

	scanner := bufio.NewScanner(ctx.Stdin)
	ok := true
	for scanner.Scan() {
		if !process(scanner.Text()) {
			ok = false
		}
	}
	if !ok {
		return command.ExitFailure
	}
	return command.ExitSuccess
}

func formatBcNumber(v float64) string {
	if v == math.Trunc(v) && !math.IsInf(v, 0) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	return strconv.FormatFloat(v, 'g', -1, 64)
}

type bcParser struct {
	s   string
	pos int
}

func evalBcExpr(s string) (float64, error) {
	p := &bcParser{s: s}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpace()
	if p.pos != len(p.s) {
		return 0, fmt.Errorf("unexpected input at %q", p.s[p.pos:])
	}
	return v, nil
}

func (p *bcParser) skipSpace() {
	for p.pos < len(p.s) && unicode.IsSpace(rune(p.s[p.pos])) {
		p.pos++
	}
}

func (p *bcParser) peek() byte {
	p.skipSpace()
	if p.pos >= len(p.s) {
		return 0
	}
	return p.s[p.pos]
}

// parseExpr handles + and -.
func (p *bcParser) parseExpr() (float64, error) {
	v, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek() {
		case '+':
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v += r
		case '-':
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v -= r
		default:
			return v, nil
		}
	}
}

// parseTerm handles *, /, %.
func (p *bcParser) parseTerm() (float64, error) {
	v, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek() {
		case '*':
			p.pos++
			r, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			v *= r
		case '/':
			p.pos++
			r, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			v /= r
		case '%':
			p.pos++
			r, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if r == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			v = math.Mod(v, r)
		default:
			return v, nil
		}
	}
}

// parsePower handles ^, right-associative.
func (p *bcParser) parsePower() (float64, error) {
	v, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	if p.peek() == '^' {
		p.pos++
		r, err := p.parsePower()
		if err != nil {
			return 0, err
		}
		return math.Pow(v, r), nil
	}
	return v, nil
}

func (p *bcParser) parseUnary() (float64, error) {
	if p.peek() == '-' {
		p.pos++
		v, err := p.parseUnary()
		return -v, err
	}
	if p.peek() == '+' {
		p.pos++
		return p.parseUnary()
	}
	return p.parseAtom()
}

func (p *bcParser) parseAtom() (float64, error) {
	if p.peek() == '(' {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.peek() != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return v, nil
	}

	p.skipSpace()
	start := p.pos
	for p.pos < len(p.s) && (unicode.IsDigit(rune(p.s[p.pos])) || p.s[p.pos] == '.') {
		p.pos++
	}
	if p.pos == start {
		return 0, fmt.Errorf("expected a number at %q", p.s[start:])
	}
	return strconv.ParseFloat(p.s[start:p.pos], 64)
}

func init() { command.Register(bcCommand{}) }
