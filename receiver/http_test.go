/*
 * skogul, complex receiver tests
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *  - Håkon Solbjørg <hakon.solbjorg@telenor.com>
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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
)

// Tests http receiver, sender and JSON parser implicitly
func TestHttp_stack(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("TestHttp encountered an assert errror")
		}
	}()
	config, err := config.Bytes([]byte(`
{
	"senders": {
		"plain_origin": {
			"type": "http",
			"url": "http://localhost:1339"
		},
		"auth_plain_origin": {
			"type": "http",
			"url": "http://god:hunter2@localhost:3000"
		},
		"auth_plain_fail1": {
			"type": "http",
			"url": "http://gad:hunter2@localhost:3000"
		},
		"auth_plain_fail2": {
			"type": "http",
			"url": "http://gad:@localhost:3000"
		},
		"auth_plain_fail3": {
			"type": "http",
			"url": "http://:hunter2@localhost:3000"
		},
		"auth_plain_fail4": {
			"type": "http",
			"url": "http://localhost:3000"
		},
		"ssl_ok": {
			"type": "http",
			"url": "https://localhost:3443",
			"insecure": true
		},
		"ssl_bad1": {
			"type": "http",
			"url": "https://localhost:3443",
			"insecure": false
		},
		"ssl_bad2": {
			"type": "http",
			"url": "http://localhost:3443"
		},
		"ssl_auth_ok": {
			"type": "http",
			"url": "https://god:hunter2@localhost:5443",
			"insecure": true
		},
		"ssl_auth_bad1": {
			"type": "http",
			"url": "https://gad:hunter2@localhost:5443",
			"insecure": true
		},
		"ssl_auth_bad2": {
			"type": "http",
			"url": "https://god:hunter3@localhost:5443",
			"insecure": true
		},
		"ssl_auth_bad3": {
			"type": "http",
			"url": "https://god:@localhost:5443",
			"insecure": true
		},
		"ssl_auth_bad4": {
			"type": "http",
			"url": "https://:hunter2@localhost:5443",
			"insecure": true
		},
		"ssl_auth_bad5": {
			"type": "http",
			"url": "http://god:hunter2@localhost:5443"
		},
		"ssl_auth_bad6": {
			"type": "http",
			"url": "https://god:hunter2@localhost:5443",
			"insecure": false
		},
		"common": {
			"type": "test"
		}
	},
	"receivers": {
		"plain": {
			"type": "http",
			"address": "localhost:1339",
			"handlers": { "/": "common"}
		},
		"auth": {
			"type": "http",
			"address": "localhost:3000",
			"handlers": { "/": "common"},
			"auth": {
				"/": {
					"username": "god",
					"password": "hunter2"
				}
			}
		},
		"ssl_noauth": {
			"type": "http",
			"address": "localhost:3443",
			"handlers": { "/": "common"},
			"certfile": "../docs/examples/basics/cacert-snakeoil.pem",
			"keyfile": "../docs/examples/basics/privkey-snakeoil.pem"
		},
		"ssl_auth": {
			"type": "http",
			"address": "localhost:5443",
			"handlers": { "/": "common"},
			"certfile": "../docs/examples/basics/cacert-snakeoil.pem",
			"keyfile": "../docs/examples/basics/privkey-snakeoil.pem",
			"auth": {
				"/": {
					"username": "god",
					"password": "hunter2"
				}
			}
		}
	},
	"handlers": {
		"common": {
			"parser": "json",
			"transformers": [],
			"sender": "common"
		}
	}
}`))
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	UpdateHttpConfigWithUsablePorts(config)

	sPlainOrigin := config.Senders["plain_origin"].Sender.(*sender.HTTP)
	sAuthPlainOrigin := config.Senders["auth_plain_origin"].Sender.(*sender.HTTP)
	sAuthPlainFail1 := config.Senders["auth_plain_fail1"].Sender.(*sender.HTTP)
	sAuthPlainFail2 := config.Senders["auth_plain_fail2"].Sender.(*sender.HTTP)
	sAuthPlainFail3 := config.Senders["auth_plain_fail3"].Sender.(*sender.HTTP)
	sAuthPlainFail4 := config.Senders["auth_plain_fail4"].Sender.(*sender.HTTP)
	sSSLOk1 := config.Senders["ssl_ok"].Sender.(*sender.HTTP)
	sSSLBad1 := config.Senders["ssl_bad1"].Sender.(*sender.HTTP)
	sSSLBad2 := config.Senders["ssl_bad2"].Sender.(*sender.HTTP)
	sSSLAuthOk1 := config.Senders["ssl_auth_ok"].Sender.(*sender.HTTP)
	sSSLAuthBad1 := config.Senders["ssl_auth_bad1"].Sender.(*sender.HTTP)
	sSSLAuthBad2 := config.Senders["ssl_auth_bad2"].Sender.(*sender.HTTP)
	sSSLAuthBad3 := config.Senders["ssl_auth_bad3"].Sender.(*sender.HTTP)
	sSSLAuthBad4 := config.Senders["ssl_auth_bad4"].Sender.(*sender.HTTP)
	sSSLAuthBad5 := config.Senders["ssl_auth_bad5"].Sender.(*sender.HTTP)
	sSSLAuthBad6 := config.Senders["ssl_auth_bad6"].Sender.(*sender.HTTP)
	sCommon := config.Senders["common"].Sender.(*sender.Test)

	rPlain := config.Receivers["plain"].Receiver.(*receiver.HTTP)
	rAuth := config.Receivers["auth"].Receiver.(*receiver.HTTP)
	rSSLNoAuth := config.Receivers["ssl_noauth"].Receiver.(*receiver.HTTP)
	rSSLAuth := config.Receivers["ssl_auth"].Receiver.(*receiver.HTTP)

	go rPlain.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	sCommon.TestQuick(t, sPlainOrigin, &validContainer, 1)
	blank := skogul.Container{}
	sCommon.TestNegative(t, sPlainOrigin, &blank)

	go rAuth.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	sCommon.TestQuick(t, sAuthPlainOrigin, &validContainer, 1)
	sCommon.TestNegative(t, sAuthPlainOrigin, &blank)
	sCommon.TestNegative(t, sAuthPlainFail1, &validContainer)
	sCommon.TestNegative(t, sAuthPlainFail2, &validContainer)
	sCommon.TestNegative(t, sAuthPlainFail3, &validContainer)
	sCommon.TestNegative(t, sAuthPlainFail4, &validContainer)
	sCommon.TestQuick(t, sAuthPlainOrigin, &validContainer, 1)

	go rSSLNoAuth.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	sCommon.TestQuick(t, sSSLOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLBad1, &validContainer)
	sCommon.TestNegative(t, sSSLBad2, &validContainer)
	sCommon.TestQuick(t, sSSLOk1, &validContainer, 1)

	go rSSLAuth.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad1, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad2, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad3, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad4, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad5, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
	sCommon.TestNegative(t, sSSLAuthBad6, &validContainer)
	sCommon.TestQuick(t, sSSLAuthOk1, &validContainer, 1)
}

func TestHttpHandlersWithAuth(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("TestHttpHandlersWithAuth encountered an assert errror")
		}
	}()
	config, err := config.Bytes([]byte(`
{
	"senders": {
		"open_endpoint_ok": {
			"type": "http",
			"url": "https://localhost:6443",
			"insecure": true
		},
		"authentication_required_correct_details": {
			"type": "http",
			"url": "https://god:hunter2@localhost:6443/auth",
			"insecure": true
		},
		"authentication_required_invalid_details": {
			"type": "http",
			"url": "https://god:wrong-pw@localhost:6443/auth",
			"insecure": true
		},
		"common": {
			"type": "test"
		}
	},
	"receivers": {
		"ssl_with_auth": {
			"type": "http",
			"address": "localhost:6443",
			"handlers": {
				"/": "common",
				"/auth": "common"
			},
			"certfile": "../docs/examples/basics/cacert-snakeoil.pem",
			"keyfile": "../docs/examples/basics/privkey-snakeoil.pem",
			"auth": {
				"/auth": {
					"username": "god",
					"password": "hunter2"
				}
			}
		}
	},
	"handlers": {
		"common": {
			"parser": "json",
			"transformers": [],
			"sender": "common"
		}
	}
}`))
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
		return
	}

	openEndpointOk := config.Senders["open_endpoint_ok"].Sender.(*sender.HTTP)
	authRequiredCorrectDetails := config.Senders["authentication_required_correct_details"].Sender.(*sender.HTTP)
	authRequiredIncorrectDetails := config.Senders["authentication_required_invalid_details"].Sender.(*sender.HTTP)
	sCommon := config.Senders["common"].Sender.(*sender.Test)

	rSSLAuthHandlers := config.Receivers["ssl_with_auth"].Receiver.(*receiver.HTTP)

	go rSSLAuthHandlers.Start()
	time.Sleep(time.Duration(100 * time.Millisecond))
	sCommon.TestQuick(t, openEndpointOk, &validContainer, 1)
	sCommon.TestQuick(t, authRequiredCorrectDetails, &validContainer, 1)
	sCommon.TestNegative(t, authRequiredIncorrectDetails, &validContainer)
}

var bConfig *config.Config

func TestMain(m *testing.M) {
	var err error
	bConfig, err = config.Bytes([]byte(`
{
	"senders": {
		"plain_origin": {
			"type": "http",
			"url": "http://localhost:1349",
			"connsperhost": 300
		},
		"plain_batch": {
			"type": "batch",
			"next": "plain_origin",
			"interval": "1h",
			"threshold": 100
		},
		"ssl_auth": {
			"type": "http",
			"url": "https://god:hunter2@localhost:5449",
			"insecure": true
		},
		"ssl_auth_batch": {
			"type": "batch",
			"next": "ssl_auth",
			"interval": "1h",
			"threshold": 100
		},
		"common": {
			"type": "test"
		}
	},
	"receivers": {
		"plain": {
			"type": "http",
			"address": "localhost:1349",
			"handlers": { "/": "common"}
		},
		"ssl_auth": {
			"type": "http",
			"address": "localhost:5449",
			"handlers": { "/": "common"},
			"certfile": "../docs/examples/basics/cacert-snakeoil.pem",
			"keyfile": "../docs/examples/basics/privkey-snakeoil.pem",
			"username": "god",
			"password": "hunter2"
		}
	},
	"handlers": {
		"common": {
			"parser": "json",
			"transformers": [],
			"sender": "common"
		}
	}
}`))
	if err != nil {
		fmt.Printf("Failed to load config: %v", err)
		os.Exit(1)
	}

	rPlain := bConfig.Receivers["plain"].Receiver.(*receiver.HTTP)
	rSSL := bConfig.Receivers["ssl_auth"].Receiver.(*receiver.HTTP)

	go rPlain.Start()
	go rSSL.Start()
	time.Sleep(time.Duration(10 * time.Millisecond))
	os.Exit(m.Run())
}

// BenchmarkHttp relies on the bConfig structure and TestMain having
// started a http receiver to be tested. It ends up benchmarking:
//
// - The HTTP sender
//
// - The HTTP receiver
//
// - The JSON parser
//
// It is also affected heavily by your TCP stack.
//
// There are 6 variants. The first two are to compare http with https, and
// both serve ten containers per iteration.
//
// The next four compare batching, non-batching and with and without SSL.
func BenchmarkHttp_json(b *testing.B) {
	sPlainOrigin := bConfig.Senders["plain_origin"].Sender.(*sender.HTTP)
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sPlainOrigin, &validContainer, 10, 10)
	}
}

func BenchmarkHttp_ssl_json(b *testing.B) {
	sPlainOrigin := bConfig.Senders["ssl_auth"].Sender.(*sender.HTTP)
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sPlainOrigin, &validContainer, 10, 10)
	}
}

func BenchmarkHttp_batch_json(b *testing.B) {
	sPlainBatch := bConfig.Senders["plain_batch"].Sender
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sPlainBatch, &validContainer, 100, 1)
	}
}

func BenchmarkHttp_nobatch_json(b *testing.B) {
	sPlainNoBatch := bConfig.Senders["plain_origin"].Sender
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sPlainNoBatch, &validContainer, 100, 100)
	}
}

func BenchmarkHttp_ssl_nobatch_json(b *testing.B) {
	sSSL := bConfig.Senders["ssl_auth"].Sender
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sSSL, &validContainer, 100, 100)
	}
}

func BenchmarkHttp_ssl_batch_json(b *testing.B) {
	sSSL := bConfig.Senders["ssl_auth_batch"].Sender
	sCommon := bConfig.Senders["common"].Sender.(*sender.Test)
	sCommon.SetSync(true)
	for i := 0; i < b.N; i++ {
		sCommon.TestSync(b, sSSL, &validContainer, 100, 1)
	}
}
