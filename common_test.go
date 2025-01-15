/*
 * skogul, tests
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

package skogul_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/parser"
	"github.com/telenornms/skogul/sender"
	"github.com/telenornms/skogul/transformer"
)

func TestHandler(t *testing.T) {
	h1 := skogul.Handler{}
	h2 := skogul.Handler{}
	h2.SetParser(parser.SkogulJSON{})
	h3 := skogul.Handler{Transformers: []skogul.Transformer{}}
	h3.SetParser(parser.SkogulJSON{})
	h4 := skogul.Handler{Transformers: []skogul.Transformer{}, Sender: &(sender.Test{})}
	h4.SetParser(parser.SkogulJSON{})
	h5 := skogul.Handler{Transformers: []skogul.Transformer{nil}, Sender: &(sender.Test{})}
	h5.SetParser(parser.SkogulJSON{})

	err := h1.Verify()
	if err == nil {
		t.Errorf("Handler verification didn't spot empty handler")
	}
	err = h2.Verify()
	if err == nil {
		t.Errorf("Handler verification didn't spot parser-only handler")
	}
	err = h3.Verify()
	if err == nil {
		t.Errorf("Handler verification didn't spot parser-and-transformers-only handler")
	}

	err = h4.Verify()
	if err != nil {
		t.Errorf("Supposedly valid handler actually failed verification: %v", err)
	}
	err = h5.Verify()
	if err == nil {
		t.Errorf("Handler verification didn't spot nil-transformer")
	}
}

func TestAssert(t *testing.T) {
	skogul.Assert(true)
	skogul.Assert(1+1 != 0)
	skogul.Assert(t != nil)
	skogul.Assert(true, "foo")
}

func TestAssert_fail(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered from assert: ", r)
		}
	}()
	skogul.Assert(false)
	t.Errorf("skogul.Assert(false) called, but execution continued.")
}

func TestAssert_fail_arg(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("Recovered from assert: ", r)
		}
	}()
	skogul.Assert(false, "something")
	t.Errorf("skogul.Assert(false,\"test\") called, but execution continued.")
}

func TestParseInvalidContainerSuccess(t *testing.T) {
	data := []byte(`{"data": 1, "ts": "2020-01-01T00:00:00.0Z"}`)

	h := skogul.Handler{}
	h.SetParser(parser.SkogulJSON{})

	// Verify that a SkogulJSON{} parser successfully parses this
	// container even though it's not on the proper format
	_, err := h.Parse(data)
	if err != nil {
		t.Error("Failed to parse json data", err)
		return
	}
}

func TestParseAndTransformInvalidContainerFails(t *testing.T) {
	data := []byte(`{"data": 1, "ts": "2020-01-01T00:00:00.0Z"}`)

	h := skogul.Handler{}
	h.SetParser(parser.SkogulJSON{})

	c, err := h.Parse(data)
	if err != nil {
		t.Error("Failed to parse json data", err)
		return
	}

	// Verify that running a transformer does not fail
	// this container even though it's on an invalid format
	err = h.Transform(c)
	if err != nil {
		t.Error("Transformation unsuccessful even though it should pass")
		return
	}
}

func TestParseTransformAndSendInvalidContainerFails(t *testing.T) {
	data := []byte(`{"data": 1, "ts": "2020-01-01T00:00:00.0Z"}`)

	h := skogul.Handler{}
	h.SetParser(parser.SkogulJSON{})

	c, err := h.Parse(data)
	if err != nil {
		t.Error("Failed to parse json data", err)
		return
	}

	err = h.Transform(c)
	if err != nil {
		t.Error("Transformation unsuccessful even though it should pass")
		return
	}

	err = h.Send(c)
	if err == nil {
		t.Error("Sending container should fail if the container is invalid")
		return
	}
}

func TestParseAndTransformInvalidContainerSuccess(t *testing.T) {
	data := []byte(`{"metrics": [{ "metadata": {"foo":"bar"}}], "template": {"data": {"a": "b"}, "timestamp": "2020-01-01T00:00:00.0Z"}}`)

	h := skogul.Handler{}
	h.SetParser(parser.SkogulJSON{})

	c, err := h.Parse(data)
	if err != nil {
		t.Error("Failed to parse json data", err)
		return
	}

	templater := transformer.Templater{}

	h.Transformers = []skogul.Transformer{&templater}

	// Verify that running a transformer validates this container
	// successfully after transforming
	err = h.Transform(c)
	if err != nil {
		t.Error("Transformation of container failed after transforming it valid", err)
		return
	}
}

func TestParseInvalidContainerAndTransformItValid(t *testing.T) {
	tformat := "2006-01-02T15:04:05Z07:00"
	parsedTimestamp, err := time.Parse(tformat, "2020-01-01T00:00:00.0Z")
	if err != nil {
		t.Error("Failed to boot-strap test-cases by parsing hard-coded date string. This should never happen unless golang is bugged or broken.")
		return
	}
	timestring := parsedTimestamp.Format(tformat)
	data := []byte(fmt.Sprintf(`{"data": 1, "ts": "%s"}`, timestring))

	h := skogul.Handler{}
	h.SetParser(parser.JSON{})

	// Parse the data using the custom JSON handler
	c, err := h.Parse(data)
	if err != nil {
		t.Error("Failed to parse json data", err)
		return
	}

	// Extract timestamp from data
	parseTimestamp := transformer.Timestamp{}
	parseTimestamp.Source = []string{"ts"}
	parseTimestamp.Format = "RFC3339"

	h.Transformers = []skogul.Transformer{&parseTimestamp}

	err = h.Transform(c)
	// Make sure the transformer validates the container successfully
	if err != nil {
		t.Error("Failed to transform container", err)
		return
	}

	if c.Metrics[0].Time.UTC().Format(tformat) != timestring {
		t.Errorf("%v not matching expected time '%v'", c.Metrics[0].Time.UTC().String(), timestring)
		return
	}
}

func TestSecretIsRedacted(t *testing.T) {
	secret := skogul.Secret("hunter2")
	if secret.String() != "<redacted>" {
		t.Errorf("Expected secret to be redacted, but got %s", secret.String())
	}
}

func TestSecretIsExposed(t *testing.T) {
	secret := skogul.Secret("hunter2")
	if secret.Expose() != "hunter2" {
		t.Errorf("Expected secret to be 'hunter2', but got %s", secret.Expose())
	}
}
