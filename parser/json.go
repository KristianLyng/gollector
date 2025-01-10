/*
 * skogul, json parser
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

package parser

import (
	"encoding/json"

	"github.com/telenornms/skogul"
)

// SkogulJSON parses a byte string-representation of a Container in the
// format Skogul produces.
type SkogulJSON struct{}

// Parse accepts a byte slice of JSON data and marshals it into a container
func (x SkogulJSON) Parse(b []byte) (*skogul.Container, error) {
	container := skogul.Container{}
	err := json.Unmarshal(b, &container)
	return &container, err
}

// JSONMetric matches encoder's EncodeMetric - reads the byte as a metric,
// not container
type JSONMetric struct{}

// Parse accepts a byte slice of JSON data and marshals it into a metric,
// then wraps it in a container
func (x JSONMetric) Parse(b []byte) (*skogul.Container, error) {
	container := skogul.Container{}
	metric := skogul.Metric{}
	err := json.Unmarshal(b, &metric)
	metrics := []*skogul.Metric{&metric}
	container.Metrics = metrics
	return &container, err
}

// JSON is schemaless JSON. If data is sent between Skogul instances, SkogulJSON should be used instead
// which retains the data structure Skogul works with. A distinction between 'Skogul' and 'JSON' is made
// because historically, Skogul accepted 'json' as a configuration option for its own JSON format, now
// named 'skogul'.
// JSON can be useful e.g. as the first step of parsing from a third party source where modifying
// the source data structure might be hard/impossible
type JSON struct{}

// Parse accepts a byte slice of JSON data and marshals it into an empty skogul.Container
func (data JSON) Parse(b []byte) (*skogul.Container, error) {
	// The Validate() func of a container expects a timestamp to be valid.
	// Better way to fix?
	time := skogul.Now()
	metric := skogul.Metric{
		Metadata: make(map[string]interface{}),
		Time:     &time,
	}

	err := json.Unmarshal(b, &metric.Data)
	if err != nil {
		// Try to marshal data in an array form
		var array []interface{}
		err = json.Unmarshal(b, &array)
		if err != nil {
			return nil, err
		}

		metric.Data = make(map[string]interface{})
		metric.Data["blob"] = array
	}

	container := skogul.Container{
		Metrics: []*skogul.Metric{&metric},
	}

	return &container, nil
}
