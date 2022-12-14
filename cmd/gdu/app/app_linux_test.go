//go:build linux

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ungtb10d/gdu/v5/internal/testdev"
	"github.com/ungtb10d/gdu/v5/internal/testdir"
	"github.com/ungtb10d/gdu/v5/pkg/device"
)

func TestNoCrossWithErr(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null", NoCross: true},
		[]string{"test_dir"},
		false,
		device.LinuxDevicesInfoGetter{MountsPath: "/xxxyyy"},
	)

	assert.Equal(t, "loading mount points: open /xxxyyy: no such file or directory", err.Error())
	assert.Empty(t, out)
}

func TestListDevicesWithErr(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	_, err := runApp(
		&Flags{LogFile: "/dev/null", ShowDisks: true},
		[]string{},
		false,
		device.LinuxDevicesInfoGetter{MountsPath: "/xxxyyy"},
	)

	assert.Equal(t, "loading mount points: open /xxxyyy: no such file or directory", err.Error())
}

func TestLogError(t *testing.T) {
	out, err := runApp(
		&Flags{LogFile: "/xyzxyz"},
		[]string{},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestOutputFileError(t *testing.T) {
	out, err := runApp(
		&Flags{LogFile: "/dev/null", OutputFile: "/xyzxyz"},
		[]string{},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Empty(t, out)
	assert.Contains(t, err.Error(), "permission denied")
}
