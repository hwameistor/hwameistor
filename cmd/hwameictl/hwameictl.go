package main

import (
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser"
	"os"
)

func main() {
	err := cmdparser.Hwameictl.Execute()
	if err != nil {
		os.Exit(1)
	}
}
