package main

import (
	"bufio"
	"context"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
	scanner *bufio.Scanner
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (c *telnetClient) Connect() error {
	dialer := &net.Dialer{
		Timeout: c.timeout,
	}
	conn, err := dialer.DialContext(context.Background(), "tcp", c.address)
	if err != nil {
		return err
	}
	c.conn = conn
	c.scanner = bufio.NewScanner(conn)
	return nil
}

func (c *telnetClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *telnetClient) Send() error {
	if c.conn == nil {
		return io.EOF
	}
	_, err := io.Copy(c.conn, c.in)
	return err
}

func (c *telnetClient) Receive() error {
	if c.conn == nil {
		return io.EOF
	}
	_, err := io.Copy(c.out, c.conn)
	return err
}
