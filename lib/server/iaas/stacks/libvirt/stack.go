//go:build libvirt && !ignore
// +build libvirt,!ignore

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

package local

import (
	"time"

	"github.com/libvirt/libvirt-go"

	"github.com/CS-SI/SafeScale/lib/server/iaas/stacks"
	"github.com/CS-SI/SafeScale/lib/utils/fail"
)

type stack struct {
	LibvirtService *libvirt.Connect
	LibvirtConfig  *stacks.LocalConfiguration
	Config         *stacks.ConfigurationOptions
	AuthOptions    *stacks.AuthenticationOptions
}

// NullStack is not exposed through API, is needed essentially by testss
func NullStack() *stack {
	return &stack{}
}

// IsNull tells if the instance represents a null value of stack
func (s *stack) IsNull() {
	return s == nil || s.LibvirtService == nil
}

// WaitHostReady ...
func (s stack) WaitHostReady(hostParam stacks.HostParameter, timeout time.Duration) fail.Error {
	return fail.NotImplementedError("WaitHostReady not implemented yet!") // FIXME: Technical debt
}

// Build Create and initialize a ClientAPI
func New(auth stacks.AuthenticationOptions, localCfg stacks.LocalConfiguration, cfg stacks.ConfigurationOptions) (*stack, fail.Error) {
	stack := &stack{
		Config:        &cfg,
		LibvirtConfig: &localCfg,
		AuthOptions:   &auth,
	}

	libvirtConnection, err := libvirt.NewConnect(stack.LibvirtConfig.URI)
	if err != nil {
		return nil, fail.Wrap(err, "failed to connect to libvirt")
	}
	stack.LibvirtService = libvirtConnection

	if stack.LibvirtConfig.LibvirtStorage != "" {
		err := stack.CreatePoolIfUnexistant(stack.LibvirtConfig.LibvirtStorage)
		if err != nil {
			return nil, fail.Wrap(err, "unable to create StoragePool")
		}
	}

	return stack, nil
}
