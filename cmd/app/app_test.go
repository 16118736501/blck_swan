package app

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/dundee/gdu/v4/device"
	"github.com/dundee/gdu/v4/internal/testapp"
	"github.com/dundee/gdu/v4/internal/testdev"
	"github.com/dundee/gdu/v4/internal/testdir"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	out, err := runApp(
		&Flags{ShowVersion: true},
		[]string{},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Contains(t, out, "Version:\t development")
	assert.Nil(t, err)
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

func TestAnalyzePath(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null"},
		[]string{"test_dir"},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Contains(t, out, "nested")
	assert.Nil(t, err)
}

func TestAnalyzePathWithGui(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null"},
		[]string{"test_dir"},
		true,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Empty(t, out)
	assert.Nil(t, err)
}

func TestAnalyzePathWithErr(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null"},
		[]string{"xxx"},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Equal(t, "", out)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestNoCross(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null", NoCross: true},
		[]string{"test_dir"},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Contains(t, out, "nested")
	assert.Nil(t, err)
}

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

func TestListDevices(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null", ShowDisks: true},
		[]string{},
		false,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Contains(t, out, "Device")
	assert.Nil(t, err)
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

func TestListDevicesWithGui(t *testing.T) {
	fin := testdir.CreateTestDir()
	defer fin()

	out, err := runApp(
		&Flags{LogFile: "/dev/null", ShowDisks: true},
		[]string{},
		true,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Nil(t, err)
	assert.Empty(t, out)
}

func TestMaxCores(t *testing.T) {
	_, err := runApp(
		&Flags{LogFile: "/dev/null", MaxCores: 1},
		[]string{},
		true,
		testdev.DevicesInfoGetterMock{},
	)

	assert.Equal(t, 1, runtime.GOMAXPROCS(0))
	assert.Nil(t, err)
}

func TestMaxCoresHighEdge(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("Skipping on a single core CPU")
	}
	out, err := runApp(
		&Flags{LogFile: "/dev/null", MaxCores: runtime.NumCPU() + 1},
		[]string{},
		true,
		testdev.DevicesInfoGetterMock{},
	)

	assert.NotEqual(t, runtime.NumCPU(), runtime.GOMAXPROCS(0))
	assert.Empty(t, out)
	assert.Nil(t, err)
}

func TestMaxCoresLowEdge(t *testing.T) {
	if runtime.NumCPU() < 2 {
		t.Skip("Skipping on a single core CPU")
	}
	out, err := runApp(
		&Flags{LogFile: "/dev/null", MaxCores: -100},
		[]string{},
		true,
		testdev.DevicesInfoGetterMock{},
	)

	assert.NotEqual(t, runtime.NumCPU(), runtime.GOMAXPROCS(0))
	assert.Empty(t, out)
	assert.Nil(t, err)
}

func runApp(flags *Flags, args []string, istty bool, getter device.DevicesInfoGetter) (string, error) {
	buff := bytes.NewBufferString("")

	app := App{
		Flags:   flags,
		Args:    args,
		Istty:   istty,
		Writer:  buff,
		TermApp: testapp.CreateMockedApp(false),
		Getter:  getter,
	}
	err := app.Run()

	return strings.TrimSpace(buff.String()), err
}
