/*
 * skogul, complex sender tests
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

package sender_test

import (
	"fmt"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/parser"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
	"math/rand"
	"testing"
	"time"
)

var validContainer = skogul.Container{}

func init() {

	now := time.Now()
	m := skogul.Metric{}
	m.Time = &now
	m.Metadata = make(map[string]interface{})
	m.Data = make(map[string]interface{})
	m.Metadata["foo"] = "bar"
	m.Data["tall"] = 5
	validContainer.Metrics = []*skogul.Metric{&m}
}

// Tests http receiver, sender and JSON parser implicitly
func TestHttp_stack(t *testing.T) {
	one := &(sender.Test{})

	port := 1337 + rand.Intn(100)
	adr := fmt.Sprintf("localhost:%d", port)
	h := skogul.Handler{Sender: one}
	h.SetParser(parser.SkogulJSON{})
	rcv := receiver.HTTP{Address: adr, Handlers: map[string]*skogul.HandlerRef{"/": {H: &h}}}
	//	rcv.Handle("/", &h)
	go rcv.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	hs := sender.HTTP{URL: fmt.Sprintf("http://%s", adr)}
	one.TestQuick(t, &hs, &validContainer, 1)
	blank := skogul.Container{}
	one.TestNegative(t, &hs, &blank)

	hs2 := sender.HTTP{URL: "http://localhost:1/foobar"}

	err := hs2.Send(&validContainer)
	if err == nil {
		t.Errorf("hs2.Send() to invalid url did not fail.")
	}
	err = hs2.Send(&skogul.Container{})
	if err == nil {
		t.Errorf("hs2.Send() with invalid container did not fail.")
	}
}

func TestHttp_rootCa(t *testing.T) {
	_, err := config.Bytes([]byte(`
{
   "senders": {
     "ok1": {
       "type": "http",
       "url": "https://localhost/write?db=foo",
       "rootca": "testdata/cacert-snakeoil.pem"
     }
   }
}`))
	if err != nil {
		t.Errorf("Failed to load config for http test: %v", err)
		return
	}
}

func TestHttp_rootCa_bad1(t *testing.T) {
	_, err := config.Bytes([]byte(`
{
   "senders": {
     "bad1": {
       "type": "http",
       "url": "https://localhost/write?db=foo",
       "rootca": "/dev/null"
     }
   }
}`))
	if err == nil {
		t.Errorf("Successfully read invalid rootca /dev/null !")
		return
	}
}
func TestHttp_rootCa_bad2(t *testing.T) {
	_, err := config.Bytes([]byte(`
{
   "senders": {
     "bad2": {
       "type": "http",
       "url": "https://localhost/write?db=foo",
       "rootca": "/dev/no/proc/such/sys/file/run/lol/kek/why/are/you/still/reading/this"
     }
   }
}`))
	if err == nil {
		t.Errorf("Successfully read invalid rootca from non-existent file!")
		return
	}
}
