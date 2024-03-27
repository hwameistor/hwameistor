package utils

import (
	"crypto/md5"
	"fmt"
)

// Hash returns the MD5 hash of a string
func Hash(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
