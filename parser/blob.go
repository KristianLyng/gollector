/*
 * skogul, blob parser
 *
 * Copyright (c) 2023 Telenor Norge AS
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

package parser

import (
	"github.com/telenornms/skogul"
)

type Blob struct{}

// Parse accepts a byte slice of arbitrary data and stores it on
// data["data"] unprocessed
func (x Blob) Parse(b []byte) (*skogul.Container, error) {
	container := skogul.Container{}
	container.Metrics = make([]*skogul.Metric, 1)
	m := skogul.Metric{}
	container.Metrics[0] = &m
	container.Metrics[0].Data = make(map[string]interface{})
	container.Metrics[0].Data["data"] = b
	return &container, nil
}
