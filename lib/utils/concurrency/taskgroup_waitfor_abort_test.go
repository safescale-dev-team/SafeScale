/*
 * Copyright 2018-2021, CS Systemes d'Information, http://csgroup.eu
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
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CS-SI/SafeScale/lib/utils/fail"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

// Well, let's run a few tasks that, after listening to ABORT signal, quit, but they quit
// after a while, let's say they do a cleanup before exiting ;)
// this is also one of those tests that when tested in a demo, it works..., so we run it 50 times to make sure we see the problem.
// it's better to run this test without race -> too many logs and warnings, hunting data races here doesn't change the outcome,
// it only clouds the real issues
// We want to Wait for our children, and when Abort actually comes, then wait until the children have finished
// and then quit
// This is not what happens (even if that's the easy case where children actually listen and don't block themselves fighting for resources)...
// Let's take a look...
func TestAbortThingsThatActuallyTakeTimeCleaningUpWhenWeAlreadyStartedWaitingFor(t *testing.T) {
	streak := 0
	enough := false
	iter := 0
	chansize := 10
	for {
		iter++
		if iter > 12 {
			break
		}
		if enough {
			break
		}

		t.Log("Next") // Each time we iterate we see this line, sometimes this doesn't fail at 1st iteration
		overlord, xerr := NewTaskGroup()
		require.NotNil(t, overlord)
		require.Nil(t, xerr)
		xerr = overlord.SetID(fmt.Sprintf("/parent-%d", iter))
		require.Nil(t, xerr)

		bailout := make(chan string, chansize) // a buffered channel
		for ind := 0; ind < chansize; ind++ {  // with the same number of tasks, good
			_, xerr = overlord.Start(
				func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
					weWereAborted := false
					for { // do some work, then look for aborted, again and again
						// some work
						time.Sleep(time.Duration(RandomInt(20, 30)) * time.Millisecond)
						if t.Aborted() {
							// Cleaning up first before leaving... ;)
							time.Sleep(time.Duration(RandomInt(100, 800)) * time.Millisecond)
							weWereAborted = true
							break
						}
					}

					// We are using the classic 'send on closed channel' trick to see if Wait actually waits until everyone is DONE.
					// If it does we will never see a panic, but if Abort doesn't mean TellYourChildrenToAbort but
					// actually means AbortYourChildrenAndQuitNOWWithoutWaiting, then we have a problem
					acha := parameters.(chan string)
					acha <- "Bailing out"

					if weWereAborted {
						return "", fail.AbortedError(nil, "we were killed")
					}

					return "who cares", nil
				}, bailout, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)),
			)
			require.Nil(t, xerr)
		}

		// after this, some tasks will already be looking for ABORT signals
		time.Sleep(time.Duration(65) * time.Millisecond)

		go func() {
			// this will actually start after wait
			time.Sleep(time.Duration(100) * time.Millisecond)

			// let's have fun
			xerr := overlord.Abort()
			require.Nil(t, xerr)

			// did we abort ?
			aborted := overlord.Aborted()
			if !aborted {
				t.Logf("We just aborted without error above..., why Aborted() says it's not ?")
			}
		}()

		_, res, xerr := overlord.WaitFor(5 * time.Second) // 100 ms after this, .Abort() should hit
		if xerr != nil {
			t.Logf("Failed to Wait: %s", xerr.Error()) // Of course, we did !!, we induced a panic !! didn't we ?
			switch xerr.(type) {
			case *fail.ErrAborted:
				cause := xerr.Cause()
				// VPL: why expecting a panic when there is no way to have one if everything is working ?
				// this code can be removed safely IF AND ONLY IF test TestAbortThingsThatActuallyTakeTimeCleaningUpAndMayPanicWhenWeAlreadyStartedWaiting
				// proves than panics intended or unintended on a Task/Subtask/TaskGroup are detected by Start/Wait/Abort functions
				// sometimes previous test works..., sometimes fails..., until panic handling is reliable, no need to remove this
				if !strings.Contains(spew.Sdump(cause), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			// or maybe we were fast enough and we are quitting only because of Abort, but no problem, we have more iterations...
			// VPL: there is no way to get a *fail.ErrRuntimePanic from TaskGroup...
			// yet..., all we have to do is, say add a new feature (that also breaks Wait behavior, mistakes happen... ) and then this case will catch the error for us...
			// remove this (and also the default case), and this test no longer protects us against unintended errors
			case *fail.ErrRuntimePanic:
				t.Errorf("That shouldn't ever happen")
				t.FailNow()
			case *fail.ErrorList:
				// VPL: again, why expecting a panic when there is no way to have one if everything is working ?
				// indeed, the test now works, but there were 2 problems:
				// -wait didn't wait until the end (and the test induced a panic to prove it); this problem is fixed
				// -panic goes unnoticed from a client point of view, only looking at logs can be noticed, that's a bug
				// still unfixed as TestAbortThingsThatActuallyTakeTimeCleaningUpAndMayPanicWhenWeAlreadyStartedWaiting proves,
				// when the latter test is fixed, we can safely remove Logf lines below
				if !strings.Contains(spew.Sdump(xerr), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			default:
				t.Errorf("Unexpected error: %v", xerr)
				t.FailNow()
			}
		} else {
			require.NotNil(t, res)
			if len(bailout) == chansize {
				streak++
				if streak > 5 {
					break
				}
				continue
			}
		}
		close(bailout) // If Wait actually waits, this is closed AFTER all Tasks filled the channel, so no panics
		// If not..., well...

		reminder := false
		if len(bailout) != chansize { // this means panic
			reminder = true
			t.Errorf("Not everyone finished on time !!, panic is coming !!, some tasks will hit a closed channel !!")
			// if we now do a t.FailNow() we already proved our point (if Wait actually waited, the channel
			// size should be chansize each time), but if we don't...
			// we will see runtime panics on our LOGS !!, but NOT in the code
			// with a t.FailNow() we also fail, but the test output is less frightening
			enough = true
		}

		time.Sleep(2 * time.Second)
		if reminder {
			t.Errorf("by now we should see panics in lines above, panics that only shows in logs and the rest of the code is unaware of")
		}
		// Well, we have a problem Waiting, now it's clear, and as a bonus we uncovered a problem communicating panics to function callers
	}
}

// Like the previous test, but adding panics into it
func TestAbortThingsThatActuallyTakeTimeCleaningUpAndMayPanicWhenWeAlreadyStartedWaitingFor(t *testing.T) {
	caught := false
	enough := false
	streak := 0
	iter := 0
	chansize := 20

	var failureCounter int32
	var cleanCounter int32

	for {
		iter++
		if iter > 20 {
			break
		}
		if enough || caught {
			break
		}

		t.Log("Next") // Each time we iterate we see this line, sometimes this doesn't fail at 1st iteration
		overlord, xerr := NewTaskGroup()
		require.NotNil(t, overlord)
		require.Nil(t, xerr)
		xerr = overlord.SetID(fmt.Sprintf("/parent-%d", iter))
		require.Nil(t, xerr)

		bailout := make(chan string, chansize) // a buffered channel
		for ind := 0; ind < chansize; ind++ {  // with the same number of tasks, good
			_, xerr = overlord.Start(
				func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
					weWereAborted := false
					for { // do some work, then look for aborted, again and again
						// some work
						time.Sleep(time.Duration(RandomInt(20, 30)) * time.Millisecond)
						if t.Aborted() {
							// Cleaning up first before leaving... ;)
							time.Sleep(time.Duration(RandomInt(100, 800)) * time.Millisecond)
							weWereAborted = true
							break
						}
					}

					// We are using the classic 'send on closed channel' trick to see if Wait actually waits until everyone is DONE.
					// If it does we will never see a panic, but if Abort doesn't mean TellYourChildrenToAbort but
					// actually means AbortYourChildrenAndQuitNOWWithoutWaiting, then we have a problem
					acha := parameters.(chan string)
					acha <- "Bailing out"

					// we throw a loaded dice, 70% of the time we should have a panic
					if RandomInt(0, 10) > 3 {
						atomic.AddInt32(&failureCounter, 1)
						panic("head")
					}
					// tails

					if weWereAborted {
						return "", fail.AbortedError(nil, "we were killed")
					}

					return "who cares", nil
				}, bailout, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)),
			)
			require.Nil(t, xerr)
		}

		// after this, some tasks will already be looking for ABORT signals
		time.Sleep(time.Duration(65) * time.Millisecond)

		go func() {
			// this will actually start after wait
			time.Sleep(time.Duration(100) * time.Millisecond)

			// let's have fun
			xerr := overlord.Abort()
			require.Nil(t, xerr)

			// did we abort ?
			aborted := overlord.Aborted()
			if !aborted {
				t.Logf("We just aborted without error above..., why Aborted() says it's not ?")
			}
		}()

		res, xerr := overlord.WaitGroup() // 100 ms after this, .Abort() should hit
		if xerr != nil {
			t.Logf("Failed to Wait: %s", xerr.Error()) // Of course, we did !!, we induced a panic !! didn't we ?
			switch xerr.(type) {
			case *fail.ErrAborted:
				cause := xerr.Cause()
				if !strings.Contains(spew.Sdump(cause), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			// or maybe we were fast enough and we are quitting only because of Abort, but no problem, we have more iterations...
			case *fail.ErrRuntimePanic: // This MUST NEVER HAPPEN in a TaskGroup; the panic should be in the ErrorList returned by Wait()
				t.Errorf("That shouldn't happen")
				t.FailNow()
			case *fail.ErrorList:
				if !strings.Contains(spew.Sdump(xerr), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				} else {
					t.Logf("We catched a panic..., good")
					caught = true
					atomic.AddInt32(&cleanCounter, 1)
					break
				}
			default:
				t.Errorf("Unexpected error: %v", xerr)
			}
		} else {
			require.NotNil(t, res)
			if len(bailout) == chansize {
				streak++
				if streak > 5 {
					break
				}
				continue
			}
		}
		close(bailout) // If Wait actually waits, this is closed AFTER all Tasks filled the channel, so no panics
		// If not..., well...

		reminder := false
		if len(bailout) != chansize { // this means panic
			reminder = true
			t.Errorf("Not everyone finished on time !!, panic is coming !!, some tasks will hit a closed channel !!")
			// if we now do a t.FailNow() we already proved our point (if Wait actually waited, the channel
			// size should be chansize each time), but if we don't...
			// we will see runtime panics on our LOGS !!, but NOT in the code
			// with a t.FailNow() we also fail, but the test output is less frightening
			enough = true
		}

		time.Sleep(800 * time.Millisecond)
		if reminder {
			t.Errorf("by now we should see panics in lines above, panics that only shows in logs and the rest of the code is unaware of")
		}
		// Well, we have a problem Waiting, now it's clear, and as a bonus we uncovered a problem communicating panics to function callers
	}

	if !caught {
		t.Errorf("We were unable to catch a panic..., and we generated %d", failureCounter)
		t.FailNow()
	}

	if failureCounter != cleanCounter {
		t.Errorf("Not all panics were caught: %d missing panics", failureCounter-cleanCounter)
		t.FailNow()
	}

	if iter > 6 {
		t.Errorf("Even if it succeeds form time to time, NO way it works as intented, compare this with next test that detects the panic in a few seconds")
		t.FailNow()
	}
}

// Like the previous test, without .Abort()
func TestThingsThatActuallyTakeTimeCleaningUpAndMayPanicWhenWeAlreadyStartedWaitingFor(t *testing.T) {
	caught := false
	enough := false
	streak := 0
	iter := 0
	chansize := 20
	for {
		iter++
		if iter > 20 {
			break
		}
		if enough || caught {
			break
		}

		t.Log("Next") // Each time we iterate we see this line, sometimes this doesn't fail at 1st iteration
		overlord, xerr := NewTaskGroup()
		require.NotNil(t, overlord)
		require.Nil(t, xerr)
		xerr = overlord.SetID(fmt.Sprintf("/parent-%d", iter))
		require.Nil(t, xerr)

		bailout := make(chan string, chansize) // a buffered channel
		for ind := 0; ind < chansize; ind++ {  // with the same number of tasks, good
			_, xerr = overlord.Start(
				func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
					weWereAborted := false
					out := 4
					for { // do some work, then look for aborted, again and again
						// some work
						time.Sleep(time.Duration(RandomInt(20, 30)) * time.Millisecond)
						if t.Aborted() {
							// Cleaning up first before leaving... ;)
							time.Sleep(time.Duration(RandomInt(100, 800)) * time.Millisecond)
							weWereAborted = true
							break
						}
						out--
						if out < 0 {
							break
						}
					}

					// We are using the classic 'send on closed channel' trick to see if Wait actually waits until everyone is DONE.
					// If it does we will never see a panic, but if Abort doesn't mean TellYourChildrenToAbort but
					// actually means AbortYourChildrenAndQuitNOWWithoutWaiting, then we have a problem
					acha := parameters.(chan string)
					acha <- "Bailing out"

					// flip a coin, true and we panic, false we don't
					if RandomInt(0, 10) > 3 {
						panic("head")
					}
					// tails

					if weWereAborted {
						return "", fail.AbortedError(nil, "we were killed")
					}

					return "who cares", nil
				}, bailout, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)),
			)
			require.Nil(t, xerr)
		}

		// after this, some tasks will already be looking for ABORT signals
		time.Sleep(time.Duration(65) * time.Millisecond)

		res, xerr := overlord.WaitGroup() // 100 ms after this, .Abort() should hit
		if xerr != nil {
			t.Logf("Failed to Wait: %s", xerr.Error()) // Of course, we did !!, we induced a panic !! didn't we ?
			switch xerr.(type) {
			case *fail.ErrAborted:
				cause := xerr.Cause()
				if !strings.Contains(spew.Sdump(cause), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			// or maybe we were fast enough and we are quitting only because of Abort, but no problem, we have more iterations...
			case *fail.ErrRuntimePanic: // This MUST NEVER HAPPEN in a TaskGroup; the panic should be in the ErrorList returned by Wait()
				t.Errorf("That shouldn't happen")
				t.FailNow()
			case *fail.ErrorList:
				if !strings.Contains(spew.Sdump(xerr), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				} else {
					t.Logf("We catched a panic..., good")
					caught = true
					break
				}
			default:
				t.Errorf("Unexpected error: %v", xerr)
			}
		} else {
			require.NotNil(t, res)
			if len(bailout) == chansize {
				streak++
				if streak > 5 {
					break
				}
				continue
			}
		}
		close(bailout) // If Wait actually waits, this is closed AFTER all Tasks filled the channel, so no panics
		// If not..., well...

		reminder := false
		if len(bailout) != chansize { // this means panic
			reminder = true
			t.Errorf("Not everyone finished on time !!, panic is coming !!, some tasks will hit a closed channel !!")
			// if we now do a t.FailNow() we already proved our point (if Wait actually waited, the channel
			// size should be chansize each time), but if we don't...
			// we will see runtime panics on our LOGS !!, but NOT in the code
			// with a t.FailNow() we also fail, but the test output is less frightening
			enough = true
		}

		time.Sleep(2000 * time.Millisecond)
		if reminder {
			t.Errorf("by now we should see panics in lines above, panics that only shows in logs and the rest of the code is unaware of")
		}
		// Well, we have a problem Waiting, now it's clear, and as a bonus we uncovered a problem communicating panics to function callers
	}

	if !caught {
		t.Errorf("We were unable to catch a panic...")
		t.FailNow()
	}
}

// Like previous tests but also with errors
func TestAbortThingsThatActuallyTakeTimeCleaningUpAndFailWhenWeAlreadyStartedWaitingFor(t *testing.T) {
	enough := false
	iter := 0
	streak := 0
	chansize := 10
	for {
		iter++
		if iter > 12 {
			break
		}
		if enough {
			break
		}

		t.Log("--- Next ---") // Each time we iterate we see this line, sometimes this doesn't fail at 1st iteration
		overlord, xerr := NewTaskGroup()
		require.NotNil(t, overlord)
		require.Nil(t, xerr)
		xerr = overlord.SetID(fmt.Sprintf("/parent-%d", iter))
		require.Nil(t, xerr)

		var failureCounter int32
		bailout := make(chan string, chansize) // a buffered channel
		for ind := 0; ind < chansize; ind++ {  // with the same number of tasks, good
			_, xerr = overlord.Start(
				func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
					weWereAborted := false
					for { // do some work, then look for aborted, again and again
						// some work
						time.Sleep(time.Duration(RandomInt(20, 30)) * time.Millisecond)
						if t.Aborted() {
							// Cleaning up first before leaving... ;)
							time.Sleep(time.Duration(RandomInt(100, 800)) * time.Millisecond)
							weWereAborted = true
							break
						}
					}

					// We are using the classic 'send on closed channel' trick to see if Wait actually waits until everything is DONE.
					// If it does we will never see a panic, but if Abort doesn't mean TellYourChildrenToAbort but
					// actually means AbortYourChildrenAndQuitNOWWithoutWaiting, then we have a problem
					acha := parameters.(chan string)
					acha <- "Bailing out"

					if weWereAborted {
						return "", fail.AbortedError(nil, "we were killed")
					}

					// flip a coin, true and we panic, false we don't
					if RandomInt(0, 2) == 1 {
						atomic.AddInt32(&failureCounter, 1)
						return "mistakes happen", fail.NewError("It was head")
					}

					return "who cares", nil
				}, bailout, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)),
			)
			require.Nil(t, xerr)
		}

		// after this, some tasks will already be looking for ABORT signals
		time.Sleep(time.Duration(65) * time.Millisecond)

		go func() {
			// this will actually start after wait
			time.Sleep(time.Duration(100) * time.Millisecond)

			// let's have fun
			xerr := overlord.Abort()
			require.Nil(t, xerr)

			// did we abort ?
			aborted := overlord.Aborted()
			if !aborted {
				t.Logf("We just aborted without error above..., why Aborted() says it's not ?")
			}
		}()

		_, res, xerr := overlord.WaitFor(5 * time.Second) // 100 ms after this, .Abort() should hit
		if xerr != nil {
			t.Logf("Wait reports a failure with %d child failures: %s", failureCounter, reflect.TypeOf(xerr).String()) // Of course, we did !!, we induced a panic !! didn't we ?
			switch cerr := xerr.(type) {
			case *fail.ErrAborted:
				cause := xerr.Cause()
				if cause != nil {
					t.Logf("TaskGroup reported error cause '%s': %v", reflect.TypeOf(cause).String(), cause)
				}
				// if it's unexpected and it happens -> error, and we can finish the test
				if strings.Contains(spew.Sdump(cause), "panic happened") {
					t.Errorf("TaskGroup reported panic in cause!!!")
					t.FailNow()
				}
				consequences := xerr.Consequences()
				if len(consequences) > 0 {
					counted := 0
					t.Log("TaskGroup children reported failures:")
					for _, v := range consequences {
						logged := false
						switch cerr := v.(type) {
						case *fail.ErrAborted:
							consequences := cerr.Consequences()
							if len(consequences) > 0 {
								t.Logf("aborted with consequence: %v (%s)", v, reflect.TypeOf(v).String())
								logged = true
							}
						default:
							counted++
						}
						if !logged {
							t.Logf("%v (%s)", v, reflect.TypeOf(v).String())
						}
					}
					if counted != int(failureCounter) {
						t.Errorf("Taskgroup returned error does not reports the effective children failure count!!!")
					}
				}
				if strings.Contains(spew.Sdump(consequences), "panic happened") {
					t.Logf("no panic reported by TaskGroup children")
				}
			// or maybe we were fast enough and we are quitting only because of Abort, but no problem, we have more iterations...
			case *fail.ErrRuntimePanic:
				t.Errorf("That shouldn't happen")
				t.FailNow()
			case *fail.ErrorList:
				errorList := cerr.ToErrorSlice()
				if len(errorList) > 0 {
					t.Logf("TaskGroup children reported failures:")
					for _, v := range errorList {
						logged := false
						switch cerr := v.(type) {
						case *fail.ErrAborted:
							consequences := cerr.Consequences()
							if len(consequences) > 0 {
								t.Logf("aborted with consequence: %v (%s)", v, reflect.TypeOf(v).String())
								logged = true
							}
						default:
						}
						if !logged {
							t.Logf("%v (%s)", v, reflect.TypeOf(v).String())
						}
					}
				}
				if strings.Contains(spew.Sdump(xerr), "panic happened") {
					t.Logf("panic reported by TaskGroup children!!!")
				}
			default:
				t.Errorf("Unexpected error: %v", xerr)
			}
		} else {
			require.NotNil(t, res)
			if len(bailout) == chansize {
				streak++
				if streak > 5 {
					break
				}
				continue
			}
		}
		close(bailout) // If Wait actually waits, this is closed AFTER all Tasks filled the channel, so no panics
		// If not..., well...

		reminder := false
		if len(bailout) != chansize { // this means panic
			reminder = true
			t.Errorf("Not everyone finished on time !!, panic is coming !!, some tasks will hit a closed channel !!")
			// if we now do a t.FailNow() we already proved our point (if Wait actually waited, the channel
			// size should be chansize each time), but if we don't...
			// we will see runtime panics on our LOGS !!, but NOT in the code
			// with a t.FailNow() we also fail, but the test output is less frightening
			enough = true
		}

		time.Sleep(2 * time.Second)
		if reminder {
			t.Errorf("by now we should see panics in lines above, panics that only shows in logs and the rest of the code is unaware of")
		}
		// Well, we have a problem Waiting, now it's clear, and as a bonus we uncovered a problem communicating panics to function callers
	}
}

// This tests the same thing as TestAbortThingsThatActuallyTakeTimeCleaningUpWhenWeAlreadyStartedWaiting, it just
// runs .Abort first, then Wait
// It fails, however it's unclear if it should work..: by design, what should happen if we abort 1st before running the wait ?
func TestAbortThingsThatActuallyTakeTimeCleaningUpAbortAndWaitForLater(t *testing.T) {
	enough := false
	iter := 0
	streak := 0
	chansize := 10
	for {
		iter++
		if iter > 12 {
			break
		}
		if enough {
			break
		}

		t.Log("Next") // Each time we iterate we see this line, sometimes this doesn't fail at 1st iteration
		overlord, xerr := NewTaskGroup()
		require.NotNil(t, overlord)
		require.Nil(t, xerr)
		xerr = overlord.SetID(fmt.Sprintf("parent-%d", iter))
		require.Nil(t, xerr)

		bailout := make(chan string, chansize) // a buffered channel
		for ind := 0; ind < chansize; ind++ {  // with the same number of tasks, good
			_, xerr = overlord.Start(
				func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
					weWereAborted := false
					for { // do some work, then look for aborted, again and again
						// some work
						time.Sleep(time.Duration(RandomInt(20, 30)) * time.Millisecond)
						if t.Aborted() {
							// Cleaning up first before leaving... ;)
							time.Sleep(time.Duration(RandomInt(100, 800)) * time.Millisecond)
							weWereAborted = true
							break
						}
					}
					// We are using the classic 'send on closed channel' trick to see if Wait actually waits until everyone is DONE.
					// If it does we will never see a panic, but if, Abort doesn't mean TellYourChildrenToAbort but
					// actually means AbortYourChildrenAndQuitNOWWithoutWaiting, then we have a problem
					acha := parameters.(chan string)
					acha <- "Bailing out"

					if weWereAborted {
						return "", fail.AbortedError(nil, "we were killed")
					}

					return "who cares", nil
				}, bailout,
				InheritParentIDOption, AmendID(fmt.Sprintf("child-%d", ind)),
			)
			require.Nil(t, xerr)
		}

		// after this, some tasks will already be looking for ABORT signals
		time.Sleep(time.Duration(65) * time.Millisecond)

		xerr = overlord.Abort()
		require.Nil(t, xerr)

		// did we abort ?
		aborted := overlord.Aborted()
		if !aborted {
			t.Errorf("We just aborted without error above..., why Aborted() says it's not ?")
		}

		_, res, xerr := overlord.WaitFor(5 * time.Second)
		if xerr != nil {
			t.Logf("Failed to Wait: %s", xerr.Error()) // Of course, we did !!, we induced a panic !! didn't we ?
			switch xerr.(type) {
			case *fail.ErrAborted:
				cause := xerr.Cause()
				if !strings.Contains(spew.Sdump(cause), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			// or maybe we were fast enough and we are quitting only because of Abort, but no problem, we have more iterations...
			case *fail.ErrRuntimePanic:
				t.Errorf("That shouldn't happen")
				t.FailNow()
			case *fail.ErrorList:
				if !strings.Contains(spew.Sdump(xerr), "panic happened") {
					t.Logf("What ?? the panic was just swallowed in the logs ??, the code making the call doesn't know ???, or we just stopped waiting even before the panic happened ??...")
				}
			default:
				t.Errorf("Unexpected error: %v", xerr)
			}
		} else {
			require.NotNil(t, res)
			if len(bailout) == chansize {
				streak++
				if streak > 5 {
					break
				}
				continue
			}
		}

		close(bailout) // If Wait actually waits, this is closed AFTER all Tasks filled the channel, so no panics
		// If not..., well...

		reminder := false
		if len(bailout) != chansize { // this means panic
			reminder = true
			t.Errorf("Not everyone finished on time !!, panic is coming !!, some tasks will hit a closed channel !!")
			// if we now do a t.FailNow() we already proved our point (if Wait actually waited, the channel
			// size should be chansize each time), but if we dont...
			// we will see runtime panics on our LOGS !!, but NOT in the code
			// with a t.FailNow() we also fail, but the test output is less frightening
			enough = true
		}

		time.Sleep(2000 * time.Millisecond)
		if reminder {
			t.Errorf("by now we should see panics in lines above, panics that only shows in logs and the rest of the code is unaware of")
		}
		// Well, we have a problem Waiting, now it's clear, and as a bonus we uncovered a problem communicating panics to function callers
	}
}

// This could be an open question
// This test runs some tasks that succeed by design, so what should happen when we abort and then wait ?
// What's troubling about this test, it's that non-determinisic...
// non-deterministic ?? what ??, yes, run it enough times and you will have different outcomes, seriously
// sometimes the wait fails, sometimes doesn't, sometimes we have an empty result map, sometimes a populated map...
func TestAbortAlreadyFinishedSuccessfullyThingsThenWaitFor(t *testing.T) {
	var previousRes map[string]TaskResult
	var previousErr error
	iter := 0
	for {
		iter++
		if iter > 15 {
			break
		}

		overlord, xerr := NewTaskGroupWithParent(nil)
		if xerr != nil {
			t.Errorf("Error creating taskGroup: %v", xerr)
		}
		require.Nil(t, xerr)
		if overlord == nil {
			t.Errorf("Error creating taskGroup")
		}
		require.NotNil(t, overlord)
		xerr = overlord.SetID("/parent")
		require.Nil(t, xerr)

		for ind := 0; ind < 10; ind++ {
			_, err := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
				time.Sleep(time.Duration(RandomInt(10, 20)) * time.Millisecond)
				return "waiting game", nil
			}, nil, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)))
			if err != nil {
				t.Errorf("Unexpected: %s", err)
				t.FailNow()
			}
		}

		time.Sleep(time.Duration(100) * time.Millisecond)
		// overlord should have finished a loooong time ago...

		// but we abort anyway
		xerr = overlord.Abort()
		if xerr != nil {
			t.Errorf("Failed to abort")
			t.FailNow()
		}

		// did we abort ?
		aborted := overlord.Aborted()
		if !aborted {
			t.Errorf("We just aborted without error above..., why Aborted() says it's not ?")
		}

		// the question here, is why we fail ?
		// and more, from a client point of view, why this failed ?
		// all we have is an aborted error
		var res map[string]TaskResult
		res, xerr = overlord.WaitGroup()
		if xerr != nil {
			t.Errorf("Failed to Wait: %v", xerr)
		}

		// check for error inconsistencies
		if iter == 1 {
			previousErr = xerr
		} else {
			if xerr != previousErr {
				t.Errorf("Not consistent, before: %v, now: %v", previousErr, xerr)
				t.FailNow()
			}
		}

		// check for result inconsistencies
		if iter == 1 {
			previousRes = res
		} else {
			if len(res) != len(previousRes) {
				t.Errorf("Not consistent, before: %d, now: %d", len(previousRes), len(res))
				t.Logf("Recovered this: %v", res)
				t.FailNow()
			}
			previousRes = res
		}
	}
}