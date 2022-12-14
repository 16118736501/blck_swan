//go:build linux

package stdout

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
	"github.com/ungtb10d/gdu/v5/pkg/device"
)

func init() {
	log.SetLevel(log.WarnLevel)
}

func TestShowDevicesWithErr(t *testing.T) {
	output := bytes.NewBuffer(make([]byte, 10))

	getter := device.LinuxDevicesInfoGetter{MountsPath: "/xyzxyz"}
	ui := CreateStdoutUI(output, false, true, false, false, false, false, false)
	err := ui.ListDevices(getter)

	assert.Contains(t, err.Error(), "no such file")
}
