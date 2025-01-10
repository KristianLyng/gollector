/*
 * skogul, backoff sender
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
	"sync/atomic"
	"time"

	"github.com/telenornms/skogul"
)

// Backoff sender will send to Next, but retry up to Retries times, with
// exponential backoff, starting with time.Duration
type Backoff struct {
	Next    skogul.SenderRef `doc:"The sender to try"`
	Base    skogul.Duration  `doc:"Initial delay after a failure. Will double for each retry"`
	Retries uint64           `doc:"Number of retries before giving up"`
	holdoff uint64
}

// Send with a delay
func (bo *Backoff) Send(c *skogul.Container) error {
	var err error
	delay := bo.Base.Duration
	t := atomic.LoadUint64(&bo.holdoff)
	if t > 0 {
		time.Sleep(delay)
	}
	for i := uint64(1); i <= bo.Retries; i++ {
		err = bo.Next.S.Send(c)
		if err == nil {
			if i > 1 {
				atomic.AddUint64(&bo.holdoff, 1-i)
			}
			return nil
		}
		atomic.AddUint64(&bo.holdoff, 1)
		time.Sleep(delay)
		delay = delay * 2
	}
	return err
}
