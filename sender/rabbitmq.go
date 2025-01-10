/*
 * skogul, rabbitmq producer/sender
 *
 * Copyright (c) 2023 Telenor Norge AS
 * Author(s):
 *  - Kamil Oracz <kamil.oracz@telenor.no>
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
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

type Rabbitmq struct {
	Username skogul.Secret     `doc:"Username for rabbitmq instance"`
	Password skogul.Secret     `doc:"Password for rabbitmq instance"`
	Host     string            `doc:"Hostname for rabbitmq instance. Fallback is localhost"`
	Port     string            `doc:"Port for rabbitmq instance. Fallback is 5672"`
	Queue    string            `doc:"Queue to write to"`
	Encoder  skogul.EncoderRef `doc:"Encoder to use. Fallback is json"`
	Timeout  int               `doc:"Timeout for rabbitmq instance connection. Fallback is 10 seconds."`
	channel  *amqp.Channel
	once     sync.Once
}

var rabbitmqLog = skogul.Logger("sender", "rabbitmq")

func (r *Rabbitmq) init() {
	if r.Port == "" {
		r.Port = "5672"
	}

	if r.Host == "" {
		r.Host = "localhost"
	}

	if r.Timeout == 0 {
		r.Timeout = 10
	}

	if r.Encoder.E == nil {
		r.Encoder.E = encoder.JSON{}
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", r.Username.Expose(), r.Password.Expose(), r.Host, r.Port))
	if err != nil {
		rabbitmqLog.WithError(err).Error("Failed initializing broker connection")
		return
	}

	ch, err := conn.Channel()
	if err != nil {
		rabbitmqLog.WithError(err).Error("Failed initializing channel")
		return
	}

	r.channel = ch

	_, err = ch.QueueDeclare(
		r.Queue,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		rabbitmqLog.WithError(err).Error("Failed to declare a queue")
		return
	}
}

func (r *Rabbitmq) Send(c *skogul.Container) error {
	r.once.Do(func() {
		r.init()
	})

	if r.channel == nil {
		return fmt.Errorf("No active rabbitmq connections")
	}

	body, err := r.Encoder.E.Encode(c)
	if err != nil {
		r.channel.Close()
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.Timeout)*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		"",
		r.Queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		},
	)
	if err != nil {
		r.channel.Close()
		return err
	}

	return nil
}

func (r *Rabbitmq) Verify() error {
	if r.Username.Expose() == "" {
		return skogul.MissingArgument("Username")
	}

	if r.Password.Expose() == "" {
		return skogul.MissingArgument("Password")
	}

	if r.Queue == "" {
		return skogul.MissingArgument("Queue")
	}

	return nil
}
