package commands

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"linuxcmd/internal/command"
	"linuxcmd/internal/parser"
)

type ncCommand struct{}

func (ncCommand) Name() string    { return "nc" }
func (ncCommand) Summary() string { return "basic TCP/UDP client and listener" }

var ncSpec = parser.Spec{
	{Short: 'l'},
	{Short: 'u'},
	{Short: 'w', HasArg: true},
}

func (ncCommand) Run(ctx *command.Context) int {
	res, err := parser.Parse(ctx.Args, ncSpec)
	if err != nil {
		fmt.Fprintf(ctx.Stderr, "nc: %s\n", err)
		return command.ExitUsage
	}
	listen := res.Bool('l', "")
	network := "tcp"
	if res.Bool('u', "") {
		network = "udp"
	}
	timeout := 10 * time.Second
	if v, ok := res.Value('w', ""); ok {
		secs, err := strconv.Atoi(v)
		if err != nil || secs < 1 {
			fmt.Fprintf(ctx.Stderr, "nc: invalid timeout '%s'\n", v)
			return command.ExitUsage
		}
		timeout = time.Duration(secs) * time.Second
	}

	var conn net.Conn
	if listen {
		if len(res.Positional) < 1 {
			fmt.Fprintln(ctx.Stderr, "usage: nc -l PORT")
			return command.ExitUsage
		}
		ln, err := net.Listen(network, ":"+res.Positional[0])
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "nc: %s\n", err)
			return command.ExitFailure
		}
		defer ln.Close()
		accepted := make(chan net.Conn, 1)
		errCh := make(chan error, 1)
		go func() {
			c, err := ln.Accept()
			if err != nil {
				errCh <- err
				return
			}
			accepted <- c
		}()
		select {
		case c := <-accepted:
			conn = c
		case err := <-errCh:
			fmt.Fprintf(ctx.Stderr, "nc: %s\n", err)
			return command.ExitFailure
		case <-time.After(timeout):
			fmt.Fprintln(ctx.Stderr, "nc: timed out waiting for a connection")
			return command.ExitFailure
		}
	} else {
		if len(res.Positional) < 2 {
			fmt.Fprintln(ctx.Stderr, "usage: nc HOST PORT")
			return command.ExitUsage
		}
		c, err := net.DialTimeout(network, net.JoinHostPort(res.Positional[0], res.Positional[1]), timeout)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "nc: %s\n", err)
			return command.ExitFailure
		}
		conn = c
	}
	defer conn.Close()

	done := make(chan struct{}, 2)
	go func() { io.Copy(conn, ctx.Stdin); done <- struct{}{} }()
	go func() { io.Copy(ctx.Stdout, conn); done <- struct{}{} }()
	<-done
	return command.ExitSuccess
}

func init() { command.Register(ncCommand{}) }
