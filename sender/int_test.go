/*
 * skogul, internal sender tests
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

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/sender"
)

func TestNull(t *testing.T) {
	c := skogul.Container{}
	x := sender.Null{}

	err := x.Send(&c)
	if err != nil {
		t.Errorf("Debug.Send returned non-nil: %v", err)
	}
}

func TestDupe(t *testing.T) {
	c := skogul.Container{}
	one := &(sender.Test{})
	two := &(sender.Test{})
	dupe := sender.Dupe{Next: []*skogul.SenderRef{{S: one}, {S: two}}}

	err := dupe.Send(&c)
	if err != nil {
		t.Errorf("dupe.Send() failed: %v", err)
	}
	if one.Received() != 1 {
		t.Errorf("dupe.Send(), sender 1 expected %d recevied, got %d", 1, one.Received())
	}
	if two.Received() != 1 {
		t.Errorf("dupe.Send(), sender 2 expected %d recevied, got %d", 1, two.Received())
	}
}

func TestFallback(t *testing.T) {
	c := skogul.Container{}
	conf, err := config.Bytes([]byte(`
{
   "senders": {
      "foo": { 
        "type": "fallback",
        "next": ["fail","ok","ignore" ]
      },
      "fail": { 
          "type": "forwardfail",
          "next": "fake-fail"
      },
      "fake-fail": {"type": "test" },
      "ok": {"type": "test" },
      "ignore": {"type": "test" }
   }
}
`))
	if err != nil {
		t.Errorf("Failed to load config for fallback test: %v", err)
		return
	}

	one := conf.Senders["fake-fail"].Sender.(*sender.Test)
	two := conf.Senders["ok"].Sender.(*sender.Test)
	three := conf.Senders["ignore"].Sender.(*sender.Test)
	fb := conf.Senders["foo"].Sender.(*sender.Fallback)

	err = fb.Send(&c)
	if err != nil {
		t.Errorf("fallback.Send() failed: %v", err)
	}
	if one.Received() != 1 {
		t.Errorf("fallback.Send(), sender 1 expected %d recevied, got %d", 1, one.Received())
	}
	if two.Received() != 1 {
		t.Errorf("fallback.Send(), sender 2 expected %d recevied, got %d", 1, two.Received())
	}
	if three.Received() != 0 {
		t.Errorf("fallback.Send(), sender 3 expected %d recevied, got %d", 0, three.Received())
	}
}

func TestFallback_fail(t *testing.T) {
	c := skogul.Container{}
	one := &(sender.Test{})
	two := &(sender.Test{})
	three := &(sender.Test{})
	faf := &(sender.ForwardAndFail{Next: skogul.SenderRef{S: one}})
	fb := sender.Fallback{Next: []*skogul.SenderRef{{S: faf}, {S: two}, {S: three}}}

	err := fb.Send(&c)
	if err != nil {
		t.Errorf("fallback.Send() failed: %v", err)
	}
	if one.Received() != 1 {
		t.Errorf("fallback.Send(), sender 1 expected %d recevied, got %d", 1, one.Received())
	}
	if two.Received() != 1 {
		t.Errorf("fallback.Send(), sender 2 expected %d recevied, got %d", 1, two.Received())
	}
	if three.Received() != 0 {
		t.Errorf("fallback.Send(), sender 3 expected %d recevied, got %d", 0, two.Received())
	}
}

func TestForwardAndFail(t *testing.T) {
	c := skogul.Container{}
	one := &(sender.Test{})
	faf := sender.ForwardAndFail{Next: skogul.SenderRef{S: one}}

	err := faf.Send(&c)
	if err == nil {
		t.Errorf("forwardandfail.Send() .... failed to fail (returned true)")
	}
	if one.Received() != 1 {
		t.Errorf("forwardandfail.Send(), sender 1 expected %d recevied, got %d", 1, one.Received())
	}
}
