package webdriver

import "fmt"

// Error contains information about a failure of a command. See the table of
// these strings at https://www.w3.org/TR/webdriver/#handling-errors .
//
// This error type is only returned by servers that implement the W3C
// specification.
type Error struct {
	// Err contains a general error string provided by the server.
	Err string `json:"error"`
	// Message is a detailed, human-readable message specific to the failure.
	Message string `json:"message"`
	// Stacktrace may contain the server-side stacktrace where the error occurred.
	Stacktrace string `json:"stacktrace"`
	// HTTPCode is the HTTP status code returned by the server.
	HTTPCode int
	// LegacyCode is the "Response Status Code" defined in the legacy Selenium
	// WebDriver JSON wire protocol. This code is only produced by older
	// Selenium WebDriver versions, Chromedriver, and InternetExplorerDriver.
	LegacyCode int
}

// TODO(minusnine): Make Stacktrace more descriptive. Selenium emits a list of
// objects that enumerate various fields. This is not standard, though.

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}
