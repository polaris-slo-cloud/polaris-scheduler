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

// StartAt sets the specified time as the start time of the Stopwatch.
func (me *Stopwatch) StartAt(startTime time.Time) {
	me.isStarted = true
	me.start = startTime
}

// Stop sets the current time as the stop time of the Stopwatch.
//
// It is explicitly supported to stop a stopwatch multiple times and read the
// duration after every stoppage to get multiple time readings.
func (me *Stopwatch) Stop() {
	me.stop = time.Now()
	me.isStopped = true
}

// IsStarted returns true if the stopwatch has been started.
//
// Note that this is not reset after the stopwatch is stopped.
// A stopped stopwatch has both IsStarted() and IsStopped() return true.
func (me *Stopwatch) IsStarted() bool {
	return me.isStarted
}

// IsStopped returns true if the stopwatch has already been stopped.
func (me *Stopwatch) IsStopped() bool {
	return me.isStopped
}

// Returns the start time of this stopwatch.
// Note that this value only makes sense, if the stopwatch has been started.
func (me *Stopwatch) StartTime() time.Time {
	return me.start
}

// Returns the stop time of this stopwatch.
// Note that this value only makes sense, if the stopwatch has been stopped.
func (me *Stopwatch) StopTime() time.Time {
	return me.stop
}

// Duration returns the duration of the time that was measured by this Stopwatch.
func (me *Stopwatch) Duration() time.Duration {
	return me.stop.Sub(me.start)
}
