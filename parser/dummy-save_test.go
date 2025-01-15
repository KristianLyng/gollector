/*
 * skogul, test json parser
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

package parser_test

import (
	"bytes"
	"os"
	"path"
	"testing"

	"github.com/telenornms/skogul/parser"
)

func TestDummyStore(t *testing.T) {
	b := []byte(`some bytes`)
	p := parser.DummyStore{}
	filename := "skogul-dummy-store-ok"
	tmpdir := os.TempDir()
	path := path.Join(tmpdir, filename)
	p.File = path

	// Ensure file does not exist already
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// Already exists..
		// Let's assume it's safe to remove ?
		os.Remove(path)
	}

	c, err := p.Parse(b)
	if err != nil {
		t.Errorf("err! %v", err)
	}
	if c == nil {
		t.Errorf("c == nil!")
	}
	newb, err := os.ReadFile(path)
	if bytes.Compare(newb, b) != 0 {
		t.Errorf("newb != b! %v != %v", newb, b)
	}
	if err != nil {
		t.Errorf("err! %v", err)
	}
}

func TestDummyStore_no(t *testing.T) {
	b := []byte(`some bytes`)
	p := parser.DummyStore{}
	filename := "skogul-dummy-store-bad/no/such/directory"
	tmpdir := os.TempDir()
	path := path.Join(tmpdir, filename)
	p.File = path

	// Ensure file does not exist already
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// Already exists..
		// Let's assume it's safe to remove ?
		os.Remove(path)
	}

	c, err := p.Parse(b)
	if err == nil {
		t.Errorf("no err!")
	}
	if c != nil {
		t.Errorf("c != nil!")
	}
}
