/*
 * skogul, configuration parsing
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

/*
Package config handles Skogul configuration parsing.
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
	"github.com/telenornms/skogul/parser"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
	"github.com/telenornms/skogul/transformer"
)

var confLog = skogul.Logger("core", "config")

// Sender wraps the skogul.Sender for configuration parsing.
type Sender struct {
	Type   string
	Sender skogul.Sender `json:"-"`
}

// Parser wraps the skogul.Parser for configuration parsing.
type Parser struct {
	Type   string
	Parser skogul.Parser `json:"-"`
}

// Receiver wraps the skogul.Receiver for configuration parsing.
type Receiver struct {
	Type     string
	Receiver skogul.Receiver `json:"-"`
}

// Encoder wraps the skogul.Encoder module-type for configuration parsing.
type Encoder struct {
	Type    string
	Encoder skogul.Encoder `json:"-"`
}

// Handler wraps skogul.Handler for configuration parsing.
type Handler struct {
	Parser                skogul.ParserRef
	Transformers          []*skogul.TransformerRef
	Sender                skogul.SenderRef
	IgnorePartialFailures bool
	Handler               skogul.Handler `json:"-"`
}

// Transformer wraps skogul.Transformer
type Transformer struct {
	Type        string
	Transformer skogul.Transformer `json:"-"`
}

// Config encapsulates all configuration for Skogul, and represent the
// top-level configuration object.
type Config struct {
	Handlers     map[string]*Handler
	Receivers    map[string]*Receiver
	Senders      map[string]*Sender
	Parsers      map[string]*Parser
	Encoders     map[string]*Encoder
	Transformers map[string]*Transformer
}

// MarshalJSON marshals Transformer config. See MarshalJSON for receiver - same
// same.
func (t *Transformer) MarshalJSON() ([]byte, error) {
	nest, err := json.Marshal(t.Transformer)
	if err != nil {
		return nil, err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(nest, &merged); err != nil {
		return nil, err
	}
	merged["type"] = t.Type
	return json.Marshal(merged)
}

// UnmarshalJSON picks up the type of the Receiver, instantiates a copy of
// that receiver, then unmarshals the remaining configuration onto that.
func (t *Transformer) UnmarshalJSON(b []byte) error {
	type tType struct {
		Type string
	}
	var myt tType
	if err := json.Unmarshal(b, &myt); err != nil {
		return err
	}
	t.Type = myt.Type
	if transformer.Auto[t.Type] == nil {
		return fmt.Errorf("unknown transformer %v", t.Type)
	}
	if transformer.Auto[t.Type].Alloc == nil {
		return fmt.Errorf("bad transformer %v", t.Type)
	}
	var ok bool
	t.Transformer, ok = (transformer.Auto[t.Type].Alloc()).(skogul.Transformer)
	skogul.Assert(ok)
	if err := json.Unmarshal(b, &t.Transformer); err != nil {
		return fmt.Errorf("transformer unmarshal: %w", err)
	}

	// Find superfluous config parameters
	var jsonConf map[string]interface{}
	json.Unmarshal(b, &jsonConf) // Assuming this works out well since it did up there ^
	VerifyOnlyRequiredConfigProps(&jsonConf, "transformer", t.Type, reflect.ValueOf(t.Transformer).Elem().Type())
	return nil
}

// MarshalJSON for a receiver marshals the actual instantiated receiver,
// then merges it to add "type". Probably not the most efficient
// implementation, since it does marshal-unmarshal-merge-marshal, but since
// this isn't really performance sensitive, that's ok.
func (r *Receiver) MarshalJSON() ([]byte, error) {
	nest, err := json.Marshal(r.Receiver)
	if err != nil {
		return nil, err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(nest, &merged); err != nil {
		return nil, err
	}
	merged["type"] = r.Type
	return json.Marshal(merged)
}

// UnmarshalJSON picks up the type of the Receiver, instantiates a copy of
// that receiver, then unmarshals the remaining configuration onto that.
func (r *Receiver) UnmarshalJSON(b []byte) error {
	type tType struct {
		Type string
	}
	var t tType
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	r.Type = t.Type
	if receiver.Auto[r.Type] == nil {
		return fmt.Errorf("unknown receiver %v", r.Type)
	}
	if receiver.Auto[r.Type].Alloc == nil {
		return fmt.Errorf("bad receiver %v", r.Type)
	}
	var ok bool
	r.Receiver, ok = (receiver.Auto[r.Type].Alloc()).(skogul.Receiver)
	skogul.Assert(ok)
	if err := json.Unmarshal(b, &r.Receiver); err != nil {
		return fmt.Errorf("receiver unmarshalling: %w", err)
	}

	// Find superfluous config parameters
	var jsonConf map[string]interface{}
	json.Unmarshal(b, &jsonConf) // Assuming this works out well since it did up there ^
	VerifyOnlyRequiredConfigProps(&jsonConf, "receiver", r.Type, reflect.ValueOf(r.Receiver).Elem().Type())
	return nil
}

// MarshalJSON marshals Parser config. See MarshalJSON for receiver - same
// same.
func (p *Parser) MarshalJSON() ([]byte, error) {
	nest, err := json.Marshal(p.Parser)
	if err != nil {
		return nil, err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(nest, &merged); err != nil {
		return nil, err
	}
	merged["type"] = p.Type
	return json.Marshal(merged)
}

// UnmarshalJSON for Parser. See UnmarshalJSON for Receiver - same same.
func (p *Parser) UnmarshalJSON(b []byte) error {
	type tType struct {
		Type string
	}
	var t tType
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	p.Type = t.Type
	if parser.Auto[p.Type] == nil {
		return fmt.Errorf("unknown parser%v", p.Type)
	}
	if parser.Auto[p.Type].Alloc == nil {
		return fmt.Errorf("bad parser %v", p.Type)
	}
	var ok bool
	p.Parser, ok = (parser.Auto[p.Type].Alloc()).(skogul.Parser)
	skogul.Assert(ok)
	if err := json.Unmarshal(b, &p.Parser); err != nil {
		return fmt.Errorf("parser unmarshalling: %w", err)
	}
	// Find superfluous config parameters
	var jsonConf map[string]interface{}
	json.Unmarshal(b, &jsonConf) // Assuming this works out well since it did up there ^
	VerifyOnlyRequiredConfigProps(&jsonConf, "parser", p.Type, reflect.ValueOf(p.Parser).Elem().Type())
	return nil
}

// MarshalJSON marshals Encoder config. See MarshalJSON for receiver - same
// same.
func (e *Encoder) MarshalJSON() ([]byte, error) {
	nest, err := json.Marshal(e.Encoder)
	if err != nil {
		return nil, err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(nest, &merged); err != nil {
		return nil, err
	}
	merged["type"] = e.Type
	return json.Marshal(merged)
}

// UnmarshalJSON for Parser. See UnmarshalJSON for Receiver - same same.
func (e *Encoder) UnmarshalJSON(b []byte) error {
	type tType struct {
		Type string
	}
	var t tType
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	e.Type = t.Type
	if encoder.Auto[e.Type] == nil {
		return fmt.Errorf("unknown encoder %v", e.Type)
	}
	if encoder.Auto[e.Type].Alloc == nil {
		return fmt.Errorf("bad encoder %v", e.Type)
	}
	var ok bool
	e.Encoder, ok = (encoder.Auto[e.Type].Alloc()).(skogul.Encoder)
	skogul.Assert(ok)
	if err := json.Unmarshal(b, &e.Encoder); err != nil {
		return fmt.Errorf("encoder unmarshalling: %w", err)
	}
	// Find superfluous config parameters
	var jsonConf map[string]interface{}
	json.Unmarshal(b, &jsonConf) // Assuming this works out well since it did up there ^
	VerifyOnlyRequiredConfigProps(&jsonConf, "encoder", e.Type, reflect.ValueOf(e.Encoder).Elem().Type())
	return nil
}

// MarshalJSON marshals Sender config. See MarshalJSON for receiver - same
// same.
func (s *Sender) MarshalJSON() ([]byte, error) {
	nest, err := json.Marshal(s.Sender)
	if err != nil {
		return nil, err
	}
	var merged map[string]interface{}
	if err := json.Unmarshal(nest, &merged); err != nil {
		return nil, err
	}
	merged["type"] = s.Type
	return json.Marshal(merged)
}

// UnmarshalJSON for Sender. See UnmarshalJSON for Receiver - same same.
func (s *Sender) UnmarshalJSON(b []byte) error {
	type tType struct {
		Type string
	}
	var t tType
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	s.Type = t.Type
	if sender.Auto[s.Type] == nil {
		return fmt.Errorf("unknown sender %v", s.Type)
	}
	if sender.Auto[s.Type].Alloc == nil {
		return fmt.Errorf("bad sender %v", s.Type)
	}
	var ok bool
	s.Sender, ok = (sender.Auto[s.Type].Alloc()).(skogul.Sender)
	skogul.Assert(ok)
	if err := json.Unmarshal(b, &s.Sender); err != nil {
		return fmt.Errorf("sender unmarshalling: %w", err)
	}
	// Find superfluous config parameters
	var jsonConf map[string]interface{}
	json.Unmarshal(b, &jsonConf) // Assuming this works out well since it did up there ^
	VerifyOnlyRequiredConfigProps(&jsonConf, "sender", s.Type, reflect.ValueOf(s.Sender).Elem().Type())
	return nil
}

func printSyntaxError(b []byte, offset int, text string) {
	start := offset
	start2 := 0
	lines := 0
	// Start by finding where we want to start. Work from offset and
	// decrement. start will represent "print start", start2 will
	// represent the start of the problematic line
	for i := offset; i >= 0 && lines < 3; i-- {
		start = i
		if len(b) <= i || b[i] == '\n' {
			if lines == 0 || (lines == 1 && start2 == offset) {
				start2 = start
			}
			lines++
		}
	}
	end := offset
	end2 := offset
	lines = 0
	// Next do things the other way around. End will be the actual end
	// of what to display, while end2 will be the first line after
	// "start2", e.g., the beginning of the line _after_ the
	// problematic one.
	for i := offset; i <= len(b) && lines < 3; i++ {
		end = i
		if i == len(b) || b[i] == '\n' {
			if lines == 0 {
				end2 = end
			}
			lines++
		}
	}
	fmt.Printf("Unable to parse JSON configuration at byte offset %d.\nError: %s\nContext:\n", offset, text)
	fmt.Println(string(b[start:end2]))
	for i := start2; i < (offset - 2); i++ {
		if b[i] == '	' {
			fmt.Print("--------")
		} else {
			fmt.Print("-")
		}
	}
	// We found the crappy part!
	fmt.Println("💩")
	end2++
	if end2 > len(b) {
		end2 = len(b)
	}
	fmt.Println(string(b[end2:end]))
}

// Bytes parses json in the provided byte array and returns a
// configuration.
//
// It does this by first doing a pass where it just does JSON
// unmarshalling, which also updates sender and handler reference tables
// globally (unfortunately...), then calling secondPass(), which resolves
// references and does a final validation.
func Bytes(b []byte) (*Config, error) {
	var jsonData map[string]interface{}
	skogul.HandlerMap = skogul.HandlerMap[0:0]
	skogul.SenderMap = skogul.SenderMap[0:0]
	skogul.ParserMap = skogul.ParserMap[0:0]
	skogul.TransformerMap = skogul.TransformerMap[0:0]
	skogul.EncoderMap = skogul.EncoderMap[0:0]
	if err := json.Unmarshal(b, &jsonData); err != nil {
		jerr, ok := err.(*json.SyntaxError)
		if ok {
			printSyntaxError(b, int(jerr.Offset), jerr.Error())
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	c := Config{}
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("valid JSON, but not valid Skogul configuration: %w", err)
	}

	return secondPass(&c)
}

// Path opens a path (file or directory) and parses the configuration.
func Path(path string) (*Config, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration path: %w", err)
	}

	if stat.IsDir() {
		return ReadFiles(path)
	}

	return File(path)
}

// File opens a config file and parses it, then returns the valid
// configuration, using Bytes()
func File(f string) (*Config, error) {
	dat, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}
	return Bytes(dat)
}

func findConfigFiles(path string) ([]string, error) {
	confLog.WithField("path", path).Debugf("Reading configuration files from %s", path)
	configFiles := make([]string, 0)
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			configFiles = append(configFiles, path)
		}
		return err
	})

	if err != nil {
		return nil, err
	}

	return configFiles, nil
}

// ReadFiles reads all JSON files (with the .JSON suffix) in a given directory
// and combines them to a configuration for the program.
func ReadFiles(p string) (*Config, error) {
	files, err := findConfigFiles(p)

	if err != nil {
		return nil, err
	}

	config := Config{}

	for _, f := range files {
		confLog.WithField("file", f).Debug("Reading file")
		b, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(b, &config)
		if err != nil {
			jerr, ok := err.(*json.SyntaxError)
			if ok {
				printSyntaxError(b, int(jerr.Offset), jerr.Error())
			}
			return nil, err
		}
	}

	return secondPass(&config)
}

// resolveSenders iterates over the skogul.SenderMap and resolves senders,
// using the provided configuration. It zeroes the senderMap upon
// completion.
func resolveSenders(c *Config) error {
	for _, s := range skogul.SenderMap {
		confLog.WithField("sender", s.Name).Debug("Resolving sender")

		if c.Senders[s.Name] == nil {
			m := sender.Auto.Lookup(s.Name)
			if m != nil {
				tmp := m.Alloc()
				tmp2 := tmp.(skogul.Sender)
				snew := Sender{}
				snew.Type = s.Name
				snew.Sender = tmp2
				if c.Senders == nil {
					c.Senders = make(map[string]*Sender)
				}
				c.Senders[s.Name] = &snew
			}
		}
		if c.Senders[s.Name] == nil {
			return fmt.Errorf("sender `%s' referenced but not defined", s.Name)
		}
		skogul.Identity[c.Senders[s.Name].Sender] = s.Name
		s.S = c.Senders[s.Name].Sender
	}
	skogul.SenderMap = skogul.SenderMap[0:0]
	return nil
}

// resolveParsers iterates over the skogul.ParserMap and resolve any
// parsers.
func resolveParsers(c *Config) error {
	for _, p := range skogul.ParserMap {
		confLog.WithField("parser", p.Name).Debug("Resolving parser")

		if c.Parsers[p.Name] == nil {
			m := parser.Auto.Lookup(p.Name)
			if m != nil {
				tmp := m.Alloc()
				tmp2 := tmp.(skogul.Parser)
				pnew := Parser{}
				pnew.Type = p.Name
				pnew.Parser = tmp2
				if c.Parsers == nil {
					c.Parsers = make(map[string]*Parser)
				}
				c.Parsers[p.Name] = &pnew
			}
		}
		if c.Parsers[p.Name] == nil {
			return fmt.Errorf("parser `%s' referenced but not defined", p.Name)
		}
		skogul.Identity[c.Parsers[p.Name].Parser] = p.Name
		p.P = c.Parsers[p.Name].Parser
	}
	skogul.ParserMap = skogul.ParserMap[0:0]
	return nil
}

// resolveEncoders iterates over the skogul.EncoderMap and resolve any
// encoders.
func resolveEncoders(c *Config) error {
	for _, e := range skogul.EncoderMap {
		confLog.WithField("encoder", e.Name).Debug("Resolving encoders")

		if c.Encoders[e.Name] == nil {
			m := encoder.Auto.Lookup(e.Name)
			if m != nil {
				tmp := m.Alloc()
				tmp2 := tmp.(skogul.Encoder)
				enew := Encoder{}
				enew.Type = e.Name
				enew.Encoder = tmp2
				if c.Encoders == nil {
					c.Encoders = make(map[string]*Encoder)
				}
				c.Encoders[e.Name] = &enew
			}
		}
		if c.Encoders[e.Name] == nil {
			return fmt.Errorf("encoder `%s' referenced but not defined", e.Name)
		}
		skogul.Identity[c.Encoders[e.Name].Encoder] = e.Name
		e.E = c.Encoders[e.Name].Encoder
	}
	skogul.EncoderMap = skogul.EncoderMap[0:0]
	return nil
}

// resolveHandlers iterates over handlers and instantiates them, since
// there is no unmarshaller (or need for one) that does this. It then
// iterates of the skogul.HandlerMap and resolves the handler references to
// the actual handlers.
//
// It then zeroes the skogul.HandlerMap
func resolveHandlers(c *Config) error {
	for _, h := range c.Handlers {
		logger := confLog.WithField("parser", h.Parser)

		h.Handler.Sender = h.Sender.S
		h.Handler.Transformers = make([]skogul.Transformer, 0)
		h.Handler.IgnorePartialFailures = h.IgnorePartialFailures
		h.Handler.SetParser(h.Parser.P)

		for _, t := range h.Transformers {
			logger = logger.WithField("transformer", t.Name)
			logger.Debug("Using predefined transformer")
			skogul.Assert(t.T != nil)
			h.Handler.Transformers = append(h.Handler.Transformers, t.T)
		}
	}
	for _, h := range skogul.HandlerMap {
		if c.Handlers[h.Name] == nil {
			return fmt.Errorf("handler `%s' referenced but not defined", h.Name)
		}
		h.H = &(c.Handlers[h.Name].Handler)
	}
	skogul.HandlerMap = skogul.HandlerMap[0:0]
	return nil
}

// resolveTransformers looks in the parsed config for transformers and initializes the
// actual transformers. Zeroizes the TransformerMap after if case a new
// config is applied without restarting.
func resolveTransformers(c *Config) error {
	logger := confLog.WithField("method", "resolveTransformers")
	for transformerName, t := range skogul.TransformerMap {
		logger = logger.WithField("transformer", transformerName)

		if c.Transformers[t.Name] == nil {
			m := transformer.Auto.Lookup(t.Name)
			if m != nil {
				tmp := m.Alloc()
				tmp2 := tmp.(skogul.Transformer)
				tnew := Transformer{}
				tnew.Type = t.Name
				tnew.Transformer = tmp2
				if c.Transformers == nil {
					c.Transformers = make(map[string]*Transformer)
				}
				c.Transformers[t.Name] = &tnew
			}
		}
		if c.Transformers[t.Name] != nil {
			logger.Debug("Using predefined transformer")
		} else {
			return fmt.Errorf("transformer `%s' referenced but not defined", t.Name)
		}
		skogul.Assert(c.Transformers[t.Name].Transformer != nil)
		skogul.Identity[c.Transformers[t.Name].Transformer] = t.Name
		t.T = c.Transformers[t.Name].Transformer
	}
	skogul.TransformerMap = skogul.TransformerMap[0:0]
	return nil
}

// identifyReceivers iterates over defined receivers and updates the
// identity map. This is required specially because the other modules each
// have module resolving logic where we do it for those modules, but there
// are no external references to receivers (yet?), so no code that iterates
// over them.
func identifyReceivers(c *Config) {
	for idx, name := range c.Receivers {
		skogul.Identity[name.Receiver] = idx
	}
}

// secondPass accepts a parsed configuration as input and resolves the
// references in it, and verifies basic integrity.
func secondPass(c *Config) (*Config, error) {
	skogul.Identity = make(map[interface{}]string)
	identifyReceivers(c)
	if err := resolveSenders(c); err != nil {
		return nil, err
	}
	if err := resolveTransformers(c); err != nil {
		return nil, err
	}
	if err := resolveParsers(c); err != nil {
		return nil, err
	}
	if err := resolveHandlers(c); err != nil {
		return nil, err
	}
	if err := resolveEncoders(c); err != nil {
		return nil, err
	}

	for idx, h := range c.Handlers {
		confLog.WithField("handler", idx).Debug("Verifying handler configuration")
		if err := verifyItem("handler", idx, h.Handler); err != nil {
			return nil, err
		}
	}
	for idx, t := range c.Transformers {
		confLog.WithField("transformer", idx).Debug("Verifying transformer configuration")
		if err := verifyItem("transformer", idx, t.Transformer); err != nil {
			return nil, err
		}
		deprecateCheck("transformer", idx, t.Transformer)
	}
	for idx, s := range c.Senders {
		confLog.WithField("sender", idx).Debug("Verifying sender configuration")
		if err := verifyItem("sender", idx, s.Sender); err != nil {
			return nil, err
		}
		deprecateCheck("sender", idx, s.Sender)
	}
	for idx, r := range c.Receivers {
		confLog.WithField("receiver", idx).Debug("Verifying receiver configuration")
		if err := verifyItem("receiver", idx, r.Receiver); err != nil {
			return nil, err
		}
		deprecateCheck("receiver", idx, r.Receiver)
	}
	for idx, e := range c.Encoders {
		confLog.WithField("encoders", idx).Debug("Verifying encoder configuration")
		if err := verifyItem("encoder", idx, e.Encoder); err != nil {
			return nil, err
		}
		deprecateCheck("encoder", idx, e.Encoder)
	}
	for idx, p := range c.Parsers {
		confLog.WithField("parsers", idx).Debug("Verifying parser configuration")
		if err := verifyItem("parser", idx, p.Parser); err != nil {
			return nil, err
		}
		deprecateCheck("parser", idx, p.Parser)
	}

	return c, nil
}

func deprecateCheck(family string, name string, item interface{}) {
	i, ok := item.(skogul.Deprecated)
	if !ok {
		return
	}
	if err := i.Deprecated(); err != nil {
		confLog.Warnf("deprecation warning for %s \"%s\": %s", family, name, err)
	}
}

// verifyItem checks if the item implements Verifier and if so, verifies
// the item. Otherwise, returns nil.
func verifyItem(family string, name string, item interface{}) error {
	i, ok := item.(skogul.Verifier)
	if !ok {
		confLog.WithFields(logrus.Fields{"family": family, "name": name}).Trace("No verifier found")
		return nil
	}
	err := i.Verify()
	if err != nil {
		return fmt.Errorf("configuration for %s `%s' doesn't verify: %w", family, name, err)
	}
	confLog.WithFields(logrus.Fields{"family": family, "name": name}).Trace("Verified OK")
	return nil
}

func findFieldsOfStruct(T reflect.Type) []string {
	requiredProps := make([]string, 0)
	switch T.Kind() {
	case reflect.Struct:
		for i := 0; i < T.NumField(); i++ {
			field := T.Field(i)
			jsonTag := field.Tag.Get("json")

			property := field.Name
			if jsonTag != "" {
				property = jsonTag
			}
			requiredProps = append(requiredProps, property)
		}
	}

	return requiredProps
}

// GetRelevantRawConfigSection is a helper function to dig down into a Config JSON
// and select the wanted family (receivers, transformers, senders) and item (foo).
func GetRelevantRawConfigSection(rawConfig *map[string]interface{}, family, section string) map[string]interface{} {
	configFamily, ok := (*rawConfig)[family].(map[string]interface{})
	if !ok {
		confLog.WithFields(logrus.Fields{
			"family":  family,
			"section": section,
		}).Warnf("Failed to cast config family to map[string]interface{}")
		return nil
	}

	configSection, ok := configFamily[section].(map[string]interface{})
	if !ok {
		confLog.WithFields(logrus.Fields{
			"family":  family,
			"section": section,
		}).Warnf("Failed to cast config section to map[string]interface{}")
		return nil
	}
	return configSection
}

// VerifyOnlyRequiredConfigProps checks for undefined configuration properties
// It can be used to identify typos or invalid configuration
// Use 'config.GetRelevantRawConfigSection' first for handler if you have a full config.
func VerifyOnlyRequiredConfigProps(componentConfig *map[string]interface{}, family, handler string, T reflect.Type) []string {
	requiredProps := findFieldsOfStruct(T)

	superfluousProperties := make([]string, 0)

	for prop := range *componentConfig {
		propertyDefined := false

		if prop == "type" {
			// Skip the type specifying what type this is
			continue
		}

		for _, requiredProp := range requiredProps {
			if strings.ToLower(prop) == strings.ToLower(requiredProp) {
				propertyDefined = true
				break
			}
		}
		if !propertyDefined {
			superfluousProperties = append(superfluousProperties, prop)
			confLog.WithFields(logrus.Fields{
				"family":   family,
				"handler":  handler,
				"property": prop,
			}).Warn("Configuration property configured but not defined in code (this property won't change anything, is it wrongly defined?)")
		}
	}

	return superfluousProperties
}
