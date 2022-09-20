/*
 * skogul, rfc 5424 structured data parser
 *
 * Copyright (c) 2020 Telenor Norge AS
 * Author(s):
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

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/telenornms/skogul"
)

var sdLog = skogul.Logger("parser", "structured_data")

// StructuredData supports parsing RFC5424 structured data through the Parse() function
// Note: This does not parse a full syslog message.
// In the resulting metrics, there may be a metadata field
// which contains the SD-ID from the message if it was present.
// By default the metadata field is named SD-ID, but it may be overridden in configuration.
// SD-ID: https://datatracker.ietf.org/doc/html/draft-ietf-syslog-protocol-23#section-6.3.2
type StructuredData struct {
	SDIDField string `doc:"The field to store the SDID in if it was present." default:"SD-ID"`
	once      sync.Once
}

// Parse converts RFC5424 Structured Data data into a skogul Container
func (sd *StructuredData) Parse(bytes []byte) (*skogul.Container, error) {
	sd.once.Do(func() {
		if sd.SDIDField == "" {
			sd.SDIDField = "sd-id"
		}
	})
	metrics, err := sd.parseStructuredData(bytes)
	if err != nil {
		return nil, err
	}

	return &skogul.Container{
		Metrics: metrics,
	}, nil
}

// parseStructuredData takes the raw input and parses it
// this takes care of splitting input on newlines etc
func (sd *StructuredData) parseStructuredData(data []byte) ([]*skogul.Metric, error) {
	lineScanner := bufio.NewScanner(bytes.NewReader(data))
	lineScanner.Split(splitStructuredDataMetrics)

	metrics := make([]*skogul.Metric, 0)

	timestamp := skogul.Now()
	for lineScanner.Scan() {
		line := lineScanner.Bytes()
		if len(line) == 0 {
			// Skip empty lines
			continue
		}

		metric := &skogul.Metric{
			Time:     &timestamp,
			Metadata: make(map[string]interface{}),
			Data:     make(map[string]interface{}),
		}

		// Set up the Key-Value scanner to extract data
		kvScanner := bufio.NewScanner(bytes.NewReader(line))
		kvScanner.Split(splitKeyValuePairs)

		for kvScanner.Scan() {
			tag := strings.Trim(kvScanner.Text(), "\u0000")
			tagValue := strings.SplitN(tag, "=", 2)

			if len(tagValue) == 1 && metric.Metadata[sd.SDIDField] == nil {
				// Set the SD-ID if it exists in the message (https://datatracker.ietf.org/doc/html/draft-ietf-syslog-protocol-23#section-6.3.2)
				metric.Metadata[sd.SDIDField] = tagValue[0]
				continue
			}

			if len(tagValue) == 1 {
				if strings.TrimSpace(tagValue[0]) == "" {
					return nil, fmt.Errorf("got invalid data in the middle of a structured data line")
				}
				continue
			}

			paramName := tagValue[0]
			paramValue := tagValue[1][1 : len(tagValue[1])-1] // remove leading and trailing "s

			// @ToDo: Support multiple paramName with different paramValue
			// if the value already exists, replace it with an array ?

			metric.Data[paramName] = paramValue
		}
		metrics = append(metrics, metric)
	}
	if len(metrics) == 0 {
		sdLog.WithField("bytes", len(data)).Warnf("RFC5424/Structured Data parser failed to parse any lines")
		return nil, fmt.Errorf("failed to parse RFC5424 lines")
	}
	return metrics, nil
}

// splitKeyValuePairs splits a section (tag key=value pairs or field key=value pairs)
func splitKeyValuePairs(data []byte, atEOF bool) (advance int, token []byte, err error) {
	fieldWidth, newData := structuredDataParser(data, true, false)

	returnChars := len(newData)

	if atEOF {
		// EOF, return with what we have left
		return returnChars + 1, newData[:returnChars], nil
	} else if returnChars == len(data) {
		// 'Soft EOF', we don't actually have more data
		// but we might have a separator char leftover.
		return returnChars, newData[:returnChars], nil
	}

	// Skip the trailing comma between each key=value pair, but still advance counter
	return fieldWidth, newData[:returnChars], nil
}

// splitStructuredDataMatrics splits a byte-stream of structured data into a list
// of metrics. This is most commonly split on newline, but multiple metrics can also
// appear on the same line, so we also split those.
func splitStructuredDataMetrics(data []byte, atEOF bool) (advance int, token []byte, err error) {
	fieldWidth, newData := structuredDataParser(data, false, true)

	returnChars := len(newData)

	if atEOF || returnChars == len(data) {
		// EOF, return with what we have left
		advance, token, err = returnChars+1, newData[:returnChars], nil
		return advance, token, err
	}

	// Skip the trailing comma between each key=value pair, but still advance counter
	advance, token, err = fieldWidth, newData[:returnChars], nil
	return advance, token, err
}

// struturedDataParser parses a structured data-line.
// A boolean flag decides whether or not escape characters should remain in the output
// or have their prepending escape character removed.
func structuredDataParser(bytes []byte, removeEscapedCharsFromResult, stopOnNewMetric bool) (tokens int, data []byte) {
	// Discard lines beginning with spaces, as this is not allowed in the RFC.
	if len(bytes) > 0 && bytes[0] == ' ' {
		return len(bytes), nil
	}

	openQuote := false
	escape := false
	escapeChars := make([]int, 0)
	escapeCharsWidth := make([]int, 0)
	tokens = 0
	for width := 0; tokens < len(bytes); tokens += width {
		var c rune
		c, width = utf8.DecodeRune(bytes[tokens:])

		if escape {
			escape = false
			continue
		}

		// If we receive an un-escaped ] or newline character, this section is done
		// and we'll restart parsing of the rest (if any) as a new section.
		if c == ']' || c == '\n' {
			break
		}

		// If there is an open quote, continue until we find the closing quote
		if openQuote {
			if c == '"' {
				// We found the closing quote, mark it and continue regular operations
				openQuote = false
				continue
			}
			// Fast forward loop until we find the closing quote
			continue
		}

		// We found the opening of a quote, continue until we find the closing one
		if c == '"' {
			openQuote = true
			continue
		}

		// Skip next char
		if c == '\\' {
			escape = true
			if removeEscapedCharsFromResult {
				escapeChars = append([]int{tokens}, escapeChars...)
				escapeCharsWidth = append([]int{width}, escapeCharsWidth...)
			}
			continue
		}

		// Stop when we reach a space, unless we're
		// instructed to only stop on new metrics,
		// in which case we will keep going until
		// we find a closing bracket.
		if !stopOnNewMetric && c == ' ' {
			break
		}
	}

	// Prepare the return value
	data = bytes[:tokens]

	for i, escapedChar := range escapeChars {
		if removeEscapedCharsFromResult {
			data = []byte(fmt.Sprintf("%s%s", data[0:escapedChar], data[escapedChar+escapeCharsWidth[i]:tokens]))
		}
		tokens += escapeCharsWidth[i]
	}

	// If we haven't skipped any chars, we need to tell the scanner to advance one position extra
	// to skip over separator of the next key=value pair
	if len(escapeChars) == 0 {
		tokens += 1
	}

	skipLeadingChars := 0
	// If the value starts with a [, we remove it from the output
	if len(bytes) >= 1 && bytes[0] == '[' {
		skipLeadingChars = 1
	}

	data = data[skipLeadingChars:]

	return tokens, data
}
