/*
 * skogul, test blob encoder
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *  - Håkon Solbjørg <hakon.solbjorg@telenor.com>
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

package encoder_test

import (
	"bytes"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

// TestBlobEncode performs a shallow test of encoding of a skogul container
// using blob.
// Tests single-metric encoding, encoding something that isn't a byte
// array and multiple (two) metrics.
func TestBlobEncode(t *testing.T) {
	c := skogul.Container{}
	c.Metrics = make([]*skogul.Metric, 1, 1)
	raw := []byte(`hei faderullan`)
	raw2 := []byte(`kjell magne bondevik uten mellomnavn`)
	raw3 := []byte(`hei faderullan:kjell magne bondevik uten mellomnavn`)

	m := skogul.Metric{
		Data: map[string]interface{}{
			"data": raw,
		},
	}

	m2 := skogul.Metric{
		Data: map[string]interface{}{
			"data": raw2,
		},
	}

	c.Metrics[0] = &m

	enc := encoder.Blob{}
	enc.Delimiter = []byte(`:`)
	b, err := enc.Encode(&c)
	if err != nil {
		t.Errorf("Encoding failed: %s", err)
	}
	if bytes.Compare(b, raw) != 0 {
		t.Errorf("Encoding failed, new and old not the same: %v vs %v", b, raw)
	}
	m.Data["data"] = "not a byte array"
	b, err = enc.Encode(&c)
	if err == nil {
		t.Errorf("Encoding failed: %s", err)
	}

	m.Data["data"] = raw
	c.Metrics = append(c.Metrics, &m2)
	b, err = enc.Encode(&c)
	if err != nil {
		t.Errorf("Encoding failed: %s", err)
	}
	if bytes.Compare(b, raw3) != 0 {
		t.Errorf("Encoding failed, the two don't compare")
	}
}
