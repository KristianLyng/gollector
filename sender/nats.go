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
	"github.com/nats-io/nats.go"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
	"strings"
	"sync"
)

var natsLog = skogul.Logger("sender", "nats")

/*
Nats sender. A small Nats, non-jetstream, publisher with:

- Authentication: Username/Password, TLS
- Authorization: Username/Password, UserCredentials/JWT

*/

type Nats struct {
	Servers       string `doc:"Comma separated list of nats URLs"`
	Subject       string `doc:"Subject to publish messages on"`
	SubjectAppend string `doc:"Append this Metadata field to subject"`
	Name          string `doc:"Client name"`
	Username      string `doc:"Client username"`
	Password      string `doc:"Client password"`
	TLSClientKey  string `doc:"TLS client key file path"`
	TLSClientCert string `doc:"TLS client cert file path"`
	TLSCACert     string `doc:"CA cert file path"`
	UserCreds     string `doc:"Nats credentials file path"`
	NKeyFile      string `doc:"Nats nkey file path"`
	Insecure      bool   `doc:"TLS InsecureSkipVerify"`
	Encoder       skogul.EncoderRef
	o             *[]nats.Option
	nc            *nats.Conn
	once          sync.Once
	init_error    error
}

// Verify configuration
func (n *Nats) Verify() error {
	if n.Subject == "" {
		return skogul.MissingArgument("Subject")
	}
	if n.Servers == "" {
		return skogul.MissingArgument("Servers")
	}

	//User Credentials, use either.
	if n.UserCreds != "" && n.NKeyFile != "" {
		return fmt.Errorf("Please configure usercreds or nkeyfile.")
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
	n.o = &[]nats.Option{nats.Name(n.Name)}

	if n.UserCreds != "" {
		*n.o = append(*n.o, nats.UserCredentials(n.UserCreds))
	}

	//Plain text passwords
	if n.Username != "" && n.Password != "" {
		if n.TLSClientKey != "" {
			natsLog.Warnf("Using plain text password over a non encrypted transport!")
		}
		*n.o = append(*n.o, nats.UserInfo(n.Username, n.Password))
	}

	//TLS authentication, Note: Fix selfsigned certificates.
	if n.TLSClientKey != "" && n.TLSClientCert != "" {
		cert, err := tls.LoadX509KeyPair(n.TLSClientCert, n.TLSClientKey)
		if err != nil {
			n.init_error = fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}

		cp, err := skogul.GetCertPool(n.TLSCACert)
		if err != nil {
			n.init_error = fmt.Errorf("Failed to initialize root CA pool")
		}

		if n.init_error == nil {
			config := &tls.Config{
				InsecureSkipVerify: n.Insecure,
				Certificates:       []tls.Certificate{cert},
				RootCAs:            cp,
			}
			*n.o = append(*n.o, nats.Secure(config))
		}
	}

	//NKey auth
	if n.NKeyFile != "" {
		opt, err := nats.NkeyOptionFromSeed(n.NKeyFile)
		if err != nil {
			natsLog.Fatal(err)
		}
		*n.o = append(*n.o, opt)
	}

	//Log disconnects
	*n.o = append(*n.o, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		natsLog.WithError(err).Error("Got disconnected!")
	}))
	//Log reconnects
	*n.o = append(*n.o, nats.ReconnectHandler(func(nc *nats.Conn) {
		natsLog.Info("Reconnected")
	}))
	//Always try reconnecting
	*n.o = append(*n.o, nats.RetryOnFailedConnect(true))
	//Keep doing reconnects
	*n.o = append(*n.o, nats.MaxReconnects(-1))

	var err error
	n.nc, err = nats.Connect(n.Servers, *n.o...)
	if err != nil {
		n.init_error = fmt.Errorf("Encountered an error while connecting to Nats: %v", err)
	}
	return err
}

func (n *Nats) Send(c *skogul.Container) error {
	n.once.Do(func() {
		n.init()
	})
	if n.init_error != nil {
		return n.init_error
	}
	for _, m := range c.Metrics {
		subject := n.Subject
		a_subject := m.Metadata[n.SubjectAppend]
		//Append metadata field to subject.
		if n.SubjectAppend != "" && a_subject.(string) != "" {
			if !strings.HasSuffix(m.Metadata[n.SubjectAppend].(string), ".") {
				subject = subject + "."
			}
			subject = subject + m.Metadata[n.SubjectAppend].(string)
		}

		b, err := n.Encoder.E.EncodeMetric(m)
		if err != nil {
			natsLog.Warnf("couldn't encode metric: %w", err)
			continue
		}
		n.nc.Publish(subject, b)
	}

	return n.nc.LastError()
}
