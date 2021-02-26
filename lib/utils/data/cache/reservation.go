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

package cache

//go:generate mockgen -destination=../mocks/mock_clonable.go -package=mocks github.com/CS-SI/SafeScale/lib/utils/data Cacheable

import (
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/data/observer"
)

// reservation is a struct to simulate a content of a Entry to "reserve" a key
type reservation struct {
	key string
}

func (rc reservation) GetID() string {
	return rc.key
}

func (rc reservation) GetName() string {
	return rc.key
}

func (rc reservation) AddObserver(task concurrency.Task, o observer.Observer) error {
	return nil
}

func (rc reservation) NotifyObservers(task concurrency.Task) error {
	return nil
}

func (rc reservation) RemoveObserver(task concurrency.Task, name string) error {
	return nil
}

func (rc reservation) Released(task concurrency.Task) {
	return //nolint
}

func (rc reservation) Destroyed(task concurrency.Task) {
	return //nolint
}
