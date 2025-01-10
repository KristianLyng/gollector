package transformer

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/telenornms/skogul"
)

type BanField struct {
	SourceData     string `doc:"Data field to ban"`
	RegexpData     string `doc:"Regex to match value of source-data field"`
	regexpData     *regexp.Regexp
	SourceMetadata string `doc:"Metadata field to ban"`
	RegexpMetadata string `doc:"Regex to match value of source-metadata field"`
	regexpMetadata *regexp.Regexp
	errData        error
	errMetadata    error
	init           sync.Once
}

func (b *BanField) Transform(c *skogul.Container) error {
	b.init.Do(func() {
		b.regexpData, b.errData = regexp.Compile(b.RegexpData)
		b.regexpMetadata, b.errMetadata = regexp.Compile(b.RegexpMetadata)
	})

	if b.errData != nil {
		return fmt.Errorf("unable to compile regexp `%s': %w", b.RegexpData, b.errData)
	}
	if b.errMetadata != nil {
		return fmt.Errorf("unable to compile regexp `%s': %w", b.RegexpMetadata, b.errMetadata)
	}

	for _, metric := range c.Metrics {
		if b.SourceData != "" {
			if str, ok := metric.Data[b.SourceData]; ok {
				if b.regexpData.Match([]byte(str.(string))) {
					delete(metric.Data, b.SourceData)
				}
			}
		}
		if b.SourceMetadata != "" {
			if str, ok := metric.Metadata[b.SourceMetadata]; ok {
				if b.regexpMetadata.Match([]byte(str.(string))) {
					delete(metric.Metadata, b.SourceMetadata)
				}
			}
		}
	}

	return nil
}

func (b *BanField) Verify() error {
	if b.SourceData != "" && b.RegexpData == "" {
		return fmt.Errorf("regexpdata field has to have a value when sourcedata is provided")
	}
	if b.SourceMetadata != "" && b.RegexpMetadata == "" {
		return fmt.Errorf("regexpmetadata field has to have a value when sourcemetadata is provided")
	}

	var err error

	_, err = regexp.Compile(b.RegexpData)
	if err != nil {
		return fmt.Errorf("failed to compile regexp for regexpdata field %v %v", b.RegexpData, err)
	}

	_, err = regexp.Compile(b.RegexpMetadata)
	if err != nil {
		return fmt.Errorf("failed to compile regexp for regexpmetadata field %v %v", b.RegexpMetadata, err)
	}
	return nil
}
