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
	"sync"

	"github.com/telenornms/skogul"
	"github.com/telenornms/skogul/encoder"
)

var fileLog = skogul.Logger("sender", "file")

const newLineChar byte = 10

/*
File sender writes data to a file in various different fashions. Typical
use will be debugging (write to disk) and writing to a FIFO for example.
*/
type File struct {
	Path    string            `doc:"Absolute path to file to write"`
	Append  bool              `doc:"Whether to append to the file when starting. If false, will empty file before starting writes. Default: false"`
	Encoder skogul.EncoderRef `doc:"Which encoder to use. Defaults to JSON."`
	ok      bool
	once    sync.Once
	f       *os.File
	c       chan []byte
}

func (f *File) init() {
	fileLog.WithField("path", f.Path).Debug("Initializing File sender")

	var err error
	var file *os.File

	if f.Encoder.Name == "" {
		f.Encoder.E = encoder.JSON{}
	}
	if f.Encoder.E == nil {
		err := fmt.Errorf("No valid encoder specified")
		fileLog.WithError(err).Errorf("No valid encoder")
		f.ok = false
		return
	}
	// Open file for append-only if it already exists and config says to append
	if finfo, err := os.Stat(f.Path); !os.IsNotExist(err) && f.Append {
		fileLog.WithField("path", f.Path).Trace("File exists, let's open it for writing")
		file, err = os.OpenFile(f.Path, os.O_APPEND|os.O_WRONLY, finfo.Mode())
	} else {
		// Otherwise, create the file (which will truncate it if it already exists)
		fileLog.WithField("path", f.Path).Trace("Creating file since it doesn't exist or we don't want to append to it")
		file, err = os.Create(f.Path)
	}
	if err != nil {
		fileLog.WithField("path", f.Path).WithError(err).Errorf("Failed to open '%s'", f.Path)
		f.ok = false
		return
	}

	// startChan() handles closing the file as this function returns
	// and consequently would close the file
	f.f = file

	// Listening to a channel is blocking so we have
	// to start the channel listening in a goroutine
	// so that init() doesn't block.
	f.c = make(chan []byte, 50)
	go f.startChan()

	f.ok = true
}

func (f *File) startChan() {
	fileLog.Trace("Starting file writer channel")
	// Making sure we close the file if this function exits
	defer f.f.Close()
	for b := range f.c {
		written, err := f.f.Write(append(b, newLineChar))
		if err != nil {
			f.ok = false
			fileLog.WithField("path", f.Path).WithError(err).Errorf("Failed to write to file. Wrote %d of %d bytes", written, len(b))
		}
		f.f.Sync()
	}
	fileLog.WithField("path", f.Path).Warning("File writer chan closed, not handling any more writes!")
}

// Send receives a skogul container and writes it to file.
func (f *File) Send(c *skogul.Container) error {
	f.once.Do(func() {
		f.init()
	})

	if !f.ok {
		e := skogul.Error{Reason: "File sender not in OK state", Source: "file sender"}
		fileLog.WithError(e).Error("Failied to initialize file sender, or an error occurred in runtime")
		return e
	}

	b, err := f.Encoder.E.Encode(c)

	if err != nil {
		fileLog.WithError(err).Error("Failed to marshal container data to json")
		return err
	}

	f.c <- b

	return nil
}

// Verify checks that the configuration options are set appropriately
func (f *File) Verify() error {
	if f.Encoder.E == nil {
		return fmt.Errorf("No valid encoder specified")
	}
	if f.Path == "" {
		err := skogul.Error{Reason: "Path name for file sender missing", Source: "file sender"}
		fileLog.WithError(err).Error("Missing path to file for file sender")
		return err
	}
	return nil
}
