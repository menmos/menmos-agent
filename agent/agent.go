package agent

import (
	"errors"
	"fmt"

	"github.com/menmos/menmos-agent/agent/native"
	"go.uber.org/zap"
)

func New(config Config, log *zap.Logger) (Agent, error) {
	switch config.AgentType {
	case TypeNative:
		nativeCfg := native.NativeAgentConfig{
			S3Bucket:        config.S3Bucket,
			S3Region:        config.S3Region,
			GithubToken:     config.GithubToken,
			Path:            config.Path,
			LocalBinaryPath: config.LocalBinaryPath,
		}
		return native.New(nativeCfg, log)

	case TypeDocker:
		return nil, errors.New("docker agent is not implemented")

	case TypeKubernetes:
		return nil, errors.New("kubernetes agent is not implemented")

	default:
		return nil, fmt.Errorf("unknown agent type: %s", config.AgentType)
	}
}
