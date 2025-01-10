/*
 * skogul, detach sender - tests
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
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/sender"
)

func TestDetacher(t *testing.T) {
	c := skogul.Container{}
	m := skogul.Metric{}

	c.Metrics = []*skogul.Metric{&m}
	tst := &(sender.Test{})
	delay := &(sender.Sleeper{Base: skogul.Duration{Duration: time.Duration(100 * time.Millisecond)}, MaxDelay: skogul.Duration{Duration: time.Duration(100 * time.Millisecond)}, Next: skogul.SenderRef{S: tst}})
	detach := &(sender.Detacher{Next: skogul.SenderRef{S: delay}})

	start := time.Now()
	tst.TestQuick(t, detach, &c, 0)

	diff := time.Since(start)
	if diff > (20 * time.Millisecond) {
		t.Errorf("Took too long sending to the detach-sender. Took more than 20ms (%v). Should be ~instant.", diff)
	}
	time.Sleep(200 * time.Millisecond)
	if tst.Received() != 1 {
		t.Errorf("Didn't get the event after time expired? Wanted %d containers, got %d", 1, tst.Received())
	}
}

func TestFanout(t *testing.T) {
	c := skogul.Container{}
	m := skogul.Metric{}

	c.Metrics = []*skogul.Metric{&m}
	tst := &(sender.Test{})
	delay := &(sender.Sleeper{Base: skogul.Duration{Duration: time.Duration(300 * time.Millisecond)}, MaxDelay: skogul.Duration{Duration: time.Duration(1 * time.Millisecond)}, Next: skogul.SenderRef{S: tst}})
	fanout := &(sender.Fanout{Next: skogul.SenderRef{S: delay}, Workers: 3})

	start := time.Now()
	// With a work queue of 3, we can send 3 at the same time and not
	// worry.
	tst.TestQuick(t, fanout, &c, 0)
	tst.TestQuick(t, fanout, &c, 0)
	tst.TestQuick(t, fanout, &c, 0)

	diff := time.Since(start)
	if diff > (20 * time.Millisecond) {
		t.Errorf("Took too long sending to the fanout-sender. Took more than 20ms (%v). Should be ~instant.", diff)
	}
	time.Sleep(400 * time.Millisecond)
	if tst.Received() != 3 {
		t.Errorf("Didn't get the event after time expired? Wanted %d containers, got %d", 3, tst.Received())
	}

	start = time.Now()

	// need a delay between them to avoid race condition when reading
	// the first non-blocking. Otherwise the testquick after these
	// three would SOME times return rcv.Received() == 2, since two
	// delays finished and same time
	tst.TestQuick(t, fanout, &c, 0)
	time.Sleep(50 * time.Millisecond)
	tst.TestQuick(t, fanout, &c, 0)
	time.Sleep(50 * time.Millisecond)
	tst.TestQuick(t, fanout, &c, 0)

	diff = time.Since(start)
	if diff > (150 * time.Millisecond) {
		t.Errorf("Took too long sending to the fanout-sender. Took more than 120ms (%v). Should be ~100ms.", diff)
	}
	// Should block, have 1 received as one worker clears up.
	tst.TestQuick(t, fanout, &c, 1)

	diff = time.Since(start)
	if diff < (100 * time.Millisecond) {
		t.Errorf("Unexpectedly fast. Expected to block.Took less than 100ms (%v).", diff)
	}
	time.Sleep(400 * time.Millisecond)
	// Now all should be done...
	if tst.Received() != 4 {
		t.Errorf("Fanout: Expected 4 received events after timer(s) expired. Got %d", tst.Received())
	}
}
