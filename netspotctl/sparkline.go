// sparkline.go

// netspotctl is not a subpackage of netspot but it is a command line client
// to manage the netspot server. termui aims to plot interactively the
// data provided by netspot
package main

import (
	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

// StatSparkline is a termui sparkline to monitor statistics
type StatSparkline struct {
	line *widgets.Sparkline
	size int
}

// NewStatSparkline creates a new instance of sparkline monioring stats
func NewStatSparkline(name string, size int) *StatSparkline {
	line := widgets.NewSparkline()
	line.Title = name
	line.Data = make([]float64, 0)
	line.LineColor = ui.ColorRed
	return &StatSparkline{
		line: line,
		size: size,
	}
}

// Update adds a new values to the sparkline
func (s *StatSparkline) Update(val float64) {
	l := len(s.line.Data)
	if l < s.size {
		s.line.Data = append(s.line.Data, val)
	} else if l == s.size {
		s.line.Data = append(s.line.Data[1:], val)
	} else {
		printERROR("sparkline overflow")
	}
}
