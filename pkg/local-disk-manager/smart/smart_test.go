package smart

import "testing"

func TestSmartHealthStatus(t *testing.T) {
	device := Device{Device: "disk0"}
	ctr := NewSMARTController(device, "-d", "nvme")

	ok, err := ctr.GetHealthStatus()
	if err != nil {
		t.Fatalf("Failed to get disk health status %v", err)
	}

	t.Logf("Disk %v health statsus: %v", device.Device, ok)
}
