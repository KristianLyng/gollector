/*
 * skogul, templating tests
 *
 * Copyright (c) 2019-2021 Telenor Norge AS
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

package transformer_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/transformer"
)

func check_m(t *testing.T, m *skogul.Metric, field string, want interface{}) {
	t.Helper()
	if m.Metadata[field] != want {
		t.Errorf("Transformer failed to enforce rule for metadata field \"%s\". Wanted \"%#v\"(%T), got \"%#v\"(%T)", field, want, want, m.Metadata[field], m.Metadata[field])
	}
}

func check_d(t *testing.T, m *skogul.Metric, field string, want interface{}) {
	t.Helper()
	if m.Data[field] != want {
		t.Errorf("Transformer failed to enforce rule for data field \"%s\". Wanted \"%#v\", got \"%#v\"", field, want, m.Data[field])
	}
}

func TestMetadata(t *testing.T) {
	// now := time.Now()

	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Metadata["set"] = "original"
	metric.Metadata["require"] = "present"
	metric.Metadata["remove"] = "not removed"
	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	metadata := transformer.Metadata{
		Set:     map[string]interface{}{"set": "new"},
		Require: []string{"require"},
		Remove:  []string{"remove"},
		Ban:     []string{"ban"},
	}

	err := metadata.Transform(&c)
	if err != nil {
		t.Errorf("Metadata() returned non-nil err: %v", err)
	}

	check_m(t, c.Metrics[0], "set", "new")
	check_m(t, c.Metrics[0], "require", "present")
	check_m(t, c.Metrics[0], "remove", nil)
	check_m(t, c.Metrics[0], "ban", nil)
}

func TestMetadata_config(t *testing.T) {
	testConfOk(t, `
	{
		"transformers": {
			"ok": {
				"type": "metadata",
				"set": {
					"this": "to that",
					"foo": "is bar"
				},
				"require": [ "reqthis" ],
				"remove": [ "gruff" ],
				"ban": [ "trash" ]
			}
		}
	}`)
	testConfOk(t, `
	{
		"transformers": {
			"ok": {
				"type": "metadata",
				"set": { },
				"require": [],
				"remove": [ ],
				"ban": [  ]
			}
		}
	}`)
	testConfBad(t, `
	{
		"transformers": {
			"ok": {
				"type": "metadata",
				"set": 5
				"require": [ "reqthis" ],
				"remove": [ "gruff" ],
				"ban": [ "trash" ]
			}
		}
	}`)
}

func testConfOk(t *testing.T, rawconf string) *config.Config {
	t.Helper()
	conf, err := config.Bytes([]byte(rawconf))
	if err != nil {
		t.Errorf("failed to parse valid transformer config: %v", err)
	}
	if conf == nil {
		t.Errorf("failed to get valid config for transformer")
	}
	return conf
}

func testConfBad(t *testing.T, rawconf string) {
	t.Helper()
	conf, err := config.Bytes([]byte(rawconf))
	if err == nil {
		t.Errorf("Didn't catch invalid config")
	}
	if conf != nil {
		t.Errorf("got config from invalid source config")
	}
}

func TestExtract(t *testing.T) {
	extracted_value_key := "extract-this"
	extracted_value := "the value"

	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": "%s"}`, extracted_value_key, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	metadata := transformer.Metadata{
		ExtractFromData: []string{extracted_value_key, "empty_key"},
	}

	err := metadata.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	if c.Metrics[0].Metadata[extracted_value_key] != extracted_value {
		t.Errorf(`Expected %s but got %s`, extracted_value, c.Metrics[0].Metadata[extracted_value_key])
	}
	if _, ok := c.Metrics[0].Data["empty_key"]; ok {
		t.Errorf(`Data key 'empty_key' is set after extraction`)
	}
	if _, ok := c.Metrics[0].Metadata["empty_key"]; ok {
		t.Errorf(`Metadata key 'empty_key' is set after extraction`)
	}
	if _, ok := c.Metrics[0].Data[extracted_value_key]; ok {
		t.Errorf(`Data key %s is still set after extraction`, extracted_value_key)
	}
}

func TestCopy(t *testing.T) {
	metric := skogul.Metric{}
	testData := `
	{
		"metadata": {
			"something": "value",
			"old_key_name": "mv_123"
		},
		"data": {
			"leavethis": "123",
			"deleteme": "456",
			"newkey": "789"
		}
	}`
	json.Unmarshal([]byte(testData), &metric)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conf := testConfOk(t, `
	{ "transformers": {
		"ok": {
			"type": "metadata",
			"copyfromdata": [
				{ "source": "leavethis", "destination": "meta1", "keep": true },
				{ "source": "deleteme", "destination": "meta2" },
				{ "source": "newkey" }
			],
			"rename": [
				{ "source": "old_key_name", "destination": "new_key_name" }
			]

		}
	}
	}`)

	metadata := conf.Transformers["ok"].Transformer

	err := metadata.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	if c.Metrics[0].Metadata["meta1"] != "123" {
		t.Errorf(`Expected 123 but got %s`, c.Metrics[0].Metadata["meta1"])
	}
	if c.Metrics[0].Metadata["meta2"] != "456" {
		t.Errorf(`Expected 456 but got %s`, c.Metrics[0].Metadata["meta2"])
	}
	if c.Metrics[0].Metadata["newkey"] != "789" {
		t.Errorf(`Expected 789 but got %s`, c.Metrics[0].Metadata["newkey"])
	}
	if c.Metrics[0].Metadata["something"] != "value" {
		t.Errorf(`Expected value but got %s`, c.Metrics[0].Metadata["something"])
	}
	if c.Metrics[0].Data["leavethis"] != "123" {
		t.Errorf(`Expected 123 but got %s`, c.Metrics[0].Data["leavethis"])
	}
	if c.Metrics[0].Metadata["old_key_name"] != nil {
		t.Errorf(`Expcted old_key_name to be removed, but still present: %s`, c.Metrics[0].Metadata["old_key_name"])
	}
	if c.Metrics[0].Metadata["new_key_name"] != "mv_123" {
		t.Errorf(`Expected new_key_name to be mv_123, but it is: %s`, c.Metrics[0].Metadata["new_key_name"])
	}
	if _, ok := c.Metrics[0].Data["deleteme"]; ok {
		t.Errorf(`Data key 'deleteme' is set after extraction`)
	}
}

func TestRenameData(t *testing.T) {
	metric := skogul.Metric{}
	testData := `
	{
		"metadata": {
			"something": "value"
		},
		"data": {
			"leavethis": "123",
			"old_key_name": "mv_123"
		}
	}`
	json.Unmarshal([]byte(testData), &metric)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	conf := testConfOk(t, `
	{ "transformers": {
		"ok": {
			"type": "data",
			"rename": [
				{ "source": "old_key_name", "destination": "new_key_name" }
			]

		}
	}
	}`)

	metadata := conf.Transformers["ok"].Transformer

	err := metadata.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	if c.Metrics[0].Data["old_key_name"] != nil {
		t.Errorf(`Expcted old_key_name to be removed, but still present: %s`, c.Metrics[0].Data["old_key_name"])
	}
	if c.Metrics[0].Data["new_key_name"] != "mv_123" {
		t.Errorf(`Expected new_key_name to be mv_123, but it is: %s`, c.Metrics[0].Data["new_key_name"])
	}
}

func TestFlattenMap(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "key"
	extracted_value := "value"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": {"%s": "%s"}, "otherData": "dataer"}`, path, extracted_value_key, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten: [][]string{{path}},
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s__%s", path, extracted_value_key)

	// Expect data to be accessible at its new location
	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}

	// Expect data be removed at its original location
	if c.Metrics[0].Data[path] != nil {
		t.Errorf(`Expected nil-value but got value`)
	}

	// Expect data unrelated to the flattening to still be accessible
	if c.Metrics[0].Data["otherData"] != "dataer" {
		t.Errorf(`Expected "%s" but got "%s"`, "dataer", c.Metrics[0].Data["otherData"])
	}
}

func TestFlattenMapDefaultSeparator(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "key"
	extracted_value := "value"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": {"%s": "%s"}, "otherData": "dataer"}`, path, extracted_value_key, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten:          [][]string{{path}},
		FlattenSeparator: "",
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s__%s", path, extracted_value_key)

	// Expect data to be accessible at its new location
	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}

	// Expect data to be removed at its original location
	if c.Metrics[0].Data[path] != nil {
		t.Errorf(`Expected nil-value but got value`)
	}

	// Expect data unrelated to the flattening to still be accessible
	if c.Metrics[0].Data["otherData"] != "dataer" {
		t.Errorf(`Expected "%s" but got "%s"`, "dataer", c.Metrics[0].Data["otherData"])
	}
}

func TestFlattenMapCustomSeparator(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "key"
	extracted_value := "value"
	separator := "!SEP!"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": {"%s": "%s"}, "otherData": "dataer"}`, path, extracted_value_key, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten:          [][]string{{path}},
		FlattenSeparator: separator,
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s%s%s", path, separator, extracted_value_key)

	// Expect data to be accessible at its new location
	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}

	// Expect data to removed at its original location
	if c.Metrics[0].Data[path] != nil {
		t.Errorf(`Expected nil-value but got value`)
	}

	// Expect data unrelated to the flattening to still be accessible
	if c.Metrics[0].Data["otherData"] != "dataer" {
		t.Errorf(`Expected "%s" but got "%s"`, "dataer", c.Metrics[0].Data["otherData"])
	}
}

func TestFlattenMapDropSeparator(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "key"
	extracted_value := "value"
	separator := "drop"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": {"%s": "%s"}, "otherData": "dataer"}`, path, extracted_value_key, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten:          [][]string{{path}},
		FlattenSeparator: separator,
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s", extracted_value_key)

	// Expect data to be accessible at its new location
	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}

	// Expect data to still be accessible at its original location
	if c.Metrics[0].Data[path] == nil {
		t.Errorf(`Expected "%s" but got "%s" in %+v`, extracted_value, c.Metrics[0].Data[path], c.Metrics[0].Data)
	}

	// Expect data unrelated to the flattening to still be accessible
	if c.Metrics[0].Data["otherData"] != "dataer" {
		t.Errorf(`Expected "%s" but got "%s"`, "dataer", c.Metrics[0].Data["otherData"])
	}
}

func TestFlattenArray(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "0"
	extracted_value := "value"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": ["%s"]}`, path, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten:      [][]string{{path}},
		KeepOriginal: false,
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s__%s", path, extracted_value_key)

	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}
}

func TestFlattenArrayOfMaps(t *testing.T) {
	path := "nestedData"
	extracted_value_key := "0"
	extracted_value_key_2 := "key"
	extracted_value := "value"

	metric := skogul.Metric{}

	metric.Data = make(map[string]interface{})
	testData := fmt.Sprintf(`{"%s": [{"%s": "%s"}, {"a": "b"}]}`, path, extracted_value_key_2, extracted_value)
	json.Unmarshal([]byte(testData), &metric.Data)

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	data := transformer.Data{
		Flatten:      [][]string{{path}},
		KeepOriginal: false,
	}

	err := data.Transform(&c)
	if err != nil {
		t.Error(err)
	}

	new_path := fmt.Sprintf("%s__%s__%s", path, extracted_value_key, extracted_value_key_2)

	if c.Metrics[0].Data[new_path] != extracted_value {
		t.Errorf(`Expected "%s" but got "%s"`, extracted_value, c.Metrics[0].Data[new_path])
	}
}
