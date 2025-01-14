/*
 * skogul, data structure tests
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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/parser"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"
)

var logBuffer bytes.Buffer

func init() {
	log.SetOutput(&logBuffer)
}

// Test the String() capability of containers. Unfortunately, this depends
// on marshaling happening in a predictable order, which there is no
// guarantee that it will. But it seems to work ok for now.
func TestString(t *testing.T) {
	want := `{
  "template": {},
  "metrics": [
    {
      "timestamp": "2019-03-15T11:08:02+01:00",
      "metadata": {
        "key": "value"
      },
      "data": {
        "integer": 5
      }
    }
  ]
}`
	b := []byte(want)
	j := parser.SkogulJSON{}
	c, err := j.Parse(b)
	if err != nil {
		t.Errorf("JSON.Parse(b) failed: %v", err)
		return
	}
	got := c.String()

	if got != want {
		t.Errorf("String() on container failed, wanted %v, got %v", want, got)
	}
}

func TestValidate(t *testing.T) {
	empty := skogul.Container{}
	err := empty.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an empty Container")
	}
	got := fmt.Sprintf("%s", err)
	want := "missing metrics[] data."
	if got != want {
		t.Errorf("Validate() expected reason %s, got %s", want, got)
	}

	noMetrics := skogul.Container{}
	noMetrics.Metrics = []*skogul.Metric{}
	err = noMetrics.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an Container with empty metrics[]")
	}
	got = fmt.Sprintf("%s", err)
	want = "empty metrics[] data."
	if got != want {
		t.Errorf("Validate() expected reason %s, got %s", want, got)
	}

	badMetrics := skogul.Container{}
	metric := skogul.Metric{}
	badMetrics.Metrics = []*skogul.Metric{&metric}
	err = badMetrics.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an Container with empty metrics[]")
	}
	got = fmt.Sprintf("%s", err)
	want = "missing timestamp for metric(0 metadatafields and 0 data-fields, time: <nil>, First 5 Metadata fields:  Data: )"
	if got != want {
		t.Errorf("Validate() expected reason %s, got %s", want, got)
	}

	now := time.Now()
	metric.Time = &now
	notimeMetrics := skogul.Container{}
	notimeMetrics.Metrics = []*skogul.Metric{&metric}
	err = notimeMetrics.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an Container with no data")
	}
	got = fmt.Sprintf("%s", err)
	want = fmt.Sprintf("missing data for metric(0 metadatafields and 0 data-fields, time: %s, First 5 Metadata fields:  Data: )", now.Format(time.RFC3339))
	if got != want {
		t.Errorf("Validate() expected reason {%s}, got {%s}", want, got)
	}

	metric.Data = make(map[string]interface{})
	notimeMetrics.Metrics = []*skogul.Metric{&metric}
	err = notimeMetrics.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an Container with no data")
	}
	got = fmt.Sprintf("%s", err)
	want = fmt.Sprintf("empty data for metric(0 metadatafields and 0 data-fields, time: %s, First 5 Metadata fields:  Data: )", now.Format(time.RFC3339))
	if got != want {
		t.Errorf("Validate() expected reason %s, got %s", want, got)
	}
	metric.Data["test"] = "foo"
	okc := skogul.Container{}
	okc.Metrics = []*skogul.Metric{&metric}
	err = okc.Validate(false)
	if err != nil {
		t.Errorf("Validate() failed when it should work: %v", err)
	}
}

func TestValidate_partial(t *testing.T) {
	p := parser.SkogulJSON{}
	c, err := p.Parse([]byte(`{
		"metrics": [
			{
				"timestamp": "2022-03-22T19:55:15+02:00",
				"metadata": { "ok": "nah" },
				"data": { "cakes": 5 }
			},
			{
				"timestamp": "2022-03-22T19:55:15+02:00",
				"metadata": { "ok": "nah" }
			},
			{
				"timestamp": "2022-03-22T19:55:15+02:00",
				"metadata": { "ok": "nah" },
				"data": { "cakes": 5 }
			}
		]
	}`))
	if err != nil {
		t.Errorf("Failed to parse basic container, err: %v", err)
	}
	err = c.Validate(false)
	if err == nil {
		t.Errorf("Validate() succeeded on an Container with bad data")
	}
	got := fmt.Sprintf("%s", err)
	want := "missing data for metric(1 metadatafields and 0 data-fields, time: 2022-03-22T19:55:15+02:00, First 5 Metadata fields:  [ok=nah] Data: )"
	if got != want {
		t.Errorf("Validate() expected reason %s, got %s", want, got)
	}
	if len(c.Metrics) != 3 {
		t.Errorf("Parsed failed to return correct number of metrics. Expected 3 prior to validation, got %d.", len(c.Metrics))
	}
	err = c.Validate(true)
	if err != nil {
		t.Errorf("Validate() failed on an Container with partially OK data")
	}
	if len(c.Metrics) != 2 {
		t.Errorf("Validate() failed to return correct number of metrics. Expected 2, got %d.", len(c.Metrics))
	}

}

func BenchmarkValidate(b *testing.B) {

	now := time.Now()
	metric := skogul.Metric{}
	metric.Data = make(map[string]interface{})
	metric.Data["test"] = "foo"
	metric.Time = &now
	okc := skogul.Container{}
	okc.Metrics = []*skogul.Metric{&metric}
	for i := 0; i < b.N; i++ {
		okc.Validate(false)
	}

}

func BenchmarkCompareText(b *testing.B) {
	data := []string{"the fox jumps over the some-variable=na something", "this is fine, nothing is on fire", "only 1337 allowed"}
	str := "only 1337 allowed"
	for i := 0; i < b.N; i++ {
		for x := 0; x < 3; x++ {
			if data[x] == str {
				continue
			}
		}
	}
}
func BenchmarkCompareRegexp(b *testing.B) {
	data := []string{"the fox jumps over the some-variable=na something", "this is fine, nothing is on fire", "only 1337 allowed"}
	exp, err := regexp.Compile(".*some-variable=.*")
	if err != nil {
		b.Fatalf("Couldn't compile regexp: %v", err)
	}
	for i := 0; i < b.N; i++ {
		for x := 0; x < 3; x++ {
			if exp.Match([]byte(data[x])) {
				continue
			}
		}
	}
}
func BenchmarkCompareSubstr(b *testing.B) {
	data := []string{"the fox jumps over the some-variable=na something", "this is fine, nothing is on fire", "only 1337 allowed"}
	for i := 0; i < b.N; i++ {
		for x := 0; x < 3; x++ {
			if strings.Contains(data[x], "1337") {
				continue
			}
		}
	}
}

func TestString_invalid(t *testing.T) {
	c := skogul.Container{}
	metric1 := skogul.Metric{}
	metric1.Data = make(map[string]interface{})
	metric1.Data["ch"] = make(chan string)
	c.Metrics = []*skogul.Metric{&metric1}

	str := c.String()
	want := ""
	if str != want {
		t.Errorf("Invalid metrics marshalled successfully. Wanted %s, got %s", want, str)
	}
}

func durationOK(t *testing.T, s string, check time.Duration) {
	t.Helper()
	d := skogul.Duration{}
	err := json.Unmarshal([]byte(s), &d)
	if err != nil {
		t.Errorf("Failed to unmarshal duration `%s'. Error: %v", s, err)
	}
	if check == 0 {
		return
	}
	if d.Duration != check {
		t.Errorf("Duration %s unmarshalled ok. But got %v and expected %v", s, d.Duration, check)
	}
}

func durationBAD(t *testing.T, s string) {
	t.Helper()
	d := skogul.Duration{}
	err := json.Unmarshal([]byte(s), &d)
	if err == nil {
		t.Errorf("Successfully unmarshalled bogus duration `%s' to %v", s, d)
	}
}

func TestDuration(t *testing.T) {
	durationOK(t, `"5h"`, time.Hour*5)
	durationOK(t, `"1ms"`, time.Millisecond)
	durationOK(t, `500`, time.Nanosecond*500)
	durationOK(t, `-1`, -time.Nanosecond)
	durationBAD(t, `chicken`)
	durationBAD(t, `"2019-09-12T15:00:00Z"`)
	durationBAD(t, "four score and seven years ago")
	durationBAD(t, `null`)
	durationBAD(t, `{`)
	durationBAD(t, `"two",fifty{"`)
}
