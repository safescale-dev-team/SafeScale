/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
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

package handlers

import (
	"context"

	"github.com/CS-SI/SafeScale/providers"
	"github.com/CS-SI/SafeScale/providers/model"
)

//go:generate mockgen -destination=../mocks/mock_imageapi.go -package=mocks github.com/CS-SI/SafeScale/broker/server/handlers ImageAPI

// TODO At service level, ve need to log before returning, because it's the last chance to track the real issue in server side

// ImageAPI defines API to manipulate images
type ImageAPI interface {
	List(ctx context.Context, all bool) ([]model.Image, error)
	Select(ctx context.Context, osfilter string) (*model.Image, error)
	Filter(ctx context.Context, osfilter string) ([]model.Image, error)
}

// NewImageHandler creates an host service
func NewImageHandler(api *providers.Service) ImageAPI {
	return &ImageHandler{
		provider: api,
	}
}

// ImageHandler image service
type ImageHandler struct {
	provider *providers.Service
}

// List returns the image list
func (srv *ImageHandler) List(ctx context.Context, all bool) ([]model.Image, error) {
	images, err := srv.provider.ListImages(all)
	return images, infraErr(err)
}

// Select selects the image that best fits osname
func (srv *ImageHandler) Select(ctx context.Context, osname string) (*model.Image, error) {
	return nil, nil
}

// Filter filters the images that do not fit osname
func (srv *ImageHandler) Filter(ctx context.Context, osname string) ([]model.Image, error) {
	return nil, nil
}
