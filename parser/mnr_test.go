/*
 * skogul, mnr parser
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
	"io/ioutil"
	"testing"
	"time"

	"github.com/telenornms/skogul/parser"
)

func TestMNRLineParsed(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=val")
	p := parser.MNR{}

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	expectedTime := time.Unix(1599730066, 0)
	if *c.Metrics[0].Time != expectedTime {
		t.Errorf("Expected time to be %s but got %s", expectedTime, *c.Metrics[0].Time)
	}
}

func TestMNRChangedLineParsed(t *testing.T) {
	line := []byte("+r	1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=val")
	p := parser.MNR{}

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	expectedTime := time.Unix(1599730066, 0)
	if *c.Metrics[0].Time != expectedTime {
		t.Errorf("Expected time to be %s but got %s", expectedTime, *c.Metrics[0].Time)
	}
}

func TestMNRExtractValue(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=val")
	p := parser.MNR{
		DefaultFieldName: "value",
	}
	value := 0.0

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	if len(c.Metrics) != 1 || c.Metrics[0].Data == nil {
		t.Error("Missing metrics in container after mnr parse")
		return
	}

	if c.Metrics[0].Data[p.DefaultFieldName] != value {
		t.Errorf("Expected data to contain the key '%s' with the value '%f', but got '%v'", p.DefaultFieldName, value, c.Metrics[0].Data[p.DefaultFieldName])
	}
}

func TestMNRExtractProperty(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=val")
	p := parser.MNR{}

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	if len(c.Metrics) != 1 || c.Metrics[0].Data == nil {
		t.Error("Missing metrics in container after mnr parse")
		return
	}

	if c.Metrics[0].Data["key"] != "val" {
		t.Errorf("Expected data to contain the key 'key' with the value 'val', but got '%s'", c.Metrics[0].Data["key"])
	}
}

func TestMNRExtractTypedValueInt(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=1")
	p := parser.MNR{}
	var val int64 = 1

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	if len(c.Metrics) != 1 || c.Metrics[0].Data == nil {
		t.Error("Missing metrics in container after mnr parse")
		return
	}

	if c.Metrics[0].Data["key"] != val {
		t.Errorf("Expected data to contain the key 'key' with the value '1', but got '%s' (%t)", c.Metrics[0].Data["key"], c.Metrics[0].Data["key"])
	}
}

func TestMNRExtractTypedValueFloat(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=1.0")
	p := parser.MNR{}

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	if len(c.Metrics) != 1 || c.Metrics[0].Data == nil {
		t.Error("Missing metrics in container after mnr parse")
		return
	}

	if c.Metrics[0].Data["key"] != 1.0 {
		t.Errorf("Expected data to contain the key 'key' with the value '1.0', but got '%s' (%t)", c.Metrics[0].Data["key"], c.Metrics[0].Data["key"])
	}
}

func TestMNRExtractValues(t *testing.T) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=val	foo=bar")
	p := parser.MNR{}

	c, err := p.Parse(line)
	if err != nil {
		t.Errorf("MnR parser errored on parsing line: %s", err)
		return
	}

	if len(c.Metrics) != 1 || c.Metrics[0].Data == nil {
		t.Error("Missing metrics in container after mnr parse")
		return
	}

	if c.Metrics[0].Data["key"] != "val" {
		t.Errorf("Expected data to contain the key 'key' with the value 'val', but got '%s'", c.Metrics[0].Data["key"])
	}
	if c.Metrics[0].Data["foo"] != "bar" {
		t.Errorf("Expected data to contain the key 'foo' with the value 'bar', but got '%s'", c.Metrics[0].Data["foo"])
	}
}

func TestMNROnDataset(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/mnr.txt")
	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}

	p := parser.MNR{}
	container, err := p.Parse(b)
	if err != nil {
		t.Errorf("MnR parser errored on parsing data: %s", err)
		return
	}
	if len(container.Metrics) == 0 {
		t.Error("Expected metrics from parsing MnR test data set, got 0")
		return
	}
}

func BenchmarkMNRParse(b *testing.B) {
	line := []byte("1599730066	group	127.0.0.1.ifXTable..1.12.RATEP.Pkts/s.820424119	0.0	key=1")
	x := parser.MNR{}
	for i := 0; i < b.N; i++ {
		x.Parse(line)
	}
}
