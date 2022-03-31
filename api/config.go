package api

// Config is the configuration entrypoint of the agent API.
type Config struct {
	Host string
	Port uint16

	// TODO: Auth.
}
