package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
	"github.com/menmos/menmos-agent/agent/menmosd"
	"github.com/menmos/menmos-agent/agent/xecute"
	"github.com/mitchellh/mapstructure"
)

func (a *MenmosAgent) createMenmosdConfig(nodeDir string, request *CreateNodeRequest, config *MenmosdConfig) (string, error) {
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

func (a *MenmosAgent) CreateNode(request *CreateNodeRequest) (CreateNodeResponse, error) {
	binPath, err := a.getBinary(request.Version, string(request.Type))
	if err != nil {
		return CreateNodeResponse{}, err
	}

	nodeID := uuid.New().String()

	nodeDir := path.Join(a.nodeDir(), nodeID)
	if err := ensureDirExists(nodeDir); err != nil {
		return CreateNodeResponse{}, err
	}

	var healthCheckURL string
	if request.Type == NodeMenmosd {
		var requestConfig MenmosdConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return CreateNodeResponse{}, err
		}
		healthCheckURL, err = a.createMenmosdConfig(nodeDir, request, &requestConfig)
		if err != nil {
			return CreateNodeResponse{}, err
		}
	} else if request.Type == NodeAmphora {
		// FIXME
		return CreateNodeResponse{}, errors.New("unimplemented")
	}

	process := xecute.NewNativeProcess(nodeDir, binPath, healthCheckURL, a.log.Named(string(request.Type)).Named(nodeID).Desugar())
	if err := process.Start(xecute.LogNormal); err != nil { // TODO: Find a way to customize this loglevel.
		return CreateNodeResponse{}, err
	}

	a.runningNodes[nodeID] = process

	// If we made it here, we commit a nodeinfo file along with the process.
	// This file contains the info required to restart the process.
	nodeInfo := nodeInfo{Version: request.Version, Binary: string(request.Type), HealthCheckURL: healthCheckURL}
	if err := jsonWrite(nodeInfo, path.Join(nodeDir, AGENT_NODE_INFO_FILE)); err != nil {
		return CreateNodeResponse{}, err
	}

	// TODO: Return the non-sensitive config also
	return CreateNodeResponse{ID: nodeID}, nil
}

func (a *MenmosAgent) StartNode(nodeID string) error {
	if _, ok := a.runningNodes[nodeID]; ok {
		return fmt.Errorf("node '%s' is already running", nodeID)
	}

	nodeDir := path.Join(a.nodeDir(), nodeID)

	// Load the agent node info.
	infoBytes, err := os.ReadFile(path.Join(nodeDir, AGENT_NODE_INFO_FILE))
	if err != nil {
		return err
	}

	var info nodeInfo
	if err := json.Unmarshal(infoBytes, &info); err != nil {
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
