/*
 * skogul, test protobuf parser
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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

package parser_test

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	// proto "github.com/golang/protobuf/proto"
	"github.com/gogo/protobuf/proto"
	"github.com/telenornms/skogul"
	junos_protobuf_telemetry "github.com/telenornms/skogul/gen/junos/telemetry"
	"github.com/telenornms/skogul/parser"
)

/*
bit patterns for 32-bit infinity. See ../common.go for details.
*/

const (
	uvneginf = 0b11111111100000000000000000000000
)

type failer interface {
	Fatalf(format string, args ...interface{})
	Helper()
}

func readProtobufFile(t failer, file string) []byte {
	t.Helper()
	b := make([]byte, 9000)
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("unable to open protobuf packet file: %v", err)
	}
	defer f.Close()
	n, err := f.Read(b)
	if err != nil {
		t.Fatalf("unable to read protobuf packet file: %v", err)
	}
	if n == 0 {
		t.Fatalf("read 0 bytes from protobuf packet file....")
	}
	return b[0:n]
}

func TestProtoBuf(t *testing.T) {
	b := readProtobufFile(t, "testdata/protobuf-packet.bin")
	x := parser.ProtoBuf{}
	c, err := x.Parse(b)
	if err != nil {
		t.Errorf("ProtoBuf.Parse(b) failed: %s", err)
	}
	if c == nil {
		t.Errorf("ProtoBuf.Parse(b) returned nil-container")
	}
}

func BenchmarkProtoBufParse(b *testing.B) {
	by := readProtobufFile(b, "testdata/protobuf-packet.bin")
	x := parser.ProtoBuf{}
	for i := 0; i < b.N; i++ {
		x.Parse(by)
	}
}

func generateJunosTelemetryStream(sensorName string, eps junos_protobuf_telemetry.EnterpriseSensors) junos_protobuf_telemetry.TelemetryStream {
	systemId := "localhost"
	now := uint64(time.Now().Unix())
	componentId := uint32(1)
	subComponentId := uint32(2)

	return junos_protobuf_telemetry.TelemetryStream{
		SystemId:       &systemId,
		Timestamp:      &now,
		ComponentId:    &componentId,
		SubComponentId: &subComponentId,
		SensorName:     &sensorName,
		Enterprise:     (*junos_protobuf_telemetry.EnterpriseSensors)(&eps),
		// Should this be used ?  Ietf:       (*junos_protobuf_telemetry.IETFSensors)(&juniperNetworksSensors),
	}
}

func generateOpticsDiag(val float32) junos_protobuf_telemetry.TelemetryStream {
	eps := junos_protobuf_telemetry.EnterpriseSensors{}
	juniperNetworksSensors := junos_protobuf_telemetry.JuniperNetworksSensors{}
	if err := proto.SetExtension(&eps, junos_protobuf_telemetry.E_JuniperNetworks, &juniperNetworksSensors); err != nil {
		fmt.Printf("Failed to set juniperNetworks extension: %v\n", err)
	}

	ifName := "ge-1/0/1"
	optics := junos_protobuf_telemetry.Optics{
		OpticsDiag: []*junos_protobuf_telemetry.OpticsInfos{
			{
				IfName: &ifName,
				OpticsDiagStats: &junos_protobuf_telemetry.OpticsDiagStats{
					OpticsLaneDiagStats: []*junos_protobuf_telemetry.OpticsDiagLaneStats{
						{
							LaneLaserReceiverPowerDbm: &val,
						},
					},
				},
			},
		},
	}
	if err := proto.SetExtension(&juniperNetworksSensors, junos_protobuf_telemetry.E_JnprOpticsExt, &optics); err != nil {
		fmt.Printf("Failed to set Optics extension: %v\n", err)
	}
	return generateJunosTelemetryStream("foo", eps)
}

func parseDiagStatsResp(data map[string]interface{}, key string) interface{} {
	opticsDiag, ok := data["Optics_diag"].([]interface{})
	if !ok {
		fmt.Printf("failed to cast")
	}
	foo, ok := opticsDiag[0].(map[string]interface{})
	if !ok {
		fmt.Printf("failed to cast 2")
	}
	opticsDiagStats := foo["optics_diag_stats"].(map[string]interface{})

	opticsLaneDiagStats := opticsDiagStats["optics_lane_diag_stats"].([]interface{})

	bar, ok := opticsLaneDiagStats[0].(map[string]interface{})
	if !ok {
		fmt.Printf("failed to cast 3")
	}

	return bar[key]
}

func TestParseJunosProtobufTelemetryStreamOptics(t *testing.T) {
	expected := float64(-40)
	telemetry := generateOpticsDiag(-40)

	bytes, err := proto.Marshal(&telemetry)
	if err != nil {
		t.Errorf("Failed to marshal protobuf message to bytes: %v", err)
		return
	}
	if bytes == nil {
		t.Error("Bytes marshalling resulted in nil")
		return
	}

	protobufParser := parser.ProtoBuf{}
	c, err := protobufParser.Parse(bytes)
	if err != nil {
		t.Errorf("Failed to parse optics diag lane stats protobuf data, err: %v", err)
	}
	if c == nil {
		t.Error("Protobuf parse returned nil-container")
	}

	got := parseDiagStatsResp(c.Metrics[0].Data, "lane_laser_receiver_power_dbm")
	if got != expected {
		t.Errorf("Expected lane_laser_receiver_power_dbm to be %T(%v), but got %T(%v)", expected, expected, got, got)
	}
}

func TestParseJunosProtobufTelemetryStreamOpticsNegativeInf(t *testing.T) {
	val := float32(math.Float32frombits(uvneginf))
	telemetry := generateOpticsDiag(val)

	bytes, err := proto.Marshal(&telemetry)
	if !skogul.IsInf(val, -1) {
		t.Errorf("Trying to test negative infinity value, but value isn't interpreted as infinity. Value: %f", val)
		return
	}
	if err != nil {
		t.Errorf("Failed to marshal protobuf message to bytes: %v", err)
		return
	}
	if bytes == nil {
		t.Error("Bytes marshalling resulted in nil")
		return
	}

	protobufParser := parser.ProtoBuf{}
	_, err = protobufParser.Parse(bytes)
	if err != nil {
		t.Errorf("Expected parsing -Inf values to NOT return an error, ref issue #194 which should now be ... resolved.")
		return
	}
}
