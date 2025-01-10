/*
 * skogul, test avro parser
 * Copyright (c) 2022 Telenor Norge AS
 * Author:
 * - Roshini Narasimha Raghavan <roshiragavi@gmail.com>
 * - Kristian Lyngstøl <kly@kly.no>
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

func (x *AVRO) Parse(b []byte) (*skogul.Container, error) {
	x.once.Do(func() {
		s, err := os.ReadFile(x.Schema)
		x.err = err
		if x.err == nil {
			x.s = avro.MustParse(string(s))
			if x.s == nil {
				x.err = fmt.Errorf("parsed schema is nil")
			}
		}
	})
	if x.err != nil {
		return nil, fmt.Errorf("unable to load schema: %w", x.err)
	}

	tmpContainer := AvroContainer{}
	err := avro.Unmarshal(x.s, b, &tmpContainer)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal avro into container %w", err)
	}

	container := skogul.Container{}
	container.Metrics = make([]*skogul.Metric, 0, len(tmpContainer.Metrics))

	for m := range tmpContainer.Metrics {
		var tmpMetric skogul.Metric
		tmpMetric.Metadata = tmpContainer.Metrics[m].Metadata
		tmpMetric.Data = tmpContainer.Metrics[m].Data
		tmpMetric.Time = &tmpContainer.Metrics[m].Time
		container.Metrics = append(container.Metrics, &tmpMetric)
	}

	return &container, err
}

func (x *AVRO) ParseMetric(m *skogul.Metric) ([]byte, error) {
	return nil, fmt.Errorf("not supported")
}
