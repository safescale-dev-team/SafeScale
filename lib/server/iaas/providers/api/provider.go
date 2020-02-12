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

package api

import (
	"github.com/CS-SI/SafeScale/lib/server/iaas/providers"
	stacks "github.com/CS-SI/SafeScale/lib/server/iaas/stacks/api"
	"github.com/CS-SI/SafeScale/lib/server/resources/abstracts"
)

//go:generate mockgen -destination=../mocks/mock_providerapi.go -package=mocks github.com/CS-SI/SafeScale/lib/server/iaas/providers/api Provider

// Provider is the interface to cloud stack
// It has to recall Stack api, to serve as Provider AND as Stack
type Provider interface {
	Build(map[string]interface{}) (Provider, error)

	stacks.Stack

	// ListImages lists available OS images
	ListImages(all bool) ([]abstracts.Image, error)

	// ListTemplates lists available host templates
	// Host templates are sorted using Dominant Resource Fairness Algorithm
	ListTemplates(all bool) ([]abstracts.HostTemplate, error)

	// AuthenticationOptions returns authentication options as a Config
	AuthenticationOptions() (providers.Config, error)

	// ConfigurationfgOpts returns configuration options as a Config
	ConfigurationOptions() (providers.Config, error)

	// Name returns the provider name
	Name() string

	// Capabilities returns the capabilities of the provider
	Capabilities() providers.Capabilities

	// TenantParameters returns the tenant parameters as read
	TenantParameters() map[string]interface{}
}
