/*
 * skogul, enrichment transformer
 *
 * Copyright (c) 2022 Telenor Norge AS
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
	"sync"

	"github.com/telenornms/skogul"
)

// keyType identifies the type the hash returns
type keyType string

// entry represents what is supposed to be added. An entry is added to a
// metric's metadatafields on a hit.
type entry map[string]interface{}

/*
Enrich transformer adds additional metadatainformation to metrics. This
is a work in progress, see #224 for on-going discussions.
*/
type Enrich struct {
	Keys  []string `doc:"Metadatafields to match, e.g.: sysName and ifName."`
	lock  sync.RWMutex
	store map[keyType]*entry
}

var eLog = skogul.Logger("transformer", "enrich")

// Hash calculates the hash for a metric using relevant settings (e.g.:
// e.Keys). Only exported for the sake of tests and benchmarks.
//
// XXX: This used to be a manually created hash, now we rely on the
// underlying map[] implementation instead.
func (e *Enrich) Hash(m skogul.Metric) keyType {
	ret := ""
	for _, meta := range e.Keys {
		ret = ret + fmt.Sprintf("%v", m.Metadata[meta])
	}
	return keyType(ret)
}

// Update uses the provided Container to update/bootstrap the enrichment
// database (e.store). It is used by the enrichmentupdater-sender.
func (e *Enrich) Update(c *skogul.Container) {
	e.lock.Lock()
	eLog.Infof("Updating enrichment database")
	if e.store == nil {
		e.store = make(map[keyType]*entry)
	}
	for _, m := range c.Metrics {
		h := e.Hash(*m)
		en := entry(m.Data)
		e.store[h] = &en
	}
	e.lock.Unlock()
}

func (e *Enrich) find(m skogul.Metric) *entry {
	if e.store == nil {
		return nil
	}
	nm := e.store[e.Hash(m)]
	if nm != nil {
		eLog.Tracef("Enrichment hit")
	} else {
		eLog.Tracef("Enrichment miss")
	}
	return nm
}

// Transform looks up a metric in the stored enrichment database (loading
// it on initial run) and adds metadata if a match is found.
func (e *Enrich) Transform(c *skogul.Container) error {
	e.lock.RLock()
	for _, m := range c.Metrics {
		hit := e.find(*m)
		if hit == nil {
			continue
		}
		for idx, field := range *hit {
			m.Metadata[idx] = field
		}
	}
	e.lock.RUnlock()
	return nil
}
