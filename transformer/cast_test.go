/*
 * skogul, cast tests
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
	"math/big"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/transformer"
)

func TestCast(t *testing.T) {
	metric := skogul.Metric{}
	metric.Metadata = make(map[string]interface{})
	metric.Data = make(map[string]interface{})
	metric.Metadata["minttostring"] = 3
	metric.Metadata["mfloattostring"] = 3.14
	metric.Metadata["mstringtostring"] = "pi"
	metric.Metadata["minttofloat"] = 3
	metric.Metadata["mfloattofloat"] = 3.14
	metric.Metadata["mstringtofloat"] = "3.14"
	metric.Metadata["minttoint"] = 3
	metric.Metadata["mfloattoint"] = 3.14
	metric.Metadata["mstringtoint"] = "3.14"
	metric.Metadata["mflatten"] = 314159265358979.0
	metric.Metadata["mipv4"] = "127.0.0.1"
	metric.Metadata["mipv6"] = "::1"
	metric.Data["dinttostring"] = 3
	metric.Data["dfloattostring"] = 3.14
	metric.Data["dstringtostring"] = "pi"
	metric.Data["dinttofloat"] = 3
	metric.Data["dfloattofloat"] = 3.14
	metric.Data["dstringtofloat"] = "3.14"
	metric.Data["dinttoint"] = 3
	metric.Data["dfloattoint"] = 3.14
	metric.Data["dstringtoint"] = "3.14"
	metric.Data["dipv4"] = "127.0.0.1"
	metric.Data["dipv6"] = "::1"

	c := skogul.Container{}
	c.Metrics = []*skogul.Metric{&metric}

	cast := transformer.Cast{
		MetadataStrings:    []string{"minttostring", "mfloattostring", "mstringtostring"},
		MetadataFloats:     []string{"minttofloat", "mfloattofloat", "mstringtofloat"},
		MetadataInts:       []string{"minttoint", "mfloattoint", "mstringtoint"},
		MetadataFlatFloats: []string{"mflatten"},
		MetadataIpToDec:    []string{"mipv4", "mipv6"},
		DataStrings:        []string{"dinttostring", "dfloattostring", "dstringtostring"},
		DataFloats:         []string{"dinttofloat", "dfloattofloat", "dstringtofloat"},
		DataInts:           []string{"dinttoint", "dfloattoint", "dstringtoint"},
		DataIpToDec:        []string{"dipv4", "dipv6"},
	}

	err := cast.Transform(&c)
	if err != nil {
		t.Errorf("Cast() returned non-nil err: %v", err)
	}

	bigIntTestIpv4 := big.NewInt(2130706433)
	bigIntTestIpv6 := big.NewInt(1)

	check_m(t, c.Metrics[0], "minttostring", "3")
	check_m(t, c.Metrics[0], "mfloattostring", "3.14")
	check_m(t, c.Metrics[0], "mstringtostring", "pi")
	check_m(t, c.Metrics[0], "minttofloat", 3.0)
	check_m(t, c.Metrics[0], "mfloattofloat", 3.14)
	check_m(t, c.Metrics[0], "mstringtofloat", 3.14)
	check_m(t, c.Metrics[0], "minttoint", 3)
	check_m(t, c.Metrics[0], "mfloattoint", 3)
	check_m(t, c.Metrics[0], "mstringtoint", 3)
	check_m(t, c.Metrics[0], "mflatten", "314159265358979")

	// check_m(t, c.Metrics[0], "mipv4", bigIntTestIpv4)
	// check_m(t, c.Metrics[0], "mipv6", big.NewInt(1).Cmp(metric.Metadata["mipv6"].(*big.Int)))
	if bigIntTestIpv4.Cmp(metric.Metadata["mipv4"].(*big.Int)) != 0 {
		t.Error("ip to dec not equal")
	}
	if bigIntTestIpv6.Cmp(metric.Metadata["mipv6"].(*big.Int)) != 0 {
		t.Error("ip to dec not equal")
	}

	check_d(t, c.Metrics[0], "dinttostring", "3")
	check_d(t, c.Metrics[0], "dfloattostring", "3.14")
	check_d(t, c.Metrics[0], "dstringtostring", "pi")
	check_d(t, c.Metrics[0], "dinttofloat", 3.0)
	check_d(t, c.Metrics[0], "dfloattofloat", 3.14)
	check_d(t, c.Metrics[0], "dstringtofloat", 3.14)
	check_d(t, c.Metrics[0], "dinttoint", 3)
	check_d(t, c.Metrics[0], "dfloattoint", 3)
	check_d(t, c.Metrics[0], "dstringtoint", 3)

	// check_d(t, c.Metrics[0], "dipv4", big.NewInt(2130706433).Cmp(metric.Metadata["dipv4"].(*big.Int)))
	// check_d(t, c.Metrics[0], "dipv6", big.NewInt(1).Cmp(metric.Metadata["dipv6"].(*big.Int)))
	if bigIntTestIpv4.Cmp(metric.Data["dipv4"].(*big.Int)) != 0 {
		t.Error("ip to dec not equal")
	}
	if bigIntTestIpv6.Cmp(metric.Data["dipv6"].(*big.Int)) != 0 {
		t.Error("ip to dec not equal")
	}
}

func TestCast_config(t *testing.T) {
	testConfOk(t, `
	{
		"transformers": {
			"ok": {
				"type": "cast",
				"MetadataStrings": [ "foo", "bar" ]
			}
		}
	}`)
}
