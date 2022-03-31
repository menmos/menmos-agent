package agent

type RunType uint8

const (
	Native RunType = iota
	Docker
)

// Config stores the configuration of a cluster agent.
type Config struct {
	AgentType RunType `json:"agent_type,omitempty"`
	Path      string  `json:"data_path,omitempty"`
}
