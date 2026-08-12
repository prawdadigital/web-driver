package webdriver

import "time"

// KeyPauseAction builds a KeyAction which pauses for the supplied duration.
func KeyPauseAction(duration time.Duration) KeyAction {
	return KeyAction{
		"type":     "pause",
		"duration": uint(duration / time.Millisecond),
	}
}

// KeyUpAction builds a KeyAction press.
func KeyUpAction(key string) KeyAction {
	return KeyAction{
		"type":  "keyUp",
		"value": key,
	}
}

// KeyDownAction builds a KeyAction which presses and holds
// the specified key.
func KeyDownAction(key string) KeyAction {
	return KeyAction{
		"type":  "keyDown",
		"value": key,
	}
}

// PointerPauseAction builds a PointerAction which pauses for the supplied duration.
func PointerPauseAction(duration time.Duration) PointerAction {
	return PointerAction{
		"type":     "pause",
		"duration": uint(duration / time.Millisecond),
	}
}

// PointerMoveAction builds a PointerAction which moves the pointer.
func PointerMoveAction(duration time.Duration, offset Point, origin PointerMoveOrigin) PointerAction {
	return PointerAction{
		"type":     "pointerMove",
		"duration": uint(duration / time.Millisecond),
		"origin":   origin,
		"x":        offset.X,
		"y":        offset.Y,
	}
}

// PointerUpAction builds an action which releases the specified pointer key.
func PointerUpAction(button MouseButton) PointerAction {
	return PointerAction{
		"type":   "pointerUp",
		"button": button,
	}
}

// PointerDownAction builds a PointerAction which presses
// and holds the specified pointer key.
func PointerDownAction(button MouseButton) PointerAction {
	return PointerAction{
		"type":   "pointerDown",
		"button": button,
	}
}

// WheelPauseAction builds a WheelAction which pauses for the supplied duration.
func WheelPauseAction(duration time.Duration) WheelAction {
	return WheelAction{
		"type":     "pause",
		"duration": uint(duration / time.Millisecond),
	}
}

// ScrollAction builds a WheelAction which scrolls by (deltaX, deltaY) pixels
// over the given duration. The scroll begins at the point (x, y) relative to
// origin, which is either FromViewport (the default when nil) or a WebElement
// to scroll from the center of.
func ScrollAction(duration time.Duration, origin interface{}, x, y, deltaX, deltaY int) WheelAction {
	if origin == nil {
		origin = FromViewport
	}
	return WheelAction{
		"type":     "scroll",
		"duration": uint(duration / time.Millisecond),
		"origin":   origin,
		"x":        x,
		"y":        y,
		"deltaX":   deltaX,
		"deltaY":   deltaY,
	}
}
