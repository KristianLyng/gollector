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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/KristianLyng/skogul"
	pb "github.com/KristianLyng/skogul/gen"
)

// ProtoBuf parses a byte string-representation of a Container
type ProtoBuf struct{}

// Parse accepts a byte slice of protobuf data and marshals it into a container
func (x ProtoBuf) Parse(b []byte) (*skogul.Container, error) {

	parsedProtoBuf, err := parseTelemetryStream(b)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse protocol buffer (err: %s)", err)
	}

	protobufTimestamp := time.Unix(int64(*parsedProtoBuf.Timestamp/1000), int64(*parsedProtoBuf.Timestamp%1000)*1000000)

	metric := skogul.Metric{
		Time:     &protobufTimestamp,
		Metadata: createMetadata(parsedProtoBuf),
		Data:     createData(parsedProtoBuf),
	}

	if metric.Metadata == nil || metric.Data == nil {
		return nil, errors.New("Metric metadata or data was nil; aborting")
	}

	container := skogul.Container{}
	container.Metrics = make([]*skogul.Metric, 1)
	container.Metrics[0] = &metric

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
func createMetadata(telemetry *pb.TelemetryStream) map[string]interface{} {
	var metadata = make(map[string]interface{})

	metadata["systemId"] = telemetry.GetSystemId()
	metadata["sensorName"] = telemetry.GetSensorName()
	metadata["componentId"] = telemetry.GetComponentId()
	metadata["subComponentId"] = telemetry.GetSubComponentId()
	return metadata
}

/* createData creates a string-interface map of skogul.Metric type Data
by first marshalling the raw protobuf data into json and then parsing
it back in to a string-interface map.
@ToDo: Make this cheaper
*/
func createData(telemetry *pb.TelemetryStream) map[string]interface{} {
	pbjsonmarshaler := jsonpb.Marshaler{}
	var out bytes.Buffer
	extension, err := proto.GetExtension(telemetry.GetEnterprise(), pb.E_JuniperNetworks)
	if err != nil {
		log.Printf("Failed to get Juniper protobuf extension, is this really a Juniper protobuf message?")
		return nil
	}

	enterpriseExtension, ok := extension.(proto.Message)
	if !ok {
		log.Printf("Failed to cast to juniper message")
		return nil
	}

	registeredExtensions := proto.RegisteredExtensions(enterpriseExtension)

	var regextensions []*proto.ExtensionDesc
	for _, ext := range registeredExtensions {
		regextensions = append(regextensions, ext)
	}

	availableExtensions, err := proto.GetExtensions(enterpriseExtension, regextensions)

	found := false
	for _, ext := range availableExtensions {
		if ext == nil {
			continue
		}

		if found {
			log.Printf("Multiple extensions found, don't know what to do!")
			return nil
		}

		messageOnly, ok := ext.(proto.Message)
		if !ok {
			log.Printf("Failed to cast to message: %v", ext)
			return nil
		}
		if err := pbjsonmarshaler.Marshal(&out, messageOnly); err != nil {
			log.Printf("Marshalling protocol buffer data to JSON failed: %s", err)
			return nil
		}

		found = true
	}

	var metrics map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &metrics); err != nil {
		log.Printf("Unmarshalling JSON data to string/interface map failed: %s", err)
		return nil
	}

	delete(metrics, "timestamp")
	delete(metrics, "sensorName")
	delete(metrics, "componentId")
	delete(metrics, "subComponentId")
	return metrics
}
