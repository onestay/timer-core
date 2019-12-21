package timer

import "time"

import "fmt"

type subtimer struct {
	Time  time.Duration
	state State
}

// AddSubTimer adds a timer with an id to the subtimer pool
// id has to be unique and can only be added when timer is in reset state
func (t *Timer) AddSubTimer(id int) error {
	if t.State != Reset {
		return fmt.Errorf("Subtimer can only be added when timer is in reset state")
	}
	if _, ok := t.subtimers[id]; ok {
		return fmt.Errorf("Subtimer with key %v already exists", id)
	}
	s := subtimer{}
	s.state = Reset
	t.subtimers[id] = &s

	return nil
}

// StopSubTimer will stop a specific subtimer
// only works when subtimer and timer are running
func (t *Timer) StopSubTimer(id int) (time.Duration, error) {
	s, ok := t.subtimers[id]
	if !ok {
		return time.Duration(0), fmt.Errorf("Subtimer with id %v does not exist", id)
	}
	if t.State == Running && s.state == Running {

	}
	s.state = Stopped
	s.Time = t.elapsed

	if t.stopOnSubtimersFinish && t.checkSubTimerFinish() {
		t.StopTimer()
	}

	return s.Time, nil
}

func (t *Timer) checkSubTimerFinish() bool {
	for _, s := range t.subtimers {
		if s.state != Stopped {
			return false
		}
	}

	return true
}

func (t *Timer) startSubTimers() {
	if len(t.subtimers) <= 0 {
		return
	}

	for _, s := range t.subtimers {
		s.state = Running
	}
}
