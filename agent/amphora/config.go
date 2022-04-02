package amphora

type DirectoryConfig struct {
	URL  string `toml:"url,omitempty"`
	Port uint16 `toml:"port,omitempty"`
}

type BlobStorageConfig struct {
	Type string `toml:"type,omitempty"`

	// Disk settings - mutually exclusive with s3
	Path string `toml:"path,omitempty"`

	// S3 settings - mutually exclusive with disk
	Bucket    string `toml:"bucket,omitempty"`
	Region    string `toml:"region,omitempty"`
	CachePath string `toml:"cache_path,omitempty"`
	CacheSize uint64 `toml:"cache_size,omitempty"`
}

type NodeConfig struct {
	Name            string            `toml:"name,omitempty"`
	DbPath          string            `toml:"db_path,omitempty"`
	EncryptionKey   string            `toml:"encryption_key,omitempty"`
	BlobStorage     BlobStorageConfig `toml:"blob_storage,omitempty"`
	MaximumCapacity *uint64           `toml:"maximum_capacity,omitempty"`

	// TODO: Advanced settings.
}

type ServerConfig struct {
	CertificateStoragePath string `toml:"certificate_storage_path,omitempty"`
	Port                   uint16 `toml:"port,omitempty"`
}

type RedirectConfig struct {
	Ip         string `toml:"ip,omitempty"`
	SubnetMask string `toml:"subnet_mask,omitempty"`
}

type Config struct {
	Directory DirectoryConfig `toml:"directory,omitempty"`
	Node      NodeConfig      `toml:"node,omitempty"`
	Server    ServerConfig    `toml:"server,omitempty"`
	Redirect  RedirectConfig  `toml:"redirect,omitempty"`
}
