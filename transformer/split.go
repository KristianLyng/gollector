/*
 * skogul, split transformer
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
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

package transformer

import (
	"fmt"

	"github.com/telenornms/skogul"
)

// Split is the configuration for the split transformer
type Split struct {
	Field        []string `doc:"Split into multiple metrics based on this field (each field denotes the path to a nested object element)."`
	MetadataName string   `doc:"If specified, the index of the array being split will be stored as the named metadata field. E.g.: The first element will have a metadata field matching MetadataName with a value of 0, the second will have a 1, and so on. If left blank, the array index will be discarded."`
	Fail         bool     `doc:"Fail the transformer entirely if split is unsuccsessful on a metric container. This will prevent successive transformers from working."`
}

type DictSplit struct {
	Field        []string `doc:"Split into multiple metrics based on this field (each field denotes the path to a nested object element). To split on the top-element, leave this blank."`
	MetadataName string   `doc:"If specified, the key of the dictionary being split will be stored as the named metadata field. E.g.: If the data is indexed by interface name, setting MetadataName to if_name will populate if_name with the ... name of the interface. If left blank, the key will be discarded."`
	Fail         bool     `doc:"Fail the transformer entirely if split is unsuccsessful on a metric container. This will prevent successive transformers from working."`
}

// Transform splits the container assuming it has an array to split
func (split *Split) Transform(c *skogul.Container) error {
	metrics := c.Metrics

	if split.Field != nil {
		splitMetrics, err := split.splitMetricsByObjectKey(&metrics)
		if err == nil {
			c.Metrics = splitMetrics
		} else if split.Fail {
			return fmt.Errorf("failed to split metrics: %v", err)
		}
	}

	return nil
}

// splitMetricsByObjectKey splits the metrics into multiple metrics based on a key in a list of sub-metrics
func (split *Split) splitMetricsByObjectKey(metrics *[]*skogul.Metric) ([]*skogul.Metric, error) {
	origMetrics := *metrics
	var newMetrics []*skogul.Metric

	for mi := range origMetrics {
		splitObj, err := skogul.ExtractNestedObject(origMetrics[mi].Data, split.Field)
		if err != nil {
			if !split.Fail {
				newMetrics = append(newMetrics, origMetrics[mi])
				continue
			}
			return nil, fmt.Errorf("failed to extract nested obj '%v' from '%v' to string/interface map", split.Field, origMetrics[mi].Data)
		}

		metrics, ok := splitObj[split.Field[len(split.Field)-1]].([]interface{})

		if !ok {
			if !split.Fail {
				newMetrics = append(newMetrics, origMetrics[mi])
				continue
			}
			return nil, fmt.Errorf("failed to cast '%v' to string/interface map on '%s'", origMetrics[mi].Data, split.Field[0])
		}

		for idx, obj := range metrics {
			// Create a new metrics object as a copy of the original one, then reassign the data field
			metricsData, ok := obj.(map[string]interface{})

			if !ok {
				if !split.Fail {
					newMetrics = append(newMetrics, origMetrics[mi])
					continue
				}
				return nil, fmt.Errorf("failed to cast '%v' to string/interface map", obj)
			}

			newMetric := *origMetrics[mi]
			newMetric.Data = metricsData
			newMetric.Metadata = make(map[string]interface{})

			for key, val := range origMetrics[mi].Metadata {
				newMetric.Metadata[key] = val
			}
			if split.MetadataName != "" {
				newMetric.Metadata[split.MetadataName] = idx
			}

			newMetrics = append(newMetrics, &newMetric)
		}
	}

	return newMetrics, nil
}

// Transform splits the container assuming it has a dictionary to split
func (split *DictSplit) Transform(c *skogul.Container) error {
	metrics := c.Metrics

	if split.Field != nil {
		splitMetrics, err := split.splitMetricsByObjectKey(&metrics)
		if err == nil {
			c.Metrics = splitMetrics
		} else if split.Fail {
			return fmt.Errorf("failed to split metrics: %v", err)
		}
	}

	return nil
}

// splitMetricsByObjectKey splits the metrics into multiple metrics based on a key in a list of sub-metrics
func (split *DictSplit) splitMetricsByObjectKey(metrics *[]*skogul.Metric) ([]*skogul.Metric, error) {
	origMetrics := *metrics
	var newMetrics []*skogul.Metric

	for mi := range origMetrics {
		var metrics map[string]interface{}
		if len(split.Field) > 0 {
			splitObj, err := skogul.ExtractNestedObject(origMetrics[mi].Data, split.Field)
			if err != nil {
				if !split.Fail {
					newMetrics = append(newMetrics, origMetrics[mi])
					continue
				}
				return nil, fmt.Errorf("failed to extract nested obj '%v' from '%v' to string/interface map", split.Field, origMetrics[mi].Data)
			}
			var ok bool
			metrics, ok = splitObj[split.Field[len(split.Field)-1]].(map[string]interface{})
			if !ok {
				if !split.Fail {
					newMetrics = append(newMetrics, origMetrics[mi])
					continue
				}
				return nil, fmt.Errorf("failed to cast '%v' to string/interface map on '%s'", origMetrics[mi].Data, split.Field[0])
			}
		} else {
			metrics = origMetrics[mi].Data
		}

		for idx, obj := range metrics {
			// Create a new metrics object as a copy of the original one, then reassign the data field
			metricsData, ok := obj.(map[string]interface{})

			if !ok {
				if !split.Fail {
					newMetrics = append(newMetrics, origMetrics[mi])
					continue
				}
				return nil, fmt.Errorf("failed to cast '%v' to string/interface map", obj)
			}

			newMetric := *origMetrics[mi]
			newMetric.Data = metricsData
			newMetric.Metadata = make(map[string]interface{})

			for key, val := range origMetrics[mi].Metadata {
				newMetric.Metadata[key] = val
			}
			if split.MetadataName != "" {
				newMetric.Metadata[split.MetadataName] = idx
			}

			newMetrics = append(newMetrics, &newMetric)
		}
	}

	return newMetrics, nil
}
