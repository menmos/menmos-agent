package docker

import "github.com/docker/docker/client"

type imageRegistry struct {
	client *client.Client
}

func newImageRegistry(client *client.Client) *imageRegistry {
	return &imageRegistry{client: client}
}
