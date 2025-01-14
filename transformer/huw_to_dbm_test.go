/*
 * Copyright (c) 2024 Telenor Norge AS
 * Author(s):
 *  - Hans Rafaelsen <hans.rafaelsen@telenor.no>
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

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/transformer"
)

func TestHUWtoDBMDefault(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["systemID"] = "xxx1"
	metric.Metadata["if_name"] = "SNMPv2-SMI::enterprises.2011.5.25.31.1.1.3.1.8.67305550"
	metric.Data = make(map[string]interface{})
	metric.Data["uW"] = float64(35)
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conv := transformer.HuWtoDBM{
		Source: "uW",
	}

	t.Logf("Container before transform:\n%v", c)
	err := conv.Transform(&c)
	if err != nil {
		t.Errorf("UWtoDBM returned non-nil err: %v", err)
	}

	t.Logf("Container after transform:\n%v", c)

	v, ok := c.Metrics[0].Data["dbm"].(float64)
	if !ok {
		t.Fatal("Failed to create 'dbm' field")
	}
	expect := -4.436974992327127
	if v != expect {
		t.Errorf("Failed to compute correct 'dbm' field. Got: %f. Expected: %f", v, expect)
	}
}

func TestHUWtoDBMDest(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["systemID"] = "xxx1"
	metric.Metadata["if_name"] = "SNMPv2-SMI::enterprises.2011.5.25.31.1.1.3.1.8.67305550"
	metric.Data = make(map[string]interface{})
	metric.Data["uW"] = float64(35)
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conv := transformer.HuWtoDBM{
		Source:      "uW",
		Destination: "foo",
	}

	t.Logf("Container before transform:\n%v", c)
	err := conv.Transform(&c)
	if err != nil {
		t.Errorf("UWtoDBM returned non-nil err: %v", err)
	}

	t.Logf("Container after transform:\n%v", c)

	v, ok := c.Metrics[0].Data["foo"].(float64)
	if !ok {
		t.Fatal("Failed to create destination field 'foo'")
	}
	expect := -4.436974992327127
	if v != expect {
		t.Errorf("Failed to compute correct 'foo' field. Got: %f. Expected: %f", v, expect)
	}
}

func TestHUWtoDBMDestTreshold(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["systemID"] = "xxx1"
	metric.Metadata["if_name"] = "SNMPv2-SMI::enterprises.2011.5.25.31.1.1.3.1.8.67305550"
	metric.Data = make(map[string]interface{})
	metric.Data["uW"] = float64(35)
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conv := transformer.HuWtoDBM{
		Source:      "uW",
		Destination: "foo",
		Treshold:    10,
	}

	t.Logf("Container before transform:\n%v", c)
	err := conv.Transform(&c)
	if err != nil {
		t.Errorf("UWtoDBM returned non-nil err: %v", err)
	}

	t.Logf("Container after transform:\n%v", c)

	v, ok := c.Metrics[0].Data["foo"].(float64)
	if !ok {
		t.Fatal("Failed to create destination field 'foo'")
	}
	expect := -8.696662315049938
	if v != expect {
		t.Errorf("Failed to compute correct 'foo' field. Got: %f. Expected: %f", v, expect)
	}
}

func TestHUWtoDBMNotFloat(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["systemID"] = "xxx1"
	metric.Metadata["if_name"] = "SNMPv2-SMI::enterprises.2011.5.25.31.1.1.3.1.8.67305550"
	metric.Data = make(map[string]interface{})
	metric.Data["uW"] = "foo"
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conv := transformer.HuWtoDBM{
		Source: "uW",
	}

	t.Logf("Container before transform:\n%v", c)
	err := conv.Transform(&c)
	if err == nil {
		t.Errorf("UWtoDBM check for error test failed")
	}

	t.Logf("Container after transform:\n%v", c)
}

func TestHUWtoDBMNegative(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["systemID"] = "xxx1"
	metric.Metadata["if_name"] = "SNMPv2-SMI::enterprises.2011.5.25.31.1.1.3.1.8.67305550"
	metric.Data = make(map[string]interface{})
	metric.Data["uW"] = float64(-35)
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conv := transformer.HuWtoDBM{
		Source: "uW",
	}

	t.Logf("Container before transform:\n%v", c)
	err := conv.Transform(&c)
	if err != nil {
		t.Errorf("UWtoDBM returned non-nil err: %v", err)
	}

	t.Logf("Container after transform:\n%v", c)

	v, ok := c.Metrics[0].Data["dbm"].(float64)
	if !ok {
		t.Fatal("Failed to create 'dbm' field")
	}
	expect := -40.0
	if v != expect {
		t.Errorf("Failed to compute correct 'dbm' field. Got: %f. Expected: %f", v, expect)
	}
}
