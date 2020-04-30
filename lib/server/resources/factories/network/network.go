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

// Package factory contains methods to load or create objects
package network

import (
	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/server/resources"
	"github.com/CS-SI/SafeScale/lib/server/resources/operations"
	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/fail"
)

// New creates an instance of resources.Network
func New(svc iaas.Service) (resources.Network, error) {
	if svc == nil {
		return nil, fail.InvalidParameterReport("svc", "cannot be nil")
	}

	return operations.NewNetwork(svc)
}

// Load loads the metadata of a network and returns an instance of resources.Network
func Load(task concurrency.Task, svc iaas.Service, ref string) (resources.Network, fail.Report) {
	if task == nil {
		return nil, fail.InvalidParameterReport("task", "cannot be nil")
	}
	if svc == nil {
		return nil, fail.InvalidParameterReport("svc", "cannot be nil")
	}
	if ref == "" {
		return nil, fail.InvalidParameterReport("ref", "cannot be empty string")
	}

	return operations.LoadNetwork(task, svc, ref)
}
