/*
 * skogul, tcpline tests
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

package sender_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
)

func TestTCPLine(t *testing.T) {
	sconf := fmt.Sprintf(`
{
  "receivers": {
    "tcpline": {
      "type": "tcp",
      "address": "localhost:1337",
      "handler": "h"
    }
  },
  "handlers": {
    "h": {
      "parser": "json",
      "transformers": [],
      "sender": "test"
    }
  },
  "senders": {
    "test": {
      "type": "test"
    },
    "net": {
      "type": "net",
      "address": "localhost:1337",
      "network": "tcp"
    },
    "bad1": {
      "type": "net",
      "address": "[hesterbestmenikkeipv6addresse]:3123",
      "network": "udp"
    },
    "bad2": {
      "type": "net",
      "address": "localhost:655655",
      "network": "udp"
    },
    "bad3": {
      "type": "net",
      "address": "localhost:3123",
      "network": "uxp"
    },
    "bad4": {
      "type": "net",
      "address": "localhost:1",
      "network": "tcp"
    },
    "udp": {
      "type": "net",
      "address": "localhost:1339",
      "network": "udp"
    }
  }
}`)

	conf, err := config.Bytes([]byte(sconf))
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	sTest := conf.Senders["test"].Sender.(*sender.Test)
	sNet := conf.Senders["net"].Sender
	sNetBad1 := conf.Senders["bad1"].Sender
	sNetBad2 := conf.Senders["bad2"].Sender
	sNetBad3 := conf.Senders["bad3"].Sender
	sNetBad4 := conf.Senders["bad4"].Sender
	sNetUdp := conf.Senders["udp"].Sender
	rcv := conf.Receivers["tcpline"].Receiver.(*receiver.TCPLine)

	if rcv == nil {
		t.Errorf("failed to get receiver")
		return
	}

	go rcv.Start()
	time.Sleep(time.Duration(10 * time.Millisecond))
	sNet.Send(&validContainer)
	time.Sleep(time.Duration(10 * time.Millisecond))
	if sTest.Received() != 1 {
		t.Errorf("Didn't receive thing on other end!")
	}
	sNet.Send(&validContainer)
	time.Sleep(time.Duration(10 * time.Millisecond))
	if sTest.Received() != 2 {
		t.Errorf("Didn't receive thing on other end!")
	}
	if err := sNetBad1.Send(&validContainer); err == nil {
		t.Errorf("Should've failed number1, but didn't")
	}
	if err := sNetBad2.Send(&validContainer); err == nil {
		t.Errorf("Should've failed number2, but didn't")
	}
	if err := sNetBad3.Send(&validContainer); err == nil {
		t.Errorf("Should've failed number3, but didn't")
	}
	if err := sNetBad4.Send(&validContainer); err == nil {
		t.Errorf("Should've failed number4, but didn't")
	}
	if err := sNetUdp.Send(&validContainer); err != nil {
		t.Errorf("UDP failed to send: %v", err)
	}
}
