package core

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/emicklei/melrose/notify"
)

type Loop struct {
	ctx        Context
	target     []Sequenceable
	isRunning  bool
	mutex      sync.RWMutex
	condition  Condition
	nextPlayAt time.Time
}

func NewLoop(ctx Context, target []Sequenceable) *Loop {
	return &Loop{
		ctx:       ctx,
		target:    target,
		condition: TrueCondition,
	}
}

func (l *Loop) Target() []Sequenceable { return l.target }

func (l *Loop) SetTarget(newTarget []Sequenceable) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.target = newTarget
}

func (l *Loop) IsRunning() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.isRunning
}

func (l *Loop) Storex() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "loop(")
	AppendStorexList(&b, true, l.target)
	fmt.Fprintf(&b, ")")
	return b.String()
}

func (l *Loop) Evaluate(ctx Context) error {
	// create and start a clone
	clone := NewLoop(l.ctx, l.target)
	cond := NoCondition
	if with, ok := ctx.(Conditional); ok {
		cond = with.Condition()
	}
	clone.condition = cond
	if IsDebug() {
		notify.Debugf("loop.eval")
	}
	clone.Play(l.ctx, time.Now())
	return nil
}

// func (l *Loop) Start(d AudioDevice) *Loop {
// 	l.mutex.Lock()
// 	defer l.mutex.Unlock()
// 	if l.isRunning || d == nil {
// 		return l
// 	}
// 	l.isRunning = true
// 	l.reschedule(d, time.Now())
// 	return l
// }

// Inspect is part of Inspectable
func (l *Loop) Inspect(i Inspection) {
	i.Properties["running"] = l.isRunning
}

// in mutex
func (l *Loop) reschedule(d AudioDevice, when time.Time) {
	if !l.isRunning {
		return
	}
	if l.condition != nil && !l.condition() {
		l.isRunning = false
		return
	}
	moment := when
	for _, each := range l.target {
		// after each other
		moment = d.Play(l.condition, each, l.ctx.Control().BPM(), moment)
	}
	if IsDebug() {
		notify.Debugf("core.loop: next=%s", moment.Format("15:04:05.00"))
	}
	// schedule the loop itself so it can play again when Handle is called
	l.nextPlayAt = moment
	d.Schedule(l, moment)
}

func (l *Loop) NextPlayAt() time.Time {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if !l.isRunning {
		return time.Time{}
	}
	return l.nextPlayAt
}

// Handle is part of TimelineEvent
func (l *Loop) Handle(tim *Timeline, when time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.isRunning {
		return
	}
	l.reschedule(l.ctx.Device(), when)
}

// Play is part of Playable
func (l *Loop) Play(ctx Context, at time.Time) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.isRunning {
		return nil
	}
	l.isRunning = true
	l.reschedule(l.ctx.Device(), at)
	return nil
}

// Stop is part of Playable
func (l *Loop) Stop(ctx Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if !l.isRunning {
		return nil
	}
	l.isRunning = false
	return nil
}
