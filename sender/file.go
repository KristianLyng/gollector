/*
 * skogul, file writer
 *
 * Copyright (c) 2019-2020 Telenor Norge AS
 * Author(s):
 *  - Håkon Solbjørg <Hakon.Solbjorg@telenor.com>
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

package sender

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

var fileLog = skogul.Logger("sender", "file")

const newLineChar byte = 10

/*
File sender writes data to a file in various different fashions. Typical
use will be debugging (write to disk) and writing to a FIFO for example.

Created file under path given by File option.

When SIGHUP signal is received File will be truncated.
In case SIGHUP is received and the file doesn't exists the file will be created.

When Append option is supplied, and this sender receives a SIGHUP data will
be appended to file, if the file exists.
*/
type File struct {
	Path    string            `doc:"Absolute path to file to write. DEPRECATED - replaced by option File (to keep options more consistent across modules)."`
	File    string            `doc:"Absolute path to file to write to."`
	Append  bool              `doc:"When sighup is received data will be appended to file. When this option is false, the file will be truncated before writing."`
	Encoder skogul.EncoderRef `doc:"Which encoder to use. Defaults to JSON."`
	ok      bool
	once    sync.Once
	f       *os.File
	c       chan []byte
	sighup  chan os.Signal
}

func (f *File) init() {
	var err error
	var file *os.File

	// To be removed
	if f.File == "" {
		f.File = f.Path
	}

	fileLog.WithField("path", f.File).Debug("Initializing File sender")

	if f.Encoder.Name == "" {
		fileLog.Info("No Encoder specified, using default: json")
		f.Encoder.E = encoder.JSON{}
	}

	if f.Encoder.E == nil {
		fileLog.Errorf("No valid encoder specified")
		f.ok = false
		return
	}

	// Open file for append-only if it already exists and config says to append
	if finfo, statErr := os.Stat(f.File); !os.IsNotExist(statErr) && f.Append {
		fileLog.WithField("path", f.File).Trace("File exists, let's open it for writing")
		file, err = os.OpenFile(f.File, os.O_APPEND|os.O_WRONLY, finfo.Mode())
	} else {
		// Otherwise, create the file (which will truncate it if it already exists)
		fileLog.WithField("path", f.File).Trace("Creating file since it doesn't exist or we don't want to append to it")
		file, err = os.Create(f.File)
	}

	if err != nil {
		fileLog.WithField("path", f.File).WithError(err).Errorf("Failed to open '%s'", f.File)
		f.ok = false
		return
	}

	// startChan() handles closing the file as this function returns
	// and consequently would close the file
	f.f = file

	f.sighup = make(chan os.Signal, 1)

	// Listening to a channel is blocking so we have
	// to start the channel listening in a goroutine
	// so that init() doesn't block.
	f.c = make(chan []byte, 50)
	go f.startChan()

	f.ok = true
}

func (f *File) startChan() {
	fileLog.Trace("Starting file writer routine")
	defer f.f.Close()

	signal.Notify(f.sighup, syscall.SIGHUP)
	for {
		select {
		case b := <-f.c:
			written, err := f.f.Write(append(b, newLineChar))
			if err != nil {
				f.ok = false
				fileLog.WithField("path", f.File).WithError(err).Errorf("Failed to write to file. Wrote %d of %d bytes", written, len(b))
			}
			f.f.Sync()
		case <-f.sighup: // Hung up
			var file *os.File
			var err error

			if f.Append {
				if finfo, statErr := os.Stat(f.File); !os.IsNotExist(statErr) && f.Append {
					fileLog.WithField("path", f.File).Trace("File exists, let's open it for writing")
					file, err = os.OpenFile(f.File, os.O_APPEND|os.O_WRONLY, finfo.Mode())
				}
			} else {
				file, err = os.Create(f.File)
			}

			if err != nil {
				f.ok = false
				// Prevent handle leak
				f.f.Close()
				file.Close()
				fileLog.WithField("path", file).WithError(err).Errorf("Error with file")
				continue
			}
			// Prevent handle leak
			// f.f is poiting to the previous file.
			f.f.Close()

			// Override file handler
			f.f = file
			f.ok = true
		}
	}
}

// Send receives a skogul container and writes it to file.
func (f *File) Send(c *skogul.Container) error {
	f.once.Do(func() {
		f.init()
	})

	if !f.ok {
		return fmt.Errorf("failed to initialize file sender, or an error occurred in runtime")
	}

	b, err := f.Encoder.E.Encode(c)
	if err != nil {
		return fmt.Errorf("file sender unable to encode: %w", err)
	}

	f.c <- b

	return nil
}

func (f *File) Deprecated() error {
	if f.Path != "" {
		return fmt.Errorf("config option Path is replaced by option File, Path will be removed in future versions.")
	}
	return nil
}

// Verify checks that the configuration options are set appropriately
func (f *File) Verify() error {
	if f.Encoder.E == nil {
		return skogul.MissingArgument("Encoder")
	}
	if f.File != "" && f.Path != "" {
		return fmt.Errorf("both File and deprecated option Path specified for file sender - which to use?")
	}

	if f.File == "" && f.Path == "" {
		return skogul.MissingArgument("File")
	}
	return nil
}
