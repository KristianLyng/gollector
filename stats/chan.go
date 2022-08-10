/*
 * skogul, internal stats
 *
 * Copyright (c) 2021 Telenor Norge AS
 * Author(s):
 *  - Håkon Solbjørg <hakon.solbjorg@telenor.no>
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
stats supports Skogul by adding internal statistics

	about how Skogul is doing. The stats are very generic
	and are sent as regular `skogul.Metric`s. By configuring
	the stats receiver all stats will be picked up and sent
	through skogul as any other metrics.
*/
package stats

import (
	"context"
	"time"

	"github.com/telenornms/skogul"
)

// DefaultInterval is the default interval used for sending stats.
var DefaultInterval = time.Second * 10

// DefaultChanSize is the default size of stats channels.
var DefaultChanSize = 100

var statsLog = skogul.Logger("stats", "chan")

// Chan is a channel which accepts skogul statistic as a skogul.Metric
// By configuring the stats receiver, this channel is drained and sent on to
// the specified handler.
var Chan chan *skogul.Metric

// DrainCtx and CancelDrain are the context and cancel functions
// for the automatically created stats.Chan.
// If a skogul stats receiver is configured, DrainCancel MUST be called
// so that statistics are not discarded.
var DrainCtx, CancelDrain = context.WithCancel(context.Background())

// init makes sure that the skogul stats channel exists at all times.
// Furthermore, it starts a goroutine to empty the channel in the case
// that the stats receiver is not configured, in which case the chan
// would end up blocking after it is filled.
func init() {
	// Create stats.Chan so we don't have components blocking on it
	if Chan == nil {
		Chan = make(chan *skogul.Metric, DefaultChanSize)
	}
	go DrainStats(DrainCtx)
}

// DrainStats drains all statistics on the stats channel.
// If the passed context is cancelled it will stop draining the channel
// so that a configured stats-receiver can listen on the channel.
func DrainStats(ctx context.Context) {
	statsLog.Debug("Starting stats drain. All stats are being dropped.")
	for {
		select {
		case <-Chan:
			continue
		case <-ctx.Done():
			statsLog.Debug("Stopping stats drain. Stats are being consumed.")
			return
		}
	}
}

// Collect stats for a skogul.Module if it has stats.
// If it doesn't have stats it does nothing.
func Collect(m interface{}) {
	module, ok := m.(skogul.Stats)
	if !ok {
		// We haven't gotten stats for this module, silently return
		return
	}

	Chan <- module.GetStats()
}
