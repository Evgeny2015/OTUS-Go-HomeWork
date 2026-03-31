package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout=10s] host port\n", os.Args[0])
		os.Exit(1)
	}
	host := args[0]
	port := args[1]
	address := net.JoinHostPort(host, port)

	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	// Connect
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Failed to connect: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	// Signal handling
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	// Goroutine for sending stdin to socket
	go func() {
		if err := client.Send(); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "...EOF\n")
			} else {
				fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
			}
			stop()
		}
	}()

	// Goroutine for receiving from socket to stdout
	go func() {
		if err := client.Receive(); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintf(os.Stderr, "...Connection was closed by peer\n")
			} else {
				fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
			}
			stop()
		}
	}()

	// Wait for either SIGINT or one of the goroutines to finish
	<-sigCtx.Done()
	client.Close()
}
