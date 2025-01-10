package sender_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/parser"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
)

func TestSnmpSenderTest(t *testing.T) {
	t.Skip()
	c, _ := loadJsonFile(t, "")

	sconf := fmt.Sprintln(`
	{
		"receivers": {
			"http": {
				"type": "http",
				"address": ":11111",
				"handlers": {
					"/": "h"
				}
			}
		},
		"transformers": {},
		"handlers": {
			"h": {
				"parser": "skogul",
				"transformers": [
					"now"
				],
				"sender": "print"
			}
		},
		"senders": {
			"snmp": {
				"type": "snmp",
				"port": 1337,
				"community": "public",
				"version": "2c",
				"target": "localhost",
				"oidmap": {}
			}
		}
	}
	`)

	conf, err := config.Bytes([]byte(sconf))
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	snmpSender := conf.Senders["snmp"].Sender.(*sender.SNMP)
	httpRcv := conf.Receivers["http"].Receiver.(*receiver.HTTP)

	for _, hr := range httpRcv.Handlers {
		hr.H.Transform(c)
	}

	if httpRcv == nil {
		t.Errorf("failed to get receiver")
		return
	}

	go httpRcv.Start()
	time.Sleep(time.Duration(5 * time.Second))

	err = snmpSender.Send(c)
	if err != nil {
		log.Println(err)
	}
}

func loadJsonFile(t *testing.T, file string) (*skogul.Container, []byte) {
	b, _ := ioutil.ReadFile(file)

	container, _ := parser.SkogulJSON{}.Parse(b)

	return container, b
}
