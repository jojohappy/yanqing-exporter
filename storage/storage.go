package storage

import (
	"fmt"
	"sort"
	"sync"

	"github.com/yanqing-exporter/container/docker"
)

type Storage interface {
	GetContainerInfo(name string) (*docker.ContainerInfo, error)
	GetAllContainerInfo() map[string]*docker.ContainerInfo
	UpdateContainerInfo(name string, cinfo *docker.ContainerInfo) error
	AddStats(name string, stats *docker.ContainerStats) error
	RemoveContainerInfo(name string) error
}

type MemoryStorage struct {
	maxStatsLength   int
	lock             sync.RWMutex
	containerInfoMap map[string]*docker.ContainerInfo
}

func (m *MemoryStorage) GetContainerInfo(name string) (*docker.ContainerInfo, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	cinfo, ok := m.containerInfoMap[name]
	if !ok {
		return nil, fmt.Errorf("unable to find data for container %v", name)
	}
	return cinfo, nil
}

func (m *MemoryStorage) GetAllContainerInfo() map[string]*docker.ContainerInfo {
	m.lock.RLock()
	defer m.lock.RUnlock()
	containers := make(map[string]*docker.ContainerInfo, len(m.containerInfoMap))

	for name, cont := range m.containerInfoMap {
		containers[name] = cont
	}
	return containers
}

func (m *MemoryStorage) UpdateContainerInfo(name string, cinfo *docker.ContainerInfo) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	c := &docker.ContainerInfo{
		ContainerReference: cinfo.ContainerReference,
		Spec:               cinfo.Spec,
	}
	if ret, ok := m.containerInfoMap[name]; ok {
		c.Stats = make([]*docker.ContainerStats, len(ret.Stats))
		copy(c.Stats, ret.Stats)
	}
	m.containerInfoMap[name] = c
	return nil
}

func (m *MemoryStorage) AddStats(name string, stats *docker.ContainerStats) error {
	var cinfo *docker.ContainerInfo
	var ok bool

	err := func() error {
		m.lock.Lock()
		defer m.lock.Unlock()
		if cinfo, ok = m.containerInfoMap[name]; !ok {
			return fmt.Errorf("unable to find data for container %v", name)
		}

		if len(cinfo.Stats) == 0 || !stats.Timestamp.Before(cinfo.Stats[len(cinfo.Stats)-1].Timestamp) {
			cinfo.Stats = append(cinfo.Stats, stats)
		} else {
			index := sort.Search(len(cinfo.Stats), func(index int) bool {
				return cinfo.Stats[index].Timestamp.After(stats.Timestamp)
			})
			cinfo.Stats = append(cinfo.Stats, &docker.ContainerStats{})
			copy(cinfo.Stats[index+1:], cinfo.Stats[index:])
			cinfo.Stats[index] = stats
		}

		if m.maxStatsLength >= 0 && len(cinfo.Stats) > m.maxStatsLength {
			startIndex := len(cinfo.Stats) - m.maxStatsLength
			cinfo.Stats = cinfo.Stats[startIndex:]
		}
		return nil
	}()

	if nil != err {
		return err
	}

	return nil
}

func (m *MemoryStorage) RemoveContainerInfo(name string) error {
	m.lock.Lock()
	delete(m.containerInfoMap, name)
	m.lock.Unlock()
	return nil
}

func New(maxStatsLength int) Storage {
	return &MemoryStorage{
		maxStatsLength:   maxStatsLength,
		containerInfoMap: make(map[string]*docker.ContainerInfo, 0),
	}
}
