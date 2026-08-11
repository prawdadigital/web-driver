// Package log provides logging-related configuration types and constants
// used to request and describe WebDriver log output. Its Type and Level
// constants populate a Capabilities map (keyed by CapabilitiesKey) in the
// session capabilities, and Message describes a single log entry returned by
// the driver.
package log

import "time"

// Type represents a component capable of logging.
type Type string

// The valid log types.
const (
	Server      Type = "server"
	Browser     Type = "browser"
	Client      Type = "client"
	Driver      Type = "driver"
	Performance Type = "performance"
	Profiler    Type = "profiler"
)

// Level represents a logging level of different components in the browser,
// the driver, or any intermediary WebDriver servers.
//
// See the documentation of each driver for what browser specific logging
// components are available.
type Level string

// The valid log levels.
const (
	Off     Level = "OFF"
	Severe  Level = "SEVERE"
	Warning Level = "WARNING"
	Info    Level = "INFO"
	Debug   Level = "DEBUG"
	All     Level = "ALL"
)

// CapabilitiesKey is the key for the logging preferences entry in the JSON
// structure representing WebDriver capabilities.
//
// Note that the W3C spec does not include logging right now, and starting with
// Chrome 75, "loggingPrefs" has been changed to "goog:loggingPrefs"
const CapabilitiesKey = "goog:loggingPrefs"

// Capabilities is the map to include in the WebDriver capabilities structure
// to configure logging.
type Capabilities map[Type]Level

// Message is a log message returned from the Log method.
type Message struct {
	// Timestamp is the time at which the entry was recorded.
	Timestamp time.Time
	// Level is the severity of the entry.
	Level Level
	// Message is the human-readable text of the entry.
	Message string
}
