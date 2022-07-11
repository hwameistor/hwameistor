package manager

// RegisterManages
var RegisterManages = make(chan interface{}, 1)

// EventType indicates the type of mount event
type EventType = string

const (
	ADD     EventType = "add"
	REMOVE  EventType = "remove"
	CHANGE  EventType = "change"
	MOVE    EventType = "move"
	ONLINE  EventType = "online"
	OFFLINE EventType = "offline"
	BIND    EventType = "bind"
	UNBIND  EventType = "unbind"
	EXIST   EventType = "exist"
)

// Event indicates the type of MountRawBlock event and the properties of the mounted disk
type Event struct {
	Type    EventType
	DevPath string
	DevName string
	DevType string
}

// Manager for disk monitor
type Manager interface {
	// ListExist list all disks exist on node
	ListExist() []Event

	// Monitor monitor all disk events(e.g. add/remove/offline)
	Monitor(chan Event)
}

// RegisterManager
func RegisterManager(manager Manager) {
	RegisterManages <- manager
}

// NewManager
func NewManager() Manager {
	return (<-RegisterManages).(Manager)
}
