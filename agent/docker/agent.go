package docker

import (
	"errors"

	"github.com/docker/docker/client"
	"github.com/menmos/menmos-agent/payload"
	"go.uber.org/zap"
)

type Agent struct {
	log *zap.SugaredLogger

	client *client.Client
	images *imageRegistry
}

func New(log *zap.Logger) (*Agent, error) {
	return &Agent{
		log: log.Sugar().Named("docker-agent"),
	}, nil
}

// GetNode gets all the information of a node.
func (a *Agent) GetNode(nodeID string) (*payload.NodeResponse, error) {
	return nil, errors.New("unimplemented")
}

// ListNode returns the information of all nodes.
func (a *Agent) ListNodes() (*payload.ListNodesResponse, error) {

	return nil, errors.New("unimplemented")
}

// CreateNode creates a new node from the given information.
func (a *Agent) CreateNode(request *payload.CreateNodeRequest) (*payload.NodeResponse, error) {
	return nil, errors.New("unimplemented")
}

// DeleteNode deletes a node. The node must be in a stopped state to be deleted.
func (a *Agent) DeleteNode(nodeID string) error {
	return errors.New("unimplemented")
}

// StartNode starts an existing node.
func (a *Agent) StartNode(nodeID string) error {

	return errors.New("unimplemented")
}

/// StopNode stops an existing node.
func (a *Agent) StopNode(nodeID string) error {

	return errors.New("unimplemented")
}

// GetNodeLogs fetches the last |nbOfLines| log lines from the node logs.
func (a *Agent) GetNodeLogs(nodeID string, nbOfLines uint) (*payload.GetLogsResponse, error) {

	return nil, errors.New("unimplemented")
}

// Shutdown stops all nodes and kills the agent.
func (a *Agent) Shutdown() {}
