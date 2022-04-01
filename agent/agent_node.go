package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/menmos/menmos-agent/agent/amphora"
	"github.com/menmos/menmos-agent/agent/menmosd"
	"github.com/menmos/menmos-agent/agent/xecute"
	"github.com/menmos/menmos-agent/payload"
	"github.com/mitchellh/mapstructure"
)

func (a *MenmosAgent) createMenmosdConfig(nodeDir string, request *payload.CreateNodeRequest, config *payload.MenmosdConfig) (string, error) {
	procConfig := menmosd.Config{
		Node: menmosd.NodeSetting{
			DbPath:           path.Join(nodeDir, "db"),
			AdminPassword:    config.NodeAdminPassword,
			EncryptionKey:    config.NodeEncryptionKey,
			RoutingAlgorithm: config.NodeRoutingAlgorithm,
		},
		Server: menmosd.ServerSetting{
			Type: config.ServerMode,
			Port: config.Port, // TODO: Find free port.
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return "", err
	}

	return procConfig.HealthCheckURL(), nil
}

func (a *MenmosAgent) createAmphoraConfig(nodeDir string, request *payload.CreateNodeRequest, config *payload.AmphoraConfig) (string, error) {

	storageConfig := amphora.BlobStorageConfig{Type: string(config.BlobStorageType)}
	if config.BlobStorageType == payload.BlobStorageDisk {
		storageConfig.Type = "Directory"
		storageConfig.Path = path.Join(nodeDir, "blob")
	} else {
		storageConfig.Type = "S3"
		storageConfig.Bucket = a.config.S3Bucket
		storageConfig.Region = a.config.S3Region
		storageConfig.CachePath = path.Join(nodeDir, "cache")
		storageConfig.CacheSize = 1024 * 1024 * 1024 //  1Gb TODO: Customize
	}

	procConfig := amphora.Config{
		Directory: amphora.DirectoryConfig{
			URL:  config.DirectoryHost,
			Port: config.DirectoryPort,
		},
		Node: amphora.NodeConfig{
			Name:            config.Name,
			DbPath:          path.Join(nodeDir, "db"),
			EncryptionKey:   config.NodeEncryptionKey,
			MaximumCapacity: config.MaximumCapacity,
			BlobStorage:     storageConfig,
		},
		Server: amphora.ServerConfig{
			CertificateStoragePath: path.Join(nodeDir, "cert"),
			Port:                   config.ServerPort,
		},
		Redirect: amphora.RedirectConfig{
			Ip:         config.RedirectIp,
			SubnetMask: config.SubnetMask,
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return "", err
	}

	return procConfig.HealthCheckURL(), nil
}

func (a *MenmosAgent) GetNode(nodeID string) (*payload.NodeResponse, error) {
	if process, ok := a.runningNodes[nodeID]; ok {
		info, err := a.getNodeInfo(nodeID)
		if err != nil {
			return nil, err
		}

		nodeResp := &payload.NodeResponse{
			ID:             nodeID,
			Binary:         string(info.Binary),
			Version:        info.Version,
			HealthCheckURL: info.HealthCheckURL,
			Status:         process.Status(),
		}

		return nodeResp, nil
	}

	return nil, nil
}

func (a *MenmosAgent) ListNodes() (*payload.ListNodesResponse, error) {
	var resp payload.ListNodesResponse

	for nodeID, process := range a.runningNodes {
		info, err := a.getNodeInfo(nodeID)
		if err != nil {
			return nil, err
		}

		nodeResp := &payload.NodeResponse{
			ID:             nodeID,
			Binary:         string(info.Binary),
			Version:        info.Version,
			HealthCheckURL: info.HealthCheckURL,
			Status:         process.Status(),
		}

		resp.Nodes = append(resp.Nodes, nodeResp)
	}

	return &resp, nil
}

func (a *MenmosAgent) CreateNode(request *payload.CreateNodeRequest) (*payload.NodeResponse, error) {
	binPath, err := a.getBinary(request.Version, string(request.Type))
	if err != nil {
		return nil, err
	}

	nodeID := uuid.New().String()

	nodeDir := path.Join(a.nodeDir(), nodeID)
	if err := ensureDirExists(nodeDir); err != nil {
		return nil, err
	}

	var healthCheckURL string
	if request.Type == payload.NodeMenmosd {
		var requestConfig payload.MenmosdConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		healthCheckURL, err = a.createMenmosdConfig(nodeDir, request, &requestConfig)
		if err != nil {
			return nil, err
		}
	} else if request.Type == payload.NodeAmphora {
		var requestConfig payload.AmphoraConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		healthCheckURL, err = a.createAmphoraConfig(nodeDir, request, &requestConfig)
		if err != nil {
			return nil, err
		}
	}

	process := xecute.NewNativeProcess(nodeDir, binPath, healthCheckURL, a.log.Named(string(request.Type)).Named(nodeID).Desugar())
	if err := process.Start(xecute.LogNormal); err != nil { // TODO: Find a way to customize this loglevel.
		return nil, err
	}

	a.runningNodes[nodeID] = process

	// If we made it here, we commit a nodeinfo file along with the process.
	// This file contains the info required to restart the process.
	nodeInfo := nodeInfo{Version: request.Version, Binary: string(request.Type), HealthCheckURL: healthCheckURL}
	if err := jsonWrite(nodeInfo, path.Join(nodeDir, AGENT_NODE_INFO_FILE)); err != nil {
		return nil, err
	}

	return &payload.NodeResponse{
		ID:             nodeID,
		Binary:         string(request.Type),
		Version:        request.Version,
		HealthCheckURL: healthCheckURL,
		Status:         process.Status(),
	}, nil
}

func (a *MenmosAgent) DeleteNode(nodeID string) error {
	if process, ok := a.runningNodes[nodeID]; ok {
		status := process.Status()
		if status == xecute.StatusStopped || status == xecute.StatusError {
			delete(a.runningNodes, nodeID)
			return os.RemoveAll(path.Join(a.nodeDir(), nodeID))
		} else {
			return fmt.Errorf("cannot delete node in '%v' state, node needs to be stopped", status)
		}
	}

	return nil
}

func (a *MenmosAgent) StopNode(nodeID string) error {
	if process, ok := a.runningNodes[nodeID]; ok {
		if err := process.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (a *MenmosAgent) getNodeInfo(nodeID string) (nodeInfo, error) {
	nodeDir := path.Join(a.nodeDir(), nodeID)

	// Load the agent node info.
	infoBytes, err := os.ReadFile(path.Join(nodeDir, AGENT_NODE_INFO_FILE))
	if err != nil {
		return nodeInfo{}, err
	}

	var info nodeInfo
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return nodeInfo{}, err
	}

	return info, nil
}

func (a *MenmosAgent) StartNode(nodeID string) error {
	if process, ok := a.runningNodes[nodeID]; ok && (process.Status() != xecute.StatusStopped && process.Status() != xecute.StatusError) {
		return fmt.Errorf("node '%s' is already running", nodeID)
	}

	nodeDir := path.Join(a.nodeDir(), nodeID)

	info, err := a.getNodeInfo(nodeID)
	if err != nil {
		return err
	}

	binPath, err := a.getBinary(info.Version, info.Binary)
	if err != nil {
		return err
	}

	process := xecute.NewNativeProcess(nodeDir, binPath, info.HealthCheckURL, a.log.Named(info.Binary).Named(nodeID).Desugar())
	if err := process.Start(xecute.LogNormal); err != nil { // TODO: Find a way to customize this loglevel.
		return err
	}
	a.runningNodes[nodeID] = process

	a.log.Infof("node '%s' started", nodeID)

	return nil
}

// TODO: Add call to get logs of a node
//		 (either JSON or text if the process crashed pre-json)
