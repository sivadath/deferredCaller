package deferredCaller

import (
	"sync"
	"time"
)

type DeferredCaller interface {
	StopAndUpdate(duration time.Duration, caller func()) DeferredCaller
	Call() DeferredCaller
}

type callerData struct {
	caller func()
	delay <-chan time.Time
	exit chan struct{}
	runningStatus	bool//to prevent multiple waits on the channel.
}

func (cd *callerData)start() {
	if cd != nil && !cd.runningStatus {
		cd.runningStatus = true
		select {
		case <-cd.delay:
			go cd.caller()
		case <-cd.exit:
		}
	}
}

func (cd *callerData)stop() {
	if cd != nil {
		close(cd.exit)
	}
}

type deferredCallData struct {
	access         sync.Mutex
	tempData       *callerData
}

//StopAndUpdate will stop previously triggered process if any, but not trigger the goroutine release. Routine is released by Start.
func (DCD *deferredCallData) StopAndUpdate(duration time.Duration, caller func()) DeferredCaller {
	DCD.access.Lock()
	DCD.tempData.stop()
	DCD.tempData = &callerData{
		delay:time.After(duration),
		caller:caller,
		exit:make(chan struct{}),
	}
	DCD.access.Unlock()
	return DCD
}

//Added function will be triggered only after this call.
func (DCD *deferredCallData)Call() DeferredCaller {
	DCD.access.Lock()
	go DCD.tempData.start()
	DCD.access.Unlock()
	return DCD
}

func New()DeferredCaller {
	return &deferredCallData{}
}
