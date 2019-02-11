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

package metadata

import (
	"github.com/CS-SI/SafeScale/providers"
	"github.com/CS-SI/SafeScale/providers/api"
	"github.com/CS-SI/SafeScale/providers/objectstorage"
	"github.com/CS-SI/SafeScale/utils"
	"github.com/CS-SI/SafeScale/utils/serialize"
)

// Item is an entry in the ObjectStorage
type Item struct {
	payload serialize.Serializable
	folder  *Folder
	lock    chan struct{}
}

// ItemDecoderCallback ...
type ItemDecoderCallback func([]byte) (serialize.Serializable, error)

// NewItem creates a new item with 'name' and in 'path'
func NewItem(svc *providers.Service, path string) *Item {
	return &Item{
		folder:  NewFolder(svc, path),
		payload: nil,
		lock:    make(chan struct{}, 1),
	}
}

// GetService returns the service used by Item
func (i *Item) GetService() *providers.Service {
	return i.folder.GetService()
}

// GetBucket returns the bucket used by Item
func (i *Item) GetBucket() objectstorage.Bucket {
	return i.folder.GetBucket()
}

// GetClient returns the bucket used by Item
func (i *Item) GetClient() api.ClientAPI {
	return i.folder.GetClient()
}

// GetPath returns the path in the Object Storage where the Item is stored
func (i *Item) GetPath() string {
	return i.folder.GetPath()
}

// Carry links metadata with cluster struct
func (i *Item) Carry(data serialize.Serializable) *Item {
	i.payload = data
	return i
}

// Reset ...
func (i *Item) Reset() {
	i.payload = nil
}

// Get returns payload in item
func (i *Item) Get() interface{} {
	return i.payload
}

// DeleteFrom removes a metadata from a folder
func (i *Item) DeleteFrom(path string, name string) error {
	if name == "" {
		panic("name is empty!")
	}
	if path == "" {
		path = "."
	}

	err := i.folder.Search(path, name)
	if err != nil {
		if _, ok := err.(utils.ErrNotFound); ok {
			// If entry not found, consider a success
			return nil
		}
		return err
	}

	return i.folder.Delete(path, name)
}

// Delete removes a metadata
func (i *Item) Delete(name string) error {
	return i.DeleteFrom(".", name)
}

// ReadFrom reads metadata of item from Object Storage in a subfolder
func (i *Item) ReadFrom(path string, name string, callback ItemDecoderCallback) error {
	var data serialize.Serializable
	err := i.folder.Read(path, name, func(buf []byte) error {
		var err error
		data, err = callback(buf)
		return err
	})
	if err != nil {
		return err
	}
	i.payload = data
	return nil
}

// Read read metadata of item from Object Storage (in current folder)
func (i *Item) Read(name string, callback ItemDecoderCallback) error {
	return i.ReadFrom(".", name, callback)
}

// WriteInto saves the content of Item in a subfolder to the Object Storage
func (i *Item) WriteInto(path string, name string) error {
	if i == nil {
		panic("i is nil!")
	}
	if i.payload == nil {
		panic("i.payload is nil!")
	}
	data, err := i.payload.Serialize()
	if err != nil {
		return err
	}
	return i.folder.Write(path, name, data)
}

// Write saves the content of Item to the Object Storage
func (i *Item) Write(name string) error {
	return i.WriteInto(".", name)
}

// BrowseInto walks through a subfolder ogf item folder and executes a callback for each entry
func (i *Item) BrowseInto(path string, callback func([]byte) error) error {
	if callback == nil {
		panic("callback is nil!")
	}

	if path == "" {
		path = "."
	}
	return i.folder.Browse(path, func(buf []byte) error {
		return callback(buf)
	})
}

// Browse walks through folder of item and executes a callback for each entry
func (i *Item) Browse(callback func([]byte) error) error {
	return i.BrowseInto(".", func(buf []byte) error {
		return callback(buf)
	})
}

// Acquire waits until the lock is available, then locks the metadata
func (i *Item) Acquire() {
	i.lock <- struct{}{}
}

// Release unlocks the metadata
func (i *Item) Release() {
	<-i.lock
}
