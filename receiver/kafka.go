/*
 * skogul, Kafka consumer/receiver
 *
 * Copyright (c) 2022 Telenor Norge AS
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

package receiver

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	kafka "github.com/segmentio/kafka-go"
	kplain "github.com/segmentio/kafka-go/sasl/plain"
	"github.com/telenornms/skogul"
)

var kafkaLog = skogul.Logger("receiver", "kafka")

/*
Kafka receiver is a MVP-variant, and further features are reasonable and expected, including but not limited to:

- Authentication (coming before release)
- Dynamic keys from metadata
- Adjustment of various timeouts
*/
type Kafka struct {
	Topic    string            `doc:"Topic to read from."`
	Brokers  []string          `doc:"Array of brokeraddresses."`
	Handler  skogul.HandlerRef `doc:"Handler to use"`
	TLS      bool              `doc:"Enable TLS, off by default."`
	Username string            `doc:"Username for SASL auth."`
	Password string            `doc:"Password for SASL auth."`
	ClientID string            `doc:"ClientID to use - uses lower-case skogul by default."`
}

// Start the Kafka receiver and never return
func (k *Kafka) Start() error {
	if k.ClientID == "" {
		k.ClientID = "skogul"
	}
	dialer := kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		ClientID:  k.ClientID,
	}
	if k.TLS {
		dialer.TLS = &tls.Config{}
	}
	if (k.Username != "" && k.Password == "") || (k.Username == "" && k.Password != "") {
		return fmt.Errorf("provided just one of Username or Password for Kafka receiver, which makes no sense. Provide both or neither")
	}
	if k.Username != "" && k.Password != "" {
		if !k.TLS {
			kafkaLog.Warnf("Using authentication and no encryption... are you sure this makes sense?")
		}
		// XXX: WARN IF NOT TLS
		mechanism := kplain.Mechanism{
			Username: k.Username,
			Password: k.Password,
		}
		dialer.SASLMechanism = mechanism
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: k.Brokers,
		Topic:   k.Topic,
		Dialer:  &dialer,
	})
	r.SetOffset(kafka.LastOffset)
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			kafkaLog.WithError(err).Warnf("Unable to read message. Sleeping for 1s and retrying.")
			time.Sleep(time.Second)
			continue
		}
		if err := k.Handler.H.Handle(m.Value); err != nil {
			kafkaLog.WithError(err).Warn("Unable to handle Kafka message")
		}
	}
}
