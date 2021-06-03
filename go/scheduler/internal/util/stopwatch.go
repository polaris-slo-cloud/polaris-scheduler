package util

import (
	"time"

	"k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	stopwatchPtr *Stopwatch
	_            framework.StateData = stopwatchPtr
)

const (
	// StopwatchStateKey can be used as a key in the scheduler's state data for storing a single Stopwatch.
	StopwatchStateKey = "Stopwatch"
)

// Stopwatch can be used to measure the time between two instants.
type Stopwatch struct {
	start time.Time
	stop  time.Time
}

// NewStopwatch creates a new Stopwatch.
func NewStopwatch() *Stopwatch {
	return &Stopwatch{}
}

// Start sets the current time as the start time of the Stopwatch.
func (me *Stopwatch) Start() {
	me.start = time.Now()
}

// Stop sets the current time as the stop time of the Stopwatch.
func (me *Stopwatch) Stop() {
	me.stop = time.Now()
}

// Duration returns the duration of the time that was measured by this Stopwatch.
func (me *Stopwatch) Duration() time.Duration {
	return me.stop.Sub(me.start)
}

// Clone creates a shallow copy of this object.
func (me *Stopwatch) Clone() framework.StateData {
	return &Stopwatch{
		start: me.start,
		stop:  me.stop,
	}
}
