package native

import (
	"path"

	"github.com/menmos/menmos-agent/agent/amphora"
	"github.com/menmos/menmos-agent/agent/creator"
	"github.com/menmos/menmos-agent/agent/menmosd"
	"github.com/menmos/menmos-agent/agent/native/artifact"
	"github.com/menmos/menmos-agent/agent/native/process"
	"github.com/menmos/menmos-agent/fs"
	"github.com/menmos/menmos-agent/payload"
	"go.uber.org/zap"
)

type NativeAgent struct {
	// Misc. management stuff.
	config NativeAgentConfig
	log    *zap.SugaredLogger

	artifacts *artifact.Repository
}

func New(config NativeAgentConfig, log *zap.Logger) (*NativeAgent, error) {
	// Using a github release fetcher by default.
	agent := &NativeAgent{
		config: config,
		log:    log.Sugar().Named("native"),
		artifacts: artifact.NewRepository(
			artifact.RepositoryParams{
				ReleaseFetcher: artifact.NewGithubFetcher(config.GithubToken),
				Log:            log,
				Path:           path.Join(config.Path, "pkg"),
			},
		),
	}

	if err := agent.initWorkspace(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *NativeAgent) pkgDir() string {
	return path.Join(a.config.Path, "pkg")
}

func (a *NativeAgent) nodeDir() string {
	return path.Join(a.config.Path, "node")
}

func (a *NativeAgent) initWorkspace() error {
	a.log.Debug("initializing menmos agent workspace")

	if err := fs.EnsureDirExists(a.config.Path); err != nil {
		return err
	}

	if err := fs.EnsureDirExists(a.pkgDir()); err != nil {
		return err
	}

	if err := fs.EnsureDirExists(a.nodeDir()); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) getBinary(version, binary string) (binaryPath string, err error) {
	if version != "" {
		return a.artifacts.Get(version, binary)
	} else if a.config.LocalBinaryPath != "" {
		// Fallback on local binaries.
		binaryPath = path.Join(a.config.LocalBinaryPath, binary)
	}

	return
}

func (a *NativeAgent) createMenmosdConfig(nodeDir string, config *payload.MenmosdConfig) error {
	procConfig := menmosd.Config{
		Node: menmosd.NodeSetting{
			DbPath:           path.Join(nodeDir, "db"),
			AdminPassword:    config.NodeAdminPassword,
			EncryptionKey:    config.NodeEncryptionKey,
			RoutingAlgorithm: config.NodeRoutingAlgorithm,
		},
		Server: menmosd.ServerSetting{
			Type: config.ServerMode,
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) createAmphoraConfig(nodeDir string, config *payload.AmphoraConfig) error {

	storageConfig := amphora.BlobStorageConfig{Type: string(config.BlobStorageType)}
	if config.BlobStorageType == payload.BlobStorageDisk {
		storageConfig.Type = "Directory"
		storageConfig.Path = path.Join(nodeDir, "blob")
	} else {
		storageConfig.Type = "S3"
		storageConfig.Bucket = a.config.S3Bucket
		storageConfig.Region = a.config.S3Region
		storageConfig.CachePath = path.Join(nodeDir, "cache")
		storageConfig.CacheSize = 1024 * 1024 * 1024 //  1Gb TODO: Customize
	}

	procConfig := amphora.Config{
		Directory: amphora.DirectoryConfig{
			URL:  config.DirectoryHost,
			Port: config.DirectoryPort,
		},
		Node: amphora.NodeConfig{
			Name:            config.Name,
			DbPath:          path.Join(nodeDir, "db"),
			EncryptionKey:   config.NodeEncryptionKey,
			MaximumCapacity: config.MaximumCapacity,
			BlobStorage:     storageConfig,
		},
		Server: amphora.ServerConfig{
			CertificateStoragePath: path.Join(nodeDir, "cert"),
		},
		Redirect: amphora.RedirectConfig{
			Ip:         config.RedirectIp,
			SubnetMask: config.SubnetMask,
		},
	}

	if err := tomlWrite(procConfig, path.Join(nodeDir, "config.toml")); err != nil {
		return err
	}

	return nil
}

func (a *NativeAgent) CreateMenmosd(id, version string, config *payload.MenmosdConfig) (creator.Node, error) {
	binPath, err := a.getBinary(version, "menmosd")
	if err != nil {
		return nil, err
	}

	nodeDir := path.Join(a.nodeDir(), id)
	if err := fs.EnsureDirExists(nodeDir); err != nil {
		return nil, err
	}

	if err := a.createMenmosdConfig(nodeDir, config); err != nil {
		return nil, err
	}

	return process.NewNativeProcess(nodeDir, binPath, a.log.Named("menmosd").Named(id).Desugar())
}

func (a *NativeAgent) CreateAmphora(id, version string, config *payload.AmphoraConfig) (creator.Node, error) {
	binPath, err := a.getBinary(version, "amphora")
	if err != nil {
		return nil, err
	}

	nodeDir := path.Join(a.nodeDir(), id)
	if err := fs.EnsureDirExists(nodeDir); err != nil {
		return nil, err
	}

	if err := a.createAmphoraConfig(nodeDir, config); err != nil {
		return nil, err
	}

	return process.NewNativeProcess(nodeDir, binPath, a.log.Named("amphora").Named(id).Desugar())
}

func (a *NativeAgent) RestoreMenmosd(id, version string) (creator.Node, error) {
	nodeDir := path.Join(a.nodeDir(), id)

	binPath, err := a.getBinary(version, "menmosd")
	if err != nil {
		return nil, err
	}

	return process.NewNativeProcess(nodeDir, binPath, a.log.Named("menmosd").Named(id).Desugar())
}

func (a *NativeAgent) RestoreAmphora(id, version string) (creator.Node, error) {
	nodeDir := path.Join(a.nodeDir(), id)

	binPath, err := a.getBinary(version, "amphora")
	if err != nil {
		return nil, err
	}

	return process.NewNativeProcess(nodeDir, binPath, a.log.Named("amphora").Named(id).Desugar())
}
