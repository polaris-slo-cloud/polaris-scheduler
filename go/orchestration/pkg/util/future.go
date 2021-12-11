package util

import (
	"sync"
)

var (
	_futureImpl *futureImpl

	_ Future = _futureImpl
)

// Callback used to provide a result for a Future.
type ResultProvider func(result interface{}, err error)

// Future allows to retrieve the result of an asynchronous operation by
// a) blocking on Get() until the result is available
// b) getting the result immediately on Get() is the operation has already been completed.
//
// Other options, e.g., registering a handler or checking the status, may be added as needed.
type Future interface {

	// Gets the result of the operation.
	//
	// If the operation is still in progress, the method blocks until the result is available.
	// If the operation is already complete, the result is returned immediately.
	//
	// An error is returned if the asynchronous operation resulted in an error.
	Get() (interface{}, error)
}

type futureImpl struct {
	waitGroup sync.WaitGroup
	opResult  interface{}
	opError   error
}

func (me *futureImpl) Get() (interface{}, error) {
	me.waitGroup.Wait()
	return me.opResult, me.opError
}

func NewFuture() (Future, ResultProvider) {
	future := futureImpl{
		waitGroup: sync.WaitGroup{},
	}
	future.waitGroup.Add(1)

	provideResult := func(result interface{}, err error) {
		future.opResult = result
		future.opError = err
		future.waitGroup.Done()
	}

	return &future, provideResult
}
