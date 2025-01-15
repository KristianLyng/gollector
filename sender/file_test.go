/*
 * skogul, file writer tests
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
 * Author(s):
 *  - Håkon Solbjørg <Hakon.Solbjorg@telenor.com>
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
	"os"
	"path"
	"syscall"
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/sender"
)

// createContainer is a simple helper func which
// creates a skogul.Container with some data
func createContainer() *skogul.Container {
	meta := make(map[string]interface{})
	meta["foo"] = "bar"
	data := make(map[string]interface{})
	data["baz"] = "qux"

	metric := skogul.Metric{
		Metadata: meta,
		Data:     data,
	}
	metrics := make([]*skogul.Metric, 0)
	metrics = append(metrics, &metric)

	return &skogul.Container{
		Metrics: metrics,
	}
}

func TestWriteToFIle(t *testing.T) {
	filename := "skogul-file-sender-existing-file-append.txt"
	path := path.Join(os.TempDir(), filename)

	sender := &sender.File{
		File:   path,
		Append: false,
	}

	c := createContainer()
	sender.Send(c)

	time.Sleep(time.Second)

	b, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	if len(b) > 0 {
		t.Log("Ok, sender has written test data")
	}

	testPid := os.Getpid()
	process, err := os.FindProcess(testPid)
	if err != nil {
		t.Errorf("Could not find requested PID %v", testPid)
		return
	}

	process.Signal(syscall.SIGHUP)

	time.Sleep(time.Second)

	b, err = os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	if len(b) == 0 && !sender.Append {
		t.Log("Ok, sender has received SIGHUP and truncated file.")
		return
	}
}

func TestAppendToFile(t *testing.T) {
	filename := "skogul-file-sender-existing-file-append.txt"
	path := path.Join(os.TempDir(), filename)

	sender := &sender.File{
		File:   path,
		Append: true,
	}

	c := createContainer()
	sender.Send(c)

	time.Sleep(time.Second)

	b, err := os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	if len(b) > 0 {
		t.Log("Ok, sender has written test data")
	}

	testPid := os.Getpid()
	process, err := os.FindProcess(testPid)
	if err != nil {
		t.Errorf("Could not find requested PID %v", testPid)
		return
	}

	process.Signal(syscall.SIGHUP)

	time.Sleep(time.Second)

	b, err = os.ReadFile(path)
	if err != nil {
		t.Error(err)
		return
	}

	if len(b) > 0 {
		t.Log("Ok, sender has received SIGHUP and appended file.")
		return
	}
}
