/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sync2

// What's in a name? Channels have all you need to emulate a counting
// semaphore with a boatload of extra functionality. However, in some
// cases, you just want a familiar API.

import (
	"context"
)

// Semaphore is a counting semaphore with the option to
// specify a timeout.
type Semaphore struct {
	slots chan struct{}
}

// NewSemaphore creates a Semaphore. The count parameter must be a positive
// number. A timeout of zero means that there is no timeout.
func NewSemaphore(count int) *Semaphore {
	sem := &Semaphore{
		slots: make(chan struct{}, count),
	}
	for i := 0; i < count; i++ {
		sem.slots <- struct{}{}
	}
	return sem
}

// Acquire returns nil on successful acquisition, and
// error on a timeout.
func (sem *Semaphore) Acquire(ctx context.Context) error {
	select {
	case <-sem.slots:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TryAcquire acquires a semaphore if it's immediately available.
// It returns false otherwise.
func (sem *Semaphore) TryAcquire() bool {
	select {
	case <-sem.slots:
		return true
	default:
		return false
	}
}

// Release releases the acquired semaphore. You must
// not release more than the number of semaphores you've
// acquired.
func (sem *Semaphore) Release() {
	sem.slots <- struct{}{}
}

// Size returns the current number of available slots.
func (sem *Semaphore) Size() int {
	return len(sem.slots)
}
