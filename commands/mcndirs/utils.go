package mcndirs

import (
	"os"
	"path/filepath"

	"github.com/thelonelyghost/p2box/libmachine/mcnutils"
)

var (
	BaseDir = os.Getenv("MACHINE_STORAGE_PATH")
)

func GetBaseDir() string {
	if BaseDir == "" {
		BaseDir = filepath.Join(mcnutils.GetHomeDir(), ".local", "machine")
	}
	return BaseDir
}

func GetMachineDir() string {
	return filepath.Join(GetBaseDir(), "machines")
}

func GetMachineCertDir() string {
	return filepath.Join(GetBaseDir(), "certs")
}
