package system

import (
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/event"
)

// ensure EventManager implem System interface
var _ core.System = (*event.EventManager)(nil)
