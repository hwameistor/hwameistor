package controller

import (
	"github.com/hwameistor/local-storage/pkg/controller/localvolumeconvert"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, localvolumeconvert.Add)
}
