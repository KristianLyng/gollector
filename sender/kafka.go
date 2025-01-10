/*
 * skogul, kafka producer/sender
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

package sender

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"

	kafka "github.com/segmentio/kafka-go"
	kplain "github.com/segmentio/kafka-go/sasl/plain"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

var kafkaLog = skogul.Logger("sender", "kafka")

/*
Kafka sender is a MVP-variant, and further features are reasonable and expected, including but not limited to:

- Authentication (coming before release)
- Better control of batching, probably
- Dynamic keys from metadata
- Adjustment of various timeouts
*/
type Kafka struct {
	Topic    string `doc:"Topic to write to."`
	Sync     bool   `doc:"Synchronous or not. By default, the sender is async."`
	Address  string `doc:"Address for the broker."`
	ClientID string `doc:"ClientID to use - uses lower-case skogul by default."`
	TLS      bool   `doc:"Enable TLS, off by default."`
	Username string `doc:"Username for SASL auth."`
	Password string `doc:"Password for SASL auth."`
	Encoder  skogul.EncoderRef
	w        *kafka.Writer
	once     sync.Once
}

func (k *Kafka) init() {
	/*
	 * We need to control batching, or rather, not do it, because
	 * Skogul does that itself and this just ends up being one weird
	 * mess. With actual batching in kafka-go, the WriteMessages will
	 * always block for the BatchTimeout period, which sort of makes
	 * sense, but I can't see how we can maintain control of go
	 * routines this way, so it's better to let skoguls regular
	 * batch sender do the batching.
	 *
	 * Sync still has a function: Without sync, errors are difficult to
	 * spot. So the effect here is that sync: true with batchsize: 1
	 * means we get errors, but we never block longer than needed.
	 */
	k.w = &kafka.Writer{
		Addr:      kafka.TCP(k.Address),
		Topic:     k.Topic,
		Async:     !k.Sync,
		BatchSize: 1,
	}
	transport := kafka.Transport{}
	if k.ClientID == "" {
		k.ClientID = "skogul"
	}
	transport.ClientID = k.ClientID
	if k.TLS {
		transport.TLS = &tls.Config{}
	}
	if (k.Username != "" && k.Password == "") || (k.Username == "" && k.Password != "") {
		kafkaLog.Warnf("Provided just one of Username or Password for Kafka receiver, which makes no sense. Provide both or neither.")
	}
	if k.Username != "" && k.Password != "" {
		if !k.TLS {
			kafkaLog.Warnf("Using authentication and no encryption... are you sure this makes sense?")
		}
		kafkaLog.Infof("Using authentication")
		mechanism := kplain.Mechanism{
			Username: k.Username,
			Password: k.Password,
		}
		transport.SASL = mechanism
	}
	hasPw := "<not configured>"
	if k.Password != "" {
		hasPw = "<configured, redacted>"
	}
	kafkaLog.Infof("Kafka settings: TLS: %v, Topic: %s, ClientID: %s, Username: %s, Password: %s", k.TLS, k.Topic, k.ClientID, k.Username, hasPw)
	k.w.Transport = &transport
	if k.Encoder.Name == "" {
		k.Encoder.E = encoder.JSON{}
	}
}

func (k *Kafka) Send(c *skogul.Container) error {
	k.once.Do(func() {
		k.init()
	})
	messages := make([]kafka.Message, 0, len(c.Metrics))
	for _, m := range c.Metrics {
		b, err := k.Encoder.E.EncodeMetric(m)
		if err != nil {
			return fmt.Errorf("couldn't encode metric: %w", err)
		}
		km := kafka.Message{
			Value: b,
		}
		messages = append(messages, km)
	}
	err := k.w.WriteMessages(context.Background(), messages...)
	return err
}
