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
