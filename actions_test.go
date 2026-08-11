package webdriver

import (
	"testing"
	"time"
)

func TestKeyActionConstructors(t *testing.T) {
	if a := KeyDownAction("a"); a["type"] != "keyDown" || a["value"] != "a" {
		t.Errorf("KeyDownAction = %v", a)
	}
	if a := KeyUpAction("b"); a["type"] != "keyUp" || a["value"] != "b" {
		t.Errorf("KeyUpAction = %v", a)
	}
	a := KeyPauseAction(250 * time.Millisecond)
	if a["type"] != "pause" || a["duration"].(uint) != 250 {
		t.Errorf("KeyPauseAction = %v, want pause/250", a)
	}
}

func TestPointerActionConstructors(t *testing.T) {
	if a := PointerDownAction(LeftButton); a["type"] != "pointerDown" || a["button"] != LeftButton {
		t.Errorf("PointerDownAction = %v", a)
	}
	if a := PointerUpAction(RightButton); a["type"] != "pointerUp" || a["button"] != RightButton {
		t.Errorf("PointerUpAction = %v", a)
	}
	if a := PointerPauseAction(100 * time.Millisecond); a["type"] != "pause" || a["duration"].(uint) != 100 {
		t.Errorf("PointerPauseAction = %v", a)
	}

	move := PointerMoveAction(50*time.Millisecond, Point{X: 3, Y: 4}, FromViewport)
	if move["type"] != "pointerMove" || move["duration"].(uint) != 50 {
		t.Errorf("PointerMoveAction type/duration = %v", move)
	}
	if move["origin"] != FromViewport || move["x"] != 3 || move["y"] != 4 {
		t.Errorf("PointerMoveAction origin/x/y = %v", move)
	}
}

func TestScrollAction(t *testing.T) {
	a := ScrollAction(200*time.Millisecond, FromViewport, 1, 2, 10, 20)
	if a["type"] != "scroll" || a["duration"].(uint) != 200 {
		t.Errorf("ScrollAction type/duration = %v", a)
	}
	if a["x"] != 1 || a["y"] != 2 || a["deltaX"] != 10 || a["deltaY"] != 20 {
		t.Errorf("ScrollAction coords = %v", a)
	}
	if a["origin"] != FromViewport {
		t.Errorf("ScrollAction origin = %v, want viewport", a["origin"])
	}

	// A nil origin defaults to the viewport.
	if a := ScrollAction(0, nil, 0, 0, 0, 5); a["origin"] != FromViewport {
		t.Errorf("ScrollAction(nil origin) origin = %v, want viewport", a["origin"])
	}
}

func TestWheelPauseAction(t *testing.T) {
	a := WheelPauseAction(75 * time.Millisecond)
	if a["type"] != "pause" || a["duration"].(uint) != 75 {
		t.Errorf("WheelPauseAction = %v, want pause/75", a)
	}
}
