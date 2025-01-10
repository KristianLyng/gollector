/*
 * skogul, split transformer tests
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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

func TestSplit(t *testing.T) {
	var c skogul.Container
	testData := `
	{
		"metrics": [
		{
			"data": {
				"data": [
				{
					"splitField": "key1",
					"data": "yes"
				},
				{
					"splitField": "key2",
					"data": "yes also"
				}
				]
			}

		},
		{
			"data": {
				"data": "bad"
			}
		},
		{
			"data": {
				"data": [
				{
					"splitField": "key3",
					"data": "2yes"
				},
				{
					"splitField": "key4",
					"data": "2yes also"
				}
				]
			}
		}
		]
	}
	`
	if err := json.Unmarshal([]byte(testData), &c); err != nil {
		t.Error(err)
		return
	}

	split_path := "data"
	metadata := transformer.Split{
		Field:        []string{split_path},
		MetadataName: "arrayidx",
	}

	if err := metadata.Transform(&c); err != nil {
		t.Error(err)
		return
	}

	if len(c.Metrics) != 5 {
		t.Errorf(`Expected c.Metrics to be of len %d but got %d`, 5, len(c.Metrics))
		return
	}

	// Verify that the data is not the same in the two objects as it might differ
	if c.Metrics[0].Data["data"] != "yes" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "yes", c.Metrics[0].Data["data"])
	}
	if c.Metrics[0].Metadata["arrayidx"] != 0 {
		t.Errorf(`Expected Metrics Metadata key arrayidx to contain key arrayidx of val '%d' but got '%v'`, 0, c.Metrics[0].Metadata["arrayidx"])
	}

	if c.Metrics[1].Data["data"] != "yes also" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "yes also", c.Metrics[1].Data["data"])
	}
	if c.Metrics[1].Metadata["arrayidx"] != 1 {
		t.Errorf(`Expected Metrics Metadata key arrayidx to contain key arrayidx of val '%d' but got '%v'`, 1, c.Metrics[1].Metadata["arrayidx"])
	}
	if c.Metrics[2].Data["data"] != "bad" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "bad", c.Metrics[2].Data["data"])
	}
	if c.Metrics[3].Data["data"] != "2yes" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "2yes", c.Metrics[3].Data["data"])
	}
	if c.Metrics[3].Metadata["arrayidx"] != 0 {
		t.Errorf(`Expected Metrics Metadata key arrayidx to contain key arrayidx of val '%d' but got '%v'`, 0, c.Metrics[3].Metadata["arrayidx"])
	}
	if c.Metrics[4].Data["data"] != "2yes also" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "2yes also", c.Metrics[4].Data["data"])
	}
	if c.Metrics[4].Metadata["arrayidx"] != 1 {
		t.Errorf(`Expected Metrics Metadata key arrayidx to contain key arrayidx of val '%d' but got '%v'`, 1, c.Metrics[4].Metadata["arrayidx"])
	}
}

func TestSplit_dict(t *testing.T) {
	var c skogul.Container
	testData := `
	{
		"metrics": [
		{
			"data": {
				"dict": {
					"fookey": {
						"name": "foo",
						"key": "value1"
					},
					"barkey": {
						"name": "bar",
						"key": "value2"
					}
				}
			}

		}
		]
	}
	`
	if err := json.Unmarshal([]byte(testData), &c); err != nil {
		t.Error(err)
		return
	}

	split_path := "dict"
	metadata := transformer.DictSplit{
		Field:        []string{split_path},
		MetadataName: "keyname",
	}

	if err := metadata.Transform(&c); err != nil {
		t.Error(err)
		return
	}

	if len(c.Metrics) != 2 {
		t.Errorf(`Expected c.Metrics to be of len %d but got %d`, 2, len(c.Metrics))
		return
	}

	// Verify that the data is not the same in the two objects as it might differ
	// Since dictionaries/hashes are unsorted, there's no guarantee
	// that c.Metrics[0] is the first data listed in the data set
	// above, though it usually has been for now. Thus we try to detect
	// which test is which. This sort of makes us verify parts of the
	// test twice, though.
	test1 := 0
	test2 := 1
	if c.Metrics[0].Data["name"] == "foo" && c.Metrics[1].Data["name"] == "bar" {
		test1 = 0
		test2 = 1
	} else if c.Metrics[1].Data["name"] == "foo" && c.Metrics[0].Data["name"] == "bar" {
		test1 = 1
		test2 = 0
	} else {
		t.Errorf(`Expected Metrics not present for dict split?`)
	}

	if c.Metrics[test1].Data["name"] != "foo" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "foo", c.Metrics[test1].Data["name"])
	}
	if c.Metrics[test1].Metadata["keyname"] != "fookey" {
		t.Errorf(`Expected Metrics Metadata key 'keyname' to have value of 'fookey', but got '%s'`, c.Metrics[test1].Metadata["keyname"])
	}
	if c.Metrics[test2].Data["name"] != "bar" {
		t.Errorf(`Expected Metrics Data to contain key of val '%s' but got '%s'`, "bar", c.Metrics[test2].Data["name"])
	}
	if c.Metrics[test2].Metadata["keyname"] != "barkey" {
		t.Errorf(`Expected Metrics Metadata key 'keyname' to have value of 'barkey', but got '%s'`, c.Metrics[test2].Metadata["keyname"])
	}
}

func TestTopDict(t *testing.T) {
	var c skogul.Container
	testData := `
	{
		"metrics": [
		{
			"data": {
				"eth0": {
					"in": 13,
					"out": 15
				},
				"eth1": {
					"in": 124,
					"out": 111
				}
			}

		}
		]
	}
	`
	if err := json.Unmarshal([]byte(testData), &c); err != nil {
		t.Error(err)
		return
	}

	metadata := transformer.DictSplit{
		Field:        []string{},
		MetadataName: "name",
	}

	if err := metadata.Transform(&c); err != nil {
		t.Error(err)
		return
	}

	if len(c.Metrics) != 2 {
		t.Errorf(`Expected c.Metrics to be of len %d but got %d`, 5, len(c.Metrics))
		return
	}

	// Verify that the data is not the same in the two objects as it might differ
	// Since dictionaries/hashes are unsorted, there's no guarantee
	// that c.Metrics[0] is the first data listed in the data set
	// above, though it usually has been for now. Thus we try to detect
	// which test is which. This sort of makes us verify parts of the
	// test twice, though.
	test1 := 0
	test2 := 1
	if c.Metrics[0].Metadata["name"] == "eth0" && c.Metrics[1].Metadata["name"] == "eth1" {
		test1 = 0
		test2 = 1
	} else if c.Metrics[1].Metadata["name"] == "eth0" && c.Metrics[0].Metadata["name"] == "eth1" {
		test1 = 1
		test2 = 0
	} else {
		t.Errorf(`Expected Metrics not present for dict split?`)
	}
	// Verify that the data is not the same in the two objects as it might differ
	if c.Metrics[test1].Metadata["name"] != "eth0" {
		t.Errorf(`Expected Metrics[test1].name == eth0, got %v`, c.Metrics[test1].Metadata["name"])
	}
	if c.Metrics[test1].Data["in"] != 13.0 {
		t.Errorf(`Expected Metrics Data to contain key of val 13 but got '%#v'`, c.Metrics[test1].Data["in"])
	}
	if c.Metrics[test2].Metadata["name"] != "eth1" {
		t.Errorf(`Expected Metrics[test1].name == eth1, got %v`, c.Metrics[test2].Metadata["name"])
	}
	if c.Metrics[test2].Data["in"] != 124.0 {
		t.Errorf(`Expected Metrics Data to contain key of val 124 but got '%#v'`, c.Metrics[test2].Data["in"])
	}
}
