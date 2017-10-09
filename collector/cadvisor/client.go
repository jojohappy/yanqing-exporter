package cadvisor

import (
	"fmt"
	"net"

	"github.com/google/cadvisor/client"
	info "github.com/google/cadvisor/info/v1"
)

type Client interface {
	GetAllDockerContainers() ([]info.ContainerInfo, error)
}

func New(host net.IP, port int) (Client, error) {
	cadvisorHost := fmt.Sprintf("http://%s:%d/", host, port)
	cadvisorRestClient, err := client.NewClient(cadvisorHost)
	if nil != err {
		return nil, err
	}
	return &cadvisorClient{
		client: cadvisorRestClient,
	}, nil
}

type cadvisorClient struct {
	client *client.Client
}

func (c *cadvisorClient) GetAllDockerContainers() ([]info.ContainerInfo, error) {
	request := info.ContainerInfoRequest{NumStats: 0}
	return c.client.AllDockerContainers(&request)
}
