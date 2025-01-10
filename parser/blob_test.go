/*
 * skogul, test blob parser
 *
 * Copyright (c) 2023 Telenor Norge AS
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

package parser_test

import (
	"bytes"
	"testing"

	"github.com/telenornms/skogul/parser"
)

// TestJSONParse tests parsing of a simple JSON document to skogul
// container
func TestBlobParse(t *testing.T) {
	b := []byte("kjell magne bondevik uten mellomnavn")
	x := parser.Blob{}
	c, err := x.Parse(b)
	if err != nil {
		t.Errorf("Blob.Parse(b) failed: %s", err)
	}
	if len(c.Metrics) != 1 {
		t.Errorf("Blob.Parse(b) returned ok, but length is not 1")
	}
	b2, ok := c.Metrics[0].Data["data"].([]byte)
	if !ok {
		t.Errorf("data[\"data\"] is not a byte slice")
	}

	if !bytes.Equal(b2, b) {
		t.Errorf("data[\"data\"] is not %v!", b)
	}

}
