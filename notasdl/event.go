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
)

type Event struct {
	Type     EventType
	WindowID uint32

	Key      Key
	MouseBtn MouseButton

	X, Y float32
}
