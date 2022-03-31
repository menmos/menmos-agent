package menmosd

import "fmt"

type NodeSetting struct {
	DbPath           string `toml:"db_path,omitempty" json:"db_path,omitempty"`
	AdminPassword    string `toml:"admin_password,omitempty" json:"admin_password,omitempty"`
	EncryptionKey    string `toml:"encryption_key,omitempty" json:"encryption_key,omitempty"`
	RoutingAlgorithm string `toml:"routing_algorithm,omitempty" json:"routing_algorithm,omitempty"`
}

type ServerSetting struct {
	// TODO: Strongly type server type.
	Type string `toml:"type,omitempty" json:"type,omitempty"`
	Port uint16 `toml:"port,omitempty" json:"port,omitempty"`

	// TODO: Support HTTP parameters.
}

// Represents a menmosd config.
type Config struct {
	Server ServerSetting `toml:"server,omitempty" json:"server,omitempty"`
	Node   NodeSetting   `toml:"node,omitempty" json:"node,omitempty"`
}

func (c *Config) HealthCheckURL() string {
	return fmt.Sprintf("%s://localhost:%d/health", c.Server.Type, c.Server.Port)
}
