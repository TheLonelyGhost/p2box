package commands

import "github.com/thelonelyghost/p2box/libmachine"

func cmdStop(c CommandLine, api libmachine.API) error {
	return runAction("stop", c, api)
}
