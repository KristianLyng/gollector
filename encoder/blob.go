package encoder

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/telenornms/skogul"
)

/*
TODO:
- handle multiple metrics
- add separator
- test cases
*/
type Blob struct {
	Field     string `doc:"Data field to read data from, defaults to 'data'."`
	Delimiter []byte `doc:"Base64 encoded delimiter. Defaults to no delimiter."`
	once      sync.Once
	err       error
}

func (x *Blob) init() {
	if x.Field == "" {
		x.Field = "data"
	}
}

func (x *Blob) Encode(c *skogul.Container) ([]byte, error) {
	x.once.Do(func() {
		x.init()
	})
	b := bytes.Buffer{}
	for i := 0; i < len(c.Metrics); i++ {
		if i > 0 && len(x.Delimiter) > 0 {
			n, err := b.Write(x.Delimiter)
			if n != len(x.Delimiter) {
				return nil, fmt.Errorf("unable to write whole delimiter, wrote %d of %d bytes", n, len(x.Delimiter))
			}
			if err != nil {
				return nil, fmt.Errorf("unable to append to encoding buffer: %w", err)
			}
		}
		b2, ok := c.Metrics[i].Data[x.Field].([]byte)
		if !ok {
			return nil, fmt.Errorf("field is not a byte array")
		}
		n, err := b.Write(b2)
		if n != len(b2) {
			return nil, fmt.Errorf("unable to write whole metric, wrote %d of %d bytes", n, len(x.Delimiter))
		}
		if err != nil {
			return nil, fmt.Errorf("unable to append to encoding buffer: %w", err)
		}
	}
	rb := b.Bytes()[:]
	return rb, nil
}

func (x *Blob) EncodeMetric(m *skogul.Metric) ([]byte, error) {
	x.once.Do(func() {
		x.init()
	})
	b, ok := m.Data[x.Field].([]byte)
	if !ok {
		return nil, fmt.Errorf("field is not a byte array")
	}
	return b, nil
}
