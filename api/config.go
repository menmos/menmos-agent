package api

// Config is the configuration entrypoint of the agent API.
type Config struct {
	Host string `json:"host" mapstructure:"HOST" toml:"host"`
	Port uint16 `json:"port" mapstructure:"PORT" toml:"port"`

	// TODO: Auth.
}
