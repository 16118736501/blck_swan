package main

import (
	"fmt"
	"os"

	"github.com/dundee/gdu/cmd"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func main() {
	rf := &cmd.RunFlags{}

	var rootCmd = &cobra.Command{
		Use:   "gdu [directory_to_scan]",
		Short: "Pretty fast disk usage analyzer written in Go",
		Long: `Pretty fast disk usage analyzer written in Go.

Gdu is intended primarily for SSD disks where it can fully utilize parallel processing.
However HDDs work as well, but the performance gain is not so huge.
`,
		Args: cobra.MaximumNArgs(1),
		Run: func(command *cobra.Command, args []string) {
			istty := isatty.IsTerminal(os.Stdout.Fd())
			cmd.Run(rf, args, istty, os.Stdout, false)
		},
	}

	flags := rootCmd.Flags()
	flags.StringVarP(&rf.LogFile, "log-file", "l", "/dev/null", "Path to a logfile")
	flags.StringSliceVarP(&rf.IgnoreDirs, "ignore-dirs", "i", []string{"/proc", "/dev", "/sys", "/run"}, "Absolute paths to ignore (separated by comma)")
	flags.BoolVarP(&rf.ShowDisks, "show-disks", "d", false, "Show all mounted disks")
	flags.BoolVarP(&rf.ShowVersion, "version", "v", false, "Print version")
	flags.BoolVarP(&rf.NoColor, "no-color", "c", false, "Do not use colorized output")
	flags.BoolVarP(&rf.NonInteractive, "non-interactive", "n", false, "Do not run in interactive mode")
	flags.BoolVarP(&rf.NoProgress, "no-progress", "p", false, "Do not show progress in non-interactive mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
