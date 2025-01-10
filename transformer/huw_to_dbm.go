/*
 * Copyright (c) 2024 Telenor Norge AS
 * Author(s):
 *  - Hans Rafaelsen <hans.rafaelsen@telenor.no>
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
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
)

var uwdbmLog = skogul.Logger("transformer", "uw_to_dbm")

// Compute from mico-watt (uW) to desibel (DBM) for Huwaei routers
type HuWtoDBM struct {
	Source      string  `doc:"Data key to read from."`
	Destination string  `doc:"Data key to write to. Default 'dbm'"`
	Treshold    float64 `doc:"Treshold for selecting which formula to use. Default 200"`
	once        sync.Once
	err         error
}

// Transform
func (uwdbm *HuWtoDBM) Transform(c *skogul.Container) error {
	uwdbm.once.Do(func() {
		if uwdbm.Destination == "" {
			uwdbm.Destination = "dbm"
		}
		if uwdbm.Treshold == 0 {
			uwdbm.Treshold = 200
		}
	})
	for _, m := range c.Metrics {
		v, ok := m.Data[uwdbm.Source].(float64)
		if !ok {
			uwdbmLog.Log(logrus.ErrorLevel, "Value is not valid float '", uwdbm.Source,
				"': '", m.Data[uwdbm.Source], "'")
			uwdbm.err = errors.New(fmt.Sprintf("Value '%s' is not valid float", uwdbm.Source))
			continue
		}

		if v > uwdbm.Treshold {
			v = 10 * 1 / math.Log10(10) * math.Log10(0.1+v/1000)
		} else {
			v = 10 * 1 / math.Log10(10) * math.Log10(0.01+v/100)
		}
		// In case we don't compute a valid value, set it to -40 to indicate a broken link
		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = -40
		}
		m.Data[uwdbm.Destination] = v
	}
	return uwdbm.err
}

// Verify checks that the required variables are set
func (uwdbm *HuWtoDBM) Verify() error {
	if uwdbm.Source == "" {
		return skogul.MissingArgument("Source")
	}
	return nil
}
