/*
 * skogul, udp message receiver
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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

package receiver

import (
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/stats"
)

const UDPMaxReadSize = 65535

var udpLog = skogul.Logger("receiver", "udp")

// UDP contains the configuration for the receiver
type UDP struct {
	Address      string            `doc:"Address and port to listen to." example:"[::1]:3306"`
	Handler      skogul.HandlerRef `doc:"Handler used to parse, transform and send data."`
	Backlog      int               `doc:"Number of queued messages that are not delivered before the receiver starts blocking. Defaults to 100. Even when this is full, the kernel will still buffer data on top of this value. Higher values give smoother performance, but slightly more memory usage. Actual memory usage is a factor of average message size * backlog-fill."`
	Threads      int               `doc:"Number of worker go routines to use, which loosely translates to parallel execution. Defaults to number of CPU threads, with a minimum of 20. There is no correct number, but the value depends on how fast your senders are."`
	PacketSize   int               `doc:"UDP Packet size note: max. UDP read size is 65535"`
	FailureLevel string            `doc:"Level to log receiver failures as. Error, Warning, Info, Debug, or Trace. (default: Error)"`
	Buffer       int               `doc:"Set kernel read buffer. Default is kernel-specific. Bumping this will make it easier to handler bursty UDP traffic."`
	EmitStats    skogul.Duration   `doc:"How often to emit internal skogul stats for this receiver. Default: 10s (from stats.DefaultInterval)"`
	ch           chan []byte       // Used to pass messages from the accept/read-loop to the worker pool/threads.
	failureLevel logrus.Level
	once         sync.Once
	stats        *udpStats
}

// udpStats is a type containing internal stats of the UDP receiver
type udpStats struct {
	Received uint64 // number of received elements. For a receiver, this is the number of received incoming data.
	Errors   uint64 // number of errors encountered. For a receiver, this could be if it received malformed data.
	Sent     uint64 // number of successful elements encountered and passed on to the next chain.
}

// process is the worker-thread responsible for handling individual
// messages.
func (ud *UDP) process() {
	ud.once.Do(func() {
		if ud.FailureLevel == "" {
			ud.failureLevel = logrus.ErrorLevel
		} else {
			ud.failureLevel = skogul.GetLogLevelFromString(ud.FailureLevel)
		}
	})
	for {
		bytes := <-ud.ch
		atomic.AddUint64(&ud.stats.Received, 1)
		if err := ud.Handler.H.Handle(bytes); err != nil {
			atomic.AddUint64(&ud.stats.Errors, 1)
			udpLog.WithError(err).Log(ud.failureLevel, "Unable to handle UDP message")
		} else {
			atomic.AddUint64(&ud.stats.Sent, 1)
		}
	}
}

// Verify verifies the configuration for the UDP receiver
func (ud *UDP) Verify() error {
	if ud.Handler.Name == "" {
		return skogul.MissingArgument("Handler")
	}
	if ud.Address == "" {
		return skogul.MissingArgument("Address")
	}
	if ud.PacketSize < 0 || ud.PacketSize > UDPMaxReadSize {
		return fmt.Errorf("invalid udp packet size, maximum udp read size is between 0 and %d", UDPMaxReadSize)
	}
	return nil
}

// Start boots up ud.Threads number of worker threads, then starts
// listening for incoming UDP messages on the configured address. Start
// never returns.
func (ud *UDP) Start() error {
	if ud.PacketSize == 0 {
		ud.PacketSize = 9000
	}
	if ud.Backlog == 0 {
		ud.Backlog = 100
	}
	if ud.Threads == 0 {
		ud.Threads = runtime.NumCPU()
		if ud.Threads < 20 {
			ud.Threads = 20
		}
	}
	if ud.EmitStats.Duration == 0 {
		ud.EmitStats.Duration = stats.DefaultInterval
	}

	ud.initStats()

	udpLog.Tracef("Got backlog size of %d and number of threads %d", ud.Backlog, ud.Threads)
	ud.ch = make(chan []byte, ud.Backlog)
	for i := 0; i < ud.Threads; i++ {
		go ud.process()
	}

	udpip, err := net.ResolveUDPAddr("udp", ud.Address)
	if err != nil {
		return fmt.Errorf("unable to resolve address %s: %w", ud.Address, err)
	}
	ln, err := net.ListenUDP("udp", udpip)
	if err != nil {
		return err
	}
	if ud.Buffer > 0 {
		ln.SetReadBuffer(ud.Buffer)
	}
	for {
		bytes := make([]byte, ud.PacketSize)
		n, err := ln.Read(bytes)
		if err != nil || n == 0 {
			udpLog.WithError(err).WithField("bytes", n).Error("Unable to read UDP message")
			continue
		}
		ud.ch <- bytes[0:n]
	}
}

// GetStats prepares a skogul metric with stats
// for the UDP receiver.
func (ud *UDP) GetStats() *skogul.Metric {
	now := skogul.Now()

	udpLog.WithField("time", now).Trace("Getting stats")

	metric := skogul.Metric{
		Time:     &now,
		Metadata: make(map[string]interface{}),
		Data:     make(map[string]interface{}),
	}
	metric.Metadata["component"] = "receiver"
	metric.Metadata["type"] = "UDP"
	metric.Metadata["identity"] = skogul.Identity[ud]

	metric.Data["received"] = ud.stats.Received
	metric.Data["errors"] = ud.stats.Errors
	metric.Data["sent"] = ud.stats.Sent
	return &metric
}

// initStats initializes up the necessary components for stats
func (ud *UDP) initStats() {
	ud.stats = &udpStats{
		Received: 0,
		Errors:   0,
		Sent:     0,
	}
}
