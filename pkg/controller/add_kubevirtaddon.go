package controller

import (
	"github.ibm.com/steve-kim-ibm/kubevirt-addon/pkg/controller/kubevirtaddon"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, kubevirtaddon.Add)
}
