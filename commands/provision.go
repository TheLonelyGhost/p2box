package commands

import "github.com/thelonelyghost/p2box/libmachine"

func cmdProvision(c CommandLine, api libmachine.API) error {
	return runAction("provision", c, api)
}
