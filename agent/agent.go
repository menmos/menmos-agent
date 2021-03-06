package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/menmos/menmos-agent/agent/artifact"
	"github.com/menmos/menmos-agent/agent/xecute"
	"github.com/pelletier/go-toml/v2"
	"go.uber.org/zap"
)

const AGENT_NODE_INFO_FILE = ".agent_node_info.json"

func ensureDirExists(path string) error {
	dirInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if err != nil {
		return err
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("agent path '%v' is not a directory", path)
	}

	return nil
}

func tomlWrite(config interface{}, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	configBytes, err := toml.Marshal(config)

	if err != nil {
		return err
	}

	_, err = file.Write(configBytes)
	return err
}

func jsonWrite(config interface{}, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	configBytes, err := json.Marshal(config)

	if err != nil {
		return err
	}

	_, err = file.Write(configBytes)
	return err
}

// A MenmosAgent manages menmos processes on a given machine.
type MenmosAgent struct {
	// Misc. management stuff.
	config Config
	log    *zap.SugaredLogger

	artifacts *artifact.Repository

	// State
	runningNodes map[string]*xecute.Native
}

// New returns a new menmos agent.
func New(config Config, log *zap.Logger) (*MenmosAgent, error) {
	// Using a github release fetcher by default.

	agent := &MenmosAgent{
		config: config,
		log:    log.Sugar().Named("agent"),
		artifacts: artifact.NewRepository(
			artifact.RepositoryParams{
				ReleaseFetcher: artifact.NewGithubFetcher(config.GithubToken),
				Log:            log,
				Path:           path.Join(config.Path, "pkg"),
			},
		),
		runningNodes: make(map[string]*xecute.Native),
	}

	if err := agent.initWorkspace(); err != nil {
		return nil, err
	}

	if err := agent.restartComponents(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *MenmosAgent) pkgDir() string {
	return path.Join(a.config.Path, "pkg")
}

func (a *MenmosAgent) nodeDir() string {
	return path.Join(a.config.Path, "node")
}

func (a *MenmosAgent) initWorkspace() error {
	a.log.Debug("initializing menmos agent workspace")

	if err := ensureDirExists(a.config.Path); err != nil {
		return err
	}

	if err := ensureDirExists(a.pkgDir()); err != nil {
		return err
	}

	if err := ensureDirExists(a.nodeDir()); err != nil {
		return err
	}

	return nil
}

func (a *MenmosAgent) restartComponents() error {
	entries, err := os.ReadDir(a.nodeDir())
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if err := a.StartNode(entry.Name()); err != nil {
			a.log.Errorf("failed to restart component '%s': %v", entry.Name(), err)
		}
	}

	return nil
}

func (a *MenmosAgent) getBinary(version, binary string) (binaryPath string, err error) {
	if version != "" {
		return a.artifacts.Get(version, binary)
	} else if a.config.LocalBinaryPath != "" {
		// Fallback on local binaries.
		binaryPath = path.Join(a.config.LocalBinaryPath, binary)
	}

	return
}

func (a *MenmosAgent) Shutdown() {
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
