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

package services

import (
	"fmt"
	"path"
	"strings"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/satori/go.uuid"

	"github.com/CS-SI/SafeScale/providers"
	"github.com/CS-SI/SafeScale/providers/metadata"
	"github.com/CS-SI/SafeScale/providers/model"
	"github.com/CS-SI/SafeScale/providers/model/enums/HostProperty"
	propsv1 "github.com/CS-SI/SafeScale/providers/model/properties/v1"
	"github.com/CS-SI/SafeScale/system/nfs"
	"github.com/CS-SI/SafeScale/utils"
)

//go:generate mockgen -destination=../mocks/mock_nasapi.go -package=mocks github.com/CS-SI/SafeScale/broker/server/services ShareAPI

// ShareAPI defines API to manipulate Shares
type ShareAPI interface {
	Create(name, host, path string) (*propsv1.HostShare, error)
	Delete(name string) error
	List() (map[string]map[string]*propsv1.HostShare, error)
	Mount(name, host, path string) (*propsv1.HostRemoteMount, error)
	Unmount(name, host string) error
	Inspect(name string) (*model.Host, *propsv1.HostShare, error)
}

// ShareService nas service
type ShareService struct {
	provider *providers.Service
}

// NewShareService creates a ShareService
func NewShareService(api *providers.Service) ShareAPI {
	return &ShareService{
		provider: api,
	}
}

func sanitize(in string) (string, error) {
	sanitized := path.Clean(in)
	if !path.IsAbs(sanitized) {
		return "", fmt.Errorf("Exposed path must be absolute")
	}
	return sanitized, nil
}

// Create a share on host
func (svc *ShareService) Create(shareName, hostName, path string) (*propsv1.HostShare, error) {
	// Check if a share already exists with the same name
	serverName, err := svc.findShare(shareName)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	if serverName != "" {
		return nil, model.ResourceAlreadyExistsError("share", shareName)
	}
	hostSvc := NewHostService(svc.provider)
	server, err := hostSvc.Get(hostName)
	if err != nil {
		return nil, err
	}

	// Sanitize path
	sharePath, err := sanitize(path)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	// Installs NFS Server software if needed
	sshSvc := NewSSHService(svc.provider)
	sshConfig, err := sshSvc.GetConfig(server)
	if err != nil {
		tbr := errors.Wrap(err, "Error getting NAS ssh config")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	nfsServer, err := nfs.NewServer(sshConfig)
	if err != nil {
		tbr := errors.Wrap(err, "Error creating NAS structure")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	serverSharesV1 := propsv1.NewHostShares()
	err = server.Properties.Get(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return nil, err
	}
	if len(serverSharesV1.ByID) == 0 {
		// Host doesn't have shares yet, so install NFS
		err = nfsServer.Install()
		if err != nil {
			tbr := errors.Wrap(err, "")
			log.Errorf("%+v", tbr)
			return nil, tbr
		}
	}
	err = nfsServer.AddShare(sharePath, "")
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	// Create share struct
	share := propsv1.NewHostShare()
	share.Name = shareName
	shareID, err := uuid.NewV4()
	if err != nil {
		tbr := errors.Wrap(err, "Error creating UUID for share")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	share.ID = shareID.String()
	share.Path = sharePath
	share.Type = "nfs"

	serverSharesV1.ByID[share.ID] = share
	serverSharesV1.ByName[share.Name] = share.ID

	// Updates Host Property propsv1.HostShares
	err = server.Properties.Set(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return nil, err
	}

	err = metadata.SaveHost(svc.provider, server)
	if err != nil {
		tbr := errors.Wrap(err, "Error saving server metadata")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	err = metadata.SaveShare(svc.provider, server.ID, server.Name, share.ID, share.Name)
	if err != nil {
		return nil, err
	}

	return share, nil
}

// Delete a share from host
func (svc *ShareService) Delete(name string) error {
	// Retrieve info about the share
	server, share, err := svc.Inspect(name)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}

	serverSharesV1 := propsv1.NewHostShares()
	err = server.Properties.Get(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return err
	}

	if len(share.ClientsByName) > 0 {
		list := []string{}
		for k := range share.ClientsByName {
			list = append(list, k)
		}
		return fmt.Errorf("host%s still using it: %s", utils.Plural(len(list)), strings.Join(list, ","))
	}

	sshSvc := NewSSHService(svc.provider)
	sshConfig, err := sshSvc.GetConfig(server.ID)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}

	nfsServer, err := nfs.NewServer(sshConfig)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}
	err = nfsServer.RemoveShare(share.Path)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}

	delete(serverSharesV1.ByID, share.ID)
	delete(serverSharesV1.ByName, share.Name)
	err = server.Properties.Set(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return err
	}

	// Save server metadata
	err = metadata.SaveHost(svc.provider, server)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}
	// Remove share metadata
	return metadata.RemoveShare(svc.provider, server.ID, server.Name, share.ID, share.Name)
}

// List return the list of all shares from all servers
func (svc *ShareService) List() (map[string]map[string]*propsv1.HostShare, error) {
	shares := map[string]map[string]*propsv1.HostShare{}

	servers := []string{}
	ms := metadata.NewShare(svc.provider)
	err := ms.Browse(func(hostName string, shareID string) error {
		servers = append(servers, hostName)
		return nil
	})
	if err != nil {
		tbr := errors.Wrap(err, "Error browsing NASes")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	// Now walks through the hosts acting as Nas
	if len(servers) == 0 {
		return nil, nil
	}

	hostSvc := NewHostService(svc.provider)
	for _, serverID := range servers {
		host, err := hostSvc.Get(serverID)
		if err != nil {
			return nil, err
		}

		hostSharesV1 := propsv1.NewHostShares()
		err = host.Properties.Get(HostProperty.SharesV1, hostSharesV1)
		if err != nil {
			return nil, err
		}

		shares[serverID] = hostSharesV1.ByID
	}
	return shares, nil
}

// Mount a share on a local directory of an host
func (svc *ShareService) Mount(shareName, hostName, path string) (*propsv1.HostRemoteMount, error) {

	// Sanitize path
	mountPath, err := sanitize(path)
	if err != nil {
		return nil, fmt.Errorf("invalid mount path '%s': '%s'", path, err)
	}

	// Retrieve info about the share
	server, _, err := svc.Inspect(shareName)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	hostSvc := NewHostService(svc.provider)
	target, err := hostSvc.Get(hostName)
	if err != nil {
		return nil, model.ResourceNotFoundError("host", hostName)
	}

	// Checks if there is no other device mounted in the path (or in subpath)
	// Checks if there is already something mounted in the path
	targetMountsV1 := propsv1.NewHostMounts()
	err = target.Properties.Get(HostProperty.MountsV1, targetMountsV1)
	if err != nil {
		return nil, err
	}
	for _, i := range targetMountsV1.LocalMountsByPath {
		if i.Path == path {
			// Can't mount a share in place of a volume (by convention, nothing technically preventing it)
			return nil, fmt.Errorf("Can't mount share '%s' to host '%s': there is already a volume in path '%s'", shareName, target.Name, path)
		}
	}
	for _, i := range targetMountsV1.RemoteMountsByPath {
		if strings.Index(i.Path, path) == 0 {
			// Can't mount a share inside another share (at least by convention, if not technically)
			return nil, fmt.Errorf("Can't mount share volume '%s' to host '%s': there is a share mounted in path '%s[/...]'", shareName, target.Name, path)
		}
	}

	// Mount the share on host
	serverSharesV1 := propsv1.NewHostShares()
	err = server.Properties.Get(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return nil, err
	}
	_, found := serverSharesV1.ByID[serverSharesV1.ByName[shareName]]
	if !found {
		return nil, fmt.Errorf("failed to find metadata about share '%s'", shareName)
	}
	shareID := serverSharesV1.ByName[shareName]
	share := serverSharesV1.ByID[shareID]
	sshSvc := NewSSHService(svc.provider)
	sshConfig, err := sshSvc.GetConfig(target)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	nfsClient, err := nfs.NewNFSClient(sshConfig)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}
	err = nfsClient.Install()
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	err = nfsClient.Mount(server.GetAccessIP(), share.Path, mountPath)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return nil, tbr
	}

	serverSharesV1.ByID[shareID].ClientsByName[target.Name] = target.ID
	serverSharesV1.ByID[shareID].ClientsByID[target.ID] = target.Name
	err = server.Properties.Set(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return nil, err
	}

	mount := propsv1.NewHostRemoteMount()
	mount.ShareID = share.ID
	mount.Export = server.GetAccessIP() + ":" + share.Path
	mount.Path = mountPath
	mount.FileSystem = "nfs"
	targetMountsV1.RemoteMountsByPath[mount.Path] = mount
	targetMountsV1.RemoteMountsByShareID[mount.ShareID] = mount.Path
	targetMountsV1.RemoteMountsByExport[mount.Export] = mount.Path
	err = target.Properties.Set(HostProperty.MountsV1, targetMountsV1)
	if err != nil {
		return nil, err
	}

	err = metadata.SaveHost(svc.provider.ClientAPI, target)
	if err != nil {
		return nil, err
	}
	err = metadata.SaveHost(svc.provider.ClientAPI, server)
	if err != nil {
		return nil, err
	}
	return mount, nil
}

// Unmount a share from local directory of an host
func (svc *ShareService) Unmount(shareName, hostName string) error {
	server, _, err := svc.Inspect(shareName)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}

	serverSharesV1 := propsv1.NewHostShares()
	err = server.Properties.Get(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return err
	}
	shareID, found := serverSharesV1.ByName[shareName]
	if !found {
		return fmt.Errorf("failed to find data about share '%s'", shareName)
	}
	share := serverSharesV1.ByID[shareID]
	remotePath := server.GetAccessIP() + ":" + share.Path

	hostSvc := NewHostService(svc.provider)
	target, err := hostSvc.Get(hostName)
	if err != nil {
		return err
	}
	if target == nil {
		return model.ResourceNotFoundError("host", hostName)
	}
	targetMountsV1 := propsv1.NewHostMounts()
	err = target.Properties.Get(HostProperty.MountsV1, targetMountsV1)
	if err != nil {
		return err
	}
	mount, found := targetMountsV1.RemoteMountsByPath[targetMountsV1.RemoteMountsByShareID[shareID]]
	if !found {
		return fmt.Errorf("share '%s' not mounted on host '%s'", remotePath, target.Name)
	}

	// Unmount share from client
	sshSvc := NewSSHService(svc.provider)
	sshConfig, err := sshSvc.GetConfig(target.ID)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}
	nfsClient, err := nfs.NewNFSClient(sshConfig)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}
	err = nfsClient.Unmount(server.GetAccessIP(), mount.Path)
	if err != nil {
		tbr := errors.Wrap(err, "")
		log.Errorf("%+v", tbr)
		return tbr
	}

	// Remove mount from mount list
	delete(targetMountsV1.RemoteMountsByShareID, mount.ShareID)
	delete(targetMountsV1.RemoteMountsByPath, mount.Path)
	err = target.Properties.Set(HostProperty.MountsV1, targetMountsV1)
	if err != nil {
		return err
	}

	// Remove host from client lists of the share
	delete(serverSharesV1.ByID[shareID].ClientsByName, target.Name)
	delete(serverSharesV1.ByID[shareID].ClientsByID, target.ID)
	err = server.Properties.Set(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return err
	}

	// Saves metadata
	err = metadata.SaveHost(svc.provider.ClientAPI, server)
	if err != nil {
		return err
	}
	err = metadata.SaveHost(svc.provider.ClientAPI, target)
	if err != nil {
		return err
	}

	return nil
}

// Inspect returns the host and share corresponding to 'shareName'
func (svc *ShareService) Inspect(shareName string) (*model.Host, *propsv1.HostShare, error) {

	hostName, err := metadata.LoadShare(svc.provider, shareName)
	if err != nil {
		tbr := errors.Wrap(err, "error loading share metadata")
		log.Errorf("%+v", tbr)
		return nil, nil, tbr
	}
	if hostName == "" {
		return nil, nil, model.ResourceNotFoundError("share", "")
	}

	hostSvc := NewHostService(svc.provider)
	server, err := hostSvc.Get(hostName)
	if err != nil {
		return nil, nil, err
	}
	serverSharesV1 := propsv1.NewHostShares()
	err = server.Properties.Get(HostProperty.SharesV1, serverSharesV1)
	if err != nil {
		return nil, nil, err
	}
	shareID, found := serverSharesV1.ByName[shareName]
	if !found {
		shareID = shareName
		_, found = serverSharesV1.ByID[shareID]
	}
	if !found {
		return nil, nil, err
	}
	return server, serverSharesV1.ByID[shareID], nil
}

func (svc *ShareService) findShare(shareName string) (string, error) {
	hostName, err := metadata.LoadShare(svc.provider, shareName)
	if err != nil {
		tbr := errors.Wrap(err, "Failed to load Share metadata")
		log.Errorf("%+v", tbr)
		return "", tbr
	}
	return hostName, nil
}
