/*
 * skogul, test avro encoder
 *
 * Copyright (c) 2022 Telenor Norge AS
 * Author:
 *  - Roshini Narasimha Raghavan <roshiragavi@gmail.com>
 *  - Kristian Lyngstøl <kly@kly.no>
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
	"io/ioutil"
	"strings"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
	"github.com/telenornms/skogul/parser"
)

// TestAVROEncode tests encoding of a simple avro format from a skogul
// container
func TestAVROEncode(t *testing.T) {
	testAVRO(t, "./testdata/avro_testdata.json", true)
}

func testAVRO(t *testing.T, file string, match bool) {
	t.Helper()
	c, orig := parseAVRO(t, file)
	b, err := encoder.AVRO{}.Encode(c)

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

	result := strings.Compare(sorig, snew)
	if result != 0 {
		t.Errorf("Encoding %s failed: original and newly encoded container doesn't match", file)
		t.Logf("orig:\n'%s'", sorig)
		t.Logf("new:\n'%s'", snew)
		t.Logf("result\n %d", result)
		return
	}

}

func parseAVRO(t *testing.T, file string) (*skogul.Container, []byte) {
	t.Helper()

	b, err := ioutil.ReadFile(file)

	if err != nil {
		t.Logf("Failed to read test data file: %v", err)
		t.FailNow()
		return nil, nil
	}

	p := parser.AVRO{
		Schema: "../docs/examples/avro/avro_schema",
	}
	container, err := p.Parse(b)

	if err != nil {
		t.Logf("Failed to parse AVRO data: %v", err)
		t.FailNow()
		return nil, nil
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Logf("Expected parsed AVRO to return a container with at least 1 metric")
		t.FailNow()
		return nil, nil
	}
	return container, b

}
