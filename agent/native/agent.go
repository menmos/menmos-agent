package native

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/google/uuid"
	"github.com/menmos/menmos-agent/agent/amphora"
	"github.com/menmos/menmos-agent/agent/menmosd"
	"github.com/menmos/menmos-agent/agent/native/artifact"
	"github.com/menmos/menmos-agent/agent/native/xecute"
	"github.com/menmos/menmos-agent/payload"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

const AGENT_NODE_INFO_FILE = ".agent_node_info.json"

type NativeAgent struct {
	// Misc. management stuff.
	config NativeAgentConfig
	log    *zap.SugaredLogger

	artifacts *artifact.Repository

	// State
	runningNodes map[string]*xecute.Native
}

func New(config NativeAgentConfig, log *zap.Logger) (*NativeAgent, error) {
	// Using a github release fetcher by default.

	agent := &NativeAgent{
		config: config,
		log:    log.Sugar().Named("native-agent"),
		artifacts: artifact.NewRepository(
			artifact.RepositoryParams{
				ReleaseFetcher: artifact.NewGithubFetcher(config.GithubToken),
				Log:            log,
				Path:           path.Join(config.Path, "pkg"),
			},
		),
		runningNodes: make(map[string]*xecute.Native),
	}

	if err := agent.initWorkspace(); err != nil {
		return nil, err
	}

	if err := agent.restartComponents(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *NativeAgent) pkgDir() string {
	return path.Join(a.config.Path, "pkg")
}

func (a *NativeAgent) nodeDir() string {
	return path.Join(a.config.Path, "node")
}

func (a *NativeAgent) initWorkspace() error {
	a.log.Debug("initializing menmos agent workspace")

	if err := ensureDirExists(a.config.Path); err != nil {
		return err
	}

	if err := ensureDirExists(a.pkgDir()); err != nil {
		return err
	}

	if err := ensureDirExists(a.nodeDir()); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) restartComponents() error {
	entries, err := os.ReadDir(a.nodeDir())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if err := a.StartNode(entry.Name()); err != nil {
			a.log.Errorf("failed to restart component '%s': %v", entry.Name(), err)
		}
	}

	return nil
}

func (a *NativeAgent) getBinary(version, binary string) (binaryPath string, err error) {
	if version != "" {
		return a.artifacts.Get(version, binary)
	} else if a.config.LocalBinaryPath != "" {
		// Fallback on local binaries.
		binaryPath = path.Join(a.config.LocalBinaryPath, binary)
	}

	return
}

func (a *NativeAgent) createMenmosdConfig(nodeDir string, request *payload.CreateNodeRequest, config *payload.MenmosdConfig) error {
	procConfig := menmosd.Config{
		Node: menmosd.NodeSetting{
			DbPath:           path.Join(nodeDir, "db"),
			AdminPassword:    config.NodeAdminPassword,
			EncryptionKey:    config.NodeEncryptionKey,
			RoutingAlgorithm: config.NodeRoutingAlgorithm,
		},
		Server: menmosd.ServerSetting{
			Type: config.ServerMode,
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) createAmphoraConfig(nodeDir string, request *payload.CreateNodeRequest, config *payload.AmphoraConfig) error {

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
		},
		Redirect: amphora.RedirectConfig{
			Ip:         config.RedirectIp,
			SubnetMask: config.SubnetMask,
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) getNodeInfo(nodeID string) (nodeInfo, error) {
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

func (a *NativeAgent) GetNode(nodeID string) (*payload.NodeResponse, error) {
	if process, ok := a.runningNodes[nodeID]; ok {
		info, err := a.getNodeInfo(nodeID)
		if err != nil {
			return nil, err
		}

		nodeResp := &payload.NodeResponse{
			ID:      nodeID,
			Binary:  string(info.Binary),
			Version: info.Version,
			Port:    process.Port(),
			Status:  process.Status(),
		}

		return nodeResp, nil
	}

	return nil, nil
}

func (a *NativeAgent) ListNodes() (*payload.ListNodesResponse, error) {
	var resp payload.ListNodesResponse

	for nodeID, process := range a.runningNodes {
		info, err := a.getNodeInfo(nodeID)
		if err != nil {
			return nil, err
		}

		nodeResp := &payload.NodeResponse{
			ID:      nodeID,
			Binary:  string(info.Binary),
			Version: info.Version,
			Port:    process.Port(),
			Status:  process.Status(),
		}

		resp.Nodes = append(resp.Nodes, nodeResp)
	}

	return &resp, nil
}

func (a *NativeAgent) CreateNode(request *payload.CreateNodeRequest) (*payload.NodeResponse, error) {
	binPath, err := a.getBinary(request.Version, string(request.Type))
	if err != nil {
		return nil, err
	}

	nodeID := uuid.New().String()

	nodeDir := path.Join(a.nodeDir(), nodeID)
	if err := ensureDirExists(nodeDir); err != nil {
		return nil, err
	}

	if request.Type == payload.NodeMenmosd {
		var requestConfig payload.MenmosdConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		if err = a.createMenmosdConfig(nodeDir, request, &requestConfig); err != nil {
			return nil, err
		}
	} else if request.Type == payload.NodeAmphora {
		var requestConfig payload.AmphoraConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		if err = a.createAmphoraConfig(nodeDir, request, &requestConfig); err != nil {
			return nil, err
		}
	}

	process, err := xecute.NewNativeProcess(nodeDir, binPath, a.log.Named(string(request.Type)).Named(nodeID).Desugar())
	if err != nil {
		return nil, err
	}

	if err := process.Start(xecute.LogNormal); err != nil { // TODO: Find a way to customize this loglevel.
		return nil, err
	}

	a.runningNodes[nodeID] = process

	// If we made it here, we commit a nodeinfo file along with the process.
	// This file contains the info required to restart the process.
	nodeInfo := nodeInfo{Version: request.Version, Binary: string(request.Type)}
	if err := jsonWrite(nodeInfo, path.Join(nodeDir, AGENT_NODE_INFO_FILE)); err != nil {
		return nil, err
	}

	return &payload.NodeResponse{
		ID:      nodeID,
		Binary:  string(request.Type),
		Version: request.Version,
		Port:    process.Port(),
		Status:  process.Status(),
	}, nil
}

func (a *NativeAgent) DeleteNode(nodeID string) error {
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

func (a *NativeAgent) StartNode(nodeID string) error {
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

	process, err := xecute.NewNativeProcess(nodeDir, binPath, a.log.Named(info.Binary).Named(nodeID).Desugar())
	if err != nil {
		return err
	}

	if err := process.Start(xecute.LogNormal); err != nil { // TODO: Find a way to customize this loglevel.
		return err
	}
	a.runningNodes[nodeID] = process

	a.log.Infof("node '%s' started", nodeID)

	return nil
}

func (a *NativeAgent) StopNode(nodeID string) error {
	if process, ok := a.runningNodes[nodeID]; ok {
		if err := process.Stop(); err != nil {
			return err
		}
	}

	return nil
}

func (a *NativeAgent) GetNodeLogs(nodeID string, nbOfLines uint) (*payload.GetLogsResponse, error) {
	if process, ok := a.runningNodes[nodeID]; ok {
		return &payload.GetLogsResponse{
			Log: process.GetLogs(nbOfLines),
		}, nil
	} else {
		return nil, fmt.Errorf("node '%s' does not exist", nodeID)
	}
}

func (a *NativeAgent) Shutdown() {
	var wg sync.WaitGroup
	for nodeID := range a.runningNodes {
		wg.Add(1)
		go func(nodeID string) {
			defer wg.Done()
			if err := a.StopNode(nodeID); err != nil {
				a.log.Errorf("failed to shutdown node '%s': %v", nodeID, err)
			}
		}(nodeID)
	}
	wg.Wait()

	a.log.Info("agent shutdown successfully")
}
