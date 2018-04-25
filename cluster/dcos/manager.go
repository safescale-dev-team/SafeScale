package dcos

import (
	"fmt"

	"github.com/SafeScale/providers"
	providerapi "github.com/SafeScale/providers/api"
	"github.com/davecgh/go-spew/spew"

	clusterapi "github.com/SafeScale/cluster/api"
	"github.com/SafeScale/cluster/api/ClusterState"
	"github.com/SafeScale/cluster/api/Complexity"
	"github.com/SafeScale/cluster/api/NodeType"
)

//go:generate rice embed-go

//Manager is the implementation for DCOS
type Manager struct {
	clusterapi.ClusterManager
}

//GetService returns the service corresponding to the tenant associated with the ClusterManager
func (m *Manager) GetService() *providers.Service {
	return m.ClusterManager.Service
}

//GetTenantName returns the tenant name associated with the ClusterManager
func (m *Manager) GetTenantName() string {
	return m.ClusterManager.Tenant
}

//CreateCluster creates a cluster following the parameters of the request
func (m *Manager) CreateCluster(req clusterapi.ClusterRequest) (clusterapi.ClusterAPI, error) {
	// Validates parameters
	if req.Name == "" {
		return nil, fmt.Errorf("Invalid parameter req.Name: can't be empty")
	}
	if req.CIDR == "" {
		return nil, fmt.Errorf("Invalid parameter req.CIDR: can't be empty")
	}

	svc := m.GetService()

	// Figures out the best image for the job (DCOS supports only REHL 7, CentOS 7 and CoreOS)
	img, err := svc.SearchImage("CentOS 7")
	if err != nil {
		return nil, fmt.Errorf("Failed to find image for CentOS 7")
	}
	spew.Dump(img)

	// Figures out the best template for Bootstrap node
	tmplBootstrap, err := svc.SelectTemplatesBySize(providerapi.SizingRequirements{
		MinCores:    2,
		MinRAMSize:  16,
		MinDiskSize: 60,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to find a template suitable for bootstrap server: %s", err.Error())
	}
	spew.Dump(tmplBootstrap)

	// Figures out the best template for Master node(s)
	tmplMaster, err := svc.SelectTemplatesBySize(providerapi.SizingRequirements{
		MinCores:    4,
		MinRAMSize:  32,
		MinDiskSize: 120,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to find a template suitable for master server: %s", err.Error())
	}
	spew.Dump(tmplMaster)

	// Create a KeyPair for the cluster
	name := "key-pair-cluster-" + req.Name
	svc.DeleteKeyPair(name)
	kp, err := svc.CreateKeyPair(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key Pair: %s", err.Error())
	}

	// Saving cluster parameters, with status 'Creating'
	cluster := &Cluster{
		definition: ClusterDefinition{
			Common: clusterapi.Cluster{
				Name:       req.Name,
				State:      ClusterState.Creating,
				Complexity: req.Complexity,
				Keypair:    kp,
			},
		},
		Manager: m,
	}
	err = cluster.SaveDefinition()
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster '%s': %s", req.Name, err.Error())
	}

	// Creates network
	network, err := svc.CreateNetwork(providerapi.NetworkRequest{
		Name: req.Name,
		CIDR: req.CIDR,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create Network '%s': %s", req.Name, err.Error())
	}
	cluster.definition.NetworkID = network.ID

	masterCount := 0

	// Creates bootstrap/upgrade server
	_, err = cluster.AddNode(NodeType.Bootstrap, providerapi.VMRequest{
		TemplateID: tmplBootstrap[0].ID,
		ImageID:    img.ID,
	})
	if err != nil {
		err = fmt.Errorf("failed to create DCOS bootstrap server: %s", err.Error())
		goto cleanNetwork
	}

	switch req.Complexity {
	case Complexity.Simple:
		masterCount = 1
	case Complexity.HighAvailability:
		masterCount = 3
	case Complexity.HighVolume:
		masterCount = 5
	}

	for i := 1; i <= masterCount; i++ {
		// Creates Master Node
		_, err = cluster.AddNode(NodeType.Master, providerapi.VMRequest{
			TemplateID: tmplMaster[0].ID,
			ImageID:    img.ID,
		})
		if err != nil {
			err = fmt.Errorf("failed to add master node %d: %s", i, err.Error())
			goto cleanMasters
		}
	}

	// Cluster created successfully, saving again to Object Storage
	cluster.definition.Common.State = ClusterState.Created
	err = cluster.SaveDefinition()
	if err != nil {
		goto cleanMasters
	}
	return cluster, nil

cleanMasters:
	for _, id := range cluster.definition.MasterIDs {
		svc.DeleteVM(id)
	}
	svc.DeleteVM(cluster.definition.BootstrapID)
cleanNetwork:
	svc.DeleteNetwork(cluster.definition.NetworkID)
	cluster.RemoveDefinition()
	return nil, err
}

// DeleteCluster deletes the infrastructure of the cluster named 'name'
func (m *Manager) DeleteCluster(name string) error {
	cluster, err := m.GetCluster(name)
	if err != nil {
		return fmt.Errorf("failed to find a cluster named '%s': %s", name, err.Error())
	}

	// Cleanup Object Storage data
	return cluster.RemoveDefinition()
}

//GetCluster returns the Cluster object corresponding to the cluster named 'name'
func (m *Manager) GetCluster(name string) (clusterapi.ClusterAPI, error) {
	cluster := &Cluster{
		definition: ClusterDefinition{
			Common: clusterapi.Cluster{
				Name: name,
			},
		},
		Manager: m,
	}
	found, err := cluster.ReadDefinition()
	if err != nil {
		return nil, fmt.Errorf("failed to get Cluster '%s': %s", name, err.Error())
	}
	if found {
		_, err = cluster.GetState()
		if err != nil {
			return nil, fmt.Errorf("failed to get state of the cluster: %s", err.Error())
		}
		return cluster, nil
	}
	return nil, nil
}

//ListClusters lists the clusters already created
func (m *Manager) ListClusters() (*[]string, error) {
	return nil, nil
}

//StartCluster starts the cluster named 'name'
func (m *Manager) StartCluster(name string) error {
	cluster, err := m.GetCluster(name)
	if err != nil {
		return fmt.Errorf("Failed to find cluster named '%s': %s", name, err.Error())
	}
	return cluster.Start()
}

//StopCluster stops the cluster named 'name'.
func (m *Manager) StopCluster(name string) error {
	cluster, err := m.GetCluster(name)
	if err != nil {
		return fmt.Errorf("Failed to find cluster named '%s': %s", name, err.Error())
	}
	return cluster.Stop()
}

//GetState returns the state of the cluster named 'name'
func (m *Manager) GetState(name string) (ClusterState.Enum, error) {
	cluster, err := m.GetCluster(name)
	if err != nil {
		return ClusterState.Error, fmt.Errorf("Failed to find cluster named '%s': %s", name, err.Error())
	}
	return cluster.GetState()
}
