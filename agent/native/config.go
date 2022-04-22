package native

type NativeAgentConfig struct {
	S3Bucket        string `json:"s3_bucket" mapstructure:"S3_BUCKET" toml:"s3_bucket"`
	S3Region        string `json:"s3_region" mapstructure:"S3_REGION" toml:"s3_region"`
	GithubToken     string `json:"github_token" mapstructure:"GH_TOKEN" toml:"github_token"`
	Path            string `json:"path" mapstructure:"PATH" toml:"path"`
	LocalBinaryPath string `json:"local_binary_path" mapstructure:"BIN_PATH" toml:"local_binary_path"`
}
