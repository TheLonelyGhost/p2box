package commands

import (
	"github.com/thelonelyghost/p2box/libmachine"
	"github.com/thelonelyghost/p2box/libmachine/log"
)

func cmdStart(c CommandLine, api libmachine.API) error {
	if err := runAction("start", c, api); err != nil {
		return err
	}

	log.Info("Started machines may have new IP addresses. You may need to re-run the `podman-machine env` command.")

	return nil
}
