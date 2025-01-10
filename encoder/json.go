package encoder

import (
	"encoding/json"

	"github.com/telenornms/skogul"
)

type JSON struct {
	Pretty bool `doc:"Pretty print/indent the json. Defaults to no - compact."`
}

func (x JSON) Encode(c *skogul.Container) ([]byte, error) {
	var b []byte
	var err error
	if x.Pretty {
		b, err = json.MarshalIndent(*c, "", "  ")
	} else {
		b, err = json.Marshal(*c)
	}
	return b, err
}

func (x JSON) EncodeMetric(m *skogul.Metric) ([]byte, error) {
	var b []byte
	var err error
	if x.Pretty {
		b, err = json.MarshalIndent(*m, "", "  ")
	} else {
		b, err = json.Marshal(*m)
	}
	return b, err
}
