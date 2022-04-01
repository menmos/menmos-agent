package agent

type RunType string

const (
	// A Native agent uses native binaries on the host machine.
	Native RunType = "native"

	// A Docker agent talks to the Docker daemon on the host to run menmos nodes in containers.
	Docker = "docker"

	// A Kubernetes agent talks to the kubernetes cluster it finds itself on and runs menmos nodes in pods.
	Kubernetes = "k8s"
)

// Config stores the configuration of a cluster agent.
type Config struct {
	AgentType RunType `json:"agent_type" mapstructure:"TYPE" toml:"agent_type"`

	// Common config
	S3Bucket    string `json:"s3_bucket" mapstructure:"S3_BUCKET" toml:"s3_bucket"`
	S3Region    string `json:"s3_region" mapstructure:"S3_REGION" toml:"s3_region"`
	GithubToken string `json:"github_token" mapstructure:"GH_TOKEN" toml:"github_token"`
	Path        string `json:"path" mapstructure:"PATH" toml:"path"`

	// Native agent settings only.
	LocalBinaryPath string `json:"local_binary_path" mapstructure:"BIN_PATH" toml:"local_binary_path"`
}
