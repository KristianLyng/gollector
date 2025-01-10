/*
 * skogul, edit transformer
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
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

package transformer

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/telenornms/skogul"
)

var repLog = skogul.Logger("transformer", "edit")

// Replace executes a regular expression replacement of metric metadata.
type Replace struct {
	Source      string `doc:"Metadata key to read from."`
	Destination string `doc:"Metadata key to write to. Defaults to overwriting the source-key if left blank. Destination key will always be overwritten, e.g., even if the source key is missing, the key located at the destination will be removed."`
	Regex       string `doc:"Regular expression to match."`
	Replacement string `doc:"Replacement text. Can also use $1, $2, etc to reference sub-matches. Defaults to empty string - remove matching items."`
	regex       *regexp.Regexp
	once        sync.Once
	err         error
}

// Transform executes the regular expression replacement
func (replace *Replace) Transform(c *skogul.Container) error {
	replace.once.Do(func() {
		if replace.Destination == "" {
			replace.Destination = replace.Source
		}
		replace.regex, replace.err = regexp.Compile(replace.Regex)
	})
	// Verify() should catch this, so there's no reasonable way this
	// should happen. But in the off chance that a regex compiles on
	// the first attempt but not the second.... (e.g.: some serious
	// bugs). It will also catch our own bugs, if, for some reason, we
	// manage to botch up Verify() under some corner case.
	skogul.Assert(replace.err == nil)

	for mi := range c.Metrics {
		if c.Metrics[mi].Metadata == nil {
			continue
		}
		if c.Metrics[mi].Metadata[replace.Source] == nil {
			delete(c.Metrics[mi].Metadata, replace.Destination)
			continue
		}
		// FIXME: This should be a type cast to allow working with
		// both text strings (as per now) and []byte strings.
		// Similar to what is done in the ban transformer.
		str, ok := c.Metrics[mi].Metadata[replace.Source].(string)
		if !ok {
			// FIXME: What to do? It's tempting to copy the
			// key, but that could mean multiple references to
			// the same memory, which can create unexpected
			// behavior if other transformers want to modify
			// just one of the headers.
			repLog.WithField("source", replace.Source).Printf("Unable to transform non-string field %s with content %v", replace.Source, c.Metrics[mi].Metadata[replace.Source])
			// This is to confirm with the documentation and
			// ensure that this isn't exploited by providing a
			// bogus Source-field only to be able to provide a
			// custom destination field.
			delete(c.Metrics[mi].Metadata, replace.Destination)
			continue
		}
		c.Metrics[mi].Metadata[replace.Destination] = string(replace.regex.ReplaceAll([]byte(str), []byte(replace.Replacement)))
	}
	return nil
}

// Verify checks that the required variables are set and that the regular
// expression compiles
func (replace *Replace) Verify() error {
	if replace.Source == "" {
		return skogul.MissingArgument("Source")
	}
	if replace.Regex == "" {
		return skogul.MissingArgument("Regex")
	}
	regex, err := regexp.Compile(replace.Regex)
	if err != nil {
		return fmt.Errorf("replace transformer regex `%s' didn't compile: %w", replace.Regex, err)
	}
	skogul.Assert(regex != nil)
	return nil
}

// ReplaceData is a copy of the Replace transformer but works on the Data field
// Replace executes a regular expression replacement of metric data.
type ReplaceData struct {
	Source      string `doc:"Data key to read from."`
	Destination string `doc:"Data key to write to. Defaults to overwriting the source-key if left blank. Destination key will always be overwritten, e.g., even if the source key is missing, the key located at the destination will be removed."`
	Regex       string `doc:"Regular expression to match."`
	Replacement string `doc:"Replacement text. Can also use $1, $2, etc to reference sub-matches. Defaults to empty string - remove matching items."`
	regex       *regexp.Regexp
	once        sync.Once
	err         error
}

// Transform executes the regular expression replacement
func (replace *ReplaceData) Transform(c *skogul.Container) error {
	replace.once.Do(func() {
		if replace.Destination == "" {
			replace.Destination = replace.Source
		}
		replace.regex, replace.err = regexp.Compile(replace.Regex)
	})
	// Verify() should catch this, so there's no reasonable way this
	// should happen. But in the off chance that a regex compiles on
	// the first attempt but not the second.... (e.g.: some serious
	// bugs). It will also catch our own bugs, if, for some reason, we
	// manage to botch up Verify() under some corner case.
	skogul.Assert(replace.err == nil)

	for mi := range c.Metrics {
		if c.Metrics[mi].Data == nil {
			continue
		}
		if c.Metrics[mi].Data[replace.Source] == nil {
			delete(c.Metrics[mi].Data, replace.Destination)
			continue
		}
		// FIXME: This should be a type cast to allow working with
		// both text strings (as per now) and []byte strings.
		// Similar to what is done in the ban transformer.
		str, ok := c.Metrics[mi].Data[replace.Source].(string)
		if !ok {
			// FIXME: What to do? It's tempting to copy the
			// key, but that could mean multiple references to
			// the same memory, which can create unexpected
			// behavior if other transformers want to modify
			// just one of the headers.
			repLog.WithField("source", replace.Source).Printf("Unable to transform non-string field %s with content %v", replace.Source, c.Metrics[mi].Data[replace.Source])
			// This is to confirm with the documentation and
			// ensure that this isn't exploited by providing a
			// bogus Source-field only to be able to provide a
			// custom destination field.
			delete(c.Metrics[mi].Data, replace.Destination)
			continue
		}
		c.Metrics[mi].Data[replace.Destination] = string(replace.regex.ReplaceAll([]byte(str), []byte(replace.Replacement)))
	}
	return nil
}

// Verify checks that the required variables are set and that the regular
// expression compiles
func (replace *ReplaceData) Verify() error {
	if replace.Source == "" {
		return skogul.MissingArgument("Source")
	}
	if replace.Regex == "" {
		return skogul.MissingArgument("Regex")
	}
	regex, err := regexp.Compile(replace.Regex)
	if err != nil {
		return fmt.Errorf("replace transformer regex `%s' didn't compile: %w", replace.Regex, err)
	}
	skogul.Assert(regex != nil)
	return nil
}
