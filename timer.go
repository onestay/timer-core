package timer

import "fmt"

import "time"

// State describes the different Timerstates
type State int

const (
	// Reset represents a timer which has been initialized but is currently not running
	Reset State = iota
	// Running represents a running timer
	Running
	// Paused represents a paused timer
	Paused
	// Stopped represents a stopped timer
	Stopped
)

const (
	defaultUpdateInterval = 10
	defaultTickerInterval = 10
)

// Timer is the main struct holding all relevant data
type Timer struct {
	// internal ticker
	updateInterval int
	tickerInterval int
	ticker         *time.Ticker
	updateTicker   *time.Ticker
	// public
	State   State
	Updates chan time.Duration
	// internal state
	startTime time.Time
	elapsed   time.Duration
	pauseTime time.Time
	subtimers map[int]*subtimer
	// internal config
	continueCountingWhenStopped bool
	stopOnSubtimersFinish       bool
}

// New initializes and returns a new timer with
func New() *Timer {
	return &Timer{
		updateInterval: defaultUpdateInterval,
		tickerInterval: defaultTickerInterval,
		State:          Stopped,
		Updates:        make(chan time.Duration),
		subtimers:      make(map[int]*subtimer),
	}
}

// SetUpdateInterval sets a new updateInterval for the timer
// Only works when timer is stopped. Setting 0 for updateInterval sets it back to the default
func (t *Timer) SetUpdateInterval(updateInterval int) error {
	if t.State != Stopped {
		return fmt.Errorf("UpdateInterval can only be changed when timer is stopped")
	}
	if updateInterval < 0 {
		return fmt.Errorf("Only positive values for updateInterval are allowed")
	}

	if updateInterval == 0 {
		t.updateInterval = defaultUpdateInterval
		return nil
	}
	t.updateInterval = updateInterval

	return nil
}

// StartTimer starts the timer
// only possible when timer is in Reset state
func (t *Timer) StartTimer() {
	t.State = Running
	t.ticker = time.NewTicker(time.Duration(t.tickerInterval) * time.Millisecond)
	t.updateTicker = time.NewTicker(time.Duration(t.updateInterval) * time.Millisecond)
	t.startTime = time.Now()
	go t.startSubTimers()
	go t.timerLoop()
}

// StopTimer stops the timer
// only possible when in Running state
func (t *Timer) StopTimer() {
	t.State = Stopped
	t.ticker.Stop()
	t.updateTicker.Stop()
}

// ResetTimer resets the timer to it's default state
// only possible when in Stopped state
func (t *Timer) ResetTimer() {
	t.subtimers = make(map[int]*subtimer)
	t.State = Reset
	t.ticker = nil
	t.updateTicker = nil
}

// PauseTimer timer pauses the timer
// only possible when in Running state
func (t *Timer) PauseTimer() {
	t.State = Paused
	t.pauseTime = time.Now()
}

// ResumeTimer resumes the timer from a paused state
// only possible when in Paused or Stopped state
func (t *Timer) ResumeTimer() {
	t.startTime = t.startTime.Add(time.Since(t.pauseTime))
	go t.timerLoop()
	t.State = Running
}

func (t *Timer) timerLoop() {
	for {
		select {
		case <-t.ticker.C:

			if t.State == Running {
				t.elapsed = time.Now().Sub(t.startTime)
				t.Updates <- t.elapsed
			}
		}
	}
}
