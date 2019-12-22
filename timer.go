package timer

import "fmt"

import "time"

// State describes the different Timerstates
type State int
type operation int

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
	resetOp = iota
	startOp
	pauseOp
	resumeOp
	stopOp
)

const (
	defaultUpdateInterval = 10
	defaultTickerInterval = 10
)

// Config allows configuring various settings when creating a new timer
type Config struct {
	// AllowContinueAfterStop will allow resumeing the timer after it has been stopped
	AllowResumeAfterStop bool
	// ContinueCountingWhenStopped sets how the timer behaves after it is resumed after stopping
	// if set to true the timer will continue counting in the background. For example timer is stopped after 1s and then resumed after 2s.
	// The timer will now continue counting from 3s instead from 1s. This option exists for accidental stopping of the timer
	ContinueCountingWhenStopped bool
	// StopOnSubtimersFinish will stop the timer when all subtimers are set to stop
	StopOnSubtimersStop bool
}

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
	allowResumeAfterStop        bool
	continueCountingWhenStopped bool
	stopOnSubtimersStop         bool
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
func (t *Timer) StartTimer() error {
	if !t.checkValidState(startOp) {
		return fmt.Errorf("StartTimer called with invalid state")
	}

	t.State = Running
	t.ticker = time.NewTicker(time.Duration(t.tickerInterval) * time.Millisecond)
	t.updateTicker = time.NewTicker(time.Duration(t.updateInterval) * time.Millisecond)
	t.startTime = time.Now()
	go t.startSubTimers()
	go t.timerLoop()

	return nil
}

// StopTimer stops the timer
// only possible when in Running state
func (t *Timer) StopTimer() error {
	if !t.checkValidState(stopOp) {
		return fmt.Errorf("StopTimer called with invalid state")
	}

	t.State = Stopped
	t.ticker.Stop()
	t.updateTicker.Stop()

	return nil
}

// ResetTimer resets the timer to it's default state
// only possible when in Stopped state
func (t *Timer) ResetTimer() error {
	if !t.checkValidState(resetOp) {
		return fmt.Errorf("ResetTimer called with invalid state")
	}

	t.subtimers = make(map[int]*subtimer)
	t.State = Reset
	t.ticker = nil
	t.updateTicker = nil

	return nil
}

// PauseTimer timer pauses the timer
// only possible when in Running state
func (t *Timer) PauseTimer() error {
	if !t.checkValidState(pauseOp) {
		return fmt.Errorf("PauseTimer called with invalid state")
	}
	t.State = Paused
	t.pauseTime = time.Now()

	return nil
}

// ResumeTimer resumes the timer from a paused state
// only possible when in Paused or Stopped state
func (t *Timer) ResumeTimer() error {
	if t.State == Paused {
		t.resumeAfterPause()
	} else if t.State == Stopped && t.allowResumeAfterStop {
		t.resumeAfterStop()
	}

	return nil
}

func (t *Timer) resumeAfterStop() {

}

func (t *Timer) resumeAfterPause() {
	t.startTime = t.startTime.Add(time.Since(t.pauseTime))
	t.State = Running
	go t.timerLoop()
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

func (t *Timer) checkValidState(op operation) bool {
	switch op {
	case resetOp:
		return t.State == Stopped
	case startOp:
		return t.State == Reset
	case pauseOp:
		return t.State == Running
	case resumeOp:
		return t.State == Paused || (t.State == Stopped && t.allowResumeAfterStop)
	case stopOp:
		return t.State == Running || t.State == Paused
	default:
		return false
	}
}
