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

package objectstorage

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gomodules.xyz/stow"

	// necessary for connect
	// _ "gomodules.xyz/stow/azure"
	_ "gomodules.xyz/stow/google"
	_ "gomodules.xyz/stow/s3"
	_ "gomodules.xyz/stow/swift"

	"github.com/CS-SI/SafeScale/lib/utils/concurrency"
	"github.com/CS-SI/SafeScale/lib/utils/debug"
	"github.com/CS-SI/SafeScale/lib/utils/scerr"
)

// object implementation of Object interface
type object struct {
	bucket *bucket
	item   stow.Item

	ID       string         `json:"id,omitempty"`
	Name     string         `json:"name,omitempty"`
	Metadata ObjectMetadata `json:"metadata,omitempty"`
	// DeleteAt     time.Time              `json:"delete_at,omitempty"`
	// LastModified time.Time              `json:"last_modified,omitempty"`
	// ContentType   string                 `json:"content_type,omitempty"`
	// ContentLength int64                  `json:"content_length,omitempty"`
	// ETag          string                 `json:"etag,omitempty"`
	// Size          int64                  `json:"size,omitempty"`
	// Date          time.Time              `json:"date,omitempty"`
	// Content       io.ReadSeeker          `json:"content,omitempty"`
}

// NewObject ...
func newObject(bucket *bucket, objectName string) (*object, error) {
	o := &object{
		bucket: bucket,
		Name:   objectName,
	}
	item, err := bucket.container.Item(objectName)
	if err == nil {
		o.item = item
	}
	return o, nil
}

// newObjectFromStow ...
func newObjectFromStow(bucket *bucket, item stow.Item) *object {
	return &object{
		bucket:   bucket,
		item:     item,
		Name:     item.Name(),
		Metadata: ObjectMetadata{},
	}
}

// Stored return true if the object exists in Object Storage
func (o *object) Stored() bool {
	return o.item != nil
}

// Reload reloads the data of the Object from the Object Storage
func (o *object) Reload() error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "").Entering().OnExitTrace()

	item, err := o.bucket.container.Item(o.Name)
	if err != nil {
		return err
	}
	return o.reloadFromItem(item)
}

// reloadFromItem reloads object instance with stow.Item
func (o *object) reloadFromItem(item stow.Item) error {
	o.item = item
	newMetadata, err := item.Metadata()
	if err != nil {
		return err
	}
	o.Metadata = newMetadata
	return nil
}

// Read reads the content of the object from Object Storage and writes it in 'target'
func (o *object) Read(target io.Writer, from, to int64) error {
	if target == nil {
		return scerr.InvalidInstanceError()
	}
	if from > to {
		return scerr.InvalidParameterError("from", "cannot be greater than 'to'")
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "(%d, %d)", from, to).Entering().OnExitTrace()

	var seekTo int64
	var length int64

	// 1st reload information about object, to be sure to have the last
	err := o.Reload()
	if err != nil {
		return err
	}

	size, err := o.GetSize()
	if err != nil {
		return scerr.Wrap(err, "failed to get bucket size")
	}
	if size < 0 {
		return scerr.NewError("unknown size of object")
	}

	length = size
	if from > 0 {
		seekTo = from
	}
	if to > 0 && to > from {
		length = to - from
	}

	source, err := o.item.Open()
	if err != nil {
		return err
	}
	defer func() {
		clerr := source.Close()
		if clerr != nil {
			logrus.Error("Error closing item")
		}
	}()

	if seekTo == 0 && length >= size {
		_, err := io.CopyN(target, source, size)
		if err != nil {
			return err
		}
	} else {
		buf := make([]byte, seekTo)
		if _, err := io.ReadAtLeast(source, buf, int(seekTo)); err != nil {
			logrus.Fatal(err)
		}

		bufbis := make([]byte, length)
		if _, err := io.ReadAtLeast(source, bufbis, int(length)); err != nil {
			logrus.Println("error ")
			logrus.Fatal(err)
		}

		readerbis := bytes.NewReader(bufbis)
		_, err := io.CopyBuffer(target, readerbis, bufbis)
		if err != nil {
			return err
		}
	}
	return nil
}

// Write the source to the object in Object Storage
func (o *object) Write(source io.Reader, sourceSize int64) error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}
	if source == nil {
		return scerr.InvalidParameterError("source", "cannot be nil")
	}
	if o.bucket == nil {
		return scerr.InvalidParameterError("o.bucket", "cannot be nil")
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "(%d)", sourceSize).Entering().OnExitTrace()

	item, err := o.bucket.container.Put(o.Name, source, sourceSize, o.SafeGetMetadata())
	if err != nil {
		return err
	}
	return o.reloadFromItem(item)
}

// WriteMultiPart writes big data to Object, by parts (also called chunks)
// Note: nothing to do with multi-chunk abilities of various object storage technologies
func (o *object) WriteMultiPart(source io.Reader, sourceSize int64, chunkSize int) error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}
	if source == nil { // If source is nil, do nothing and don't trigger an error
		return nil
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "(%d, %d)", sourceSize, chunkSize).Entering().OnExitTrace()

	metadataCopy := o.SafeGetMetadata().Clone()

	var chunkIndex int
	remaining := sourceSize
	for {
		if remaining < int64(chunkSize) {
			chunkSize = int(remaining)
		}
		err := writeChunk(o.bucket.container, o.Name, source, chunkSize, metadataCopy, chunkIndex)
		if err != nil {
			return err
		}
		remaining -= int64(chunkSize)
		// client.NbItem = client.NbItem + 1
		if remaining <= 0 {
			break
		}
		chunkIndex++
	}
	return nil
}

// writeChunk writes a chunk of data for object
func writeChunk(
	container stow.Container, objectName string,
	source io.Reader, nBytesToRead int,
	metadata ObjectMetadata,
	chunkIndex int,
) error {

	buf := make([]byte, nBytesToRead)
	nBytesRead, err := source.Read(buf)
	if err == io.EOF {
		msg := fmt.Sprintf("failed to read data from source to write in chunk of object '%s' in bucket '%s'", objectName, container.Name())
		logrus.Errorf(msg)
		return scerr.NewError(msg)
	}
	r := bytes.NewReader(buf)
	objectNamePart := objectName + strconv.Itoa(chunkIndex)
	metadata["Split"] = objectName
	_, err = container.Put(objectNamePart, r, int64(nBytesRead), metadata)
	if err != nil {
		return scerr.Wrap(err, "failed to write in chunk of object '%s' in bucket '%s'", objectName, container.Name())
	}
	logrus.Debugf("written chunk #%d (%d bytes) of data in object '%s:%s'", nBytesRead, chunkIndex, container.Name(), objectName)
	return err
}

// Delete deletes the object from Object Storage
func (o *object) Delete() error {
	if o.item == nil {
		return scerr.InvalidInstanceError()
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "").Entering().OnExitTrace()

	err := o.bucket.container.RemoveItem(o.Name)
	if err != nil {
		return err
	}
	o.item = nil
	return nil
}

// ForceAddMetadata overwrites the metadata entries of the object by the ones provided in parameter
func (o *object) ForceAddMetadata(newMetadata ObjectMetadata) error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "").Entering().OnExitTrace()

	for k, v := range newMetadata {
		o.Metadata[k] = v
	}
	return nil
}

// AddMetadata adds missing entries in object metadata
func (o *object) AddMetadata(newMetadata ObjectMetadata) error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}

	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "").Entering().OnExitTrace()

	for k, v := range newMetadata {
		_, found := o.Metadata[k]
		if !found {
			o.Metadata[k] = v
		}
	}
	return nil
}

// ReplaceMetadata replaces object metadata with the ones provided in parameter
func (o *object) ReplaceMetadata(newMetadata ObjectMetadata) error {
	if o == nil {
		return scerr.InvalidInstanceError()
	}
	defer concurrency.NewTracer(nil, debug.ShouldTrace("objectstorage"), "").Entering().OnExitTrace()

	o.Metadata = newMetadata
	return nil
}

// GetLastUpdate returns the date of last update
func (o *object) GetLastUpdate() (time.Time, error) {
	if o == nil {
		return time.Time{}, scerr.InvalidInstanceError()
	}
	if o.item == nil {
		return time.Time{}, scerr.InvalidInstanceContentError("o.item", "cannot be nil")
	}
	t, err := o.item.LastMod()
	if err != nil {
		return time.Time{}, scerr.NewError(err, nil, "")
	}
	return t, nil
}

// SafeGetLastUpdate returns the date of last update
// Intended to be used when o, o.item are notoriously not nil (because previously checked)
func (o *object) SafeGetLastUpdate() time.Time {
	t, _ := o.GetLastUpdate()
	return t
}

// GetMetadata returns the metadata of the object in Object Storage
func (o *object) GetMetadata() (ObjectMetadata, error) {
	if o == nil {
		return ObjectMetadata{}, scerr.InvalidInstanceError()
	}
	return o.Metadata.Clone(), nil
}

// SafeGetMetadata returns the metadata of the object in Object Storage
// Intended to be used when o, o.item are notoriously not nil (because previously checked)
func (o *object) SafeGetMetadata() ObjectMetadata {
	md, _ := o.GetMetadata()
	return md
}

// GetSize returns the size of the content of the object
func (o *object) GetSize() (int64, error) {
	if o == nil {
		return -1, scerr.InvalidInstanceError()
	}
	if o.item == nil {
		return -1, scerr.InvalidInstanceContentError("o.item", "cannot be nil")
	}
	size, err := o.item.Size()
	if err != nil {
		return -1, err
	}
	return size, nil
}

// SafeGetSize returns the size of the content of the object
func (o *object) SafeGetSize() int64 {
	size, _ := o.GetSize()
	return size
}

// GetETag returns the value of the ETag (+/- md5sum of the content...)
func (o *object) GetETag() (string, error) {
	if o == nil {
		return "", scerr.InvalidInstanceError()
	}
	if o.item == nil {
		return "", scerr.InvalidInstanceContentError("o.item", "cannot be nil")
	}
	etag, err := o.item.ETag()
	if err != nil {
		return "", err
	}
	return etag, nil
}

// SafeGetETag returns the value of the ETag (+/- md5sum of the content...)
// Intended to be used when o, o.item are notoriously not nil (because previously checked)
func (o *object) SafeGetETag() string {
	etag, _ := o.GetETag()
	return etag
}

// GetID ...
func (o *object) GetID() (string, error) {
	if o == nil {
		return "", scerr.InvalidInstanceError()
	}
	if o.item == nil {
		return "", scerr.InvalidInstanceContentError("o.item", "cannot be nil")
	}
	return o.item.ID(), nil
}

// SafeGetID ...
// Intended to be used when o, o.item are notoriously not nil (because previously checked)
func (o *object) SafeGetID() string {
	id, _ := o.GetID()
	return id
}

// GetName returns the name of the object
func (o *object) GetName() (string, error) {
	if o == nil {
		return "", scerr.InvalidInstanceError()
	}
	return o.Name, nil
}

// SafeGetName returns the name of the object
// Intended to be used when o, o.item are notoriously not nil (because previously checked)
func (o *object) SafeGetName() string {
	n, _ := o.GetName()
	return n
}
