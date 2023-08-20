package utils

import "time"

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
