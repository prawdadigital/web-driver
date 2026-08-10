package webdriver

import "time"

// Condition is a function passed to WebDriver.Wait that reports whether the
// awaited state has been reached.
type Condition func(wd WebDriver) (bool, error)

const (
	// DefaultWaitInterval is the default polling interval for WebDriver.Wait.
	DefaultWaitInterval = 100 * time.Millisecond

	// DefaultWaitTimeout is the default timeout for WebDriver.Wait.
	DefaultWaitTimeout = 60 * time.Second
)
