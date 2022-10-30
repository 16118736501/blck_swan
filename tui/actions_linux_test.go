//go:build linux
// +build linux

package tui

import (
	"bytes"
	"testing"

ungtb10d
ungtb10d
	"github.com/stretchr/testify/assert"
)

func TestShowDevicesWithError(t *testing.T) {
	app, simScreen := testapp.CreateTestAppWithSimScreen(50, 50)
	defer simScreen.Fini()

	getter := device.LinuxDevicesInfoGetter{MountsPath: "/xyzxyz"}

	ui := CreateUI(app, simScreen, &bytes.Buffer{}, false, false, false, false, false)
	err := ui.ListDevices(getter)

	assert.Contains(t, err.Error(), "no such file")
}
