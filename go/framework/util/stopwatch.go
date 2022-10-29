package util

import (
	"time"
)

const (
	// StopwatchStateKey can be used as a key in the scheduler's state data for storing a single Stopwatch.
	StopwatchStateKey = "polaris-internal.stopwatch"
)

// Stopwatch can be used to measure the time between two instants.
type Stopwatch struct {
	start     time.Time
	stop      time.Time
	isStarted bool
	isStopped bool
}

// NewStopwatch creates a new Stopwatch.
func NewStopwatch() *Stopwatch {
	return &Stopwatch{}
}

// Start sets the current time as the start time of the Stopwatch.
func (me *Stopwatch) Start() {
	me.isStarted = true
	me.start = time.Now()
}

// Stop sets the current time as the stop time of the Stopwatch.
func (me *Stopwatch) Stop() {
	me.stop = time.Now()
	me.isStopped = true
}

// IsStopped returns true if the stopwatch has already been stopped.
func (me *Stopwatch) IsStopped() bool {
	return me.isStopped
}

// Duration returns the duration of the time that was measured by this Stopwatch.
func (me *Stopwatch) Duration() time.Duration {
	return me.stop.Sub(me.start)
}
