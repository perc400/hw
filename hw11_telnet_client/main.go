package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var timeoutStr string

func init() {
	flag.StringVar(&timeoutStr, "timeout", "60s", "timeout before connection")
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=60s] host port")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	in := os.Stdin
	out := os.Stdout

	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse timeout: %v", err)
		return
	}
	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, timeout, in, out)

	err = client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connection error: %v", err)
		return
	}
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	errSendCh := make(chan error)
	errReceiveCh := make(chan error)

	go func() {
		err := client.Send()
		errSendCh <- err
	}()

	go func() {
		err := client.Receive()
		errReceiveCh <- err
	}()

	select {
	case <-ctx.Done():
		stop()
		client.Close()
	case err := <-errReceiveCh:
		if err != nil {
			fmt.Fprintf(os.Stderr, "...Connection was closed by peer: %v", err)
		}
		client.Close()
	case err := <-errSendCh:
		if err == nil {
			fmt.Fprint(os.Stderr, "...EOF")
		} else {
			fmt.Fprintf(os.Stderr, "...Connection was closed by peer: %v", err)
		}
		client.Close()
	}
}
