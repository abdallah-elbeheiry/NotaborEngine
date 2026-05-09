package notasdl

type EventType int

const (
	EventQuit EventType = iota
	EventWindowClose
	EventKeyDown
	EventKeyUp
	EventMouseDown
	EventMouseUp
	EventMouseMove
	EventGamepadButtonDown
	EventGamepadButtonUp
	EventGamepadAxisMotion
)

type Event struct {
	Type     EventType
	WindowID uint32

	Key         Key
	MouseBtn    MouseButton
	GamepadBtn  GamepadButton
	GamepadAxis GamepadAxis
	AxisValue   float32

	X, Y float32
}

type EventHandler func(Event)

type EventEmitter interface {
	SubscribeEvents(EventHandler)
}
