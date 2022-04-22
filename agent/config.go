package agent

type RunType string

const (
	// A Native agent uses native binaries on the host machine.
	TypeNative RunType = "native"

	// A Docker agent talks to the Docker daemon on the host to run menmos nodes in containers.
	TypeDocker RunType = "docker"

	// A Kubernetes agent talks to the kubernetes cluster it finds itself on and runs menmos nodes in pods.
	TypeKubernetes RunType = "k8s"
)

// Config stores the configuration of a cluster agent.
type Config struct {
	AgentType RunType `json:"type" mapstructure:"TYPE" toml:"type"`

	// Common config
	S3Bucket string `json:"s3_bucket" mapstructure:"S3_BUCKET" toml:"s3_bucket"`
	S3Region string `json:"s3_region" mapstructure:"S3_REGION" toml:"s3_region"`
	Path     string `json:"path" mapstructure:"PATH" toml:"path"`

	// Native agent settings only.
	GithubToken     string `json:"github_token" mapstructure:"GITHUB_TOKEN" toml:"github_token"`
	LocalBinaryPath string `json:"bin_path" mapstructure:"BIN_PATH" toml:"bin_path"`
}
