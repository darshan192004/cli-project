package progress

import (
	"fmt"
	"strings"
	"time"

	"github.com/gookit/color"
)

type Progress struct {
	Total       int64
	Current     int64
	StartTime   time.Time
	LastUpdate  time.Time
	Width       int
	ShowSpeed   bool
	ShowETA     bool
	ShowPercent bool
}

func New(total int64) *Progress {
	return &Progress{
		Total:       total,
		StartTime:   time.Now(),
		LastUpdate:  time.Now(),
		Width:       50,
		ShowSpeed:   true,
		ShowETA:     true,
		ShowPercent: true,
	}
}

func (p *Progress) Increment() {
	p.Current++
	p.maybeRender()
}

func (p *Progress) Set(current int64) {
	p.Current = current
	p.maybeRender()
}

func (p *Progress) maybeRender() {
	now := time.Now()
	if now.Sub(p.LastUpdate) < 100*time.Millisecond && p.Current < p.Total {
		return
	}
	p.Render()
	p.LastUpdate = now
}

func (p *Progress) Render() {
	elapsed := time.Since(p.StartTime)

	percent := float64(0)
	if p.Total > 0 {
		percent = float64(p.Current) / float64(p.Total) * 100
	}

	filled := int(float64(p.Width) * float64(p.Current) / float64(p.Total))
	if filled > p.Width {
		filled = p.Width
	}
	empty := p.Width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	var speed, eta string
	if p.ShowSpeed && elapsed > 0 {
		recordsPerSec := float64(p.Current) / elapsed.Seconds()
		speed = fmt.Sprintf("%.0f rows/s", recordsPerSec)
	}

	if p.ShowETA && p.Current > 0 && p.Total > 0 {
		remaining := p.Total - p.Current
		rate := float64(p.Current) / elapsed.Seconds()
		if rate > 0 {
			remainingTime := time.Duration(float64(remaining) / rate * float64(time.Second))
			eta = fmt.Sprintf("ETA: %s", remainingTime.Round(time.Second))
		}
	}

	fmt.Printf("\r\033[K")

	if p.ShowPercent {
		color.Cyan.Printf("[%s] %.1f%%", bar, percent)
	} else {
		color.Cyan.Printf("[%s]", bar)
	}

	fmt.Printf(" %d/%d", p.Current, p.Total)

	if speed != "" {
		color.Gray.Printf(" | %s", speed)
	}

	if eta != "" {
		color.Gray.Printf(" | %s", eta)
	}

	if p.Current >= p.Total {
		fmt.Println()
	}
}

func (p *Progress) Finish() {
	p.Current = p.Total
	p.Render()
}

type MultiProgress struct {
	Items []*Progress
}

func NewMulti(items int64) *MultiProgress {
	return &MultiProgress{
		Items: make([]*Progress, 0),
	}
}

func (mp *MultiProgress) Add(total int64) *Progress {
	p := New(total)
	mp.Items = append(mp.Items, p)
	return p
}

type Spinner struct {
	message  string
	frames   []string
	current  int
	interval time.Duration
	running  bool
	stop     chan bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message:  message,
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 100 * time.Millisecond,
		stop:     make(chan bool),
	}
}

func (s *Spinner) Start() {
	s.running = true
	go func() {
		for s.running {
			color.Cyan.Printf("\r\033[K%s %s", s.frames[s.current], s.message)
			s.current = (s.current + 1) % len(s.frames)
			time.Sleep(s.interval)
		}
	}()
}

func (s *Spinner) Stop() {
	s.running = false
	fmt.Printf("\r\033[K")
}

func (s *Spinner) Success(message string) {
	s.Stop()
	color.Green.Printf("✓ %s\n", message)
}

func (s *Spinner) Error(message string) {
	s.Stop()
	color.Red.Printf("✗ %s\n", message)
}

type IndentWriter struct {
	indent string
}

func NewIndent(indent string) *IndentWriter {
	return &IndentWriter{indent: indent}
}

func (iw *IndentWriter) Print(format string, args ...interface{}) {
	fmt.Printf(iw.indent+format, args...)
}

func (iw *IndentWriter) Println(message string) {
	fmt.Println(iw.indent + message)
}

func (iw *IndentWriter) Success(format string, args ...interface{}) {
	color.Green.Printf(iw.indent+"✓ "+format+"\n", args...)
}

func (iw *IndentWriter) Error(format string, args ...interface{}) {
	color.Red.Printf(iw.indent+"✗ "+format+"\n", args...)
}

func (iw *IndentWriter) Warning(format string, args ...interface{}) {
	color.Yellow.Printf(iw.indent+"⚠ "+format+"\n", args...)
}

func (iw *IndentWriter) Info(format string, args ...interface{}) {
	color.Cyan.Printf(iw.indent+"ℹ "+format+"\n", args...)
}
