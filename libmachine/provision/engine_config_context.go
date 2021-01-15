package provision

import (
	"github.com/thelonelyghost/p2box/libmachine/auth"
	"github.com/thelonelyghost/p2box/libmachine/engine"
)

type EngineConfigContext struct {
	AuthOptions      auth.Options
	EngineOptions    engine.Options
	EngineOptionsDir string
}
