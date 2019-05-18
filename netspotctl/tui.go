// tui.go

// netspotctl is not a subpackage of netspot but it is a command line client
// to manage the netspot server. termui aims to plot interactively the
// data provided by netspot
package main

import (
	"time"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

var (
	sparklineMap = make(map[string]*StatSparkline)
)

// UpdateUI updates sparklines according
// to the values received by netspot.
func UpdateUI() []*widgets.Sparkline {
	values := make(map[string]float64)
	i := 0
	call("Netspot.StatValues", &i, &values)

	sl := make([]*widgets.Sparkline, 0)
	for s, v := range values {
		if _, exists := sparklineMap[s]; exists {
			sparklineMap[s].Update(v)
			sl = append(sl, sparklineMap[s].line)
		}
	}
	return sl
}

// InitUI creates the sparlines according to the loaded statistics
func InitUI() {
	ls := loadedStats()
	for _, s := range ls {
		if _, exists := sparklineMap[s]; !exists {
			sparklineMap[s] = NewStatSparkline(s, 15)
		}
	}
}

// RunUI starts the rendering
func RunUI() {
	if err := ui.Init(); err != nil {
		printfERROR("failed to initialize termui: %v", err)
		return
	}
	InitUI()
	// sl := make([]*widgets.Sparkline, 0)
	// for _, s := range sparklineMap {
	// 	sl = append(sl, s.line)
	// }

	// group := widgets.NewSparklineGroup(sl...)
	// ui.Render(group)
	uiEvents := ui.PollEvents()
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			sl := UpdateUI()
			ui.Render(widgets.NewSparklineGroup(sl...))
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Close()
				return
			}
		}
	}
}

// RunUI2 starts the rendering
func RunUI2() {
	width := 30
	if err := ui.Init(); err != nil {
		printfERROR("failed to initialize termui: %v", err)
		return
	}

	stats := loadedStats()
	sl := make([]*widgets.Sparkline, len(stats))
	for i, s := range stats {
		line := widgets.NewSparkline()
		line.Title = s
		line.Data = make([]float64, 0, width)
		line.LineColor = ui.ColorRed
		sl[i] = line
		// if _, exists := sparklineMap[s]; !exists {
		// 	sparklineMap[s] = NewStatSparkline(s, 15)
		// 	sl[i] = sparklineMap[s].line
		// }
	}

	group := widgets.NewSparklineGroup(sl...)
	group.SetRect(0, 0, width, 20)

	display := func() {
		values := make(map[string]float64)
		i := 0
		call("Netspot.StatValues", &i, &values)

		for _, line := range group.Sparklines {
			l := len(line.Data)
			if l < width {
				line.Data = append(line.Data, values[line.Title])
			} else if l == width {
				line.Data = append(line.Data[1:], values[line.Title])
			} else {
				printERROR("sparkline overflow")
			}
		}

		ui.Render(group)
	}
	uiEvents := ui.PollEvents()
	ticker := time.Tick(1 * time.Second)
	for {
		select {
		case <-ticker:
			display()
			// sl := UpdateUI()
			// ui.Render(widgets.NewSparklineGroup(sl...))
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Close()
				return
			}
		}
	}
}
