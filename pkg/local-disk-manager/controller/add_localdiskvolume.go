package controller

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/controller/localdiskvolume"
)

func init() {
	// AddToNodeManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToNodeManagerFuncs = append(AddToNodeManagerFuncs, localdiskvolume.Add)
}
