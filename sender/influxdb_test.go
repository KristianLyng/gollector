/*
 * skogul, influxdb tests
 *
 * Copyright (c) 2019 Telenor Norge AS
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

package sender_test

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/telenornms/skogul/config"
)

func TestInfluxDB(t *testing.T) {
	_, err := config.Bytes([]byte(`
{
   "senders": {
     "ok1": {
       "type": "influxdb",
       "url": "http://localhost/write?db=foo",
       "measurement": "bar"
     },
     "ok2": {
       "type": "influxdb",
       "url": "http://localhost/write?db=foo",
       "measurementfrommetadata": "bar"
     },
     "ok1": {
       "type": "influxdb",
       "url": "http://localhost/write?db=foo",
       "measurement": "bar",
       "measurementfrommetadata": "baz"
     }
   }
}`))
	if err != nil {
		t.Errorf("Failed to load config for influxdb test: %v", err)
		return
	}
}

func TestInfluxDB_fail(t *testing.T) {
	log.SetLevel(log.TraceLevel)
	if _, err := config.Bytes([]byte(`
{
   "senders": {
     "bad": {
       "type": "influxdb",
       "url": "http://localhost/write?db=foo"
     }
   }
}`)); err == nil {
		t.Errorf("Loaded config despite it being invalid.")
		return
	}
	if _, err := config.Bytes([]byte(`
{
   "senders": {
     "bad": {
       "type": "influxdb"
     }
   }
}`)); err == nil {
		t.Errorf("Loaded config despite it being invalid.")
		return
	}
	if _, err := config.Bytes([]byte(`
{
   "senders": {
     "bad": {
       "type": "influxdb",
       "url": "http://localhost/write?db=foo",
       "measurement": { "foo": "bar" }
     }
   }
}`)); err == nil {
		t.Errorf("Loaded config despite it being invalid.")
		return
	}
}
