/*
 * skogul, logging utilities
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

package skogul

import (
	"io"
	"strings"

	"github.com/sirupsen/logrus"
)

var skogulLogger = logrus.New()

// ConfigureLogger sets up the logger based on calling parameters
func ConfigureLogger(requestedLoglevel string, logtimestamp bool, logFormat string) {
	// Disable output from the root logger; all logging is delegated through hooks
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetOutput(io.Discard)

	loglevel := GetLogLevelFromString(requestedLoglevel)
	skogulLogger.SetLevel(loglevel)

	if strings.ToLower(logFormat) == "json" {
		skogulLogger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		skogulLogger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: !logtimestamp,
		})
	}

	copyHook := LoggerCopyHook{
		Writer: skogulLogger,
	}

	// Add a hook to a new logger which outputs the logs to stdout
	logrus.AddHook(&copyHook)
}

// GetLogLevelFromString returns the matching logrus.Level from a string
func GetLogLevelFromString(requestedLevel string) logrus.Level {
	switch strings.ToLower(requestedLevel) {
	case "e", "error":
		return logrus.ErrorLevel
	case "w", "warn":
		return logrus.WarnLevel
	case "i", "info":
		return logrus.InfoLevel
	case "d", "debug":
		return logrus.DebugLevel
	case "v", "verbose", "t", "trace":
		return logrus.TraceLevel
	default:
		{
			skogulLogger.Warnf("Invalid loglevel '%s', defaulting to 'warn'", requestedLevel)
			return logrus.WarnLevel
		}
	}
}

// LoggerCopyHook is simply a wrapper around a logrus logger
type LoggerCopyHook struct {
	Writer *logrus.Logger
}

// Fire logs the log entry onto a copied logger to stdout
func (l *LoggerCopyHook) Fire(entry *logrus.Entry) error {
	skogulLogger.WithFields(entry.Data).Log(entry.Level, entry.Message)
	return nil
}

// Levels returns the levels this hook will support
func (l *LoggerCopyHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
