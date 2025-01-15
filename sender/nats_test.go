package sender_test

import (
	"testing"

	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/sender"
)

func TestNats(t *testing.T) {
	config, err := config.Bytes([]byte(`
	{
	  "receivers": {
	    "nats_r": {
	      "type": "nats",
	      "servers": "nats://0.0.0.0:4222",
	      "queue": "skogul_queue",
	      "subject": "test.subject",
	      "handler": "test_h"
	    }
	  },
	  "handlers": {
	    "test_h": {
	      "parser": "skogul",
	      "transformers": [],
	      "sender": "nats_s"
	    }
	  },
	  "senders": {
	    "nats_s": {
	      "type": "nats",
	      "servers": "nats://0.0.0.0:4222",
	      "subject": "nats.sender"
	    }
	  }
	}`))
	if err != nil {
		t.Errorf("Failed to load config: %s", err)
	}

	ns := config.Senders["nats_s"].Sender.(*sender.Nats)

	if ns == nil {
		t.Error("Bad config")
	}
}
