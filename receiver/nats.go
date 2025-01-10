/*
 * skogul, nats receiver/subscriber
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

package receiver

import (
	"crypto/tls"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/telenornms/skogul"
)

var natsLog = skogul.Logger("receiver", "nats")

/*
Nats basic pub/sub receiver implementing all Authentication & Authorization
features in the nats golang client. Basic queue groups is also supported.
*/
type Nats struct {
	Handler       skogul.HandlerRef `doc:"Handler used to parse, transform and send data."`
	Servers       string            `doc:"Comma separated list of nats URLs"`
	Subject       string            `doc:"Subject to subscribe to messages on"`
	Queue         string            `doc:"Worker queue to distribute messages on"`
	Name          string            `doc:"Client name"`
	Username      string            `doc:"Client username"`
	Password      string            `doc:"Client password"`
	TLSClientKey  string            `doc:"TLS client key file path"`
	TLSClientCert string            `doc:"TLS client cert file path"`
	TLSCACert     string            `doc:"CA cert file path"`
	UserCreds     string            `doc:"Nats credentials file path"`
	NKeyFile      string            `doc:"Nats nkey file path"`
	Insecure      bool              `doc:"TLS InsecureSkipVerify"`
	conOpts       *[]nats.Option
	natsCon       *nats.Conn
	wg            sync.WaitGroup
}

// Verify configuration
func (n *Nats) Verify() error {
	if n.Handler.Name == "" {
		return skogul.MissingArgument("Handler")
	}
	if n.Subject == "" {
		return skogul.MissingArgument("Subject")
	}
	if n.Servers == "" {
		return skogul.MissingArgument("Servers")
	}
	// User Credentials
	if n.UserCreds != "" && n.NKeyFile != "" {
		//Cred file contains nkey.
		return fmt.Errorf("Please configure usercreds or nkeyfile.")
	}

	return nil
}

func (n *Nats) Start() error {
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
	// TLS authentication, Note: Fix selfsigned certificates.
	if n.TLSClientKey != "" && n.TLSClientCert != "" {
		cert, err := tls.LoadX509KeyPair(n.TLSClientCert, n.TLSClientKey)
		if err != nil {
			return fmt.Errorf("error parsing X509 certificate/key pair: %v", err)
		}

		cp, err := skogul.GetCertPool(n.TLSCACert)
		if err != nil {
			return fmt.Errorf("Failed to initialize root CA pool")
		}

		config := &tls.Config{
			InsecureSkipVerify: n.Insecure,
			Certificates:       []tls.Certificate{cert},
			RootCAs:            cp,
		}
		*n.conOpts = append(*n.conOpts, nats.Secure(config))
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
	// Always try to reconnect
	*n.conOpts = append(*n.conOpts, nats.RetryOnFailedConnect(true))
	// Try to reconnect forever
	*n.conOpts = append(*n.conOpts, nats.MaxReconnects(-1))

	var err error
	n.natsCon, err = nats.Connect(n.Servers, *n.conOpts...)
	if err != nil {
		natsLog.Errorf("Encountered an error while connecting to Nats: %v", err)
	}

	cb := func(msg *nats.Msg) {
		natsLog.Debugf("Received message on %v", msg.Subject)
		if err := n.Handler.H.Handle(msg.Data); err != nil {
			natsLog.WithError(err).Warn("Unable to handle Nats message")
		}
	}

	n.wg.Add(1)
	if n.Queue == "" {
		n.natsCon.QueueSubscribe(n.Subject, n.Queue, cb)
	} else {
		n.natsCon.Subscribe(n.Subject, cb)
	}
	n.wg.Wait()
	return n.natsCon.LastError()
}
