package payload

// The type of a node (menmosd or amphora).
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

// Menmosd is the confifg sent to an agent to create a menmosd instance.
type MenmosdConfig struct {
	// Node configuration.
	NodeAdminPassword    string `mapstructure:"node_admin_password,omitempty"`
	NodeEncryptionKey    string `mapstructure:"node_encryption_key,omitempty"`
	NodeRoutingAlgorithm string `mapstructure:"node_routing_algorithm,omitempty"`

	// Server configuration
	ServerMode string `mapstructure:"server_mode,omitempty"` // TODO: Validate `http` or `https`
	Port       uint16 `mapstructure:"port,omitempty"`
}

type BlobStorageType string

const (
	BlobStorageDisk = "disk"
	BlobStorageS3   = "s3"
)

// Amphora is the confifg sent to an agent to create a amphora instance.
type AmphoraConfig struct {
	Name              string          `mapstructure:"name,omitempty"`
	DirectoryHost     string          `mapstructure:"directory_host,omitempty"`
	DirectoryPort     uint16          `mapstructure:"directory_port,omitempty"`
	NodeEncryptionKey string          `mapstructure:"node_encryption_key,omitempty"`
	MaximumCapacity   *uint64         `mapstructure:"maximum_capacity,omitempty"`
	BlobStorageType   BlobStorageType `mapstructure:"blob_storage_type"`

	ServerPort uint16 `mapstructure:"server_port,omitempty"`
	RedirectIp string `mapstructure:"redirect_ip,omitempty"`
	SubnetMask string `mapstructure:"subnet_mask,omitempty"`
}

type NodeResponse struct {
	ID             string `json:"id,omitempty"`
	Binary         string `json:"binary,omitempty"`
	Version        string `json:"version,omitempty"`
	HealthCheckURL string `json:"health_check_url,omitempty"`
	Status         string `json:"status,omitempty"`
}

type ListNodesResponse struct {
	Nodes []*NodeResponse `json:"nodes,omitempty"`
}
