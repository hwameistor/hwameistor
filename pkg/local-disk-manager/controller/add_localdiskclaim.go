package controller

import (
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/controller/localdiskclaim"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, localdiskclaim.Add)
}
