package docker

import (
	"time"

	info "github.com/google/cadvisor/info/v1"

	"github.com/yanqing-exporter/collector/types"
)

type ContainerInfo struct {
	info.ContainerReference
	Spec  ContainerSpec     `json:"spec,omitempty"`
	Stats []*ContainerStats `json:"stats,omitempty"`
}

type ContainerSpec struct {
	Image        string    `json:"image,omitempty"`
	Pid          int       `json:"pid,omitempty"`
	CreationTime time.Time `json:"creation_time,omitempty"`
}

type ContainerStats struct {
	Timestamp    time.Time             `json:"timestamp"`
	Tcp          info.TcpStat          `json:"tcp"`
	Udp          info.UdpStat          `json:"udp"`
	Tcp6         info.TcpStat          `json:"tcp6"`
	Udp6         info.UdpStat          `json:"udp6"`
	TcpExt       types.TcpExtStat      `json:"tcpext"`
	TcpWithPort  types.TcpStatWithPort `json:"tcpwithport"`
	Tcp6WithPort types.TcpStatWithPort `json:"tcp6withport"`
}
