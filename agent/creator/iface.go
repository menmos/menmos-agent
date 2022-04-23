package creator

import (
	"github.com/menmos/menmos-agent/agent/xecute"
	"github.com/menmos/menmos-agent/payload"
)

// A Node manages the execution of a single menmos process.
type Node interface {
	Start(level xecute.LogLevel) error
	Stop() error
	Delete() error

	Logs(numberOfLines uint) []interface{}
	Port() uint16
	Status() xecute.Status
}

// A NodeCreator creates and restores nodes from a given configuration.
type NodeCreator interface {
	CreateMenmosd(id, version string, config *payload.MenmosdConfig) (Node, error)
	CreateAmphora(id, version string, config *payload.AmphoraConfig) (Node, error)

	RestoreMenmosd(id, version string) (Node, error)
	RestoreAmphora(id, version string) (Node, error)
}

/*
type Agent interface {
	// GetNode gets all the information of a node.
	GetNode(nodeID string) (*payload.NodeResponse, error)

	// ListNode returns the information of all nodes.
	ListNodes() (*payload.ListNodesResponse, error)

	// CreateNode creates a new node from the given information.
	CreateNode(request *payload.CreateNodeRequest) (*payload.NodeResponse, error)

	// DeleteNode deletes a node. The node must be in a stopped state to be deleted.
	DeleteNode(nodeID string) error

	// StartNode starts an existing node.
	StartNode(nodeID string) error

	/// StopNode stops an existing node.
	StopNode(nodeID string) error

	// GetNodeLogs fetches the last |nbOfLines| log lines from the node logs.
	GetNodeLogs(nodeID string, nbOfLines uint) (*payload.GetLogsResponse, error)

	// Shutdown stops all nodes and kills the agent.
	Shutdown()
}
*/
