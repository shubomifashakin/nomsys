package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shubomifashakin/nomsys/internal/models"
	"github.com/shubomifashakin/nomsys/pkg/utils"
)


func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "nomsys - A terminal based system resource monitor\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: nomsys [flags]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	
	delay:=flag.Int("d",2, "Set the delay between updates(in seconds)")
	totalIterations:=flag.Int("n",0,"Exit nomsys after NUMBER iterations/frame updates")
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
						tables.UpdateMemTable(memRes)
					}
					if cpuErr == nil {
						tables.UpdateCpuTable(cpuRes)
					}
					if uptimeErr == nil {
						tables.UpdateUptimeTable(uptimeRes)
					}
					if networkErr == nil {
						tables.UpdateNetworkTable(networkRes)
					}
					if processErr == nil {
						tables.UpdateTopProcessByMemTable(processRes.Top5ProcessesByMem, "% Mem", func(p *process.Process) string {
							pct, _ := p.MemoryPercent()
							return fmt.Sprintf("%.2f%%", pct)
						})

						tables.UpdateTopProcessByCpuTable(processRes.Top5ProcessesByCpu, "% CPU", func(p *process.Process) string {
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


func createRoot() (*tview.Flex, *models.Tables){
	tables:=models.NewTables()

	tables.TopCpuTable.SetFocusFunc(func() {
		tables.TopCpuTable.SetBorderColor(tcell.ColorYellow)
	}).SetBlurFunc(func() {
		tables.TopCpuTable.SetBorderColor(tcell.ColorWhite)
	})
	
	tables.TopMemTable.SetFocusFunc(func() {
		tables.TopMemTable.SetBorderColor(tcell.ColorYellow)
	}).SetBlurFunc(func() {
		tables.TopMemTable.SetBorderColor(tcell.ColorWhite)
	})

	topRow := tview.NewFlex().SetDirection(tview.FlexColumn).
	AddItem(tables.UptimeTable, 0, 1, false).
		AddItem(tables.MemTable, 0, 1, false)

	middleRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tables.CpuTable, 0, 2, false).
		AddItem(tables.NetworkTable, 0, 1, false)

	// focus should be given to the cpu table when the bottom row is in focus
	bottomRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(tables.TopMemTable, 0, 1, false).
		AddItem(tables.TopCpuTable, 0, 1, true)


	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(topRow, 10, 0, false).
		AddItem(middleRow, 10, 0, false).
		AddItem(bottomRow, 0, 1, true)

	return root, tables
}