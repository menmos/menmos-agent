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
	S3Bucket  string  `json:"s3_bucket,omitempty"`
	S3Region  string  `json:"s3_region,omitempty"`
}
