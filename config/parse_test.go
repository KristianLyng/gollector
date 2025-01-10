/*
 * skogul  config module help generation
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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
package config_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
)

func TestFile(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("File() paniced")
		}
	}()
	c, err := config.File("testdata/test.json")
	if err != nil {
		t.Errorf("File() failed: %v", err)
	}
	if c == nil {
		t.Errorf("File() returned nil config")
	}
}

func TestByte_ok(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("Byte() paniced")
		}
	}()
	var okData []byte
	okData = []byte(`
{
  "senders": {
    "tnet_alarms": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INERT INTO tnet_Alarms values(${foo})"
    },
    "mysql_log": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INSERT INTO liksomlog VALUES(${timestmap.timestamp},${metadata.name},${key})"
    },
    "forward": {
	    "type": "http",
	    "url": "http://localhost:8888"
    },
    "duplicate": {
	    "type": "dupe",
	    "next": ["forward","mysql_log","tnet_alarms"]
    },
    "batch": {
      "type": "batch",
      "next": "duplicate",
      "interval": "5s"
    },
    "det": {
	    "type": "detacher",
	    "next": "batch"
    }
  },
  "handlers": {
    "plain": {
      "parser": "json",
      "sender": "batch",
      "transformers": ["foo"]
    }
  },
  "transformers": {
	  "foo": {
		  "type": "templater"
	  }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }
}
`)
	c, err := config.Bytes(okData)
	if err != nil {
		t.Errorf("Bytes() failed: %v", err)
	}
	if c == nil {
		t.Errorf("Bytes() returned nil config")
	}
	badData := []byte(`{ "senders": { "x": { "type": "sql", "ConnStr": 5 } } }`)
	c, err = config.Bytes(badData)
	if err == nil {
		t.Errorf("Bytes() test 2 failed, sent bad data, didn't get error.")
	}
	if c != nil {
		t.Errorf("Bytes() with bad data returned valid config.")
	}
	noURL := []byte(`{ "senders": { "x" : { "type": "http" }}}`)
	c, err = config.Bytes(noURL)
	if err == nil {
		t.Errorf("Bytes() test 3 failed, http sender with no URL didn't get error.")
	}
	if c != nil {
		t.Errorf("Bytes() with bad data returned valid config.")
	}
}

func TestDefaultModules(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("Byte() paniced")
		}
	}()
	var okData []byte
	okData = []byte(`
{
  "handlers": {
    "plain": {
      "parser": "json",
      "sender": "debug",
      "transformers": ["templater"]
    }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }
}
`)
	c, err := config.Bytes(okData)
	if err != nil {
		t.Errorf("Bytes() failed: %v", err)
	}
	if c == nil {
		t.Errorf("Bytes() returned nil config")
	}
}

func TestUndefinedParser(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("Byte() paniced")
		}
	}()
	var okData []byte
	okData = []byte(`
{
  "handlers": {
    "plain": {
      "parser": "bondevik",
      "sender": "debug",
      "transformers": ["templater"]
    }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }
}
`)
	_, err := config.Bytes(okData)
	if err == nil {
		t.Errorf("Bytes() didn't fail, despite undefined parser.")
	}
}

func TestNamedParser(t *testing.T) {
	defer func() {
		if skogul.AssertErrors > 0 {
			t.Errorf("Byte() paniced")
		}
	}()
	var okData []byte
	okData = []byte(`
{
  "parsers": {
    "jens": {
      "type": "json"
    }
  },
  "handlers": {
    "plain": {
      "parser": "jens",
      "sender": "debug",
      "transformers": ["templater"]
    }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }
}
`)
	c, err := config.Bytes(okData)
	if err != nil {
		t.Errorf("Bytes() failed: %v", err)
	}
	if c == nil {
		t.Errorf("Bytes() returned nil config")
	}
}

func TestHelpModule(t *testing.T) {
	_, err := config.HelpModule(sender.Auto, "sql")
	if err != nil {
		t.Errorf("HelpModule(sender.Auto,\"sql\") didn't work: %v", err)
	}
}

func testBadConf(t *testing.T, badData string) {
	t.Helper()
	_, err := config.Bytes([]byte(badData))
	if err == nil {
		t.Errorf("Bytes() was ok, despite bad data")
	}
}

// Useful for visual confirmation - e.g. - check that the arrow points at
// the right thing. Important to test errors early in a config, in the
// middle and towards the end.
func Test_syntaxError(t *testing.T) {
	testBadConf(t,
		`{
  "senders": {
    "tnet_alarms":: {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INERT INTO tnet_Alarms values(${foo})"
    }
  }
}`)

	testBadConf(t,
		`x{
  "senders": {
    "tnet_alarms":: {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INERT INTO tnet_Alarms values(${foo})"
    }
  }
}`)
	testBadConf(t,
		`{
  "senders": {
    "tnet_alarms": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INERT INTO tnet_Alarms values(${foo})"
    },
    "mysql_log": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INSERT INTO liksomlog VALUES(${timestmap.timestamp},${metadata.name},${key})"
    },
    "forward": {
	    "type": "http",
	    "url": "http://localhost:8888"
    },
    "duplicate": {
	    "type": "dupe",
	    "next": ["forward","mysql_log","tnet_alarms"]
    },
    "batch": {
      "type" "batch",
      "next": "duplicate",
      "interval": "5s"
    },
    "det": {
	    "type": "detacher",
	    "next": "batch"
    }
  },
  "handlers": {
    "plain": {
      "parser": "json",
      "sender": "batch",
      "transformers": []
    }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }
}`)
	testBadConf(t,
		`{
  "senders": {
    "tnet_alarms": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INERT INTO tnet_Alarms values(${foo})"
    },
    "mysql_log": {
      "type": "sql",
      "Driver": "mysql",
      "ConnStr": "root:lol@/mydb",
      "Query": "INSERT INTO liksomlog VALUES(${timestmap.timestamp},${metadata.name},${key})"
    },
    "forward": {
	    "type": "http",
	    "url": "http://localhost:8888"
    },
    "duplicate": {
	    "type": "dupe",
	    "next": ["forward","mysql_log","tnet_alarms"]
    },
    "batch": {
      "type": "batch",
      "next": "duplicate",
      "interval": "5s"
    },
    "det": {
	    "type": "detacher",
	    "next": "batch"
    }
  },
  "handlers": {
    "plain": {
      "parser": "json",
      "sender": "batch",
      "transformers": []
    }
  },
  "receivers": {
    "http": {
      "type": "http",
      "handlers": {
	      "/": "plain"
      }
    }
  }`)
	testBadConf(t,
		`{
	"senders": {
		"tnet_alarms":: {
			"type": "sql",
			"Driver": "mysql",
			"ConnStr":: "root:lol@/mydb",
			"Query": "INERT INTO tnet_Alarms values(${foo})"
		}
	}
}`)

	_, err := config.Bytes([]byte(`{
    "receivers": {
      "udp": {
        "type": "udp",
        "address": "[::1]:5015",
        "handler": "protobuf",
		"packetsize": 9000
      }
    },
    "handlers": {
      "protobuf": {
        "parser": "protobuf",
        "transformers": ["remove", "remove2", "remove3", "remove4", "remove5"],
        "sender": "print"
      }
    },
    "transformers": {
      "sw": {
        "type": "switch",
        "cases": [
          {
            "when": "sensorName",
            "is": "junos_system_linecard_intf-exp:/junos/system/linecard/intf-exp/:/junos/system/linecard/intf-exp/:PFE",
            "transformers": ["remove"]
          }
        ]
      },
      "remove": {
        "type": "data",
        "remove": ["interfaceExp_stats"]
      },
      "remove2": {
        "type": "data",
        "remove": ["interfaceExp_stats"]
      },
      "remove3": {
        "type": "data",
        "remove": ["interfaceExp_stats"]
      },
      "remove4": {
        "type": "data",
        "remove": ["interfaceExp_stats"]
      },
      "remove5": {
        "type": "data",
        "remove": ["interfaceExp_stats"]
      }
    },
    "senders": {
      "print": {
        "type": "debug",
        "prefix": "DEBUG"
      }
    }
  }`))
	if err != nil {
		t.Errorf("Expected config to pass: %v", err)
	}
}

func TestFindSuperfluousReceiverConfigPropertiesFromFullConfig(t *testing.T) {
	rawConfig := []byte(`{"receivers": {
		"foo": {
		  "type": "udp",
		  "address": "[::1]:5015",
		  "packetsize": 9000,
		  "superfluousField": "this is not needed"
		}
	  }
	}`)

	var parsedConfig map[string]interface{}
	err := json.Unmarshal(rawConfig, &parsedConfig)

	relevantConfig := config.GetRelevantRawConfigSection(&parsedConfig, "receivers", "foo")

	if err != nil {
		t.Error("Failed to parse config")
	}

	configStruct := reflect.TypeOf(receiver.UDP{})
	superfluousProperties := config.VerifyOnlyRequiredConfigProps(&relevantConfig, "receivers", "foo", configStruct)

	if len(superfluousProperties) != 1 {
		t.Errorf("Expected 1 superfluous property but got %d", len(superfluousProperties))
	}

	if superfluousProperties[0] != "superfluousField" {
		t.Errorf("Expected to find '%s' in the superfluous fields list", "superfluousField")
	}
}

func TestFindSuperfluousReceiverConfigProperties(t *testing.T) {
	rawConfig := []byte(`{
		"type": "udp",
		"address": "[::1]:5015",
		"packetsize": 9000,
		"superfluousField": "this is not needed"
	}`)

	var c map[string]interface{}
	err := json.Unmarshal(rawConfig, &c)
	if err != nil {
		t.Error("Failed to parse config")
	}

	configStruct := reflect.TypeOf(receiver.UDP{})
	superfluousProperties := config.VerifyOnlyRequiredConfigProps(&c, "receivers", "foo", configStruct)

	if len(superfluousProperties) != 1 {
		t.Errorf("Expected 1 superfluous property but got %d", len(superfluousProperties))
	}

	if superfluousProperties[0] != "superfluousField" {
		t.Errorf("Expected to find '%s' in the superfluous fields list", "superfluousField")
	}
}

func TestBytesWorksWithSuperfluousReceiverConfigProperties(t *testing.T) {
	rawConfig := []byte(`{"receivers": {
		"foo": {
		  "type": "udp",
		  "handler": "dummy",
		  "address": "[::1]:5015",
			"packetsize": 9000,
		  "superfluousField": "this is not needed"
		}
	  },
	  "handlers": { "dummy": { "parser": "json", "sender": "null"} }
	}`)

	_, err := config.Bytes(rawConfig)
	if err != nil {
		t.Errorf("Failed to Bytes() config: %s", err)
	}
}

func TestReadConfigWithoutSuperfluousParamsNoSuperfluousParams(t *testing.T) {
	rawConfig := []byte(`{
    "receivers": {
      "foo": {
        "type": "stdin",
        "handler": "bar"
      }
    },
    "handlers": {
      "bar": {
	"parser": "json",
        "sender": "baz"
      }
    },
    "senders": {
      "baz": {
        "type": "null"
      }
    }
  }`)

	var c map[string]interface{}
	err := json.Unmarshal(rawConfig, &c)
	if err != nil {
		t.Errorf("Failed to unmarshal json: %s", err)
	}

	superfluousProperties := make([]string, 0)

	configStruct := reflect.TypeOf(receiver.Stdin{})
	c1 := config.GetRelevantRawConfigSection(&c, "receivers", "foo")
	superfluousProperties = append(superfluousProperties, config.VerifyOnlyRequiredConfigProps(&c1, "receiver", "foo", configStruct)...)

	configStruct = reflect.TypeOf(sender.Debug{})
	c2 := config.GetRelevantRawConfigSection(&c, "senders", "baz")
	superfluousProperties = append(superfluousProperties, config.VerifyOnlyRequiredConfigProps(&c2, "sender", "baz", configStruct)...)

	if len(superfluousProperties) > 0 {
		t.Errorf("Expected 0 superfluous config params, received %d (%s)", len(superfluousProperties), superfluousProperties)
	}

	_, err = config.Bytes(rawConfig)
	if err != nil {
		t.Errorf("Failed to Bytes() config: %s", err)
	}
}

func TestReadConfigFiles(t *testing.T) {
	c, err := config.ReadFiles("testdata/configs")
	if err != nil {
		t.Error("Error from config read files", err)
	}

	if c.Receivers["foo"] == nil || c.Receivers["bar"] == nil {
		t.Error("Missing a receiver which should be configured")
	}
}
