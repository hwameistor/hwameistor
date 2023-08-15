package main

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/hwameictl/cmdparser"
	"os"
)

func main() {
	err := cmdparser.Hwameictl.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
