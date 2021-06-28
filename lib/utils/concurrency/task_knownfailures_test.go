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
	"strings"
	"testing"
	"time"

	"github.com/CS-SI/SafeScale/lib/utils/fail"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

// This imitates some of the code found in cluster.go
func TestRealCharge(t *testing.T) {
	overlord, err := NewTaskGroupWithParent(nil)
	require.NotNil(t, overlord)
	require.Nil(t, err)

	theID, err := overlord.GetID()
	require.Nil(t, err)
	require.NotEmpty(t, theID)

	gorrs := 800
	abortOccurred := false
	started := 0
	for ind := 0; ind < gorrs; ind++ {
		_, xerr := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
			time.Sleep(time.Duration(RandomInt(50, 250)) * time.Millisecond)
			return "waiting game", nil
		}, nil)
		if xerr != nil {
			if !overlord.Aborted() {
				t.Errorf("Unexpected: %s", xerr)
			}
		} else {
			started++
		}
		if RandomInt(50, 250) > 200 {
			xerr = overlord.Abort()
			if xerr != nil {
				t.Errorf("What, Cannot abort ??")
				t.FailNow()
			}
			abortOccurred = true
		}
	}

	res, err := overlord.Wait()
	require.NotEmpty(t, res)
	var abortState string
	if abortOccurred {
		abortState = " before Abort"
	}
	t.Logf("Started %d TaskActions%s", started, abortState)
}

// This imitates some of the code found in cluster.go
func TestRealCharges(t *testing.T) {
	overlord, err := NewTaskGroupWithParent(nil)
	require.NotNil(t, overlord)
	require.Nil(t, err)

	theID, err := overlord.GetID()
	require.Nil(t, err)
	require.NotEmpty(t, theID)

	gorrs := 800

	for ind := 0; ind < gorrs; ind++ {
		_, err := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
			time.Sleep(time.Duration(RandomInt(50, 250)) * time.Millisecond)
			return "waiting game", nil
		}, nil)
		if err != nil {
			t.Errorf("Unexpected: %s", err)
		}
		if RandomInt(50, 250) > 200 {
			aErr := overlord.Abort()
			if aErr != nil {
				t.Errorf("What, Cannot abort ??")
				t.FailNow()
			}
			break
		}
	}

	fast, res, err := overlord.WaitFor(280 * time.Millisecond)
	require.NotEmpty(t, res) // recovering partial records lead to a race condition, should we try ?
	require.True(t, fast)
}

// This imitates some of the code found in cluster.go
func TestRealTryCharges(t *testing.T) {
	overlord, err := NewTaskGroupWithParent(nil)
	require.NotNil(t, overlord)
	require.Nil(t, err)

	xerr := overlord.SetID("/parent")
	require.Nil(t, xerr)

	theID, err := overlord.GetID()
	require.Nil(t, err)
	require.NotEmpty(t, theID)

	gorrs := 800

	for ind := 0; ind < gorrs; ind++ {
		_, err := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
			time.Sleep(time.Duration(RandomInt(200, 250)) * time.Millisecond)
			return "waiting game", nil
		}, nil, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)))
		if err != nil {
			t.Errorf("Unexpected: %s", err)
		}
	}

	for {
		done, res, xerr := overlord.TryWaitGroup()
		if !done {
			require.Nil(t, xerr)
			require.Nil(t, res)
			require.False(t, done)
		} else {
			require.Nil(t, xerr)
			break
		}
	}
}

// This imitates some of the code found in cluster.go
func TestTryWaitRecoversErrorContent(t *testing.T) {
	overlord, xerr := NewTaskGroupWithParent(nil)
	require.NotNil(t, overlord)
	require.Nil(t, xerr)

	xerr = overlord.SetID("/parent")
	require.Nil(t, xerr)

	theID, xerr := overlord.GetID()
	require.Nil(t, xerr)
	require.NotEmpty(t, theID)

	gorrs := 800

	for ind := 0; ind < gorrs; ind++ {
		_, xerr := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
			time.Sleep(time.Duration(RandomInt(200, 250)) * time.Millisecond)
			return "waiting game", nil
		}, nil, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)))
		if xerr != nil {
			t.Errorf("Unexpected: %s", xerr)
		}
	}

	_, xerr = overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
		time.Sleep(time.Duration(RandomInt(10, 15)) * time.Millisecond)
		return "waiting game", fail.NewError("Ouch")
	}, nil, InheritParentIDOption, AmendID("/ill-child"))
	if xerr != nil {
		t.Errorf("Unexpected: %s", xerr)
	}

	time.Sleep(40 * time.Millisecond)

	for {
		done, res, xerr := overlord.TryWaitGroup()
		if !done {
			require.Nil(t, xerr)
			require.Nil(t, res)
			require.False(t, done)
		} else {
			require.NotNil(t, xerr)
			// t.Logf("%s", spew.Sdump(xerr))
			require.True(t, strings.Contains(spew.Sdump(xerr), "Ouch"))
			break
		}
	}
}

// This imitates some of the code found in cluster.go
func TestTryWaitRecoversErrorContentAlsoWhenRunningWithTimeout(t *testing.T) {
	overlord, xerr := NewTaskGroupWithParent(nil)
	require.NotNil(t, overlord)
	require.Nil(t, xerr)

	xerr = overlord.SetID("/parent")
	require.Nil(t, xerr)

	theID, xerr := overlord.GetID()
	require.Nil(t, xerr)
	require.NotEmpty(t, theID)

	gorrs := 800

	for ind := 0; ind < gorrs; ind++ {
		_, xerr := overlord.Start(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
			time.Sleep(time.Duration(RandomInt(200, 250)) * time.Millisecond)
			return "waiting game", nil
		}, nil, InheritParentIDOption, AmendID(fmt.Sprintf("/child-%d", ind)))
		if xerr != nil {
			t.Errorf("Unexpected: %s", xerr)
		}
	}

	_, xerr = overlord.StartWithTimeout(func(t Task, parameters TaskParameters) (TaskResult, fail.Error) {
		time.Sleep(time.Duration(RandomInt(10, 15)) * time.Millisecond)
		return "waiting game", fail.NewError("Ouch")
	}, nil, 12*time.Millisecond, InheritParentIDOption, AmendID("/ill-child"))
	if xerr != nil {
		t.Errorf("Unexpected: %s", xerr)
	}

	time.Sleep(40 * time.Millisecond)

	for {
		done, res, xerr := overlord.TryWaitGroup()
		if !done {
			require.Nil(t, xerr)
			require.Nil(t, res)
			require.False(t, done)
		} else {
			require.NotNil(t, xerr)
			t.Logf("%s", spew.Sdump(xerr))
			require.True(t, strings.Contains(spew.Sdump(xerr), "Ouch"))
			break
		}
	}
}
