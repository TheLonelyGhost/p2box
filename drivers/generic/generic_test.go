package generic

import (
	"testing"

	"github.com/thelonelyghost/p2box/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"generic-ip-address": "localhost",
			"generic-ssh-key":    "path",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}
