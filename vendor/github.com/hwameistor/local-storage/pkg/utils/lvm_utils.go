package utils

import (
	"errors"
	"strconv"
)

var (
	ErrNotLVMByteNum = errors.New("LVM byte format unrecognised")
)

func ConvertLVMBytesToNumeric(lvmbyte string) (int64, error) {
	if len(lvmbyte) == 0 || lvmbyte[len(lvmbyte)-1] != 'B' {
		return 0, ErrNotLVMByteNum
	}
	num, err := strconv.Atoi(lvmbyte[:len(lvmbyte)-1])
	if err != nil {
		return 0, err
	}

	return int64(num), nil
}

func ConvertNumericToLVMBytes(num int64) string {
	lvmSizeInBytes := NumericToLVMBytes(num)
	return strconv.Itoa(int(lvmSizeInBytes)) + "B"
}

func NumericToLVMBytes(bytes int64) int64 {
	peSize := int64(4 * 1024 * 1024)
	if bytes <= peSize {
		return peSize
	}
	if bytes%peSize == 0 {
		return bytes
	}
	return (bytes/peSize + 1) * peSize
}
