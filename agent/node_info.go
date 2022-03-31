package agent

type nodeInfo struct {
	Binary         string `json:"binary,omitempty"`
	Version        string `json:"version,omitempty"`
	HealthCheckURL string `json:"health_check_url,omitempty"`
}
