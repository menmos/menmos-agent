package store

// A StateStore is used at the agent level to store which nodes are known to the agent as well as their information.
type StateStore interface {
	Get(id string) (*NodeInfo, error)
	Save(info NodeInfo) error
	ListNodes() ([]NodeInfo, error)
	DeleteNode(id string) error
}
