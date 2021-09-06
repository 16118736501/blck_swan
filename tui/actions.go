package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/dundee/gdu/v5/build"
	"github.com/dundee/gdu/v5/pkg/analyze"
	"github.com/dundee/gdu/v5/pkg/device"
	"github.com/dundee/gdu/v5/report"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const defaultLinesCount = 500
const linesTreshold = 20

// ListDevices lists mounted devices and shows their disk usage
func (ui *UI) ListDevices(getter device.DevicesInfoGetter) error {
	var err error
	ui.devices, err = getter.GetDevicesInfo()
	if err != nil {
		return err
	}

	ui.showDevices()

	return nil
}

// AnalyzePath analyzes recursively disk usage in given path
func (ui *UI) AnalyzePath(path string, parentDir *analyze.Dir) error {
	ui.progress = tview.NewTextView().SetText("Scanning...")
	ui.progress.SetBorder(true).SetBorderPadding(2, 2, 2, 2)
	ui.progress.SetTitle(" Scanning... ")
	ui.progress.SetDynamicColors(true)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 10, 1, false).
			AddItem(ui.progress, 8, 1, false).
			AddItem(nil, 10, 1, false), 0, 50, false).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("progress", flex, true, true)
	ui.table.SetSelectedFunc(ui.fileItemSelected)

	go ui.updateProgress()

	go func() {
		currentDir := ui.Analyzer.AnalyzeDir(path, ui.CreateIgnoreFunc())
		runtime.GC()

		if parentDir != nil {
			currentDir.Parent = parentDir
			parentDir.Files = parentDir.Files.RemoveByName(currentDir.Name)
			parentDir.Files.Append(currentDir)

			links := make(analyze.AlreadyCountedHardlinks, 10)
			ui.topDir.UpdateStats(links)
		} else {
			ui.topDirPath = path
			ui.topDir = currentDir
		}

		ui.app.QueueUpdateDraw(func() {
			ui.currentDir = currentDir
			ui.showDir()
			ui.pages.RemovePage("progress")
		})

		if ui.done != nil {
			ui.done <- struct{}{}
		}
	}()

	return nil
}

// ReadAnalysis reads analysis report from JSON file
func (ui *UI) ReadAnalysis(input io.Reader) error {
	ui.progress = tview.NewTextView().SetText("Reading analysis from file...")
	ui.progress.SetBorder(true).SetBorderPadding(2, 2, 2, 2)
	ui.progress.SetTitle(" Reading... ")
	ui.progress.SetDynamicColors(true)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 10, 1, false).
			AddItem(ui.progress, 8, 1, false).
			AddItem(nil, 10, 1, false), 0, 50, false).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("progress", flex, true, true)

	go func() {
		var err error
		ui.currentDir, err = report.ReadAnalysis(input)
		if err != nil {
			ui.app.QueueUpdateDraw(func() {
				ui.pages.RemovePage("progress")
				ui.showErr("Error reading file", err)
			})
			if ui.done != nil {
				ui.done <- struct{}{}
			}
			return
		}
		runtime.GC()

		ui.topDirPath = ui.currentDir.GetPath()
		ui.topDir = ui.currentDir

		links := make(analyze.AlreadyCountedHardlinks, 10)
		ui.topDir.UpdateStats(links)

		ui.app.QueueUpdateDraw(func() {
			ui.showDir()
			ui.pages.RemovePage("progress")
		})

		if ui.done != nil {
			ui.done <- struct{}{}
		}
	}()

	return nil
}

func (ui *UI) deleteSelected(shouldEmpty bool) {
	row, column := ui.table.GetSelection()
	selectedItem := ui.table.GetCell(row, column).GetReference().(analyze.Item)

	var action, acting string
	if shouldEmpty {
		action = "empty "
		acting = "emptying"
	} else {
		action = "delete "
		acting = "deleting"
	}
	modal := tview.NewModal().SetText(strings.Title(acting) + " " + selectedItem.GetName() + "...")
	ui.pages.AddPage(acting, modal, true, true)

	var currentDir *analyze.Dir
	var deleteItems []analyze.Item
	if shouldEmpty && selectedItem.IsDir() {
		currentDir = selectedItem.(*analyze.Dir)
		for _, file := range currentDir.Files {
			deleteItems = append(deleteItems, file)
		}
	} else {
		currentDir = ui.currentDir
		deleteItems = append(deleteItems, selectedItem)
	}

	var deleteFun func(*analyze.Dir, analyze.Item) error
	if shouldEmpty && !selectedItem.IsDir() {
		deleteFun = ui.emptier
	} else {
		deleteFun = ui.remover
	}
	go func() {
		for _, item := range deleteItems {
			if err := deleteFun(currentDir, item); err != nil {
				msg := "Can't " + action + selectedItem.GetName()
				ui.app.QueueUpdateDraw(func() {
					ui.pages.RemovePage(acting)
					ui.showErr(msg, err)
				})
				if ui.done != nil {
					ui.done <- struct{}{}
				}
				return
			}
		}

		ui.app.QueueUpdateDraw(func() {
			ui.pages.RemovePage(acting)
			ui.showDir()
			ui.table.Select(min(row, ui.table.GetRowCount()-1), 0)
		})

		if ui.done != nil {
			ui.done <- struct{}{}
		}
	}()
}

func (ui *UI) showFile() *tview.TextView {
	if ui.currentDir == nil {
		return nil
	}

	row, column := ui.table.GetSelection()
	selectedFile := ui.table.GetCell(row, column).GetReference().(analyze.Item)
	if selectedFile.IsDir() {
		return nil
	}

	f, err := os.Open(selectedFile.GetPath())
	if err != nil {
		ui.showErr("Error opening file", err)
		return nil
	}

	totalLines := 0
	scanner := bufio.NewScanner(f)

	file := tview.NewTextView()
	ui.currentDirLabel.SetText("[::b] --- " +
		strings.TrimPrefix(selectedFile.GetPath(), build.RootPathPrefix) +
		" ---").SetDynamicColors(true)

	readNextPart := func(linesCount int) int {
		var err error
		readLines := 0
		for scanner.Scan() && readLines <= linesCount {
			_, err = file.Write(scanner.Bytes())
			if err != nil {
				ui.showErr("Error reading file", err)
				return 0
			}
			_, err = file.Write([]byte("\n"))
			if err != nil {
				ui.showErr("Error reading file", err)
				return 0
			}
			readLines++
		}
		return readLines
	}
	totalLines += readNextPart(defaultLinesCount)

	file.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Key() == tcell.KeyESC {
			err = f.Close()
			if err != nil {
				ui.showErr("Error closing file", err)
				return event
			}
			ui.currentDirLabel.SetText("[::b] --- " +
				strings.TrimPrefix(ui.currentDirPath, build.RootPathPrefix) +
				" ---").SetDynamicColors(true)
			ui.pages.RemovePage("file")
			ui.app.SetFocus(ui.table)
			return event
		}

		switch {
		case event.Rune() == 'j':
			fallthrough
		case event.Rune() == 'G':
			fallthrough
		case event.Key() == tcell.KeyDown:
			fallthrough
		case event.Key() == tcell.KeyPgDn:
			_, _, _, height := file.GetInnerRect()
			row, _ := file.GetScrollOffset()
			if height+row > totalLines-linesTreshold {
				totalLines += readNextPart(defaultLinesCount)
			}
		}
		return event
	})

	grid := tview.NewGrid().SetRows(1, 1, 0, 1).SetColumns(0)
	grid.AddItem(ui.header, 0, 0, 1, 1, 0, 0, false).
		AddItem(ui.currentDirLabel, 1, 0, 1, 1, 0, 0, false).
		AddItem(file, 2, 0, 1, 1, 0, 0, true).
		AddItem(ui.footerLabel, 3, 0, 1, 1, 0, 0, false)

	ui.pages.HidePage("background")
	ui.pages.AddPage("file", grid, true, true)

	return file
}

func (ui *UI) showInfo() {
	if ui.currentDir == nil {
		return
	}

	var content, numberColor string
	row, column := ui.table.GetSelection()
	selectedFile := ui.table.GetCell(row, column).GetReference().(analyze.Item)

	if selectedFile == ui.currentDir.Parent {
		return
	}

	if ui.UseColors {
		numberColor = "[#e67100::b]"
	} else {
		numberColor = "[::b]"
	}

	text := tview.NewTextView().SetDynamicColors(true)
	text.SetBorder(true).SetBorderPadding(2, 2, 2, 2)
	text.SetBorderColor(tcell.ColorDefault)
	text.SetTitle(" Item info ")

	content += "[::b]Name:[::-] " + selectedFile.GetName() + "\n"
	content += "[::b]Path:[::-] " +
		strings.TrimPrefix(selectedFile.GetPath(), build.RootPathPrefix) + "\n"
	content += "[::b]Type:[::-] " + selectedFile.GetType() + "\n\n"

	content += "   [::b]Disk usage:[::-] "
	content += numberColor + ui.formatSize(selectedFile.GetUsage(), false, true)
	content += fmt.Sprintf(" (%s%d[-::] B)", numberColor, selectedFile.GetUsage()) + "\n"
	content += "[::b]Apparent size:[::-] "
	content += numberColor + ui.formatSize(selectedFile.GetSize(), false, true)
	content += fmt.Sprintf(" (%s%d[-::] B)", numberColor, selectedFile.GetSize()) + "\n"

	text.SetText(content)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(text, 13, 1, false).
			AddItem(nil, 0, 1, false), 80, 1, false).
		AddItem(nil, 0, 1, false)

	ui.pages.AddPage("info", flex, true, true)
}
