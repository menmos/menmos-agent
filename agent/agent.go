package agent

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/menmos/menmos-agent/agent/creator"
	"github.com/menmos/menmos-agent/agent/native"
	"github.com/menmos/menmos-agent/agent/store"
	"github.com/menmos/menmos-agent/agent/xecute"
	"github.com/menmos/menmos-agent/payload"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Agent struct {
	cfg Config
	log *zap.SugaredLogger

	creator      creator.NodeCreator
	runningNodes map[string]creator.Node
	state        store.StateStore
}

func New(config Config, log *zap.Logger) (*Agent, error) {
	var cr creator.NodeCreator
	var err error

	sugared := log.Sugar().Named("agent")

	switch config.AgentType {
	case TypeNative:
		nativeCfg := native.NativeAgentConfig{
			S3Bucket:        config.S3Bucket,
			S3Region:        config.S3Region,
			GithubToken:     config.GithubToken,
			Path:            config.Path,
			LocalBinaryPath: config.LocalBinaryPath,
		}
		cr, err = native.New(nativeCfg, sugared.Desugar())
	case TypeDocker:
		return nil, errors.New("docker agent is not implemented")

	case TypeKubernetes:
		return nil, errors.New("kubernetes agent is not implemented")

	default:
		return nil, fmt.Errorf("unknown agent type: %s", config.AgentType)
	}

	if err != nil {
		return nil, err
	}

	store, err := store.NewJSONStore(filepath.Join(config.Path, "db"))
	if err != nil {
		return nil, err
	}

	agent := &Agent{
		cfg: config,
		log: log.Sugar().Named("agent"),

		creator:      cr,
		runningNodes: make(map[string]creator.Node),
		state:        store,
	}

	if err := agent.restore(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *Agent) restore() error {
	var node creator.Node
	var err error

	nodes, err := a.state.ListNodes()
	if err != nil {
		return err
	}

	for _, info := range nodes {
		switch info.Binary {
		case "menmosd":
			node, err = a.creator.RestoreMenmosd(info.ID, info.Version)

		case "amphora":
			node, err = a.creator.RestoreAmphora(info.ID, info.Version)

		default:
			return errors.New("unknown binary")
		}

		if err != nil {
			return err
		}
		a.runningNodes[info.ID] = node
		if err := node.Start(xecute.LogNormal); err != nil {
			return err
		}
	}
	return nil
}

// GetNode gets all the information of a node.
func (a *Agent) GetNode(nodeID string) (*payload.NodeResponse, error) {
	node, ok := a.runningNodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node '%s' does not exist", nodeID)
	}

	info, err := a.state.Get(nodeID)
	if err != nil {
		return nil, err
	}

	return &payload.NodeResponse{
		ID:      nodeID,
		Binary:  string(info.Binary),
		Version: info.Version,
		Port:    node.Port(),
		Status:  node.Status(),
	}, nil
}

// ListNode returns the information of all nodes.
func (a *Agent) ListNodes() (*payload.ListNodesResponse, error) {
	nodes := make([]*payload.NodeResponse, len(a.runningNodes))

	i := 0
	for nodeID, node := range a.runningNodes {
		info, err := a.state.Get(nodeID)
		if err != nil {
			return nil, fmt.Errorf("failed to read state entry: %w", err)
		}

		nodes[i] = &payload.NodeResponse{
			ID:      nodeID,
			Binary:  string(info.Binary),
			Version: info.Version,
			Port:    node.Port(),
			Status:  node.Status(),
		}

		i++
	}

	return &payload.ListNodesResponse{Nodes: nodes}, nil
}

// CreateNode creates a new node from the given information and starts it.
func (a *Agent) CreateNode(request *payload.CreateNodeRequest) (*payload.NodeResponse, error) {

	id := uuid.New().String()

	var node creator.Node
	var err error

	if request.Type == payload.NodeMenmosd {
		var requestConfig payload.MenmosdConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		node, err = a.creator.CreateMenmosd(id, request.Version, &requestConfig)
	} else if request.Type == payload.NodeAmphora {
		var requestConfig payload.AmphoraConfig
		if err := mapstructure.Decode(request.Config, &requestConfig); err != nil {
			return nil, err
		}
		node, err = a.creator.CreateAmphora(id, request.Version, &requestConfig)
	}

	if err != nil {
		return nil, err
	}

	if err := node.Start(xecute.LogNormal); err != nil {
		return nil, err
	}

	a.runningNodes[id] = node

	info := store.NodeInfo{
		ID:      id,
		Binary:  string(request.Type),
		Version: request.Version,
	}
	if err := a.state.Save(info); err != nil {
		return nil, err
	}

	return &payload.NodeResponse{
		ID:      id,
		Binary:  string(request.Type),
		Version: request.Version,
		Port:    node.Port(),
		Status:  node.Status(),
	}, nil
}

// DeleteNode deletes a node. The node must be in a stopped state to be deleted.
func (a *Agent) DeleteNode(nodeID string) error {
	if node, ok := a.runningNodes[nodeID]; ok {
		if err := node.Delete(); err != nil {
			return err
		}
		delete(a.runningNodes, nodeID)
	}
	return nil
}

// StartNode starts an existing node.
func (a *Agent) StartNode(nodeID string) error {
	if node, ok := a.runningNodes[nodeID]; ok {
		if err := node.Start(xecute.LogNormal); err != nil {
			return err
		}
	}

	return nil
}

/// StopNode stops an existing node.
func (a *Agent) StopNode(nodeID string) error {
	if node, ok := a.runningNodes[nodeID]; ok {
		if err := node.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// GetNodeLogs fetches the last |nbOfLines| log lines from the node logs.
func (a *Agent) GetNodeLogs(nodeID string, nbOfLines uint) (*payload.GetLogsResponse, error) {
	if node, ok := a.runningNodes[nodeID]; ok {
		return &payload.GetLogsResponse{
			Log: node.Logs(nbOfLines),
		}, nil
	}

	return nil, errors.New("node not found")
}

// Shutdown stops all nodes and kills the agent.
func (a *Agent) Shutdown() {
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
