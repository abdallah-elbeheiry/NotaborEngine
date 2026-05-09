package notacore

import (
	"NotaborEngine/notasdl"
	"sync"

	"github.com/Zyko0/go-sdl3/sdl"
)

type StateInput int

const (
	KeySpace StateInput = iota
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
	KeyMenu

	KeyMediaPlayPause
	KeyMediaStop
	KeyMediaNext
	KeyMediaPrev
	KeyVolumeUp
	KeyVolumeDown

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
)

var sdlKeyMap = map[notasdl.Key]StateInput{
	notasdl.Key(sdl.K_SPACE):                KeySpace,
	notasdl.Key(sdl.K_APOSTROPHE):           KeyApostrophe,
	notasdl.Key(sdl.K_COMMA):                KeyComma,
	notasdl.Key(sdl.K_MINUS):                KeyMinus,
	notasdl.Key(sdl.K_PERIOD):               KeyPeriod,
	notasdl.Key(sdl.K_SLASH):                KeySlash,
	notasdl.Key(sdl.K_0):                    Key0,
	notasdl.Key(sdl.K_1):                    Key1,
	notasdl.Key(sdl.K_2):                    Key2,
	notasdl.Key(sdl.K_3):                    Key3,
	notasdl.Key(sdl.K_4):                    Key4,
	notasdl.Key(sdl.K_5):                    Key5,
	notasdl.Key(sdl.K_6):                    Key6,
	notasdl.Key(sdl.K_7):                    Key7,
	notasdl.Key(sdl.K_8):                    Key8,
	notasdl.Key(sdl.K_9):                    Key9,
	notasdl.Key(sdl.K_SEMICOLON):            KeySemicolon,
	notasdl.Key(sdl.K_EQUALS):               KeyEqual,
	notasdl.Key(sdl.K_A):                    KeyA,
	notasdl.Key(sdl.K_B):                    KeyB,
	notasdl.Key(sdl.K_C):                    KeyC,
	notasdl.Key(sdl.K_D):                    KeyD,
	notasdl.Key(sdl.K_E):                    KeyE,
	notasdl.Key(sdl.K_F):                    KeyF,
	notasdl.Key(sdl.K_G):                    KeyG,
	notasdl.Key(sdl.K_H):                    KeyH,
	notasdl.Key(sdl.K_I):                    KeyI,
	notasdl.Key(sdl.K_J):                    KeyJ,
	notasdl.Key(sdl.K_K):                    KeyK,
	notasdl.Key(sdl.K_L):                    KeyL,
	notasdl.Key(sdl.K_M):                    KeyM,
	notasdl.Key(sdl.K_N):                    KeyN,
	notasdl.Key(sdl.K_O):                    KeyO,
	notasdl.Key(sdl.K_P):                    KeyP,
	notasdl.Key(sdl.K_Q):                    KeyQ,
	notasdl.Key(sdl.K_R):                    KeyR,
	notasdl.Key(sdl.K_S):                    KeyS,
	notasdl.Key(sdl.K_T):                    KeyT,
	notasdl.Key(sdl.K_U):                    KeyU,
	notasdl.Key(sdl.K_V):                    KeyV,
	notasdl.Key(sdl.K_W):                    KeyW,
	notasdl.Key(sdl.K_X):                    KeyX,
	notasdl.Key(sdl.K_Y):                    KeyY,
	notasdl.Key(sdl.K_Z):                    KeyZ,
	notasdl.Key(sdl.K_LEFTBRACKET):          KeyLeftBracket,
	notasdl.Key(sdl.K_BACKSLASH):            KeyBackslash,
	notasdl.Key(sdl.K_RIGHTBRACKET):         KeyRightBracket,
	notasdl.Key(sdl.K_GRAVE):                KeyGraveAccent,
	notasdl.Key(sdl.K_ESCAPE):               KeyEscape,
	notasdl.Key(sdl.K_RETURN):               KeyEnter,
	notasdl.Key(sdl.K_TAB):                  KeyTab,
	notasdl.Key(sdl.K_BACKSPACE):            KeyBackspace,
	notasdl.Key(sdl.K_INSERT):               KeyInsert,
	notasdl.Key(sdl.K_DELETE):               KeyDelete,
	notasdl.Key(sdl.K_RIGHT):                KeyRight,
	notasdl.Key(sdl.K_LEFT):                 KeyLeft,
	notasdl.Key(sdl.K_DOWN):                 KeyDown,
	notasdl.Key(sdl.K_UP):                   KeyUp,
	notasdl.Key(sdl.K_PAGEUP):               KeyPageUp,
	notasdl.Key(sdl.K_PAGEDOWN):             KeyPageDown,
	notasdl.Key(sdl.K_HOME):                 KeyHome,
	notasdl.Key(sdl.K_END):                  KeyEnd,
	notasdl.Key(sdl.K_CAPSLOCK):             KeyCapsLock,
	notasdl.Key(sdl.K_SCROLLLOCK):           KeyScrollLock,
	notasdl.Key(sdl.K_NUMLOCKCLEAR):         KeyNumLock,
	notasdl.Key(sdl.K_PRINTSCREEN):          KeyPrintScreen,
	notasdl.Key(sdl.K_PAUSE):                KeyPause,
	notasdl.Key(sdl.K_F1):                   KeyF1,
	notasdl.Key(sdl.K_F2):                   KeyF2,
	notasdl.Key(sdl.K_F3):                   KeyF3,
	notasdl.Key(sdl.K_F4):                   KeyF4,
	notasdl.Key(sdl.K_F5):                   KeyF5,
	notasdl.Key(sdl.K_F6):                   KeyF6,
	notasdl.Key(sdl.K_F7):                   KeyF7,
	notasdl.Key(sdl.K_F8):                   KeyF8,
	notasdl.Key(sdl.K_F9):                   KeyF9,
	notasdl.Key(sdl.K_F10):                  KeyF10,
	notasdl.Key(sdl.K_F11):                  KeyF11,
	notasdl.Key(sdl.K_F12):                  KeyF12,
	notasdl.Key(sdl.K_F13):                  KeyF13,
	notasdl.Key(sdl.K_F14):                  KeyF14,
	notasdl.Key(sdl.K_F15):                  KeyF15,
	notasdl.Key(sdl.K_F16):                  KeyF16,
	notasdl.Key(sdl.K_F17):                  KeyF17,
	notasdl.Key(sdl.K_F18):                  KeyF18,
	notasdl.Key(sdl.K_F19):                  KeyF19,
	notasdl.Key(sdl.K_F20):                  KeyF20,
	notasdl.Key(sdl.K_F21):                  KeyF21,
	notasdl.Key(sdl.K_F22):                  KeyF22,
	notasdl.Key(sdl.K_F23):                  KeyF23,
	notasdl.Key(sdl.K_F24):                  KeyF24,
	notasdl.Key(sdl.K_KP_0):                 KeyKP0,
	notasdl.Key(sdl.K_KP_1):                 KeyKP1,
	notasdl.Key(sdl.K_KP_2):                 KeyKP2,
	notasdl.Key(sdl.K_KP_3):                 KeyKP3,
	notasdl.Key(sdl.K_KP_4):                 KeyKP4,
	notasdl.Key(sdl.K_KP_5):                 KeyKP5,
	notasdl.Key(sdl.K_KP_6):                 KeyKP6,
	notasdl.Key(sdl.K_KP_7):                 KeyKP7,
	notasdl.Key(sdl.K_KP_8):                 KeyKP8,
	notasdl.Key(sdl.K_KP_9):                 KeyKP9,
	notasdl.Key(sdl.K_KP_PERIOD):            KeyKPDecimal,
	notasdl.Key(sdl.K_KP_DIVIDE):            KeyKPDivide,
	notasdl.Key(sdl.K_KP_MULTIPLY):          KeyKPMultiply,
	notasdl.Key(sdl.K_KP_MINUS):             KeyKPSubtract,
	notasdl.Key(sdl.K_KP_PLUS):              KeyKPAdd,
	notasdl.Key(sdl.K_KP_ENTER):             KeyKPEnter,
	notasdl.Key(sdl.K_KP_EQUALS):            KeyKPEqual,
	notasdl.Key(sdl.K_LSHIFT):               KeyLeftShift,
	notasdl.Key(sdl.K_LCTRL):                KeyLeftControl,
	notasdl.Key(sdl.K_LALT):                 KeyLeftAlt,
	notasdl.Key(sdl.K_LGUI):                 KeyLeftSuper,
	notasdl.Key(sdl.K_RSHIFT):               KeyRightShift,
	notasdl.Key(sdl.K_RCTRL):                KeyRightControl,
	notasdl.Key(sdl.K_RALT):                 KeyRightAlt,
	notasdl.Key(sdl.K_RGUI):                 KeyRightSuper,
	notasdl.Key(sdl.K_MENU):                 KeyMenu,
	notasdl.Key(sdl.K_MEDIA_PLAY_PAUSE):     KeyMediaPlayPause,
	notasdl.Key(sdl.K_MEDIA_STOP):           KeyMediaStop,
	notasdl.Key(sdl.K_MEDIA_NEXT_TRACK):     KeyMediaNext,
	notasdl.Key(sdl.K_MEDIA_PREVIOUS_TRACK): KeyMediaPrev,
	notasdl.Key(sdl.K_VOLUMEUP):             KeyVolumeUp,
	notasdl.Key(sdl.K_VOLUMEDOWN):           KeyVolumeDown,
}

var sdlMouseButtonMap = map[notasdl.MouseButton]StateInput{
	notasdl.MouseButton(sdl.BUTTON_LEFT):   MouseLeft,
	notasdl.MouseButton(sdl.BUTTON_RIGHT):  MouseRight,
	notasdl.MouseButton(sdl.BUTTON_MIDDLE): MouseMiddle,
	notasdl.MouseButton(sdl.BUTTON_X1):     MouseButton4,
	notasdl.MouseButton(sdl.BUTTON_X2):     MouseButton5,
	notasdl.MouseButton(6):                 MouseButton6,
	notasdl.MouseButton(7):                 MouseButton7,
	notasdl.MouseButton(8):                 MouseButton8,
}

var sdlGamepadButtonMap = map[notasdl.GamepadButton]StateInput{
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_SOUTH):          PadA,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_EAST):           PadB,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_WEST):           PadX,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_NORTH):          PadY,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_LEFT_SHOULDER):  PadLB,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_RIGHT_SHOULDER): PadRB,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_BACK):           PadBack,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_START):          PadStart,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_GUIDE):          PadGuide,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_LEFT_STICK):     PadLeftThumb,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_RIGHT_STICK):    PadRightThumb,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_DPAD_UP):        PadDpadUp,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_DPAD_RIGHT):     PadDpadRight,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_DPAD_DOWN):      PadDpadDown,
	notasdl.GamepadButton(sdl.GAMEPAD_BUTTON_DPAD_LEFT):      PadDpadLeft,
}

var sdlGamepadAxisMap = map[notasdl.GamepadAxis]StateInput{
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_LEFTX):         PadAxisLeftX,
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_LEFTY):         PadAxisLeftY,
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_RIGHTX):        PadAxisRightX,
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_RIGHTY):        PadAxisRightY,
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_LEFT_TRIGGER):  PadAxisLeftTrigger,
	notasdl.GamepadAxis(sdl.GAMEPAD_AXIS_RIGHT_TRIGGER): PadAxisRightTrigger,
}

type InputManager struct {
	ctx     *InputContext
	signals map[string]*InputSignal
	mu      sync.RWMutex
}

// NewInputManager creates a new input manager
func NewInputManager() *InputManager {
	return &InputManager{
		ctx:     NewInputContext(),
		signals: make(map[string]*InputSignal),
	}
}

// GetContext returns the input context for direct access if needed
func (im *InputManager) GetContext() *InputContext {
	return im.ctx
}

// Get retrieves a previously bound signal by name
func (im *InputManager) Get(name string) *InputSignal {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.signals[name]
}

// BeginFrame should be called at the start of each frame to update input states
func (im *InputManager) BeginFrame() {
	im.ctx.BeginFrame()
}

// HandleEvent feeds SDL-originated input events into the input context
func (im *InputManager) HandleEvent(event notasdl.Event) {
	switch event.Type {

	case notasdl.EventKeyDown:
		if input, ok := sdlKeyMap[event.Key]; ok {
			im.ctx.RecordKeyDown(input)
		}

	case notasdl.EventKeyUp:
		if input, ok := sdlKeyMap[event.Key]; ok {
			im.ctx.RecordKeyUp(input)
		}

	case notasdl.EventMouseDown:
		if input, ok := sdlMouseButtonMap[event.MouseBtn]; ok {
			im.ctx.RecordKeyDown(input)
		}

	case notasdl.EventMouseUp:
		if input, ok := sdlMouseButtonMap[event.MouseBtn]; ok {
			im.ctx.RecordKeyUp(input)
		}

	case notasdl.EventGamepadButtonDown:
		if input, ok := sdlGamepadButtonMap[event.GamepadBtn]; ok {
			im.ctx.RecordKeyDown(input)
		}

	case notasdl.EventGamepadButtonUp:
		if input, ok := sdlGamepadButtonMap[event.GamepadBtn]; ok {
			im.ctx.RecordKeyUp(input)
		}
	}
}
