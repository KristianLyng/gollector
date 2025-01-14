/*
 * skogul, linefile tests
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

package receiver_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
)

func deleteFile(t *testing.T, file string) {
	err := os.Remove(file)
	if err != nil {
		t.Errorf("Failed to remove old test file %s: %v", file, err)
	}
}

func lfMakeFile(t *testing.T) (string, error) {
	t.Helper()
	file := fmt.Sprintf("%s/skogul-linefiletest-%d-%d", os.TempDir(), os.Getpid(), rand.Int())

	_, err := os.Stat(file)

	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Error statting tmp file %s: %v", file, err)
		return "", err
	}

	if !os.IsNotExist(err) {
		t.Errorf("File possibly exists already: %s", file)
		return "", err
	}

	err = syscall.Mkfifo(file, 0600)

	if err != nil {
		t.Errorf("Unable to make fifo %s: %v", file, err)
		return "", err
	}
	t.Logf("Tmp file: %s", file)
	return file, nil
}

func TestLineFile(t *testing.T) {
	file, err := lfMakeFile(t)
	if err != nil {
		return
	}
	defer deleteFile(t, file)

	// Note the delay. Since we do not stop the receiver when we're
	// done, but we DO remove the file, it will throw errors
	// continuously, trashing the log and other tests (such as the Log
	// receiver)
	sconf := fmt.Sprintf(`
{
  "receivers": {
    "linefile": {
      "type": "fifo",
      "file": "%s",
      "handler": "h",
      "delay": "1h"
    }
  },
  "handlers": {
    "h": {
      "parser": "json",
      "transformers": [],
      "sender": "test"
    }
  },
  "senders": {
    "test": {
      "type": "test"
    }
  }
}`, file)

	conf, err := config.Bytes([]byte(sconf))

	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	sTest := conf.Senders["test"].Sender.(*sender.Test)
	rcv := conf.Receivers["linefile"].Receiver.(*receiver.LineFile)

	if rcv == nil {
		t.Errorf("failed to get receiver")
		return
	}

	go rcv.Start()
	b, err := json.Marshal(validContainer)
	if err != nil {
		t.Errorf("Failed to marshal container: %v", err)
		return
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		t.Errorf("Unable to open file/fifo for writing: %v", err)
		return
	}
	defer func() {
		f.Close()
	}()
	f.WriteString(fmt.Sprintf("%s\n", b))
	time.Sleep(time.Duration(10 * time.Millisecond))
	if sTest.Received() != 1 {
		t.Errorf("Didn't receive thing on other end!")
	}
	f.WriteString(fmt.Sprintf("%s\n", b))
	time.Sleep(time.Duration(10 * time.Millisecond))
	if sTest.Received() != 2 {
		t.Errorf("Didn't receive thing on other end!")
	}
	sTest.Set(0)
	f.WriteString(fmt.Sprintf("bad idea♥\n"))
	if sTest.Received() != 0 {
		t.Errorf("Receive thing on other end despite bogus data")
	}
}

func TestLineFileAdvanced(t *testing.T) {
	file, err := lfMakeFile(t)
	if err != nil {
		return
	}

	sconf := fmt.Sprintf(`
		{
			"receivers": {
					"x": {
						"type": "fileadvanced",
            "file": "%s",
            "handler": "kek",
            "delay": "1s",
            "newfile": "%s_copy.json",
            "shell": "/bin/bash",
            "post": "mv %s_copy.json %s_copy-archive-\"$(date)\".json"
					}
			},
			"handlers": {
					"kek": {
							"parser": "skogulmetric",
							"transformers": [
									"now"
							],
							"sender": "test"
					}
			},
			"senders": {
				"test": {
					"type": "test"
				}
			}
	}`, file, file, file, file)

	conf, err := config.Bytes([]byte(sconf))

	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	sTest := conf.Senders["test"].Sender.(*sender.Test)
	rcv := conf.Receivers["x"].Receiver.(*receiver.LineFileAdvanced)

	if rcv == nil {
		t.Errorf("failed to get receiver")
		return
	}

	skogulMetric := map[string]interface{}{
		"data": map[string]interface{}{
			"Test": "data",
			"foo":  "bar",
		},
	}

	go rcv.Start()
	b, err := json.Marshal(skogulMetric)
	if err != nil {
		t.Errorf("Failed to marshal container: %v", err)
		return
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		t.Errorf("Unable to open file/fifo for writing: %v", err)
		return
	}
	defer func() {
		f.Close()
	}()

	f.WriteString(fmt.Sprintf("%s\n", b))
	time.Sleep(time.Duration(10 * time.Millisecond))

	f.WriteString(fmt.Sprintf("%s\n", b))
	time.Sleep(time.Duration(10 * time.Millisecond))

	if sTest.Received() != 2 {
		t.Errorf("Didn't receive thing on other end!")
	}
}
