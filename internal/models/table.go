package models

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shubomifashakin/nomsys/pkg/timeutil"
)

type Tables struct {
	MemTable *tview.Table
	CpuTable *tview.Table
	NetworkTable *tview.Table
	UptimeTable *tview.Table
	TopMemTable *tview.Table
	TopCpuTable *tview.Table
}

func header(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetTextColor(tcell.ColorHotPink).
		SetAlign(tview.AlignCenter).
		SetSelectable(false).
		SetExpansion(1).
		SetAttributes(tcell.AttrBold)
}

func cell(text string) *tview.TableCell {
	return tview.NewTableCell(text).
		SetAlign(tview.AlignCenter).
		SetExpansion(1)
}

func (h *Tables)UpdateMemTable(s MemStats) {
	h.MemTable.Clear()
	h.MemTable.SetCell(0, 0, header("Used (MB)"))
	h.MemTable.SetCell(0, 1, header("Free (MB)"))
	h.MemTable.SetCell(0, 2, header("Total (MB)"))
	h.MemTable.SetCell(0, 3, header("% Used"))
	h.MemTable.SetCell(1, 0, cell(fmt.Sprintf("%d", s.UseRamMb)))
	h.MemTable.SetCell(1, 1, cell(fmt.Sprintf("%d", s.AvailableRamMb)))
	h.MemTable.SetCell(1, 2, cell(fmt.Sprintf("%d", s.TotalRamMb)))
	h.MemTable.SetCell(1, 3, cell(fmt.Sprintf("%.1f%%", s.PercentUsed)))
}

func (h *Tables)UpdateCpuTable(s CpuStats) {
	h.CpuTable.Clear()
	for col, pct := range s.PercentUsed {
		h.CpuTable.SetCell(0, col, header(fmt.Sprintf("Core %d", col)))
		h.CpuTable.SetCell(1, col, cell(fmt.Sprintf("%.1f%%", pct)))
	}
}

func (h *Tables)UpdateUptimeTable(s UptimeStats) {
	h.UptimeTable.Clear()
	h.UptimeTable.SetCell(0, 0, header("Days"))
	h.UptimeTable.SetCell(0, 1, header("Hours"))
	h.UptimeTable.SetCell(0, 2, header("Minutes"))
	h.UptimeTable.SetCell(1, 0, cell(fmt.Sprintf("%d", s.Days)))
	h.UptimeTable.SetCell(1, 1, cell(fmt.Sprintf("%d", s.Hours)))
	h.UptimeTable.SetCell(1, 2, cell(fmt.Sprintf("%d", s.Minutes)))
}

func (h *Tables)UpdateNetworkTable(s NetworkStats) {
	h.NetworkTable.Clear()
	h.NetworkTable.SetCell(0, 0, header("Sent (MB)"))
	h.NetworkTable.SetCell(0, 1, header("Recv (MB)"))
	h.NetworkTable.SetCell(0, 2, header("TCP Conns"))
	if len(s.Stats) > 0 {
		stat := s.Stats[0]
		h.NetworkTable.SetCell(1, 0, cell(fmt.Sprintf("%d", stat.BytesSent>>20)))
		h.NetworkTable.SetCell(1, 1, cell(fmt.Sprintf("%d", stat.BytesRecv>>20)))
		h.NetworkTable.SetCell(1, 2, cell(fmt.Sprintf("%d", len(s.TotalTcpConnections))))
	}
}

func (h *Tables) UpdateTopProcessByCpuTable(processes []*process.Process, metricLabel string, getMetric func(*process.Process) string) {
	populateProcessTable(h.TopCpuTable, processes, metricLabel, getMetric)
}

func (h *Tables) UpdateTopProcessByMemTable(processes []*process.Process, metricLabel string, getMetric func(*process.Process) string) {
	populateProcessTable(h.TopMemTable, processes, metricLabel, getMetric)
}

func populateProcessTable(t *tview.Table, processes []*process.Process, metricLabel string, getMetric func(*process.Process) string) {
	r, c := t.GetSelection()

	t.Clear()
	t.SetCell(0, 0, header("Name"))
	t.SetCell(0, 1, header("PID"))
	t.SetCell(0, 2, header("User"))
	t.SetCell(0, 3, header(metricLabel))
	t.SetCell(0, 4, header("Time"))
	t.SetCell(0, 5, header("Status"))

	for i, p := range processes {
		name, _ := p.Name()
		status, _ := p.Status()
		user, _ := p.Username()
		createdTime, _ := p.CreateTime()

		if len(status) > 0 {
			for i, v := range status {
				trimmed := strings.TrimSpace(v)
				if trimmed == "" {
					continue
				}
				status[i] = strings.ToUpper(string(trimmed[0]))
			}
		}

		t.SetCell(i+1, 0, cell(name))
		t.SetCell(i+1, 1, cell(fmt.Sprintf("%d", p.Pid)))
		t.SetCell(i+1, 2, cell(user))
		t.SetCell(i+1, 3, cell(getMetric(p)))
		t.SetCell(i+1, 4, cell(timeutil.ConvertMsToTime(createdTime)))
		t.SetCell(i+1, 5, cell(strings.Join(status, ",")))
	}

	t.Select(r, c)
}

func createTable(title string, rowSelectable bool) *tview.Table {
	t := tview.NewTable().SetBorders(true)
	t.SetBorder(true).SetTitle(" " + title + " ").SetTitleAlign(tview.AlignCenter).SetTitleColor(tcell.ColorYellow).SetBackgroundColor(tcell.NewRGBColor(20, 22, 30))
	t.SetSelectable(rowSelectable, false)
	t.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorDeepPink).Foreground(tcell.ColorYellow).Bold(true))
	t.SetFixed(1, 0)
	return t
}

func NewTables() *Tables {
	return &Tables{
		MemTable:     createTable("Memory", false),
		CpuTable:     createTable("CPU", false),
		NetworkTable: createTable("Network", false),
		UptimeTable:  createTable("Uptime", false),
		TopMemTable:  createTable("Top 20 by Memory", true),
		TopCpuTable:  createTable("Top 20 by CPU", true),
	}
}