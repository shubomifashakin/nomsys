package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shubomifashakin/sysmon/internal/models"
	"github.com/shubomifashakin/sysmon/pkg/utils"
)


func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "sysmon - A terminal based system resource monitor\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: sysmon [flags]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	
	delay:=flag.Int("d",2, "Set the delay between updates(in seconds)")
	totalIterations:=flag.Int("n",0,"Exit sysmon after NUMBER iterations/frame updates")
	flag.Parse()

	ctx,cancel:=context.WithCancel(context.Background())
	defer cancel()

	root,tables:=createRoot()
	app:=tview.NewApplication()

	
	ticker:=time.NewTicker(time.Duration(*delay) * time.Second)

	app=app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}
		switch event.Key() {
		case tcell.KeyCtrlC:
			app.Stop()
			return nil
			
		case tcell.KeyTab:
			if app.GetFocus() == tables.TopCpuTable {
				app.SetFocus(tables.TopMemTable)
			} else {
				app.SetFocus(tables.TopCpuTable)
			}
			return nil
		}
		return event
	})
	
	var wg sync.WaitGroup
	iterations:=0

	wg.Go(func(){
		for  {
			select{
				case <- ctx.Done():
					return
				case <- ticker.C:
				iterations++

				memChan:=make(chan models.MemStats)
				memErrChan:=make(chan error)
		
				go func(){
					memStats,err:=utils.GetMemStats()
					if err!= nil {
						memErrChan<-err
						return
					}
		
					memChan <- memStats
				}()

				cpuStatsChan:=make(chan models.CpuStats)
				cpuErrChan:=make(chan error)
		
				go func(){
					cpuStats,err:=utils.GetCpuStats()
		
					if err!= nil {
						cpuErrChan<-err
						return
					}

					cpuStatsChan<-cpuStats
				}()

				uptimeStatsChan:=make(chan models.UptimeStats)
				uptimeErrChan:=make(chan error)
		
				go func(){
					uptimeStats,err:=utils.GetUptimeStats()
					if err!= nil {
						uptimeErrChan<-err
						return
					}

					uptimeStatsChan<-uptimeStats
				}()

				networkStatsChan:=make(chan models.NetworkStats)
				networkErrChan:=make(chan error)
		
				go func(){
					netStats,err:=utils.GetNetworkStats()
					if err!= nil {
						networkErrChan<-err
						return
					}

					networkStatsChan<-netStats
				}()

				processStatsChan:=make(chan models.ProcessStats)
				processErrChan:=make(chan error)
		
				go func(){
					top5Processes,err:=utils.GetTop20ProcessesByCpuAndMem()
					if err!= nil {
						processErrChan<-err
						return
					}

					processStatsChan<-top5Processes
				}()
		
				var memRes models.MemStats
				var memErr error
				select {
				case memRes = <-memChan:
				case memErr = <-memErrChan:
				}

				var cpuRes models.CpuStats
				var cpuErr error
				select {
				case cpuRes = <-cpuStatsChan:
				case cpuErr = <-cpuErrChan:
				}

				var uptimeRes models.UptimeStats
				var uptimeErr error
				select {
				case uptimeRes = <-uptimeStatsChan:
				case uptimeErr = <-uptimeErrChan:
				}

				var networkRes models.NetworkStats
				var networkErr error
				select {
				case networkRes = <-networkStatsChan:
				case networkErr = <-networkErrChan:
				}

				var processRes models.ProcessStats
				var processErr error
				select {
				case processRes = <-processStatsChan:
				case processErr = <-processErrChan:
				}
				
			// when the app closes dont schedul any update
			if ctx.Err() == nil {
				app.QueueUpdateDraw(func(){
					if memErr == nil {
						updateMemTable(tables.MemTable, memRes)
					}
					if cpuErr == nil {
						updateCpuTable(tables.CpuTable, cpuRes)
					}
					if uptimeErr == nil {
						updateUptimeTable(tables.UptimeTable, uptimeRes)
					}
					if networkErr == nil {
						updateNetworkTable(tables.NetworkTable, networkRes)
					}
					if processErr == nil {
						updateProcessTable(tables.TopMemTable, processRes.Top5ProcessesByMem, "% Mem", func(p *process.Process) string {
							pct, _ := p.MemoryPercent()
							return fmt.Sprintf("%.2f%%", pct)
						})

						updateProcessTable(tables.TopCpuTable, processRes.Top5ProcessesByCpu, "% CPU", func(p *process.Process) string {
							pct, _ := p.CPUPercent()
							return fmt.Sprintf("%.2f%%", pct)
						})
					}
				})
			}

			if *totalIterations > 0 && iterations >= *totalIterations {
				app.Stop()
				cancel()
				ticker.Stop()
				return
			}
			}
		}
	})

	if err := app.SetRoot(root, true).Run(); err != nil {
		log.Fatalln(err)
	}

	cancel()
	ticker.Stop()

	wg.Wait()
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

func updateMemTable(t *tview.Table, s models.MemStats) {
	t.Clear()
	t.SetCell(0, 0, header("Used (MB)"))
	t.SetCell(0, 1, header("Free (MB)"))
	t.SetCell(0, 2, header("Total (MB)"))
	t.SetCell(0, 3, header("% Used"))
	t.SetCell(1, 0, cell(fmt.Sprintf("%d", s.UseRamMb)))
	t.SetCell(1, 1, cell(fmt.Sprintf("%d", s.AvailableRamMb)))
	t.SetCell(1, 2, cell(fmt.Sprintf("%d", s.TotalRamMb)))
	t.SetCell(1, 3, cell(fmt.Sprintf("%.1f%%", s.PercentUsed)))
}

func updateCpuTable(t *tview.Table, s models.CpuStats) {
	t.Clear()
	for col, pct := range s.PercentUsed {
		t.SetCell(0, col, header(fmt.Sprintf("Core %d", col)))
		t.SetCell(1, col, cell(fmt.Sprintf("%.1f%%", pct)))
	}
}

func updateUptimeTable(t *tview.Table, s models.UptimeStats) {
	t.Clear()
	t.SetCell(0, 0, header("Days"))
	t.SetCell(0, 1, header("Hours"))
	t.SetCell(0, 2, header("Minutes"))
	t.SetCell(1, 0, cell(fmt.Sprintf("%d", s.Days)))
	t.SetCell(1, 1, cell(fmt.Sprintf("%d", s.Hours)))
	t.SetCell(1, 2, cell(fmt.Sprintf("%d", s.Minutes)))
}

func updateNetworkTable(t *tview.Table, s models.NetworkStats) {
	t.Clear()
	t.SetCell(0, 0, header("Sent (MB)"))
	t.SetCell(0, 1, header("Recv (MB)"))
	t.SetCell(0, 2, header("TCP Conns"))
	if len(s.Stats) > 0 {
		stat := s.Stats[0]
		t.SetCell(1, 0, cell(fmt.Sprintf("%d", stat.BytesSent>>20)))
		t.SetCell(1, 1, cell(fmt.Sprintf("%d", stat.BytesRecv>>20)))
		t.SetCell(1, 2, cell(fmt.Sprintf("%d", len(s.TotalTcpConnections))))
	}
}

func updateProcessTable(t *tview.Table, processes []*process.Process, metricLabel string, getMetric func(*process.Process) string) {
	r,c:=t.GetSelection()

	t.Clear()
	t.SetCell(0, 0, header("Name"))
	t.SetCell(0, 1, header("PID"))
	t.SetCell(0, 2, header("User"))
	t.SetCell(0, 3, header(metricLabel))
	t.SetCell(0, 4, header("Time"))
	t.SetCell(0,5,header("Status"))

	for i, p := range processes {
		name, _ := p.Name()
		status,_:=p.Status()
		user, _ := p.Username()
		createdTime, _ := p.CreateTime()
		
		if len(status)>0{
			for i,v:=range status{
				trimmed:=strings.TrimSpace(v)
				if trimmed == ""{
					continue
				}
				status[i] = strings.ToUpper(string(trimmed[0]))
			}
		}

		t.SetCell(i+1, 0, cell(name))
		t.SetCell(i+1, 1, cell(fmt.Sprintf("%d", p.Pid)))
		t.SetCell(i+1, 2, cell(user))
		t.SetCell(i+1, 3, cell(getMetric(p)))
		t.SetCell(i+1, 4, cell(utils.ConvertMsToTime(createdTime)))
		t.SetCell(i+1, 5, cell(strings.Join(status,",")))
	}

	t.Select(r,c)
}

func createTable(title string,rowSelectable bool) *tview.Table {
	t := tview.NewTable().SetBorders(true)
	t.SetBorder(true).SetTitle(" " + title + " ").SetTitleAlign(tview.AlignCenter).SetTitleColor(tcell.ColorYellow).SetBackgroundColor(tcell.NewRGBColor(20, 22, 30))
	t.SetSelectable(rowSelectable,false)
	t.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorDeepPink).Foreground(tcell.ColorYellow).Bold(true))
	
	// headers should always be at the top
	t.SetFixed(1,0)
	return t
}

func createRoot() (*tview.Flex,models.Tables){
	memTable    := createTable("Memory",false)
	cpuTable    := createTable("CPU",false)
	networkTable := createTable("Network",false)
	uptimeTable := createTable("Uptime",false)
	topMemTable := createTable("Top 20 by Memory",true)
	topCpuTable := createTable("Top 20 by CPU",true)

	topCpuTable.SetFocusFunc(func() {
		topCpuTable.SetBorderColor(tcell.ColorYellow)
	}).SetBlurFunc(func() {
		topCpuTable.SetBorderColor(tcell.ColorWhite)
	})
	
	topMemTable.SetFocusFunc(func() {
		topMemTable.SetBorderColor(tcell.ColorYellow)
	}).SetBlurFunc(func() {
		topMemTable.SetBorderColor(tcell.ColorWhite)
	})
	

	topRow := tview.NewFlex().SetDirection(tview.FlexColumn).
	AddItem(uptimeTable, 0, 1, false).
		AddItem(memTable, 0, 1, false)

	middleRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(cpuTable, 0, 2, false).
		AddItem(networkTable, 0, 1, false)

	// focus should be given to the cpu table when the bottom row is in focus
	bottomRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(topMemTable, 0, 1, false).
		AddItem(topCpuTable, 0, 1, true)


	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topRow, 10, 0, false).
		AddItem(middleRow, 10, 0, false).
		AddItem(bottomRow, 0, 1, true)

	return root, models.Tables{
		MemTable:     memTable,
		CpuTable:     cpuTable,
		NetworkTable: networkTable,
		UptimeTable:  uptimeTable,
		TopMemTable:  topMemTable,
		TopCpuTable:  topCpuTable,
	}
}