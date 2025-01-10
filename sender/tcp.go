/*
 * skogul, tcp sender
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
 * 02110-1301  USA
 */

package sender

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

var tcpLog = skogul.Logger("sender", "tcp")

// TCP sender is optimized around TCP sockets, though it is by no means
// perfect. Ideally, use a properly stateful protocol like HTTP.
//
// The main issue with the TCP sender is buffering and error detection. It
// is difficult to be both performant and detect errors as they happen, and
// this isn't made easier by the abstractions of Go.
//
// As such, this sender works well, but does not offer a guarantee that
// errors are actually detected when they happen.
type TCP struct {
	Address      string            `doc:"Address to send data to" example:"192.168.1.99:1234"`
	MaxRetries   int               `doc:"Maximum number of retries if sending fails, defaults to 1. Should be at least 1 to account for closed connections. Set to -1 to disable retries."`
	Encoder      skogul.EncoderRef `doc:"Encoder to use. Defaults to JSON-encoding."`
	DialTimeout  skogul.Duration   `doc:"Timeout for dialing. Includes DNS lookups and tcp connect."`
	WriteTimeout skogul.Duration   `doc:"Write timeout. Strongly advised to set this to single-digit seconds."`
	KeepAlive    skogul.Duration   `doc:"Keepalive timer for TCP."`
	Delimiter    []byte            `doc:"Optional delimiter, base64-encoded. Appended after every message. Should match delimiter of any encoder as well, possibly."`
	Threads      int               `doc:"Number of threads to start up, matches number of max connection held open. Currently used in round-robin...ish, until one thread blocks, then the others pick up the slack."`
	dialer       net.Dialer
	queue        chan tcpMessage
	once         sync.Once
}

// tcpMessage is a single encoded message to be passed, it is used between
// the Send() thread and the worker pool, allowing send() to wait for a
// reply on the answer channel.
type tcpMessage struct {
	b      []byte
	answer chan error
}

// tcpWorker is used per thread and largely contains copies of data from
// TCP struct, with the exception being the connection itself and its
// state.
type tcpWorker struct {
	conn      *net.TCPConn
	connected bool

	wt        time.Duration
	delimiter []byte
	address   string
	dialer    net.Dialer
}

// connect dials the far end, closes the read side and sets state.
func (w *tcpWorker) connect() error {
	conn, err := w.dialer.Dial("tcp", w.address)
	if err == nil {
		w.connected = true
	} else {
		w.connected = false
		return err
	}
	var ok bool
	w.conn, ok = conn.(*net.TCPConn)
	if !ok {
		w.connected = false
		conn.Close()
		return fmt.Errorf("unable to cast TCP connection to tcp connection state, makes no sense.")
	}
	err = w.conn.CloseRead()
	if err != nil {
		w.connected = false
		w.conn.Close()
		return fmt.Errorf("unable to close read-side")
	}
	return nil
}

// write to a connection, check for error, close the connection on fail.
func (w *tcpWorker) write(b []byte) error {
	n, err := w.conn.Write(b)
	if err != nil {
		w.conn.Close()
		w.connected = false
		return err
	}
	// XXX: How should we handle this? Close the connection? It's kinda
	// wonky.
	if n != len(b) {
		return fmt.Errorf("write succeeded but only %d of %d bytes sent", n, len(b))
	}
	return nil
}

// handle deals with a single tcpMessage
func (w *tcpWorker) handle(m tcpMessage) error {
	if !w.connected {
		if err := w.connect(); err != nil {
			return fmt.Errorf("unable to connect to %s: %w", w.address, err)
		}
	}
	if err := w.conn.SetWriteDeadline(time.Now().Add(w.wt)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	if err := w.write(m.b); err != nil {
		return fmt.Errorf("regular write: %w", err)
	}
	// XXX: Even with an empty delimiter, this one has proven
	// important: for reasons unknown to me (Kristian), possibly
	// related to buffering, the Write() never fails immediately even
	// if the connection is timed out, but it WILL clean up the
	// connection and thus fail on the _next_ write, so we essentially
	// need two writes.
	if err := w.write(w.delimiter); err != nil {
		return fmt.Errorf("delimiter write: %w", err)
	}

	return nil
}

// worker is the main-loop for the worker threads. Sets up basics and
// listens for messages.
func (t *TCP) worker() {
	w := tcpWorker{}
	w.address = t.Address
	w.dialer = t.dialer
	w.delimiter = t.Delimiter
	w.wt = t.WriteTimeout.Duration
	for msg := range t.queue {
		retries := 0
		var err error
		for retries = 0; retries <= t.MaxRetries; retries++ {
			err = w.handle(msg)
			if err == nil {
				break
			}
		}
		if err != nil {
			msg.answer <- fmt.Errorf("send failed after %d retries: %w", retries-1, err)
		} else {
			if retries > 0 {
				tcpLog.Infof("send ok after %d retries", retries)
			}
			msg.answer <- nil
			close(msg.answer)
		}
	}
}

// basic setup.
func (t *TCP) setup() {
	if t.MaxRetries == 0 {
		t.MaxRetries = 1
	}
	if t.MaxRetries < 0 {
		t.MaxRetries = 0
	}
	if t.Threads < 1 {
		t.Threads = 1
	}
	if t.Encoder.Name == "" {
		t.Encoder.Name = "json"
		t.Encoder.E = encoder.JSON{}
	}
	t.dialer = net.Dialer{
		Timeout:   t.DialTimeout.Duration,
		KeepAlive: t.KeepAlive.Duration,
	}
	t.queue = make(chan tcpMessage, 100)
	for i := 0; i < t.Threads; i++ {
		go t.worker()
	}
}

// Send sends metrics over TCP
func (t *TCP) Send(c *skogul.Container) error {
	t.once.Do(func() {
		t.setup()
	})

	msg := tcpMessage{}
	var err error
	msg.b, err = t.Encoder.E.Encode(c)
	if err != nil {
		return fmt.Errorf("encoding failed: %w", err)
	}
	msg.answer = make(chan error)
	t.queue <- msg
	err = <-msg.answer
	return err
}

func (t *TCP) Verify() error {
	if t.Address == "" {
		return skogul.MissingArgument("Address")
	}
	return nil
}
