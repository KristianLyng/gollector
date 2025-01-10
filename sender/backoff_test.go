/*
 * skogul, backoff sender tests
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

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/sender"
)

// BackTester will fail until it has failed fails times.
type BackTester struct {
	fails int
}

func (bt *BackTester) Send(c *skogul.Container) error {
	if bt.fails > 0 {
		bt.fails--
		return fmt.Errorf("still failing")
	}
	return nil
}

// TestBackoff tests if backoff works at least a little bit
func TestBackoff(t *testing.T) {
	te := BackTester{fails: 1}
	bo := sender.Backoff{
		Next:    skogul.SenderRef{S: &te},
		Base:    skogul.Duration{Duration: time.Duration(time.Millisecond * 10)},
		Retries: 2,
	}
	err := bo.Send(&validContainer)
	if err != nil {
		t.Errorf("Got error from bo.Send(): %v", err)
	}
	te.fails = 10
	err = bo.Send(&validContainer)
	if err == nil {
		t.Errorf("Didn't get error from bo.Send()")
	}
}
