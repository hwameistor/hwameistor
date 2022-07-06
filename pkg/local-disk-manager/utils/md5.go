package utils

import (
	"crypto/md5"
	"fmt"
)

// Hash
func Hash(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}
