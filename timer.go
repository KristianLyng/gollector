/*
 * skogul, time syncer
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

package skogul

import (
	"sync/atomic"
	"time"
)

const syncPeriod = time.Millisecond * 100

var ltime struct {
	sec  int64
	nsec int64
}

func init() {
	go syncTime()
}

func syncTime() {
	f := func() {
		t := time.Now()
		s := int64(t.Unix())
		ns := int64(t.Nanosecond())
		atomic.StoreInt64(&ltime.sec, s)
		atomic.StoreInt64(&ltime.nsec, ns)
	}
	f()
	myticker := time.NewTicker(syncPeriod)
	for range myticker.C {
		f()
	}
}

/*
Now returns an approximation of time.Now atomically. It his is achieved by
syncing time in a separate go routine to reduce overhead. Currently synced
10 times per second. The reason you may want this is:

	BenchmarkTimeNow-8     	130269282	        45.6 ns/op
	BenchmarkSkogulNow-8   	1000000000	         0.878 ns/op

If ten times per second isn't accurate enough, we may consider making this
configurable. The primary use case here is for transformers or parsers that
need to insert time, but might get thousands of metrics per second.
*/
func Now() time.Time {
	return time.Unix(ltime.sec, ltime.nsec)
}
