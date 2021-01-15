package commands

import "github.com/thelonelyghost/p2box/libmachine"

func cmdIP(c CommandLine, api libmachine.API) error {
	return runAction("ip", c, api)
}
