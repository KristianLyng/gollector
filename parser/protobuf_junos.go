/*
 * skogul, protocol buffers parser
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

package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	//"github.com/golang/protobuf/proto"
	"github.com/gogo/protobuf/proto"

	"github.com/telenornms/skogul"
	pb "github.com/telenornms/skogul/gen/junos/telemetry"
)

var pbLog = skogul.Logger("parser", "protobuf")

// ProtoBuf parses a byte string-representation of a Container
type ProtoBuf struct {
	Debug bool `doc:"Logs the entire protobuf-packet if marshaling fails"`
	once  sync.Once
	stats *protobufStats
}

type protobufStats struct {
	Received                     uint64 // Received parse calls
	ParseErrors                  uint64 // Failure to parse the bytes using the protobuf definitions provided
	MissingExtension             uint64 // Missing Protobuf extension
	FailedToCastToJuniperMessage uint64 // We assumed it was a Juniper TelemetryStream message, but it failed to cast to it.
	FailedToJsonMarshal          uint64 // Failed to marshal protobuf data to json (this might fail if the data is not representable in JSON, such as the value '-Inf' as float64)
	FailedToJsonUnmarshal        uint64 // Failed to marshal JSON data back into skogul.Metric
	Parsed                       uint64 // Successful parses
}

func (x *ProtoBuf) initStats() {
	x.stats = &protobufStats{
		Received:                     0,
		ParseErrors:                  0,
		MissingExtension:             0,
		FailedToCastToJuniperMessage: 0,
		FailedToJsonMarshal:          0,
		FailedToJsonUnmarshal:        0,
		Parsed:                       0,
	}
}

// Parse accepts a byte slice of protobuf data and marshals it into a container
func (x *ProtoBuf) Parse(b []byte) (*skogul.Container, error) {
	x.once.Do(x.initStats)
	atomic.AddUint64(&x.stats.Received, 1)
	parsedProtoBuf, err := parseTelemetryStream(b)
	if err != nil {
		atomic.AddUint64(&x.stats.ParseErrors, 1)
		return nil, fmt.Errorf("initial parsing failed: %w", err)
	}

	protobufTimestamp := time.Unix(int64(*parsedProtoBuf.Timestamp/1000), int64(*parsedProtoBuf.Timestamp%1000)*1000000)

	metric := skogul.Metric{}

	metric.Time = &protobufTimestamp
	metric.Metadata, err = x.createMetadata(parsedProtoBuf)
	if err != nil {
		systemID := parsedProtoBuf.GetSystemId()
		sensorName := parsedProtoBuf.GetSensorName()
		return nil, fmt.Errorf("unable to extract metadata from protobuf packet. SystemID: %v SensorName: %v: %w", systemID, sensorName, err)
	}
	metric.Data, err = x.createData(parsedProtoBuf)
	if err != nil {
		systemID := parsedProtoBuf.GetSystemId()
		sensorName := parsedProtoBuf.GetSensorName()
		return nil, fmt.Errorf("unable to extract data from protobuf packet. SystemID: %v SensorName: %v: %w", systemID, sensorName, err)
	}

	container := skogul.Container{}
	container.Metrics = make([]*skogul.Metric, 1)
	container.Metrics[0] = &metric

	atomic.AddUint64(&x.stats.Parsed, 1)
	return &container, err
}

// parseTelemetryStream parses a protocol buffer with the Juniper TelemetryStream
// protobuf definitions
func parseTelemetryStream(protobuffer []byte) (*pb.TelemetryStream, error) {
	telemetrystream := &pb.TelemetryStream{}
	if err := proto.Unmarshal(protobuffer, telemetrystream); err != nil {
		// @ToDo: Consider what to do if failing to unmarshal the protobuf
		// Reasons: Invalid proto spec, invalid data, invalid version of proto spec (?)
		// not necessary to return here if we dont log or anything
		return telemetrystream, err
	}

	return telemetrystream, nil
}

// createMetadata extracts the fields containing metadata from the protocol buffer
// and stores them in a string-interface map to be consumed at a later stage.
func (x *ProtoBuf) createMetadata(telemetry *pb.TelemetryStream) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	metadata["systemId"] = telemetry.GetSystemId()
	metadata["sensorName"] = telemetry.GetSensorName()
	metadata["componentId"] = telemetry.GetComponentId()
	metadata["subComponentId"] = telemetry.GetSubComponentId()
	return metadata, nil
}

/*
applyJuniperHaxx adjusts incoming telemetry packets to make them JSON-compatible, sometimes "mangling" the data.

So far, it's mainly about fixing -Inf.
*/
func applyJuniperHaxx(messageOnly proto.Message) {
	var foo float32 = -40.0
	optics, ok := messageOnly.(*pb.Optics)
	if !ok {
		return
	}
	for _, odiags := range optics.OpticsDiag {
		if odiags.OpticsDiagStats == nil {
			continue
		}
		if odiags.OpticsDiagStats.LaserOutputPowerHighAlarmThresholdDbm != nil {
			if math.IsInf(*odiags.OpticsDiagStats.LaserOutputPowerHighAlarmThresholdDbm, 0) {
				odiags.OpticsDiagStats.LaserOutputPowerHighAlarmThresholdDbm = nil
			}
		}
		if odiags.OpticsDiagStats.LaserOutputPowerLowAlarmThresholdDbm != nil {
			if math.IsInf(*odiags.OpticsDiagStats.LaserOutputPowerLowAlarmThresholdDbm, 0) {
				odiags.OpticsDiagStats.LaserOutputPowerLowAlarmThresholdDbm = nil
			}
		}
		if odiags.OpticsDiagStats.LaserOutputPowerHighWarningThresholdDbm != nil {
			if math.IsInf(*odiags.OpticsDiagStats.LaserOutputPowerHighWarningThresholdDbm, 0) {
				odiags.OpticsDiagStats.LaserOutputPowerHighWarningThresholdDbm = nil
			}
		}
		if odiags.OpticsDiagStats.LaserOutputPowerLowWarningThresholdDbm != nil {
			if math.IsInf(*odiags.OpticsDiagStats.LaserOutputPowerLowWarningThresholdDbm, 0) {
				odiags.OpticsDiagStats.LaserOutputPowerLowWarningThresholdDbm = nil
			}
		}
		for _, lane := range odiags.OpticsDiagStats.OpticsLaneDiagStats {
			if lane.LaneLaserReceiverPowerDbm != nil {
				if skogul.IsInf(*lane.LaneLaserReceiverPowerDbm, -1) {
					lane.LaneLaserReceiverPowerDbm = &foo
				}
			}
			if lane.LaneLaserOutputPowerDbm != nil {
				if skogul.IsInf(*lane.LaneLaserOutputPowerDbm, -1) {
					lane.LaneLaserOutputPowerDbm = &foo
				}
			}
		}
	}
}

/*
createData creates a string-interface map of skogul.Metric type Data
by first marshalling the protobuf message into json and then parsing
it back in to a string-interface map.
*/
func (x *ProtoBuf) createData(telemetry *pb.TelemetryStream) (map[string]interface{}, error) {
	extension, err := proto.GetExtension(telemetry.GetEnterprise(), pb.E_JuniperNetworks)
	if err != nil {
		atomic.AddUint64(&x.stats.MissingExtension, 1)
		return nil, fmt.Errorf("failed to get Juniper protobuf extension: %w", err)
	}

	enterpriseExtension, ok := extension.(proto.Message)
	if !ok {
		atomic.AddUint64(&x.stats.FailedToCastToJuniperMessage, 1)
		return nil, fmt.Errorf("failed to cast to juniper message")
	}

	registeredExtensions := proto.RegisteredExtensions(enterpriseExtension)

	var regextensions []*proto.ExtensionDesc
	for _, ext := range registeredExtensions {
		regextensions = append(regextensions, ext)
	}

	availableExtensions, err := proto.GetExtensions(enterpriseExtension, regextensions)
	if err != nil {
		return nil, err
	}

	var jsonMessage []byte
	found := false
	for _, ext := range availableExtensions {
		if ext == nil {
			continue
		}

		if found {
			return nil, fmt.Errorf("multiple protobuf extensions found, don't know what to do!")
		}

		messageOnly, ok := ext.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("failed to cast to message: %v", ext)
		}
		applyJuniperHaxx(messageOnly)

		jsonMessage, err = json.Marshal(messageOnly)
		if err != nil {
			if x.Debug {
				pbLog.WithError(err).Infof("failed to marshal protobuf. data: %v", messageOnly)
			}
			atomic.AddUint64(&x.stats.FailedToJsonMarshal, 1)
			return nil, err
		}

		found = true
	}

	if !found {
		if x.Debug {
			pbLog.Infof("no valid extensions found. availableExtensions: %v, registered: %v, extensions: %v, telemetry: %v, regextensions: %v", availableExtensions, registeredExtensions, extension, telemetry, regextensions)
		}
		return nil, fmt.Errorf("found no valid extensions")
	}

	var metrics map[string]interface{}
	if err = json.Unmarshal(jsonMessage, &metrics); err != nil {
		atomic.AddUint64(&x.stats.FailedToJsonUnmarshal, 1)
		target := 500
		data := ""
		if len(jsonMessage) < 500 {
			target = len(jsonMessage) - 1
		}
		if len(jsonMessage) > 0 {
			data = string(jsonMessage[:target])
		} else {
			target = 0
			data = ""
		}

		return nil, fmt.Errorf("unmarshalling %d bytes of JSON data to string/interface map failed: %w. First %d bytes: %s", len(jsonMessage), err, target, data)
	}

	delete(metrics, "timestamp")
	delete(metrics, "sensorName")
	delete(metrics, "componentId")
	delete(metrics, "subComponentId")

	atomic.AddUint64(&x.stats.Parsed, 1)
	return metrics, nil
}

// GetStats prepares a skogul metric with stats
// for the protobuf parser.
func (x *ProtoBuf) GetStats() *skogul.Metric {
	now := skogul.Now()
	metric := skogul.Metric{
		Time:     &now,
		Metadata: make(map[string]interface{}),
		Data:     make(map[string]interface{}),
	}
	metric.Metadata["component"] = "parser"
	metric.Metadata["type"] = "protobuf"
	metric.Metadata["identity"] = skogul.Identity[x]

	// Ensure we init the stats struct in case we havent received a message yet.
	x.once.Do(x.initStats)

	metric.Data["received"] = x.stats.Received
	metric.Data["parse_errors"] = x.stats.ParseErrors
	metric.Data["missing_protobuf_extension"] = x.stats.MissingExtension
	metric.Data["failed_to_cast_to_juniper_message"] = x.stats.FailedToCastToJuniperMessage
	metric.Data["failed_to_json_marshal"] = x.stats.FailedToJsonMarshal
	metric.Data["failed_to_json_unmarshal"] = x.stats.FailedToJsonUnmarshal
	metric.Data["parsed"] = x.stats.Parsed
	return &metric
}
