/*
 * skogul, nats producer/sender
 *
 * Author(s):
 *  - Niklas Holmstedt <n.holmstedt@gmail.com>
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
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

var natsLog = skogul.Logger("sender", "nats")

/*
Nats sender. A small Nats, non-jetstream, publisher implementing all
Authentication & Authorization features in the nats golang client.
*/

type Nats struct {
	Servers       string   `doc:"Comma separated list of nats URLs"`
	Subject       string   `doc:"Subject to publish messages on"`
	SubjectAppend []string `doc:"Append theese Metadata fields to subject"`
	Name          string   `doc:"Client name"`
	Username      string   `doc:"Client username"`
	Password      string   `doc:"Client password"`
	TLSClientKey  string   `doc:"TLS client key file path"`
	TLSClientCert string   `doc:"TLS client cert file path"`
	TLSCACert     string   `doc:"CA cert file path"`
	UserCreds     string   `doc:"Nats credentials file path"`
	NKeyFile      string   `doc:"Nats nkey file path"`
	Insecure      bool     `doc:"TLS InsecureSkipVerify"`
	Encoder       skogul.EncoderRef
	conOpts       *[]nats.Option
	natsCon       *nats.Conn
	once          sync.Once
	initError     error
}

// Verify configuration
func (n *Nats) Verify() error {
	if n.Subject == "" {
		return skogul.MissingArgument("Subject")
	}
	if n.Servers == "" {
		return skogul.MissingArgument("Servers")
	}

	// User Credentials, use either.
	if n.UserCreds != "" && n.NKeyFile != "" {
		return fmt.Errorf("please configure usercreds or nkeyfile")
	}

	return nil
}

func (n *Nats) init() error {
	if n.Encoder.Name == "" {
		n.Encoder.E = encoder.JSON{}
	}

	if n.Name == "" {
		n.Name = "skogul"
	}
	n.conOpts = &[]nats.Option{nats.Name(n.Name)}

	if n.UserCreds != "" {
		*n.conOpts = append(*n.conOpts, nats.UserCredentials(n.UserCreds))
	}

	// Plain text passwords
	if n.Username != "" && n.Password != "" {
		if n.TLSClientKey != "" {
			natsLog.Warnf("Using plain text password over a non encrypted transport!")
		}
		*n.conOpts = append(*n.conOpts, nats.UserInfo(n.Username, n.Password))
	}

	// TLS authentication
	if n.TLSClientKey != "" && n.TLSClientCert != "" {
		cert, err := tls.LoadX509KeyPair(n.TLSClientCert, n.TLSClientKey)
		if err != nil {
			n.initError = fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}

		cp, err := skogul.GetCertPool(n.TLSCACert)
		if err != nil {
			n.initError = fmt.Errorf("failed to initialize root CA pool")
		}

		if n.initError == nil {
			config := &tls.Config{
				InsecureSkipVerify: n.Insecure,
				Certificates:       []tls.Certificate{cert},
				RootCAs:            cp,
			}
			*n.conOpts = append(*n.conOpts, nats.Secure(config))
		}
	}

	// NKey auth
	if n.NKeyFile != "" {
		opt, err := nats.NkeyOptionFromSeed(n.NKeyFile)
		if err != nil {
			natsLog.Fatal(err)
		}
		*n.conOpts = append(*n.conOpts, opt)
	}

	// Log disconnects
	*n.conOpts = append(*n.conOpts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		natsLog.WithError(err).Error("Got disconnected!")
	}))
	// Log reconnects
	*n.conOpts = append(*n.conOpts, nats.ReconnectHandler(func(nc *nats.Conn) {
		natsLog.Info("Reconnected")
	}))
	// Always try reconnecting
	*n.conOpts = append(*n.conOpts, nats.RetryOnFailedConnect(true))
	// Keep doing reconnects
	*n.conOpts = append(*n.conOpts, nats.MaxReconnects(-1))

	var err error
	n.natsCon, err = nats.Connect(n.Servers, *n.conOpts...)
	if err != nil {
		n.initError = fmt.Errorf("encountered an error while connecting to Nats: %v", err)
	}
	return err
}

func (n *Nats) Send(c *skogul.Container) error {
	n.once.Do(func() {
		n.init()
	})
	if n.initError != nil {
		return n.initError
	}
	for i, m := range c.Metrics {
		subject := n.Subject
		// Append metadata fields to subject.
		for _, value := range n.SubjectAppend {
			if appSubject, ok := m.Metadata[value]; ok {
				if appSubject.(string) == "" {
					continue
				}
				if !strings.HasPrefix(appSubject.(string), ".") {
					appSubject = "." + appSubject.(string)
				}
				subject = subject + appSubject.(string)
			}
		}

		b, err := n.Encoder.E.EncodeMetric(m)
		if err != nil {
			natsLog.WithError(err).Warnf("Couldn't send metric[%v], encountered and error while encoding.", i)
			natsLog.WithError(err).Debugf("Metadata of incorrect metric: %v", m.Metadata)
			continue
		}
		n.natsCon.Publish(subject, b)
	}

	return n.natsCon.LastError()
}
