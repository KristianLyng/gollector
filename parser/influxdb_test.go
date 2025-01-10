/*
 * skogul, influxdb parser tests
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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
package parser_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/telenornms/skogul/parser"
)

func TestInfluxDBLineParse(t *testing.T) {
	b := []byte("system,host=testhost uptime=5464i 1585737340000000000")

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with 1 metric")
		return
	}

	if container.Metrics[0].Metadata["measurement"] != "system" {
		t.Error("Expected parsed data to contain measurement 'system'")
	}

	if container.Metrics[0].Metadata["host"] != "testhost" {
		t.Error("Expected parsed data to contain metadata field 'host'='testhost'")
	}

	uptime, castOk := container.Metrics[0].Data["uptime"].(int64)

	if !castOk {
		t.Errorf("Failed to cast value in 'uptime' data field to int64")
		return
	}

	if uptime != 5464 {
		t.Error("Expected parsed data to contain data field 'uptime'='5464'")
	}

	correctTime := time.Unix(0, 1585737340000000000)

	if err != nil {
		t.Errorf("Parsing correct time for verification failed: %s", err)
		return
	}

	if *container.Metrics[0].Time != correctTime {
		t.Errorf("Time parse failure: expected '%s' but got '%s'", correctTime, *&container.Metrics[0].Time)
	}
}

func TestInfluxDBLineParseWithoutTimestamp(t *testing.T) {
	b := []byte("system,host=testhost uptime=5464i")

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with 1 metric")
		return
	}

	if container.Metrics[0].Time == nil {
		t.Errorf("Expected container to add own timestamp")
	}

	isNowish := container.Metrics[0].Time.UnixNano() - time.Now().UnixNano()

	// Arbitrary value for difference between when timestamp was created in test and the
	// one that should have been added in the parser
	if isNowish > 100 {
		t.Errorf("Expected container time to be reasonably close to timestamp generated in test, expected <=100 but got '%d'", isNowish)
	}
}

func TestInfluxDBParseFile(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/influxdb.txt")

	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with at least 1 metric")
		return
	}
}

func TestInfluxDBLineParseQuotedString(t *testing.T) {
	b := []byte("system,host=testhost,foo=bar text=\"sometext\"")

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with 1 metric")
		return
	}

	if container.Metrics[0].Data["text"] != "sometext" {
		t.Errorf("Expected 'sometext' but got '%s'", container.Metrics[0].Data["text"])
	}
}

func TestInfluxDBLineParseQuotedStringWithSpace(t *testing.T) {
	b := []byte("system,host=testhost text=\"some text\"")

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with 1 metric")
		return
	}

	if container.Metrics[0].Data["text"] != "some text" {
		t.Errorf("Expected 'some text' but got '%s'", container.Metrics[0].Data["text"])
	}
}

func TestInfluxDBLineParseEscapedChars(t *testing.T) {
	b := []byte(`system,foo=bar,host=test\,host,host\,name=test\ host,hexed=my\x2dtag text=some\,text,other\,text=moretext,final=0`)

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with 1 metric")
		return
	}

	if container.Metrics[0].Metadata["host"] != "test,host" {
		t.Errorf("Expected 'test,host' but got '%s'", container.Metrics[0].Metadata["host"])
	}

	if container.Metrics[0].Metadata["host,name"] != "test host" {
		t.Errorf("Expected 'test,host' but got '%s'", container.Metrics[0].Metadata["host,name"])
	}
	if container.Metrics[0].Metadata["hexed"] != `my\x2dtag` {
		t.Errorf("Expected 'my\x2dtag' but got '%s'", container.Metrics[0].Metadata["hexed"])
	}

	if container.Metrics[0].Data["text"].(string) != "some,text" {
		t.Errorf("Expected 'some,text' but got '%s'", container.Metrics[0].Data["text"])
	}

	if container.Metrics[0].Data["other,text"] != "moretext" {
		t.Errorf("Expected 'moretext' but got '%s'", container.Metrics[0].Data["other,text"])
	}
}

func TestInfluxDBParseLineEscapedChars(t *testing.T) {
	tag1 := "host=my-hostname.example.org"
	tag2 := `cmd=/usr/usr\ bin/java\ -logpath\=/var/log`
	b := []byte(fmt.Sprintf("procstat,%s,%s,foo=bar cpu=1 1593610640000000000", tag1, tag2))

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Error(err)
		return
	}
	if container.Metrics[0].Metadata["cmd"] == nil {
		t.Errorf("Expected 'cmd' tag in series")
	} else {
		orig := strings.ReplaceAll((strings.SplitN(tag2, "=", 2)[1]), "\\", "")
		gen := container.Metrics[0].Metadata["cmd"].(string)
		if len(gen) != len(orig) {
			t.Errorf("Length of value for 'cmd' in container differs from test case (%d vs %d)", len(gen), len(orig))
		}
	}
	if container.Metrics[0].Metadata["host"] == nil {
		t.Errorf("Expected 'host' tag in series")
	} else {
		// Verifies that the length of the tag is the same as the expected, because this tag contains
		// escaped characters
		if len(container.Metrics[0].Metadata["host"].(string)) != len(strings.SplitN(tag1, "=", 2)[1]) {
			t.Errorf("Length of value for 'host' in container differs from test case")
		}
	}
	if container.Metrics[0].Metadata["foo"] == nil {
		t.Errorf("Expected 'foo' tag in series")
	}
}

func TestInfluxDBParseTelegrafCmdLine(t *testing.T) {
	b := []byte(`procstat,cmdline=/usr/bin/Java/bin/version/bin/java\ -Xms64m\ -Xmx2048m\ -javaagent:/some/path/to/a/.runtime/service/1.13u3/agent.jar\ -Djava.util.logging.config.file\=/var/log/service/you/get-the/gist-of-it/conf/logging.properties,host=host-name-prod.dc1.example.org,server_group=some-server-group cpu_time_irq=0 1593610640000000000`)

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Error(err)
		return
	}

	if container.Metrics[0].Metadata["cmdline"] == nil {
		t.Errorf("Expected 'cmdline' tag in series")
	}
	if container.Metrics[0].Metadata["host"] == nil {
		t.Errorf("Expected 'host' tag in series")
	}
	if container.Metrics[0].Metadata["server_group"] == nil {
		t.Errorf("Expected 'group' tag in series")
	}
}

func TestInfluxDBParseTelegrafCmdLines(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/influxdb_procstat.txt")

	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with at least 1 metric")
		return
	}

	for i, metric := range container.Metrics {
		if metric.Metadata["measurement"] == "procstat_lookup" {
			continue
		}
		if metric.Metadata["cmdline"] == nil {
			t.Errorf("Expected 'cmdline' tag in metric %d", i)
		}
		if metric.Metadata["host"] == nil {
			t.Errorf("Expected 'host' tag in metric %d", i)
		}
		if metric.Metadata["group"] == nil {
			t.Errorf("Expected 'group' tag in metric %d", i)
		}
	}
}

func BenchmarkInfluxDBLineParse(b *testing.B) {
	by := []byte(`disk,device=sda1,fstype=fat32,host=testhost,mode=rw,path=/private/var/vm free=98896670720i,used=1073762304i,used_percent=1.0740798769394355,inodes_total=4882452880i,total=499963174912i 1585737350000000000`)
	x := parser.InfluxDB{}
	for i := 0; i < b.N; i++ {
		x.Parse(by)
	}
}

func BenchmarkInfluxDBLineParseWithoutTimestamp(b *testing.B) {
	by := []byte(`disk,device=sda1,fstype=fat32,host=testhost,mode=rw,path=/private/var/vm free=98896670720i,used=1073762304i,used_percent=1.0740798769394355,inodes_total=4882452880i,total=499963174912i`)
	x := parser.InfluxDB{}
	for i := 0; i < b.N; i++ {
		x.Parse(by)
	}
}

func TestInfluxDBParseDoubleBackslashEscape(t *testing.T) {
	b := []byte(`systemd_units,host=foo,name=systemd-fsck@dev-mapper-vg00\\x2dlog.service,bar=foo active_code=0i 1619523738000000000`)

	_container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("expected to parse influx line protocol containing two backslashes: %v", err)
	}

	name := "systemd-fsck@dev-mapper-vg00\\x2dlog.service"
	if _container.Metrics[0].Metadata["name"] != name {
		t.Errorf("expected parsed tag to equal %s, got %s", name, _container.Metrics[0].Metadata["name"])
	}
}
func TestInfluxDBParseTelegrafSystemdUnitLines(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/influxdb_systemd_units.txt")

	if err != nil {
		t.Errorf("Failed to read test data file: %v", err)
		return
	}

	container, err := parser.InfluxDB{}.Parse(b)
	if err != nil {
		t.Errorf("Failed to parse data as influx line protocol: %v", err)
		return
	}

	if container == nil || container.Metrics == nil || len(container.Metrics) == 0 {
		t.Errorf("Expected parsed influx data to return a container with at least 1 metric")
		return
	}

	for i, metric := range container.Metrics {
		// fmt.Printf("line:%v\n", metric)
		if metric.Metadata["host"] == nil {
			t.Errorf("Expected 'host' tag in metric %d", i)
		}
	}
}
