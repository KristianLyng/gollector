/*
 * skogul, switch transformer tests
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
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/transformer"
)

const (
	testMetadataA = `{ "sensor": "a" }`
	testMetadataB = `{ "sensor": "b" }`
	testData      = `{ "bannable_field": "someValue", "removable_field": "someOtherValue", "data": "42" }`
)

func generateContainer() skogul.Container {
	metric := skogul.Metric{}
	json.Unmarshal([]byte(testMetadataA), &metric.Metadata)
	json.Unmarshal([]byte(testData), &metric.Data)

	container := skogul.Container{
		Metrics: []*skogul.Metric{&metric},
	}

	return container
}

func TestSwitchTransformerRunsWithoutError(t *testing.T) {
	case1 := transformer.Case{
		When: "sensor",
		Is:   "a",
	}

	sw := transformer.Switch{
		Cases: []transformer.Case{case1},
	}

	container := generateContainer()
	err := sw.Transform(&container)
	if err != nil {
		t.Errorf("Switch transformer returned err: %v", err)
	}
}

func TestSwitchTransformerRunsSpecifiedTransformer(t *testing.T) {
	conf := testConfOk(t, `
	{
		"transformers": {
			"switch": {
				"type": "switch",
				"cases": [
					{
						"when": "sensor",
						"is": "a",
						"transformers": ["remove"]
					}
				]
			},
			"remove": {
				"type": "data",
				"remove": ["removable_field"]
			}
		}
	}`)

	container := generateContainer()

	err := conf.Transformers["switch"].Transformer.Transform(&container)
	if err != nil {
		t.Errorf("Switch transformer returned error %v", err)
	}

	if container.Metrics[0].Data["removable_field"] != nil {
		t.Errorf("Failed to remove field using switch transformer, 'removable_field' should be removed but is '%v'", container.Metrics[0].Data["removable_field"])
	}
}

func TestSwitchTransformerDoesNotRunNonSpecifiedTransformer(t *testing.T) {
	conf := testConfOk(t, `
	{
		"transformers": {
			"switch": {
				"type": "switch",
				"cases": [
					{
						"when": "sensor",
						"is": "a",
						"transformers": ["require"]
					}
				]
			},
			"require": {
				"type": "data",
				"require": ["data"]
			}
		}
	}`)

	container := generateContainer()

	err := conf.Transformers["switch"].Transformer.Transform(&container)
	if err != nil {
		t.Errorf("Switch transformer returned error %v", err)
	}

	if container.Metrics[0].Data["data"] != "42" {
		t.Errorf("Expected transformer to not modify metrics, 'data' should be '42' but is '%v'", container.Metrics[0].Data["data"])
	}

	if container.Metrics[0].Data["removable_field"] != "someOtherValue" {
		t.Errorf("Exptected transformer to not modify metrics, 'removable_field' should be 'someOtherValue' but is is '%v'", container.Metrics[0].Data["removable_field"])
	}
}

func TestSwitchOnNestedField(t *testing.T) {
	data := `{"metrics": [{"metadata": {"foo": {"bar": "baz"}, "remove": "me"}}]}`
	caseTransformer := transformer.Metadata{
		Remove: []string{"remove"},
	}
	tref := skogul.TransformerRef{
		T: &caseTransformer,
	}
	transform := transformer.Switch{
		Cases: []transformer.Case{
			{
				When:         "/foo/bar",
				Is:           "baz",
				Transformers: []*skogul.TransformerRef{&tref},
			},
		},
	}

	c := skogul.Container{}
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Errorf("failed to parse test case data: %s", err)
		return
	}

	if err := transform.Transform(&c); err != nil {
		t.Errorf("failed to run transform: %s", err)
		return
	}

	if c.Metrics[0].Metadata["remove"] != nil {
		t.Error("failed to run switch transformer based on nested field")
	}
}

func TestSwitchCaseConditionNonString(t *testing.T) {
	data := `{"metrics": [{"metadata": {"foo": true, "remove": "me"}}]}`
	caseTransformer := transformer.Metadata{
		Remove: []string{"remove"},
	}
	tref := skogul.TransformerRef{
		T: &caseTransformer,
	}
	transform := transformer.Switch{
		Cases: []transformer.Case{
			{
				When:         "/foo",
				Is:           true,
				Transformers: []*skogul.TransformerRef{&tref},
			},
		},
	}

	c := skogul.Container{}
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Errorf("failed to parse test case data: %s", err)
		return
	}

	if err := transform.Transform(&c); err != nil {
		t.Errorf("failed to run transform: %s", err)
		return
	}

	if c.Metrics[0].Metadata["remove"] != nil {
		t.Error("failed to run switch transformer based on nested field")
	}
}

func TestSwitchCaseFieldExists(t *testing.T) {
	data := `{"metrics": [{"metadata": {"foo": true, "remove": "me"}}]}`
	caseTransformer := transformer.Metadata{
		Remove: []string{"remove"},
	}
	tref := skogul.TransformerRef{
		T: &caseTransformer,
	}
	transform := transformer.Switch{
		Cases: []transformer.Case{
			{
				When:         "/foo",
				Exists:       true,
				Transformers: []*skogul.TransformerRef{&tref},
			},
		},
	}

	c := skogul.Container{}
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Errorf("failed to parse test case data: %s", err)
		return
	}

	if err := transform.Transform(&c); err != nil {
		t.Errorf("failed to run transform: %s", err)
		return
	}

	if c.Metrics[0].Metadata["remove"] != nil {
		t.Error("switch transformer did not remove field based on 'exists' option")
	}
}
