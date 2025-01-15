/*
 * skogul, rfc 5424 structured data parser
 *
 * Copyright (c) 2020 Telenor Norge AS
 * Author(s):
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

package parser_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/telenornms/skogul/parser"
)

func TestStructuredDataParseExample1(t *testing.T) {
	b := []byte(`[exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"]`)

	p := parser.StructuredData{}

	c, err := p.Parse(b)
	if err != nil {
		t.Error("Failed to parse valid format")
		return
	}

	sdID := "exampleSDID@32473"
	if c.Metrics[0].Metadata["sd-id"] != sdID {
		t.Errorf("Expected SD-ID of '%s', got '%s'", sdID, c.Metrics[0].Metadata["sd-id"])
	}

	data := c.Metrics[0].Data
	if data["iut"] != "3" || data["eventSource"] != "Application" || data["eventID"] != "1011" {
		t.Errorf("Failed to parse one or more params from the structured data, got: %v", c.Metrics[0].Data)
	}
}

func TestStructuredDataParseExample2(t *testing.T) {
	b := []byte(`[exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"][examplePriority@32473 class="high"]`)

	p := parser.StructuredData{}

	c, err := p.Parse(b)
	if err != nil {
		t.Error("Failed to parse valid format")
		return
	}

	sdID := "exampleSDID@32473"
	if c.Metrics[0].Metadata["sd-id"] != sdID {
		t.Errorf("Expected SD-ID of '%s', got '%s'", sdID, c.Metrics[0].Metadata["sd-id"])
	}

	data := c.Metrics[0].Data
	expected := make(map[string]interface{})
	expected["iut"] = "3"
	expected["eventSource"] = "Application"
	expected["eventID"] = "1011"
	for k, v := range expected {
		if data[k] != expected[k] {
			t.Errorf("Expected '%s' to be '%s', got '%s'", k, v, data[k])
		}
	}
	if c.Metrics[1].Data["class"] != "high" {
		t.Errorf("Expected '%s' to be '%s', got '%s'", "class", "high", c.Metrics[1].Data["class"])
	}
}

func TestStructuredDataParseExample3Fails(t *testing.T) {
	b := []byte(`[exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] [examplePriority@32473 class="high"]`)
	p := parser.StructuredData{}

	c, err := p.Parse(b)
	if err != nil {
		t.Errorf("Expected parser to parse up until the invalid format, got err: %v", err)
		return
	}

	if len(c.Metrics) != 1 {
		t.Errorf("Expected to parse 1 metric, got %d", len(c.Metrics))
	}
}

func TestStructuredDataParseExample4(t *testing.T) {
	allowBracketSpace := true // https://datatracker.ietf.org/doc/html/draft-ietf-syslog-protocol-23#section-6.3.5 Example 4 *may* be allowed.
	b := []byte(`[ exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"][examplePriority@32473 class="high"]`)
	p := parser.StructuredData{}

	if allowBracketSpace {
		c, err := p.Parse(b)
		if err != nil {
			t.Errorf("Expected parser to handle invalid format, err:%v", err)
			return
		}
		if len(c.Metrics) != 2 {
			t.Errorf("Expected parser to handle invalid format and parse 2 metrics, got %d", len(c.Metrics))
		}
	} else {
		if _, err := p.Parse(b); err == nil {
			t.Error("Expected parser to error on invalid format")
			return
		}
	}
}

func TestStructuredDataParseNoHostnameAllowed(t *testing.T) {
	b := []byte(`[iut="3" eventSource="Application" eventID="1011"]`)
	p := parser.StructuredData{}

	c, err := p.Parse(b)
	if err != nil {
		t.Error("Expected parser to parse even though no hostname in data")
		return
	}
	if c.Metrics[0].Data["iut"] != "3" {
		t.Errorf("Expected structured data parser to return 3 for iut, but got %v", c.Metrics[0].Data["iut"])
	}
}

func TestStructuredDataParseNoHostnameAllowedMultipleMetrics(t *testing.T) {
	b := []byte(`[iut="3" eventSource="Application" eventID="1011"]`)
	p := parser.StructuredData{}

	c, err := p.Parse([]byte(fmt.Sprintf("%s%s", b, b)))
	if err != nil {
		t.Error("Expected parser to parse even though no hostname in data")
		return
	}
	if c.Metrics[0].Data["iut"] != "3" {
		t.Errorf("Expected structured data parser to return 3 for iut, but got %v", c.Metrics[0].Data["iut"])
	}
}

func TestStructuredDataParseNoContentResultsInOneMetric(t *testing.T) {
	b := []byte(`[exampleSDID@32473]`)
	p := parser.StructuredData{}

	c, err := p.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data: %v", err)
		return
	}

	if len(c.Metrics) != 1 {
		t.Errorf("Expected SD-ID-only metric to return 1 metric, got %d", len(c.Metrics))
	}
}

func TestStructuredDataOnDataset(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/structured_data.txt")
	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}

	p := parser.StructuredData{}
	container, err := p.Parse(b)
	if err != nil {
		t.Errorf("Structured data parser errored on parsing data: %s", err)
		return
	}
	if len(container.Metrics) == 0 {
		t.Error("Expected metrics from parsing structured data test data set, got 0")
		return
	}
}
