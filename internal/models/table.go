package models

import "github.com/rivo/tview"

type Tables struct {
	MemTable *tview.Table
	CpuTable *tview.Table
	NetworkTable *tview.Table
	UptimeTable *tview.Table
	TopMemTable *tview.Table
	TopCpuTable *tview.Table
}