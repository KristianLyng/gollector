/*
 * skogul, tcpline receiver
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

package receiver

import (
	"bufio"
	"fmt"
	"net"

	"github.com/telenornms/skogul"
)

var tcpLog = skogul.Logger("receiver", "tcp")

/*
TCPLine listens on a IP:TCP port specified in the Address string and accepts
one container per line to be sent to the parser.

Example usage, assuming JSON parser:

	$ cat payloads/simple.json  | jq -c . | nc '::1' '1234'

Since this is not possible to secure, it should be avoided where possible and
placed as close to the data source. A good use of this model is to use a TCPLine
receiver on the same box that needs to write to it, combined with
skogul.senders.HTTP to forward over a more sensible channel.
*/
type TCPLine struct {
	Address string            `doc:"Address and port to listen to." example:"[::1]:3306"`
	Handler skogul.HandlerRef `doc:"Handler used to parse, transform and send data."`
}

/*
Start the TCP line receiver and run forever.

We close the write-side of the connection leaving it to the other side to
finish up. We should probably add a read-timeout in the future.
*/
func (tl *TCPLine) Start() error {
	tcpip, err := net.ResolveTCPAddr("tcp", tl.Address)
	if err != nil {
		return fmt.Errorf("unable to resolve address %s: %w", tl.Address, err)
	}
	ln, err := net.ListenTCP("tcp", tcpip)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			tcpLog.WithError(err).Error("Unable to accept connection")
			continue
		}
		go tl.handleConnection(conn)
	}
}

func (tl *TCPLine) handleConnection(conn *net.TCPConn) {
	scanner := bufio.NewScanner(conn)
	conn.CloseWrite()
	defer conn.CloseRead()
	for scanner.Scan() {
		bytes := scanner.Bytes()
		if err := tl.Handler.H.Handle(bytes); err != nil {
			tcpLog.WithError(err).Error("Unable to parse JSON")
		}
	}
	if err := scanner.Err(); err != nil {
		tcpLog.WithError(err).Error("Error reading line")
		return
	}
}
