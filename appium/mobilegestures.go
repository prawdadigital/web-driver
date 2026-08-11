package appium

import (
	"time"

	webdriver "github.com/prawdadigital/web-driver"
)

// ExecuteMobile invokes an Appium "mobile:" extension command with the given
// options map, i.e. ExecuteScript("mobile: <command>", [options]).
func (m *mobileWD) ExecuteMobile(command string, options map[string]interface{}) (interface{}, error) {
	if options == nil {
		options = map[string]interface{}{}
	}
	return m.ExecuteScript("mobile: "+command, []interface{}{options})
}

// areaOptions converts a screen rectangle to the left/top/width/height options
// shared by the UiAutomator2 area-based gestures.
func areaOptions(area webdriver.Rect) map[string]interface{} {
	return map[string]interface{}{
		"left":   area.X,
		"top":    area.Y,
		"width":  area.Width,
		"height": area.Height,
	}
}

func (m *mobileWD) SwipeGesture(area webdriver.Rect, direction string, percent float64) error {
	opts := areaOptions(area)
	opts["direction"] = direction
	opts["percent"] = percent
	_, err := m.ExecuteMobile("swipeGesture", opts)
	return err
}

func (m *mobileWD) ScrollGesture(area webdriver.Rect, direction string, percent float64) error {
	opts := areaOptions(area)
	opts["direction"] = direction
	opts["percent"] = percent
	_, err := m.ExecuteMobile("scrollGesture", opts)
	return err
}

func (m *mobileWD) PinchOpenGesture(area webdriver.Rect, percent float64) error {
	opts := areaOptions(area)
	opts["percent"] = percent
	_, err := m.ExecuteMobile("pinchOpenGesture", opts)
	return err
}

func (m *mobileWD) PinchCloseGesture(area webdriver.Rect, percent float64) error {
	opts := areaOptions(area)
	opts["percent"] = percent
	_, err := m.ExecuteMobile("pinchCloseGesture", opts)
	return err
}

func (m *mobileWD) LongClickGesture(x, y int, duration time.Duration) error {
	_, err := m.ExecuteMobile("longClickGesture", map[string]interface{}{
		"x":        x,
		"y":        y,
		"duration": int(duration / time.Millisecond),
	})
	return err
}
