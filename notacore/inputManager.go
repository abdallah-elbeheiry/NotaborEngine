package notacore

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Input int

const (
	KeySpace Input = iota
	KeyApostrophe
	KeyComma
	KeyMinus
	KeyPeriod
	KeySlash

	Key0
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9

	KeySemicolon
	KeyEqual

	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ

	KeyLeftBracket
	KeyBackslash
	KeyRightBracket
	KeyGraveAccent

	KeyEscape
	KeyEnter
	KeyTab
	KeyBackspace
	KeyInsert
	KeyDelete

	KeyRight
	KeyLeft
	KeyDown
	KeyUp

	KeyPageUp
	KeyPageDown
	KeyHome
	KeyEnd

	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause

	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25

	KeyKP0
	KeyKP1
	KeyKP2
	KeyKP3
	KeyKP4
	KeyKP5
	KeyKP6
	KeyKP7
	KeyKP8
	KeyKP9
	KeyKPDecimal
	KeyKPDivide
	KeyKPMultiply
	KeyKPSubtract
	KeyKPAdd
	KeyKPEnter
	KeyKPEqual

	KeyLeftShift
	KeyLeftControl
	KeyLeftAlt
	KeyLeftSuper
	KeyRightShift
	KeyRightControl
	KeyRightAlt
	KeyRightSuper
	KeyLeftCommand  //does not work yet
	KeyRightCommand // does not work yet
	KeyOptionLeft   //does not work yet
	KeyOptionRight  //does not work yet
	KeyFn           //does not work yet
	KeyMenu

	KeyMediaPlayPause //does not work yet
	KeyMediaStop      //does not work yet
	KeyMediaNext      //does not work yet
	KeyMediaPrev      //does not work yet
	KeyVolumeUp       //does not work yet
	KeyVolumeDown     //does not work yet
	KeyBrightnessUp   //does not work yet
	KeyBrightnessDown //does not work yet

	MouseLeft
	MouseRight
	MouseMiddle
	MouseButton4
	MouseButton5
	MouseButton6
	MouseButton7
	MouseButton8

	PadA
	PadB
	PadX
	PadY
	PadLB
	PadRB
	PadBack
	PadStart
	PadGuide
	PadLeftThumb
	PadRightThumb
	PadDpadUp
	PadDpadRight
	PadDpadDown
	PadDpadLeft

	PadAxisLeftX
	PadAxisLeftY
	PadAxisRightX
	PadAxisRightY
	PadAxisLeftTrigger
	PadAxisRightTrigger

	//TODO: what is below is not working yet, OS based API needed

	Touch1
	Touch2
	Touch3
	Touch4
	Touch5
	ForceTouch

	SwipeUp1
	SwipeDown1
	SwipeLeft1
	SwipeRight1
	SwipeUp2
	SwipeDown2
	SwipeLeft2
	SwipeRight2
	SwipeUp3
	SwipeDown3
	SwipeLeft3
	SwipeRight3
	SwipeUp4
	SwipeDown4
	SwipeLeft4
	SwipeRight4
	SwipeUp5
	SwipeDown5
	SwipeLeft5
	SwipeRight5

	PinchIn
	PinchOut
	RotateCW
	RotateCCW

	MobileTouchBegin
	MobileTouchMove
	MobileTouchEnd
	MobileTouchCancel
	MobileBack
	MobileHome
	MobileVolumeUp
	MobileVolumeDown
	AccelerometerX
	AccelerometerY
	AccelerometerZ
	GyroX
	GyroY
	GyroZ
	OrientationPitch
	OrientationYaw
	OrientationRoll
)

var glfwKeyMap = map[Input]glfw.Key{
	KeySpace:        glfw.KeySpace,
	KeyApostrophe:   glfw.KeyApostrophe,
	KeyComma:        glfw.KeyComma,
	KeyMinus:        glfw.KeyMinus,
	KeyPeriod:       glfw.KeyPeriod,
	KeySlash:        glfw.KeySlash,
	Key0:            glfw.Key0,
	Key1:            glfw.Key1,
	Key2:            glfw.Key2,
	Key3:            glfw.Key3,
	Key4:            glfw.Key4,
	Key5:            glfw.Key5,
	Key6:            glfw.Key6,
	Key7:            glfw.Key7,
	Key8:            glfw.Key8,
	Key9:            glfw.Key9,
	KeySemicolon:    glfw.KeySemicolon,
	KeyEqual:        glfw.KeyEqual,
	KeyA:            glfw.KeyA,
	KeyB:            glfw.KeyB,
	KeyC:            glfw.KeyC,
	KeyD:            glfw.KeyD,
	KeyE:            glfw.KeyE,
	KeyF:            glfw.KeyF,
	KeyG:            glfw.KeyG,
	KeyH:            glfw.KeyH,
	KeyI:            glfw.KeyI,
	KeyJ:            glfw.KeyJ,
	KeyK:            glfw.KeyK,
	KeyL:            glfw.KeyL,
	KeyM:            glfw.KeyM,
	KeyN:            glfw.KeyN,
	KeyO:            glfw.KeyO,
	KeyP:            glfw.KeyP,
	KeyQ:            glfw.KeyQ,
	KeyR:            glfw.KeyR,
	KeyS:            glfw.KeyS,
	KeyT:            glfw.KeyT,
	KeyU:            glfw.KeyU,
	KeyV:            glfw.KeyV,
	KeyW:            glfw.KeyW,
	KeyX:            glfw.KeyX,
	KeyY:            glfw.KeyY,
	KeyZ:            glfw.KeyZ,
	KeyLeftBracket:  glfw.KeyLeftBracket,
	KeyBackslash:    glfw.KeyBackslash,
	KeyRightBracket: glfw.KeyRightBracket,
	KeyGraveAccent:  glfw.KeyGraveAccent,

	KeyEscape:      glfw.KeyEscape,
	KeyEnter:       glfw.KeyEnter,
	KeyTab:         glfw.KeyTab,
	KeyBackspace:   glfw.KeyBackspace,
	KeyInsert:      glfw.KeyInsert,
	KeyDelete:      glfw.KeyDelete,
	KeyRight:       glfw.KeyRight,
	KeyLeft:        glfw.KeyLeft,
	KeyDown:        glfw.KeyDown,
	KeyUp:          glfw.KeyUp,
	KeyPageUp:      glfw.KeyPageUp,
	KeyPageDown:    glfw.KeyPageDown,
	KeyHome:        glfw.KeyHome,
	KeyEnd:         glfw.KeyEnd,
	KeyCapsLock:    glfw.KeyCapsLock,
	KeyScrollLock:  glfw.KeyScrollLock,
	KeyNumLock:     glfw.KeyNumLock,
	KeyPrintScreen: glfw.KeyPrintScreen,
	KeyPause:       glfw.KeyPause,

	KeyF1:  glfw.KeyF1,
	KeyF2:  glfw.KeyF2,
	KeyF3:  glfw.KeyF3,
	KeyF4:  glfw.KeyF4,
	KeyF5:  glfw.KeyF5,
	KeyF6:  glfw.KeyF6,
	KeyF7:  glfw.KeyF7,
	KeyF8:  glfw.KeyF8,
	KeyF9:  glfw.KeyF9,
	KeyF10: glfw.KeyF10,
	KeyF11: glfw.KeyF11,
	KeyF12: glfw.KeyF12,
	KeyF13: glfw.KeyF13,
	KeyF14: glfw.KeyF14,
	KeyF15: glfw.KeyF15,
	KeyF16: glfw.KeyF16,
	KeyF17: glfw.KeyF17,
	KeyF18: glfw.KeyF18,
	KeyF19: glfw.KeyF19,
	KeyF20: glfw.KeyF20,
	KeyF21: glfw.KeyF21,
	KeyF22: glfw.KeyF22,
	KeyF23: glfw.KeyF23,
	KeyF24: glfw.KeyF24,
	KeyF25: glfw.KeyF25,

	KeyKP0:        glfw.KeyKP0,
	KeyKP1:        glfw.KeyKP1,
	KeyKP2:        glfw.KeyKP2,
	KeyKP3:        glfw.KeyKP3,
	KeyKP4:        glfw.KeyKP4,
	KeyKP5:        glfw.KeyKP5,
	KeyKP6:        glfw.KeyKP6,
	KeyKP7:        glfw.KeyKP7,
	KeyKP8:        glfw.KeyKP8,
	KeyKP9:        glfw.KeyKP9,
	KeyKPDecimal:  glfw.KeyKPDecimal,
	KeyKPDivide:   glfw.KeyKPDivide,
	KeyKPMultiply: glfw.KeyKPMultiply,
	KeyKPSubtract: glfw.KeyKPSubtract,
	KeyKPAdd:      glfw.KeyKPAdd,
	KeyKPEnter:    glfw.KeyKPEnter,
	KeyKPEqual:    glfw.KeyKPEqual,

	KeyLeftShift:    glfw.KeyLeftShift,
	KeyLeftControl:  glfw.KeyLeftControl,
	KeyLeftAlt:      glfw.KeyLeftAlt,
	KeyLeftSuper:    glfw.KeyLeftSuper,
	KeyRightShift:   glfw.KeyRightShift,
	KeyRightControl: glfw.KeyRightControl,
	KeyRightAlt:     glfw.KeyRightAlt,
	KeyRightSuper:   glfw.KeyRightSuper,

	KeyMenu: glfw.KeyMenu,

	// Notes:
	// KeyLeftCommand/KeyRightCommand/KeyOptionLeft/KeyOptionRight/KeyFn are not part of GLFW's key enum.
	// GLFW does not expose media/volume/brightness keys in a cross-platform way.
}

var glfwMouseButtonMap = map[Input]glfw.MouseButton{
	MouseLeft:    glfw.MouseButtonLeft,
	MouseRight:   glfw.MouseButtonRight,
	MouseMiddle:  glfw.MouseButtonMiddle,
	MouseButton4: glfw.MouseButton4,
	MouseButton5: glfw.MouseButton5,
	MouseButton6: glfw.MouseButton6,
	MouseButton7: glfw.MouseButton7,
	MouseButton8: glfw.MouseButton8,
}

var glfwGamepadButtonMap = map[Input]glfw.GamepadButton{
	PadA:          glfw.ButtonA,
	PadB:          glfw.ButtonB,
	PadX:          glfw.ButtonX,
	PadY:          glfw.ButtonY,
	PadLB:         glfw.ButtonLeftBumper,
	PadRB:         glfw.ButtonRightBumper,
	PadBack:       glfw.ButtonBack,
	PadStart:      glfw.ButtonStart,
	PadGuide:      glfw.ButtonGuide,
	PadLeftThumb:  glfw.ButtonLeftThumb,
	PadRightThumb: glfw.ButtonRightThumb,
	PadDpadUp:     glfw.ButtonDpadUp,
	PadDpadRight:  glfw.ButtonDpadRight,
	PadDpadDown:   glfw.ButtonDpadDown,
	PadDpadLeft:   glfw.ButtonDpadLeft,
}

var glfwGamepadAxisMap = map[Input]glfw.GamepadAxis{
	PadAxisLeftX:        glfw.AxisLeftX,
	PadAxisLeftY:        glfw.AxisLeftY,
	PadAxisRightX:       glfw.AxisRightX,
	PadAxisRightY:       glfw.AxisRightY,
	PadAxisLeftTrigger:  glfw.AxisLeftTrigger,
	PadAxisRightTrigger: glfw.AxisRightTrigger,
}

type InputManager struct {
	inputToSignal  map[Input][]*InputSignal
	signalToAction map[*InputSignal][]*Action

	mu     sync.RWMutex
	active map[Input]bool // latest captured GLFW state
}

// UpdateSignals should be called once per logic tick (FixedHzLoop).
// It snapshots last state and applies the latest captured state.
func (im *InputManager) UpdateSignals() {
	im.mu.RLock()
	defer im.mu.RUnlock()

	for input, signals := range im.inputToSignal {
		active := im.active[input]
		for _, sig := range signals {
			if sig == nil {
				continue
			}
			sig.Snapshot()
			sig.Set(active)
		}
	}
}

func isInputActive(win *Window, input Input, gamepads []*glfw.GamepadState) bool {
	if key, ok := glfwKeyMap[input]; ok {
		return win.GLFW().GetKey(key) == glfw.Press
	}

	if btn, ok := glfwMouseButtonMap[input]; ok {
		return win.GLFW().GetMouseButton(btn) == glfw.Press
	}

	if len(gamepads) == 0 {
		return false
	}

	if btn, ok := glfwGamepadButtonMap[input]; ok {
		for _, st := range gamepads {
			if st != nil && st.Buttons[btn] == glfw.Press {
				return true
			}
		}
		return false
	}

	if axis, ok := glfwGamepadAxisMap[input]; ok {
		const deadzone = float32(0.2)
		for _, st := range gamepads {
			if st == nil {
				continue
			}
			v := st.Axes[axis]
			if v > deadzone || v < -deadzone {
				return true
			}
		}
		return false
	}

	return false
}

func connectedGamepads() []*glfw.GamepadState {
	gamepads := make([]*glfw.GamepadState, 0, 4)
	for joyID := glfw.Joystick1; joyID <= glfw.JoystickLast; joyID++ {
		if !joyID.Present() || !joyID.IsGamepad() {
			continue
		}
		state := joyID.GetGamepadState()
		if state == nil {
			continue
		}
		gamepads = append(gamepads, state)
	}
	return gamepads
}
func (im *InputManager) CaptureInputs(windows []*Window) {
	im.mu.Lock()
	defer im.mu.Unlock()

	gamepads := connectedGamepads()

	for input := range im.inputToSignal {
		im.active[input] = false // Reset

		for _, win := range windows {
			if win == nil || win.ShouldClose() {
				continue
			}
			if isInputActive(win, input, gamepads) {
				im.active[input] = true
				break
			}
		}
	}
}

func (im *InputManager) BindInput(input Input, sig *InputSignal) {
	if im.inputToSignal == nil {
		im.inputToSignal = make(map[Input][]*InputSignal)
	}
	im.inputToSignal[input] = append(im.inputToSignal[input], sig)
	if im.active == nil {
		im.active = make(map[Input]bool)
	}
}

func NewInputManager() *InputManager {
	return &InputManager{
		inputToSignal:  make(map[Input][]*InputSignal),
		signalToAction: make(map[*InputSignal][]*Action),
		active:         make(map[Input]bool),
	}
}

func (im *InputManager) Tick() {
	im.UpdateSignals()

	im.mu.RLock()
	defer im.mu.RUnlock()

	for _, actions := range im.signalToAction {
		for _, action := range actions {
			action.RunWhenShould()
		}
	}
}

func (im *InputManager) BindAction(sig *InputSignal, action *Action) {
	im.mu.Lock()
	defer im.mu.Unlock()

	action.BindSignal(sig)

	im.signalToAction[sig] = append(im.signalToAction[sig], action)
}
