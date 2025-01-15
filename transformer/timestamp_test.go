/*
 * skogul, timestamp transformer
 *
 * Copyright (c) 2019 Telenor Norge AS
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

package transformer_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/transformer"
)

func TestTimestampParse(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/data-with-timestamp.json")
	if err != nil {
		t.Errorf("Could not read json data file: %v", err)
		return
	}

	var jsonData map[string]interface{}
	err = json.Unmarshal(b, &jsonData)
	if err != nil {
		t.Errorf("Could not parse JSON data: %v", err)
		return
	}

	jTimestamp, ok := jsonData["timestamp"].(string)

	if !ok {
		t.Errorf("Failed to cast original timestamp from JSON file to string (%s)", jTimestamp)
		return
	}

	jsonTimestamp, err := time.Parse(time.RFC3339, jTimestamp)
	if err != nil {
		t.Errorf("Failed to parse original timestamp from JSON file: %v (%s)", err, jTimestamp)
		return
	}

	now := time.Now()

	c := skogul.Container{
		Metrics: []*skogul.Metric{
			{
				Time:     &now,
				Metadata: make(map[string]interface{}),
				Data:     jsonData,
			},
		},
	}

	transformer := transformer.Timestamp{
		Source: []string{"timestamp"},
		Format: "rfc3339",
		Fail:   false,
	}

	transformer.Transform(&c)

	if (*c.Metrics[0].Time).String() != jsonTimestamp.String() {
		t.Errorf("Expected timestamps from JSON tile to match the one from the container. Expected: %s, Received: %s", jsonTimestamp, c.Metrics[0].Time)
		return
	}
}
