package appium

import (
	"time"

	webdriver "github.com/prawdadigital/web-driver"
)

// touchButton is the pointer "button" used for a touch contact (a single
// finger). The W3C actions protocol uses button 0 for the primary contact.
const touchButton = webdriver.LeftButton

func moveTo(duration time.Duration, x, y int) webdriver.PointerAction {
	return webdriver.PointerMoveAction(duration, webdriver.Point{X: x, Y: y}, webdriver.FromViewport)
}

// finger stores a single touch-pointer input source under inputID.
func (m *mobileWD) finger(inputID string, actions ...webdriver.PointerAction) {
	m.StorePointerActions(inputID, webdriver.TouchPointer, actions...)
}

func (m *mobileWD) Tap(x, y int) error {
	m.finger("finger",
		moveTo(0, x, y),
		webdriver.PointerDownAction(touchButton),
		webdriver.PointerPauseAction(50*time.Millisecond),
		webdriver.PointerUpAction(touchButton),
	)
	return m.PerformActions()
}

func (m *mobileWD) DoubleTap(x, y int) error {
	m.finger("finger",
		moveTo(0, x, y),
		webdriver.PointerDownAction(touchButton),
		webdriver.PointerPauseAction(40*time.Millisecond),
		webdriver.PointerUpAction(touchButton),
		webdriver.PointerPauseAction(40*time.Millisecond),
		webdriver.PointerDownAction(touchButton),
		webdriver.PointerPauseAction(40*time.Millisecond),
		webdriver.PointerUpAction(touchButton),
	)
	return m.PerformActions()
}

func (m *mobileWD) LongPress(x, y int, duration time.Duration) error {
	m.finger("finger",
		moveTo(0, x, y),
		webdriver.PointerDownAction(touchButton),
		webdriver.PointerPauseAction(duration),
		webdriver.PointerUpAction(touchButton),
	)
	return m.PerformActions()
}

func (m *mobileWD) Swipe(startX, startY, endX, endY int, duration time.Duration) error {
	m.finger("finger",
		moveTo(0, startX, startY),
		webdriver.PointerDownAction(touchButton),
		moveTo(duration, endX, endY),
		webdriver.PointerUpAction(touchButton),
	)
	return m.PerformActions()
}

// twoFinger performs a symmetric two-finger horizontal gesture: each finger
// starts at x±fromOffset and ends at x±toOffset, at height y.
func (m *mobileWD) twoFinger(x, y, fromOffset, toOffset int, duration time.Duration) error {
	m.finger("fingerA",
		moveTo(0, x-fromOffset, y),
		webdriver.PointerDownAction(touchButton),
		moveTo(duration, x-toOffset, y),
		webdriver.PointerUpAction(touchButton),
	)
	m.finger("fingerB",
		moveTo(0, x+fromOffset, y),
		webdriver.PointerDownAction(touchButton),
		moveTo(duration, x+toOffset, y),
		webdriver.PointerUpAction(touchButton),
	)
	return m.PerformActions()
}

func (m *mobileWD) Zoom(x, y, radius int, duration time.Duration) error {
	// Fingers start near the center and move apart.
	return m.twoFinger(x, y, 5, radius, duration)
}

func (m *mobileWD) Pinch(x, y, radius int, duration time.Duration) error {
	// Fingers start apart and move toward the center.
	return m.twoFinger(x, y, radius, 5, duration)
}
