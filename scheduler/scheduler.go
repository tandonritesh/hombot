package scheduler

import (
	"hombot/errors"
	"time"
)

type ITimer interface {
	SetSeconds(seconds int)
	SetMilliSeconds(milliseconds int)
	Start(args interface{}, cb func(time.Time, interface{})) (err int)
	Stop()
}

type sTimer struct {
	timer     *time.Timer
	timerChan <-chan time.Time
	duration  time.Duration
	args      interface{}
}

/*
 * Creates a new instance of Timer object
 */
func NewSchdTimer() ITimer {
	return new(sTimer)
}

func (s *sTimer) timerFunc(args interface{}, cb func(time.Time, interface{})) {
	var t *time.Timer = time.NewTimer(s.duration)
	s.timer = t
	s.timerChan = t.C
	var tm time.Time = <-t.C
	cb(tm, args)
}

func (s *sTimer) SetSeconds(seconds int) {
	s.duration = time.Duration(seconds) * time.Second
}

func (s *sTimer) SetMilliSeconds(milliseconds int) {
	s.duration = time.Duration(milliseconds) * time.Millisecond
}

func (s *sTimer) Start(args interface{}, cb func(time.Time, interface{})) (err int) {
	if s.duration == 0 {
		return errors.TIMER_INVALID_DURATION
	}
	go s.timerFunc(args, cb)
	return errors.SUCCESS
}

func (s *sTimer) Stop() {
	if !s.timer.Stop() {
		<-s.timerChan
	}
}
