package selenium

import (
	webdriver "github.com/prawdadigital/web-driver"
	"github.com/prawdadigital/web-driver/remote"
)

// NewRemote starts a new WebDriver session against an already-running server at
// urlPrefix and returns the driver. It is a convenience wrapper around
// remote.NewRemote for the common browser case, so callers can use a single
// package alongside NewSeleniumService.
func NewRemote(capabilities webdriver.Capabilities, urlPrefix string) (webdriver.WebDriver, error) {
	return remote.NewRemote(capabilities, urlPrefix)
}

// DeleteSession deletes an existing session at the given server by its ID.
func DeleteSession(urlPrefix, id string) error {
	return remote.DeleteSession(urlPrefix, id)
}

// SetDebug enables or disables verbose logging of the WebDriver protocol.
func SetDebug(debug bool) {
	remote.SetDebug(debug)
}
