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

package encoder

import (
	"fmt"
	"os"
	"sync"
	"time"

	avro "github.com/hamba/avro/v2"
	"github.com/telenornms/skogul"
)

type AvroContainer struct {
	Template *AvroMetric
	Metrics  []*AvroMetric
}

type AvroMetric struct {
	Time     time.Time
	Metadata map[string]interface{}
	Data     map[string]interface{}
}

type AVRO struct {
	Schema string
	s      avro.Schema
	err    error
	once   sync.Once
}

func (x *AVRO) Encode(c *skogul.Container) ([]byte, error) {
	x.once.Do(func() {
		b, err := os.ReadFile(x.Schema)
		x.err = err
		if x.err == nil {
			x.s = avro.MustParse(string(b))
			if x.s == nil {
				x.err = fmt.Errorf("parsed schema is nil")
			}
		}
	})
	if x.err != nil {
		return nil, x.err
	} // same as before
	tmpContainer := AvroContainer{}
	// This allocates the Metrics array with a length of 0 but a
	// *capacity* (size) that matches the original container.
	// Otherwise, Go would have to dynamically grow the array when we
	// append to it, which isn't super-expensive, but unnecessary when
	// we know exactly how large it will be)
	tmpContainer.Metrics = make([]*AvroMetric, 0, len(c.Metrics))

	for m := range c.Metrics {
		var tmpMetric AvroMetric
		tmpMetric.Metadata = c.Metrics[m].Metadata
		tmpMetric.Data = c.Metrics[m].Data
		tmpMetric.Time = *c.Metrics[m].Time
		tmpContainer.Metrics = append(tmpContainer.Metrics, &tmpMetric)
	}

	return avro.Marshal(x.s, &tmpContainer)
}

func (x *AVRO) EncodeMetric(m *skogul.Metric) ([]byte, error) {
	return nil, fmt.Errorf("not supported")
}
