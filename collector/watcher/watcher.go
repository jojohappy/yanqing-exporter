package watcher

import (
	"flag"
	"time"

	dclient "github.com/docker/engine-api/client"
	dtypes "github.com/docker/engine-api/types"
	"github.com/golang/glog"

	"github.com/yanqing-exporter/collector/cadvisor"
	"github.com/yanqing-exporter/container/docker"
	"github.com/yanqing-exporter/storage"
)

var interval = time.Second * 5
var argDockerEndpoint = flag.String("docker", "unix:///var/run/docker.sock", "docker endpoint")

type Watcher interface {
	Start() error
	Stop() error
}

func NewWatcher(cacheStorage storage.Storage, cadvisorClient cadvisor.Client) (Watcher, error) {
	dockerClient, err := docker.Client(*argDockerEndpoint)
	if nil != err {
		return nil, err
	}

	return &watcher{
		cacheStorage:   cacheStorage,
		cadvisorClient: cadvisorClient,
		dockerClient:   dockerClient,
		ticker:         time.NewTicker(interval),
		stopWatcher:    make(chan error),
	}, nil
}

type watcher struct {
	cacheStorage   storage.Storage
	cadvisorClient cadvisor.Client
	dockerClient   *dclient.Client
	ticker         *time.Ticker
	stopWatcher    chan error
}

func (w *watcher) Start() error {
	go func() {
		for {
			select {
			case <-w.stopWatcher:
				w.ticker.Stop()
				w.stopWatcher <- nil
				return
			case <-w.ticker.C:
				err := w.getContainerInfo()
				if nil != err {
					glog.Errorf("Failed to get container info: %v", err)
				}
			}
		}
	}()

	return nil
}

func (w *watcher) Stop() error {
	w.stopWatcher <- nil
	return <-w.stopWatcher
}

func (w *watcher) getContainerInfo() error {
	var cinfo *docker.ContainerInfo
	var ctnr dtypes.ContainerJSON
	var creationTime time.Time

	allDockerContainerInfo, err := w.cadvisorClient.GetAllDockerContainers()
	if nil != err {
		return err
	}

	for _, container := range allDockerContainerInfo {
		ctnr, err = w.getContainerInspect(container.Id)
		if nil != err {
			glog.Errorf("Failed to inspect container %q: %v", container.Id, err)
			continue
		}
		creationTime, err = time.Parse(time.RFC3339Nano, ctnr.Created)
		if err != nil {
			glog.Errorf("failed to parse the create timestamp %q for container %q: %v", ctnr.Created, container.Id, err)
			continue
		}
		cinfo = &docker.ContainerInfo{
			ContainerReference: container.ContainerReference,
			Spec: docker.ContainerSpec{
				Image:        ctnr.Config.Image,
				Pid:          ctnr.State.Pid,
				CreationTime: creationTime,
			},
		}
		w.cacheStorage.UpdateContainerInfo(container.Name, cinfo)
	}
	return nil
}

func (w *watcher) getContainerInspect(id string) (dtypes.ContainerJSON, error) {
	ctnr, err := w.dockerClient.ContainerInspect(id)
	if err != nil {
		return dtypes.ContainerJSON{}, err
	}
	return ctnr, nil
}
