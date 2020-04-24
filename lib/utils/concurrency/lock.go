/*
 * Copyright 2018-2020, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package concurrency

import (
	"fmt"

	"sync"

	"github.com/CS-SI/SafeScale/lib/utils/debug"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
	"github.com/sirupsen/logrus"
)

//go:generate mockgen -destination=../mocks/mock_taskedlock.go -package=mocks github.com/CS-SI/SafeScale/lib/utils/concurrency TaskedLock

type TaskedLockHelpers interface {
	SafeLock(Task)    // To be used when instance and parameter are notoriously not nil
	SafeUnlock(Task)  // To be used when instance and parameter are notoriously not nil
	SafeRLock(Task)   // To be used when instance and parameter are notoriously not nil
	SafeRUnlock(Task) // To be used when instance and parameter are notoriously not nil
}

// TaskedLock ...
type TaskedLock interface {
	RLock(Task) error
	RUnlock(Task) error
	Lock(Task) error
	Unlock(Task) error

	IsRLocked(Task) (bool, error)
	IsLocked(Task) (bool, error)

	TaskedLockHelpers
}

// taskedLock ...
type taskedLock struct {
	mu         *sync.Mutex
	rwmutex    *sync.RWMutex
	readLocks  map[string]uint64
	writeLocks map[string]uint64
}

// NewTaskedLock ...
func NewTaskedLock() TaskedLock {
	return &taskedLock{
		mu:         &sync.Mutex{},
		rwmutex:    &sync.RWMutex{},
		readLocks:  map[string]uint64{},
		writeLocks: map[string]uint64{},
	}
}

// RLock locks for read in the context if:
// 1. registers the mu for read only if a mu for write is already registered in the context
// 2. registers the mu for read AND effectively mu for read otherwise
func (tm *taskedLock) RLock(task Task) error {
	if tm == nil {
		return scerr.InvalidInstanceError()
	}
	if task == nil {
		return scerr.InvalidParameterError("task", "cannot be nil!")
	}

	tracer := NewTracer(task, debug.ShouldTrace("concurrency.lock"), "")
	defer tracer.Entering().OnExitTrace()

	tid, err := task.GetID()
	if err != nil {
		return err
	}

	// logrus.Warnf("Calling rlock from %d, with tid %s", goid(), tid)
	tm.mu.Lock()
	access := false
	defer func() {
		tm.mu.Unlock()
		if access {
			tm.rwmutex.RLock()
		}
	}()

	if _, ok := tm.readLocks[tid]; ok {
		tm.readLocks[tid]++
		return nil
	}
	tm.readLocks[tid] = 1
	if _, ok := tm.writeLocks[tid]; !ok {
		tracer.Trace("really RLocking...")
		access = true
		return nil
	}
	tracer.Trace("using running write mu...")
	return nil
}

// SafeRUnlock ...
func (tm *taskedLock) SafeRLock(task Task) {
	err := tm.RLock(task)
	if err != nil {
		logrus.Errorf(scerr.Wrap(err, "cannot use SafeRLock() when obviously it's not safe").Error())
	}
}

// RUnlock unregisters the mu for read for the context and unlock for read
// only if no mu for write is registered for the context
func (tm *taskedLock) RUnlock(task Task) (err error) {
	tracer := NewTracer(task, debug.ShouldTrace("concurrency.lock"), "").Entering()
	defer tracer.OnExitTrace()
	defer scerr.OnExitLogError("", &err)

	if tm == nil {
		return scerr.InvalidInstanceError()
	}
	if task == nil {
		return scerr.InvalidParameterError("task", "cannot be nil")
	}

	tid, err := task.GetID()
	if err != nil {
		return err
	}

	// logrus.Warnf("Calling runlock from %d, with tid %s", goid(), tid)
	tm.mu.Lock()
	access := false
	defer func() {
		tm.mu.Unlock()
		if access {
			tm.rwmutex.RUnlock()
		}
	}()

	if _, ok := tm.readLocks[tid]; !ok {
		return scerr.ForbiddenError(fmt.Sprintf("cannot RUnlock task %s: not RLocked", tid))
	}
	tm.readLocks[tid]--
	if tm.readLocks[tid] == 0 {
		delete(tm.readLocks, tid)
		// If not locked for write, actively unlock for read the RWMutex
		if _, ok := tm.writeLocks[tid]; ok {
			tracer.Trace("in running write mu, doing nothing")
		} else {
			tracer.Trace("really RUnlocking...")
			access = true
		}
	}

	return nil
}

// SafeRUnlock ...
func (tm *taskedLock) SafeRUnlock(task Task) {
	err := tm.RUnlock(task)
	if err != nil {
		logrus.Errorf(scerr.Wrap(err, "cannot use SafeRUnlock() when obviously it's not safe").Error())
	}
}

// Lock acquires a write mu.
func (tm *taskedLock) Lock(task Task) error {
	tracer := NewTracer(task, debug.ShouldTrace("concurrency.lock"), "").Entering()
	defer tracer.OnExitTrace()

	if task == nil {
		return scerr.InvalidParameterError("task", "cannot be nil")
	}

	tid, err := task.GetID()
	if err != nil {
		return err
	}

	// logrus.Warnf("Calling mu from %d, with tid %s", goid(), tid)
	tm.mu.Lock()
	access := false
	defer func() {
		tm.mu.Unlock()
		if access {
			tm.rwmutex.Lock()
		}
	}()

	// If already locked for write, increments counter for the task
	if _, ok := tm.writeLocks[tid]; ok {
		tm.writeLocks[tid]++
		return nil
	}
	// If already mu for read, returns an error
	if _, ok := tm.readLocks[tid]; ok {
		tracer.Trace("Cannot Lock, already RLocked")
		taskID, _ := task.GetID()
		return scerr.ForbiddenError(fmt.Sprintf("cannot Lock task '%s': already RLocked", taskID))
	}
	// registers mu for read for the task and actively mu the RWMutex
	tm.writeLocks[tid] = 1
	access = true
	return nil
}

func (tm *taskedLock) SafeLock(task Task) {
	err := tm.Lock(task)
	if err != nil {
		logrus.Errorf(scerr.Wrap(err, "cannot use SafeLock() when obviously it's not safe").Error())
	}
}

// Unlock releases a write mu
func (tm *taskedLock) Unlock(task Task) error {
	tracer := NewTracer(task, debug.ShouldTrace("concurrency.lock"), "").Entering()
	defer tracer.OnExitTrace()

	if task == nil {
		return scerr.InvalidParameterError("task", "cannot be nil!")
	}

	tid, err := task.GetID()
	if err != nil {
		return err
	}

	// logrus.Warnf("Calling unlock from %d, with tid %s", goid(), tid)
	tm.mu.Lock()
	access := false
	defer func() {
		tm.mu.Unlock()
		if access {
			tm.rwmutex.Unlock()
		}
	}()

	// a TaskedLock can be Locked then RLocked without problem,
	// but RUnlocks must have been done before Unlock.
	if _, ok := tm.readLocks[tid]; ok {
		return scerr.ForbiddenError(fmt.Sprintf("cannot Unlock task '%s': %d remaining RLock inside", tid, tm.readLocks[tid]))
	}
	if _, ok := tm.writeLocks[tid]; !ok {
		return scerr.ForbiddenError(fmt.Sprintf("cannot Unlock task '%s': not Locked", tid))
	}
	tm.writeLocks[tid]--
	if tm.writeLocks[tid] == 0 {
		delete(tm.writeLocks, tid)
		access = true
	}
	return nil
}

// SafeUnlock unlocks a previously lock without possible failure, but will log if something is wrong
func (tm *taskedLock) SafeUnlock(task Task) {
	err := tm.Unlock(task)
	if err != nil {
		logrus.Errorf(scerr.Wrap(err, "cannot use SafeUnlock() when obviously it's not safe").Error())
	}
}

// IsRLocked tells if the task is owning a read lock
func (tm *taskedLock) IsRLocked(task Task) (bool, error) {
	if tm == nil {
		return false, scerr.InvalidInstanceError()
	}
	if task == nil {
		return false, scerr.InvalidParameterError("task", "cannot be nil")
	}

	tid, err := task.GetID()
	if err != nil {
		return false, err
	}
	tm.rwmutex.RLock()
	defer tm.rwmutex.RUnlock()
	_, ok := tm.readLocks[tid]
	return ok, nil
}

// IsLocked tells if the task is owning a write lock
func (tm *taskedLock) IsLocked(task Task) (bool, error) {
	if tm == nil {
		return false, scerr.InvalidInstanceError()
	}
	if task == nil {
		return false, scerr.InvalidParameterError("task", "cannot be nil")
	}

	tid, err := task.GetID()
	if err != nil {
		return false, err
	}

	tm.rwmutex.RLock()
	defer tm.rwmutex.RUnlock()
	_, ok := tm.writeLocks[tid]
	return ok, nil
}
