/*
 * skogul, Prometheus parser test
 *
 * Copyright (c) 2022 Telenor Norge AS
 * Author(s):
 *  - Roshini Narasimha Raghavan <roshiragavi@gmail.com>
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
	"os"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/parser"
)

func TestPrometheus(t *testing.T) {
	b, err := os.ReadFile("./testdata/prometheus_testdata")
	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}
	// convert byte array to json
	p := parser.Prometheus{}

	container, err := p.Parse(b)
	if err != nil {
		t.Logf("Error while parsing the data. %v", err)
		t.FailNow()
	}

	// expectedTime := 1608520832877
	metricKey1 := "dialer_name"
	metricValue1 := "federate"
	metricKey2 := "instance"
	metricValue2 := "localhost:9090"
	metricKey3 := "job"
	metricValue3 := "prometheus"
	var data1Value, data2Value, data3Value float64
	data1Value = 1
	data1Key := "net_conntrack_dialer_conn_attempted_total"
	data2Key := "conntrack_dialer_conn_attempted_total"
	data2Value = 1
	data3Key := "dialer_conn_attempted_total"
	data3Value = 1

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Logf("Expected parsed prometheus to return a container with at least 1 metric. Container: %v", container.Describe())
		t.FailNow()
	}
	if len(container.Metrics) != 3 {
		t.Errorf("Expected exactly 3 metrics, got %d", len(container.Metrics))
	}

	var m1 *skogul.Metric
	var m2 *skogul.Metric
	var m3 *skogul.Metric
	for _, m := range container.Metrics {
		if m.Data["net_conntrack_dialer_conn_attempted_total"] != nil {
			m1 = m
		}
		if m.Data["conntrack_dialer_conn_attempted_total"] != nil {
			m2 = m
		}
		if m.Data["dialer_conn_attempted_total"] != nil {
			m3 = m
		}
	}

	if m1 == nil || m2 == nil || m3 == nil {
		t.Errorf("missing metrics?")
	}
	if m1.Metadata[metricKey1] != metricValue1 {
		t.Logf("Expected parsed prometheus to return a metadata field value")
		t.FailNow()
	}

	if m1.Metadata[metricKey2] != metricValue2 {
		t.Logf("Expected parsed prometheus to return a metadata field value")
		t.FailNow()
	}

	if m1.Metadata[metricKey3] != metricValue3 {
		t.Logf("Expected parsed prometheus to return a metadata field value")
		t.FailNow()
	}
	if m1.Data[data1Key] != data1Value {
		t.Logf("Expected parsed prometheus to return a data field %s value, got %v. ", data1Key, m1.Data[data1Key])
		t.Logf("container: %s", container.Describe())
		t.FailNow()
	}
	if m2.Data[data2Key] != data2Value {
		t.Logf("Expected parsed prometheus to return a data field value %v", m2.Data[data2Key])
		t.FailNow()
	}
	if m3.Data[data3Key] != data3Value {
		t.Logf("Expected parsed prometheus to return a data field value %v", m3.Data[data3Key])
		t.FailNow()
	}
	if len(m1.Metadata) != 3 || len(m2.Metadata) != 0 || len(m3.Metadata) != 3 {
		t.Logf("container: %s", container.Describe())
		t.Logf("Length of the container Metrics Metadata fields are not correct %v, %v, %v", len(m1.Metadata), len(m2.Metadata), len(m3.Metadata))
		t.FailNow()
	}
	if len(container.Metrics[0].Data) != 1 || len(container.Metrics[1].Data) != 1 || len(container.Metrics[2].Data) != 1 {
		t.Logf("container: %s", container.Describe())
		t.Logf("Length of the container Metrics data are not correct %v, %v, %v", len(container.Metrics[0].Data), len(container.Metrics[1].Data), len(container.Metrics[2].Data))
		t.FailNow()
	}
}
