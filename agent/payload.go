package agent

type NodeType string

const (
	NodeMenmosd = "menmosd"
	NodeAmphora = "amphora"
)

type CreateNodeRequest struct {
	Version string `json:"version"`
	Type    NodeType

	// Config can be either MenmosdConfig if Type == "menmosd", or AmphoraConfig if type == "amphora"
	Config map[string]interface{}
}

// CreateMenmosdRequest is the payload sent to an agent to create a menmosd instance.
type MenmosdConfig struct {
	// Node configuration.
	NodeAdminPassword    string `mapstructure:"node_admin_password,omitempty"`
	NodeEncryptionKey    string `mapstructure:"node_encryption_key,omitempty"`
	NodeRoutingAlgorithm string `mapstructure:"node_routing_algorithm,omitempty"`

	// Server configuration
	ServerMode string `mapstructure:"server_mode,omitempty"` // TODO: Validate `http` or `https`
	Port       uint16 `mapstructure:"port,omitempty"`
}

type CreateNodeResponse struct {
	ID string `json:"id,omitempty"`
}
