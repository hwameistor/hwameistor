package hwameistor

import "testing"

func TestDeviceShortName(t *testing.T) {
	tests := []struct {
		name       string
		devicePath string
		want       string
	}{
		{name: "regular dev path", devicePath: "/dev/sdb", want: "sdb"},
		{name: "empty path", devicePath: "", want: ""},
		{name: "already short", devicePath: "sdb", want: "sdb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deviceShortName(tt.devicePath); got != tt.want {
				t.Fatalf("deviceShortName(%q) = %q, want %q", tt.devicePath, got, tt.want)
			}
		})
	}
}
