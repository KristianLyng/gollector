/*
 * skogul, transformer automation
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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
	"github.com/telenornms/skogul"
)

// Auto maps names to Transformers to allow auto configuration
var Auto skogul.ModuleMap

func init() {
	Auto.Add(skogul.Module{
		Name:     "templater",
		Aliases:  []string{"template", "templating"},
		Alloc:    func() interface{} { return &Templater{} },
		Help:     "Executes metric templating. See separate documentation for how skogul templating works.",
		AutoMake: true,
	})
	Auto.Add(skogul.Module{
		Name:    "metadata",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Metadata{} },
		Help:    "Enforces custom-rules on metadata of metrics.",
		Extras:  []interface{}{SourceDestination{}},
	})
	Auto.Add(skogul.Module{
		Name:    "data",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Data{} },
		Help:    "Enforces custom-rules for data fields of metrics.",
	})
	Auto.Add(skogul.Module{
		Name:    "cast",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Cast{} },
		Help:    "Casts fields to specific data types, where possible. If the fields are already the correct type, the CPU cost is negligible, however, it is better to fix senders where possible.",
	})
	Auto.Add(skogul.Module{
		Name:    "parse",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Parse{} },
		Help:    "Parses a metric using a skogul parser. Useful if data encoding is nested, e.g.: Original data is json, but contains a text field with influx line protocol data.",
	})
	Auto.Add(skogul.Module{
		Name:    "split",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Split{} },
		Help:    "Split an array inside a metric into multiple metrics, one for each array element.",
	})
	Auto.Add(skogul.Module{
		Name:    "dictsplit",
		Aliases: []string{},
		Alloc:   func() interface{} { return &DictSplit{} },
		Help:    "Split a dictionary/hash inside a metric into multiple metrics, one for each dictinoary/hash element.",
	})
	Auto.Add(skogul.Module{
		Name:    "replace",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Replace{} },
		Help:    "Uses a regular expression to replace the content of a metadata key, storing it to either a different metadata key, or overwriting the original.",
	})
	Auto.Add(skogul.Module{
		Name:    "replacedata",
		Aliases: []string{},
		Alloc:   func() interface{} { return &ReplaceData{} },
		Help:    "Uses a regular expression to replace the content of a data key, storing it to either a different data key, or overwriting the original.",
	})
	Auto.Add(skogul.Module{
		Name:    "switch",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Switch{} },
		Help:    "Conditionally apply transformers.",
		Extras:  []interface{}{Case{}},
	})
	Auto.Add(skogul.Module{
		Name:    "timestamp",
		Aliases: []string{},
		Alloc:   func() interface{} { return &Timestamp{} },
		Help:    "Extract a timestamp from the container data.",
	})
	Auto.Add(skogul.Module{
		Name:     "now",
		Aliases:  []string{"dummytimestamp"},
		Alloc:    func() interface{} { return &DummyTimestamp{} },
		Help:     "Forcibly set a timestamp (set to now/current time) on all metrics to ensure the container is valid even if the source doesn't provide a timestamp.",
		AutoMake: true,
	})
	Auto.Add(skogul.Module{
		Name:     "enrich",
		Aliases:  []string{},
		Alloc:    func() interface{} { return &Enrich{} },
		Help:     "PROTOTYPE/ALPHA: Static enrichment. Use a json-structured source document to enrich incoming metrics with additional data. See the docs/examples/ directory for an actual example. This is NOT production ready, and should only be used if you are prepared to file bug reports and update your setup as the transformer matures.",
		AutoMake: true,
	})
	Auto.Add(skogul.Module{
		Name:     "unflatten",
		Aliases:  []string{},
		Alloc:    func() interface{} { return &Unflatten{} },
		Help:     "Create structured data from flat keys, e.g.: foo.bar.zoo: 1, foo.baz.zoo: 2 to { foo: { bar: { zoo: 1 }, baz: { zoo: 2 } } }",
		AutoMake: true,
	})
	Auto.Add(skogul.Module{
		Name:     "ban",
		Aliases:  []string{},
		Alloc:    func() interface{} { return &Ban{} },
		Help:     "Ban values from nested structure using a path e.g. Path looking like this foo.bar.1 has a structure looking like this { foo: { bar: { 1: hello } } }. The last element in the path will get removed from the tree in this case 1: hello, you will end up having a tree looking like this { foo: { bar: {} } }.",
		AutoMake: false,
	})
	Auto.Add(skogul.Module{
		Name:     "banfield",
		Aliases:  []string{},
		Alloc:    func() interface{} { return &BanField{} },
		Help:     "Remove single fields in a metric based on a regular expression criteria",
		AutoMake: false,
	})
	Auto.Add(skogul.Module{
		Name:    "huwtodbm",
		Aliases: []string{},
		Alloc:   func() interface{} { return &HuWtoDBM{} },
		Help:    "Translate from uW to DBM.",
	})
}
