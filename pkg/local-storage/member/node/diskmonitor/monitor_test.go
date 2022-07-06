package diskmonitor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterUdevEvent(t *testing.T) {
	rawEvent := `KERNEL[8221865.810325] add      /devices/pci0000:00/0000:00:15.0/0000:03:00.0/host0/target0:0:4/0:0:4:0/block/sdt (block)
ACTION=add
DEVNAME=/dev/sdt
DEVPATH=/devices/pci0000:00/0000:00:15.0/0000:03:00.0/host0/target0:0:4/0:0:4:0/block/sdt
DEVTYPE=disk
MAJOR=65
MINOR=48
SEQNUM=81164
SUBSYSTEM=block`
	segments := strings.Split(rawEvent, "\n")
	dmI := New(NewEventQueue("DiskEvents"))
	dm, ok := dmI.(*diskMonitor)
	assert.True(t, ok, "New() does not create a *diskMonitor")
	event, err := dm.filterUdevEvent(segments)
	assert.Nil(t, err)
	assert.NotNil(t, event)
	assert.Equal(t, ActionAdd, event.Action)
	assert.Equal(t, "/dev/sdt", event.DevName)
	assert.Equal(t, "65", event.Major)
	assert.Equal(t, "48", event.Minor)
}
