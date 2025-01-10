/*
 * skogul, timer benchmarks
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

package skogul_test

import (
	"testing"
	"time"

	"github.com/telenornms/skogul"
)

func wait() {
	myt := time.NewTimer(time.Millisecond * 200)
	<-myt.C
}

func TestNow(t *testing.T) {
	start := skogul.Now()
	wait()
	two := skogul.Now()
	if !start.Before(two) {
		t.Errorf("skogul.Now() in the middle was not after skogul.Now() from the start")
	}
	middle_now := time.Now()
	if !start.Before(middle_now) {
		t.Errorf("time.Now() in the middle was not after skogul.Now() from the start")
	}
	wait()
	three := skogul.Now()
	if !middle_now.Before(three) {
		t.Errorf("time.Now() in the middle was not before skogul.Now() from the end: middle_now: %v three: %v", middle_now, three)
	}
}

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now()
	}
}

func BenchmarkSkogulNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		skogul.Now()
	}
}
