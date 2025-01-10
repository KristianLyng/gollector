package transformer

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/dolmen-go/jsonptr"
	"github.com/telenornms/skogul"
)

type Ban struct {
	LookupData      map[string]interface{} `doc:"Map of key value pairs to lookup in metrics. Looks in data fields. Key is json pointer, value any. E.g. /foo/bar: \"bar\". This is an exact match and can use any data type."`
	LookupMetadata  map[string]interface{} `doc:"Map of key value pairs to lookup in metrics. Looks in metadata fields. Key is json pointer, value any. E.g. /foo/bar: \"bar\". This is an exact match and can use any data type."`
	RegexpData      map[string]string      `doc:"Map of key value pairs to lookup in metrics. Looks in data fields. Key is json pointer, value any. E.g. /foo/bar: \"bar\". This uses regular expression and only works on strings or byte arrays."`
	RegexpMetadata  map[string]string      `doc:"Map of key value pairs to lookup in metrics. Looks in metadata fields. Key is json pointer, value any. E.g. /foo/bar: \"bar\". This uses a regular expression and only works on strings or byte arrays."`
	dataRegexps     map[string]*regexp.Regexp
	metadataRegexps map[string]*regexp.Regexp
	err             error
	init            sync.Once
}

func (b *Ban) Transform(c *skogul.Container) error {
	b.init.Do(func() {
		b.dataRegexps = make(map[string]*regexp.Regexp)
		b.metadataRegexps = make(map[string]*regexp.Regexp)
		for pathKey, pathValue := range b.RegexpData {
			b.dataRegexps[pathKey], b.err = regexp.Compile(pathValue)
			if b.err != nil {
				return
			}
		}
		for pathKey, pathValue := range b.RegexpMetadata {
			b.metadataRegexps[pathKey], b.err = regexp.Compile(pathValue)
			if b.err != nil {
				return
			}
		}
	})

	if b.err != nil {
		return fmt.Errorf("unable to compile regexp: %w", b.err)
	}

	for pathKey, pathValue := range b.LookupData {
		newMetrics := make([]*skogul.Metric, 0, len(c.Metrics))
		for _, mi := range c.Metrics {
			var ptr interface{}
			ptr, _ = jsonptr.Get(mi.Data, pathKey)
			if ptr != pathValue {
				newMetrics = append(newMetrics, mi)
			}
		}
		c.Metrics = newMetrics
	}

	for pathKey, pathValue := range b.LookupMetadata {
		newMetrics := make([]*skogul.Metric, 0, len(c.Metrics))
		for _, mi := range c.Metrics {
			var ptr interface{}
			ptr, _ = jsonptr.Get(mi.Metadata, pathKey)
			if ptr != pathValue {
				newMetrics = append(newMetrics, mi)
			}
		}
		c.Metrics = newMetrics
	}

	for pathKey, pathValue := range b.dataRegexps {
		newMetrics := make([]*skogul.Metric, 0, len(c.Metrics))
		for _, mi := range c.Metrics {
			var ptr interface{}
			ptr, _ = jsonptr.Get(mi.Data, pathKey)
			var sptr []byte
			switch ptr.(type) {
			case string:
				sptr = []byte(ptr.(string))
			case []byte:
				sptr = ptr.([]byte)
			default:
				newMetrics = append(newMetrics, mi)
				continue
			}
			if !pathValue.Match(sptr) {
				newMetrics = append(newMetrics, mi)
			}
		}
		c.Metrics = newMetrics
	}

	for pathKey, pathValue := range b.metadataRegexps {
		newMetrics := make([]*skogul.Metric, 0, len(c.Metrics))
		for _, mi := range c.Metrics {
			var ptr interface{}

			ptr, _ = jsonptr.Get(mi.Metadata, pathKey)
			var sptr []byte
			switch ptr.(type) {
			case string:
				sptr = []byte(ptr.(string))
			case []byte:
				sptr = ptr.([]byte)
			default:
				newMetrics = append(newMetrics, mi)
				continue
			}
			if !pathValue.Match(sptr) {
				newMetrics = append(newMetrics, mi)
			}
		}
		c.Metrics = newMetrics
	}
	newMetrics := make([]*skogul.Metric, len(c.Metrics))
	copy(newMetrics, c.Metrics)
	c.Metrics = newMetrics
	return nil
}

func (b *Ban) Verify() error {
	for _, pathValue := range b.RegexpData {
		_, err := regexp.Compile(pathValue)
		if err != nil {
			return fmt.Errorf("unable to compile regexp `%s': %w", pathValue, err)
		}
	}
	for _, pathValue := range b.RegexpMetadata {
		_, err := regexp.Compile(pathValue)
		if err != nil {
			return fmt.Errorf("unable to compile regexp `%s': %w", pathValue, err)
		}
	}
	return nil
}
