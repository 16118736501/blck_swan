package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dundee/gdu/v4/cmd/app"
	"github.com/dundee/gdu/v4/device"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

var af *app.Flags

var rootCmd = &cobra.Command{
	Use:   "gdu [directory_to_scan]",
	Short: "Pretty fast disk usage analyzer written in Go",
	Long: `Pretty fast disk usage analyzer written in Go.

Gdu is intended primarily for SSD disks where it can fully utilize parallel processing.
However HDDs work as well, but the performance gain is not so huge.
`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE:         runE,
}

func init() {
	af = &app.Flags{}
	flags := rootCmd.Flags()
	flags.StringVarP(&af.LogFile, "log-file", "l", "/dev/null", "Path to a logfile")
	flags.StringSliceVarP(&af.IgnoreDirs, "ignore-dirs", "i", []string{"/proc", "/dev", "/sys", "/run"}, "Absolute paths to ignore (separated by comma)")
	flags.IntVarP(&af.MaxCores, "max-cores", "m", runtime.NumCPU(), fmt.Sprintf("Set max cores that GDU will use. %d cores available", runtime.NumCPU()))
	flags.BoolVarP(&af.ShowDisks, "show-disks", "d", false, "Show all mounted disks")
	flags.BoolVarP(&af.ShowApparentSize, "show-apparent-size", "a", false, "Show apparent size")
	flags.BoolVarP(&af.ShowVersion, "version", "v", false, "Print version")
	flags.BoolVarP(&af.NoColor, "no-color", "c", false, "Do not use colorized output")
	flags.BoolVarP(&af.NonInteractive, "non-interactive", "n", false, "Do not run in interactive mode")
	flags.BoolVarP(&af.NoProgress, "no-progress", "p", false, "Do not show progress in non-interactive mode")
	flags.BoolVarP(&af.NoCross, "no-cross", "x", false, "Do not cross filesystem boundaries")
}

func runE(command *cobra.Command, args []string) error {
	istty := isatty.IsTerminal(os.Stdout.Fd())

	// we are not able to analyze disk usage on Windows and Plan9
	if runtime.GOOS == "windows" || runtime.GOOS == "plan9" {
		af.ShowApparentSize = true
	}
	if runtime.GOOS == "windows" && af.LogFile == "/dev/null" {
		af.LogFile = "nul"
	}

	var termApp *tview.Application = nil

	if !af.ShowVersion && !af.NonInteractive && istty {
		screen, err := tcell.NewScreen()
		if err != nil {
			return fmt.Errorf("Error creating screen: %w", err)
		}
		screen.Init()
		defer screen.Clear()
		defer screen.Fini()

		termApp = tview.NewApplication()
		termApp.SetScreen(screen)
	}

	a := app.App{
		Flags:   af,
		Args:    args,
		Istty:   istty,
		Writer:  os.Stdout,
		TermApp: termApp,
		Getter:  device.Getter,
	}
	return a.Run()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
