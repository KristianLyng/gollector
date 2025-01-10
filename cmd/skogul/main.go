/*
 * skogul, using config file
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
cmd/skogul parses a json-based config file and starts skogul.
*/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"plugin"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/config"
	"github.com/telenornms/skogul/encoder"
	"github.com/telenornms/skogul/parser"
	"github.com/telenornms/skogul/receiver"
	"github.com/telenornms/skogul/sender"
	"github.com/telenornms/skogul/stats"
	"github.com/telenornms/skogul/transformer"
)

// versionNo gets set by passing the -X flag to ld like this
// go build -ldflags "-X main.versionNo=0.1.0" ./cmd/skogul
var versionNo string

var (
	ffile      = flag.String("f", "/etc/skogul/conf.d/", "Path to skogul config to read. Either a file or a directory of .json files.")
	fconfigDir = flag.String("d", "", "Path to skogul configuration files. Deprecated, use -f.")
	fhelp      = flag.Bool("help", false, "Print more help")
	fconf      = flag.Bool("show", false, "Print the parsed JSON config instead of starting")
	fman       = flag.Bool("make-man", false, "Output RST documentation suited for rst2man")
	flogformat = flag.String("logformat", "auto", "Log format (auto, json, default: auto)")
	floglevel  = flag.String("loglevel", "warn", "Minimum loglevel to display ([e]rror, [w]arn, [i]nfo, [d]ebug, [t]race/[v]erbose)")
	ftimestamp = flag.Bool("timestamp", true, "Include timestamp in log entries")
	fversion   = flag.Bool("version", false, "Print skogul version")
	fprofile   = flag.String("pprof", "", "Enable profiling over HTTP, value is http endpoint, e.g: localhost:6060")
	fplugins   = flag.String("experimental-plugins", "", "Comma-separated list of .so files to load as plugins. This is completely unsupported tech preview to get experience with it.")
)

// Console width :D
const helpWidth = 66

// prettyPrint is a relic that wraps lines in a table.
func prettyPrint(scheme string, desc string) {
	fmt.Printf("%11s |", scheme)
	fields := strings.Fields(desc)
	l := 0
	for _, w := range fields {
		if (l + len(w)) > helpWidth {
			l = 0
			fmt.Printf("\n%11s |", "")
		}
		fmt.Printf(" %s", w)
		l += len(w) + 1
	}
	fmt.Printf("\n")
}

// modhelp iterates over modules in a ModuleMap and prints short-hand help
func modhelp(name string, mp skogul.ModuleMap) {
	fmt.Printf("\n%s:\n", name)
	for idx, mod := range mp {
		if idx != mod.Name {
			continue
		}
		prettyPrint(idx, mod.Help)
	}
}

// help prints the regular command line usage, and lists all receivers and
// senders.
func help() {
	flag.Usage()
	modhelp("Senders", sender.Auto)
	modhelp("Receivers", receiver.Auto)
	modhelp("Transformers", transformer.Auto)
	modhelp("Parsers", parser.Auto)
	modhelp("Encoders", encoder.Auto)
	fmt.Println("\nSee the skogul(1) man page for details and usage. Some modules also have aliases.")
}

func printVersion() {
	skogulV := versionNo
	if len(versionNo) == 0 {
		// Since versionNo has to be explicitly set compile-time
		// provide a fallback in case it is not.
		skogulV = "unknown"
	}
	fmt.Println("Skogul", skogulV, "built using", runtime.Version())
}

func loadPlugins() error {
	if *fplugins == "" {
		return nil
	}
	toLoad := strings.Split(*fplugins, ",")
	for _, p := range toLoad {
		_, err := plugin.Open(p)
		if err != nil {
			return fmt.Errorf("plugin %s failed to load: %v", p, err)
		}
	}
	return nil
}

func main() {
	flag.Parse()

	skogul.ConfigureLogger(*floglevel, *ftimestamp, *flogformat)
	log := skogul.Logger("cmd", "main")
	err := loadPlugins()
	if err != nil {
		fmt.Printf("Unable to load plugins: %v\n", err)
		os.Exit(1)
	}

	if *fversion {
		printVersion()
		os.Exit(0)
	}
	if *fhelp {
		help()
		os.Exit(0)
	}
	if *fman {
		man()
		os.Exit(0)
	}

	configPath := ""

	if *fconfigDir != "" {
		log.Warnf("Using -d is deprecated, use -f instead - it has the exact same functionality.")
		configPath = *fconfigDir
	} else {
		configPath = *ffile
	}

	c, err := config.Path(configPath)
	if err != nil {
		log.WithError(err).Fatal("Failed to configure Skogul")
	}

	if *fconf {
		out, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			fmt.Println("Configuration failed to marshal:", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		os.Exit(0)
	}
	if *fprofile != "" {
		log.Warnf("Enabling profiling on %s", *fprofile)
		go func() {
			http.ListenAndServe(*fprofile, nil)
		}()
	}
	log.Info("Starting skogul")

	exitInt := 0
	var wg sync.WaitGroup
	for name, r := range c.Receivers {
		wg.Add(1)
		go func(name string, r *config.Receiver) {
			if inerr := r.Receiver.Start(); inerr != nil {
				exitInt = 1
				fmt.Printf("Receiver \"%s\" failed: %v\n", name, inerr)
			} else {
				fmt.Printf("Receiver \"%s\" returned successfully.\n", name)
			}
			wg.Done()
		}(name, r)
	}

	go startStats(c)

	wg.Wait()
	os.Exit(exitInt)
}

// startStats starts a forever-running loop which fetches
// stats from each module at the configured interval.
func startStats(c *config.Config) {
	statsLogger := skogul.Logger("main", "stats")

	ticker := time.NewTicker(stats.DefaultInterval)

	for range ticker.C {
		statsLogger.Trace("Gathering stats")
		for _, r := range c.Receivers {
			stats.Collect(r.Receiver)
		}
		for _, p := range c.Parsers {
			stats.Collect(p.Parser)
		}
		for _, t := range c.Transformers {
			stats.Collect(t.Transformer)
		}
		for _, s := range c.Senders {
			stats.Collect(s.Sender)
		}
	}
}
