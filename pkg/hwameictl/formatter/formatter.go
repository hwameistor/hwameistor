package formatter

import (
	"fmt"
	"time"
)

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

func FormatBytesToSize(bytes int64) string {
	const (
		Byte = 1.0
		KB   = 1024 * Byte
		MB   = 1024 * KB
		GB   = 1024 * MB
		TB   = 1024 * GB
	)

	var unit string
	var size float64

	switch {
	case bytes >= TB:
		unit = "TB"
		size = float64(bytes / TB)
	case bytes >= GB:
		unit = "GB"
		size = float64(bytes / GB)
	default:
		unit = "MB"
		size = float64(bytes / MB)
	}
	return fmt.Sprintf("%.2f %s", size, unit)
}

func FormatPercentString(v1, v2 int64) string {
	return fmt.Sprintf("%.2f%%", (float64(v1)/float64(v2))*100)
}
