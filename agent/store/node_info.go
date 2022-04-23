package store

type NodeInfo struct {
	ID      string `json:"id,omitempty"`
	Binary  string `json:"binary,omitempty"`
	Version string `json:"version,omitempty"`
}
