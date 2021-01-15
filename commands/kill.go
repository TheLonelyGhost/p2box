package commands

import "github.com/thelonelyghost/p2box/libmachine"

func cmdKill(c CommandLine, api libmachine.API) error {
	return runAction("kill", c, api)
}
