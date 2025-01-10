/*
 * skogul, logrus receiver
 *
 * Copyright (c) 2019 Telenor Norge AS
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

package receiver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
)

// LogrusLog configures the logrus log receiver
type LogrusLog struct {
	Loglevel string
	Handler  skogul.HandlerRef
}

var logrusLogLogger = logrus.New()

var logMetadataFields *[]string

func (lg *LogrusLog) configureLogger() error {
	// Set up internal logger so we don't cause recursive logging in case we log errors
	logrusLogLogger.SetOutput(io.Writer(os.Stdout))
	logrusLogLogger.WithFields(logrus.Fields{
		"category": "receiver",
		"receiver": "logrus",
	})
	loglevel := skogul.GetLogLevelFromString(lg.Loglevel)
	logrusLogLogger.SetLevel(loglevel)

	metadataFields := []string{"category", "level"}
	a := make([]string, 0)
	logMetadataFields = &a
	for _, field := range metadataFields {
		*logMetadataFields = append(*logMetadataFields, field)
	}
	return nil
}

// Write logrus logs as skogul.Containers to a handler
func (lg *LogrusLog) Write(bytes []byte) (int, error) {
	metadata := make(map[string]interface{})

	var data map[string]interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		logrusLogLogger.Error("Failed to unmarshal logrus log for sending it to log receiver")
		return 0, err
	}

	// Extract metadata fields from data
	for _, field := range *logMetadataFields {
		metadata[field] = data[field]
	}

	timestampData := data["time"].(string)
	timestamp, err := time.Parse(time.RFC3339, timestampData)
	if err != nil {
		// @ToDo: Do we want to error out if this fails?
		logrusLogLogger.Errorf("Failed to parse timestamp '%s' '%s'\n", "time", timestampData)
	}

	m := skogul.Metric{
		Time:     &timestamp,
		Metadata: metadata,
		Data:     data,
	}
	c := skogul.Container{
		Metrics: []*skogul.Metric{&m},
	}

	lg.Handler.H.TransformAndSend(&c)
	return len(bytes), nil
}

// Start initializes the logger and sets up required facilities
func (lg *LogrusLog) Start() error {
	logrusLogLogger.Debug("Starting logger")
	lg.configureLogger()

	h := LogrusSkogulHook{
		Writer: lg,
	}

	logrus.AddHook(&h)
	return nil
}

// LogrusSkogulHook is a logrus.Hook made for skogul
type LogrusSkogulHook struct {
	Writer *LogrusLog
	A      string
}

// Fire implements the logrus.Hook interface and handles each log entry
func (hook *LogrusSkogulHook) Fire(entry *logrus.Entry) error {
	entr := entry.Data

	entr["message"] = entry.Message
	entr["time"] = entry.Time
	entr["level"] = entry.Level

	data, err := json.Marshal(entr)
	if err != nil {
		fmt.Println("Failed to convert log entry to json")
		return err
	}

	_, err = (*hook.Writer).Write([]byte(data))
	if err != nil {
		fmt.Println("Write to handler failed", err)
		return err
	}

	return nil
}

// Levels returns the log levels the hook should care about
func (hook *LogrusSkogulHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
