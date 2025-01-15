/*
 * skogul, logging utilities tests
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - hakon.solbjorg@telenor.com
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

package skogul_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
)

func TestLogLevelForConfiguredAsInfo(t *testing.T) {
	levelString := "info"
	levelType := logrus.InfoLevel
	level := skogul.GetLogLevelFromString(levelString)

	if level != levelType {
		t.Errorf("Failed to set log level to info, got %v", level)
	}
}

func TestLogLevelForConfiguredAsInfoShorthand(t *testing.T) {
	levelString := "i"
	levelType := logrus.InfoLevel
	level := skogul.GetLogLevelFromString(levelString)

	if level != levelType {
		t.Errorf("Failed to set log level to info, got %v", level)
	}
}

func TestLogLevelForInvalid(t *testing.T) {
	levelString := "abcdefgh"
	levelType := logrus.WarnLevel
	level := skogul.GetLogLevelFromString(levelString)

	if level != levelType {
		t.Errorf("Expected loglevel to default to warn, got %v", level)
	}
}
