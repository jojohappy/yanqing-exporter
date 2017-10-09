package docker

import (
	"sync"

	dclient "github.com/docker/engine-api/client"
)

var (
	dockerClient     *dclient.Client
	dockerClientErr  error
	dockerClientOnce sync.Once
)

func Client(dockerEndpoint string) (*dclient.Client, error) {
	dockerClientOnce.Do(func() {
		dockerClient, dockerClientErr = dclient.NewClient(dockerEndpoint, "", nil, nil)
	})
	return dockerClient, dockerClientErr
}
