package transformer

import (
	"fmt"
	"strings"

	"github.com/telenornms/skogul"
)

type Unflatten struct {
	Separator string `doc:"Separator for path strings. Default fallback is ."`
}

// Transform created a container of metrics
func (u *Unflatten) Transform(c *skogul.Container) error {
	var err error
	for mi := range c.Metrics {
		c.Metrics[mi], err = u.convertValues(c.Metrics[mi])
	}

	return err
}

/*
Creates a nested structure from keys passed as a an array.

		The keys can have variable depth from 0 to n.

		Example:
		    Let's assume we have a list of keys looking like this:
		        [
		            "foo": "bar",           // This key has depth 0, root key is foo.
		            "bar.baz": "barbaz",    // This key has depth 1, root key is bar.
		            "foo.1.bar": "foo1bar"  // This key has depth 2, root key is foo.
		        ]

		        The output will be:
		        {
		            "foo": "bar",
		            "bar": {
		                "baz": "barbaz",
		            },
		            "foo": {
		                "1": {
		                    "bar": "foo1bar",
		                }
		            }
		        }

	    IMPORTANT!
	        If you have a set of keys looking like this
	            [
	                "bar.1.baz": "foo",
	                "bar.1.baz.1.bar": "bar.1.baz.1.bar",
	            ]

	        As you can notice there's two keys, one ending with "baz" and the seconds one has deeper structure, but both contain "baz" in the same place.
	        The algorithm will proceed and write "foo" to "bar.1.baz" key.
	        The next iteration will genererate an error because bar.1.baz is a string.
*/
func (u *Unflatten) convertValues(d *skogul.Metric) (*skogul.Metric, error) {
	var err error
	tmp := map[string]interface{}{}
	newMetric := skogul.Metric{
		Time:     d.Time,
		Metadata: d.Metadata,
		Data:     nil,
	}

	// Fallback to default separator
	if u.Separator == "" {
		u.Separator = "."
	}

	keys := make([]string, 0, len(d.Data))

	// Create a list of keys
	for k := range d.Data {
		keys = append(keys, k)
	}

	// Populate keys to make sure they exist before we try writing to them.
	for _, k := range keys {
		s := strings.Split(k, u.Separator)
		tmp[s[0]] = map[string]interface{}{}
	}

	/*
	   Recursively creates nested structure for each key, if a key needs one.
	      Notice that since we have already populated the map with first part of the key, we skip it.
	*/
	for _, k := range keys {
		s := strings.Split(k, u.Separator)
		tmp[s[0]], err = u.recursivelyCreateMap(tmp[s[0]].(map[string]interface{}), s[1:], d.Data[k], 0)
		if err != nil {
			return nil, err
		}
	}

	newMetric.Data = tmp

	return &newMetric, err
}

func (u *Unflatten) recursivelyCreateMap(root map[string]interface{}, keys []string, value interface{}, pos int) (interface{}, error) {
	var err error
	// Last key is reached put the value in it.
	if pos == len(keys) {
		return value, nil
	}

	if _, ok := root[keys[pos]]; ok { // Key already exists, continue creating key-value pairs in the same map until we reach length of keys.
		key, okk := root[keys[pos]].(map[string]interface{})
		if !okk {
			return root, fmt.Errorf("trying to write a value with type `%v` to a key without depth", key)
		}
		root[keys[pos]], err = u.recursivelyCreateMap(key, keys, value, pos+1)
	} else {
		root[keys[pos]], err = u.recursivelyCreateMap(map[string]interface{}{}, keys, value, pos+1)
	}

	return root, err
}
