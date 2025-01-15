/*
 * skogul, templating tests
 *
 * Copyright (c) 2019 Telenor Norge AS
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

package transformer_test

import (
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/transformer"
)

func TestTemplate(t *testing.T) {
	now := time.Now()

	metric := skogul.Metric{}
	metric.Data = make(map[string]interface{})
	metric.Data["test"] = "foo"

	metric2 := skogul.Metric{}
	metric2.Metadata = make(map[string]interface{})
	metric2.Metadata["test"] = "int"

	template := skogul.Metric{}
	template.Time = &now
	template.Metadata = make(map[string]interface{})
	template.Data = make(map[string]interface{})
	template.Metadata["test"] = "foo"
	template.Data["test"] = "temp"
	template.Data["blank"] = "temp"

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric, &metric2}
	c.Template = &template

	templater := transformer.Templater{}

	err := templater.Transform(&c)
	if err != nil {
		t.Errorf("Templater() returned non-nil err: %v", err)
	}
	if c.Metrics[0].Time != &now {
		t.Errorf("Templater() did not expand time. Wanted %v got %v", &now, c.Metrics[0].Time)
	}
	if c.Metrics[0].Metadata["test"] != "foo" {
		t.Errorf("Templater() did not expand Metadata correctly. Wanted Metadata[\"test\"]=\"foo\", got Metadata[\"test\"]=%v", c.Metrics[0].Metadata["test"])
	}
	if c.Metrics[0].Data["test"] != "foo" {
		t.Errorf("Templater() Data[\"test\"]=\"foo\", got Data[\"test\"]=%v", c.Metrics[0].Data["test"])
	}
	if c.Metrics[0].Data["blank"] != "temp" {
		t.Errorf("Templater() Data[\"blank\"]=\"temp\", got Data[\"blank\"]=%v", c.Metrics[0].Data["blank"])
	}
	if c.Metrics[1].Metadata["test"] != "int" {
		t.Errorf("Templater() overwrote metric metadata Metrics[1].Metadata[\"test\"], expected \"int\", got \"%v\"", c.Metrics[1].Metadata["test"])
	}
	if c.Template != nil {
		t.Errorf("Templater() left template intact: %v", c.Template)
	}
}

func TestTemplate_blank(t *testing.T) {
	now := time.Now()

	metric := skogul.Metric{}
	metric.Data = make(map[string]interface{})
	metric.Data["test"] = "foo"
	metric.Time = &now

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	templater := transformer.Templater{}

	err := templater.Transform(&c)
	if err != nil {
		t.Errorf("Templater() returned non-nil err: %v", err)
	}
	if c.Metrics[0].Time != &now {
		t.Errorf("Templater() wanted %v got %v", &now, c.Metrics[0].Time)
	}
	if c.Metrics[0].Data["test"] != "foo" {
		t.Errorf("Templater() Wanted Data[\"test\"]=\"foo\", got Data[\"test\"]=%v", c.Metrics[0].Metadata["test"])
	}
	if c.Template != nil {
		t.Errorf("Templater() left template intact: %v", c.Template)
	}
}
