/*
 * skogul, test json encoder
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
	"os"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
	"github.com/telenornms/skogul/parser"
)

// TestJSONEncode tests encoding of a simple JSON document from a skogul
// container
func TestJSONEncode(t *testing.T) {
	testJSON(t, "./testdata/simple_container.json", true)
}

func testJSON(t *testing.T, file string, match bool) {
	t.Helper()
	c, orig := parseJSON(t, file)
	b, err := encoder.JSON{}.Encode(c)
	if err != nil {
		t.Errorf("Encoding %s failed: %v", file, err)
		return
	}
	if len(b) <= 0 {
		t.Errorf("Encoding %s failed: zero length data", file)
		return
	}
	if !match {
		return
	}
	sorig := string(orig)
	snew := string(b)
	if len(sorig) < 2 {
		t.Logf("Encoding %s failed: original pre-encoded length way too short. Shouldn't happen.", file)
		t.FailNow()
	}
	// strip trailing new-line
	sorig = sorig[:len(sorig)-1]
	if sorig != snew {
		t.Errorf("Encoding %s failed: original and newly encoded container doesn't match", file)
		t.Logf("orig:\n'%s'", sorig)
		t.Logf("new:\n'%s'", snew)
		return
	}

}

func parseJSON(t *testing.T, file string) (*skogul.Container, []byte) {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Logf("Failed to read test data file: %v", err)
		t.FailNow()
		return nil, nil
	}
	container, err := parser.SkogulJSON{}.Parse(b)

	if err != nil {
		t.Logf("Failed to parse JSON data: %v", err)
		t.FailNow()
		return nil, nil
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Logf("Expected parsed JSON to return a container with at least 1 metric")
		t.FailNow()
		return nil, nil
	}
	return container, b

}
