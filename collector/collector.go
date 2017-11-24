package collector

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	info "github.com/google/cadvisor/info/v1"

	"github.com/yanqing-exporter/collector/cadvisor"
	"github.com/yanqing-exporter/collector/watcher"
	"github.com/yanqing-exporter/container/docker"
	"github.com/yanqing-exporter/storage"
)

var (
	interval             = flag.Duration("collect_interval", time.Second*10, "Interval between collectings")
	rootFs               = flag.String("collector_procfs", "/host/", "Path of host proc")
	housekeepingInterval = flag.Duration("housekeeping_interval", 1*time.Minute, "Interval between housekeepings")
)

type Collector interface {
	Start() error
	Stop() error
}

func NewCollector(cacheStorage storage.Storage, cadvisorClient cadvisor.Client) (Collector, error) {
	dockerWatcher, err := watcher.NewWatcher(cacheStorage, cadvisorClient)
	if nil != err {
		return nil, err
	}
	return &collector{
		watcher:      dockerWatcher,
		cacheStorage: cacheStorage,
	}, nil
}

type collector struct {
	watcher      watcher.Watcher
	cacheStorage storage.Storage
	quitChannels []chan error
}

func (c *collector) Start() error {
	c.watcher.Start()

	quitCollector := make(chan error)
	c.quitChannels = append(c.quitChannels, quitCollector)
	c.startCollector(quitCollector)

	quitHousekeeping := make(chan error)
	c.quitChannels = append(c.quitChannels, quitHousekeeping)
	go c.housekeeping(quitHousekeeping)
	return nil
}

func (c *collector) Stop() error {
	c.watcher.Stop()

	for i, ch := range c.quitChannels {
		ch <- nil
		err := <-ch
		if err != nil {
			c.quitChannels = c.quitChannels[i:]
			return err
		}
	}
	c.quitChannels = make([]chan error, 0, 2)
	return nil
}

func (c *collector) startCollector(quit chan error) {
	var err error
	var wg sync.WaitGroup
	var tcpStat info.TcpStat
	var udpStat info.UdpStat
	var tcp6Stat info.TcpStat
	var udp6Stat info.UdpStat

	containerInfos := make(map[string]*docker.ContainerInfo)
	ticker := time.NewTicker(*interval)
	go func() {
		for {
			select {
			case <-quit:
				quit <- nil
				return
			case <-ticker.C:
				containerInfos = c.cacheStorage.GetAllContainerInfo()
				for name, container := range containerInfos {
					wg.Add(1)
					go func(name string, container *docker.ContainerInfo) {
						defer wg.Done()
						var tcpError = false
						var udpError = false
						tcpStat, err = tcpStatsFromProc(*rootFs, container.Spec.Pid, "net/tcp")
						if err != nil {
							glog.V(2).Infof("Unable to get tcp stats from pid %d: %v", container.Spec.Pid, err)
							tcpError = true
						}

						udpStat, err = udpStatsFromProc(*rootFs, container.Spec.Pid, "net/udp")
						if err != nil {
							glog.V(2).Infof("Unable to get udp stats from pid %d: %v", container.Spec.Pid, err)
							udpError = true
						}

						tcp6Stat, err = tcpStatsFromProc(*rootFs, container.Spec.Pid, "net/tcp6")
						if err != nil {
							glog.V(2).Infof("Unable to get tcp6 stats from pid %d: %v", container.Spec.Pid, err)
							tcpError = true
						}

						udp6Stat, err = udpStatsFromProc(*rootFs, container.Spec.Pid, "net/udp6")
						if err != nil {
							glog.V(2).Infof("Unable to get udp6 stats from pid %d: %v", container.Spec.Pid, err)
							udpError = true
						}

						if !udpError && !tcpError {
							containerStats := &docker.ContainerStats{
								Timestamp: time.Now(),
								Tcp:       tcpStat,
								Udp:       udpStat,
								Tcp6:      tcp6Stat,
								Udp6:      udp6Stat,
							}
							c.cacheStorage.AddStats(container.Name, containerStats)
						}
					}(name, container)
				}
				wg.Wait()
			}
		}
	}()
}

func (c *collector) housekeeping(quit chan error) {
	longHousekeeping := 100 * time.Millisecond

	containerInfos := make(map[string]*docker.ContainerInfo)
	ticker := time.Tick(*housekeepingInterval)
	for {
		select {
		case t := <-ticker:
			start := time.Now()
			containerInfos = c.cacheStorage.GetAllContainerInfo()
			for name, container := range containerInfos {
				// Delete container when recent stats is out of housekeepingInterval
				stats := container.Stats
				statsLength := len(stats)
				if statsLength == 0 {
					continue
				}
				if start.Sub(stats[statsLength-1].Timestamp) > *housekeepingInterval {
					c.cacheStorage.RemoveContainerInfo(name)
					continue
				}
			}
			duration := time.Since(start)
			if duration >= longHousekeeping {
				glog.V(3).Infof("Housekeeping(%d) took %s", t.Unix(), duration)
			}
		case <-quit:
			quit <- nil
			glog.Infof("Exiting housekeeping thread")
			return
		}
	}
}

func tcpStatsFromProc(rootFs string, pid int, file string) (info.TcpStat, error) {
	tcpStatsFile := path.Join(rootFs, "proc", strconv.Itoa(pid), file)

	tcpStats, err := scanTcpStats(tcpStatsFile)
	if err != nil {
		return tcpStats, fmt.Errorf("couldn't read tcp stats: %v", err)
	}

	return tcpStats, nil
}

func scanTcpStats(tcpStatsFile string) (info.TcpStat, error) {
	var stats info.TcpStat

	data, err := ioutil.ReadFile(tcpStatsFile)
	if err != nil {
		return stats, fmt.Errorf("failure opening %s: %v", tcpStatsFile, err)
	}

	tcpStateMap := map[string]uint64{
		"01": 0, //ESTABLISHED
		"02": 0, //SYN_SENT
		"03": 0, //SYN_RECV
		"04": 0, //FIN_WAIT1
		"05": 0, //FIN_WAIT2
		"06": 0, //TIME_WAIT
		"07": 0, //CLOSE
		"08": 0, //CLOSE_WAIT
		"09": 0, //LAST_ACK
		"0A": 0, //LISTEN
		"0B": 0, //CLOSING
	}

	reader := strings.NewReader(string(data))
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanLines)

	if b := scanner.Scan(); !b {
		return stats, scanner.Err()
	}

	for scanner.Scan() {
		line := scanner.Text()

		state := strings.Fields(line)
		tcpState := state[3]
		_, ok := tcpStateMap[tcpState]
		if !ok {
			return stats, fmt.Errorf("invalid TCP stats line: %v", line)
		}
		tcpStateMap[tcpState]++
	}

	stats = info.TcpStat{
		Established: tcpStateMap["01"],
		SynSent:     tcpStateMap["02"],
		SynRecv:     tcpStateMap["03"],
		FinWait1:    tcpStateMap["04"],
		FinWait2:    tcpStateMap["05"],
		TimeWait:    tcpStateMap["06"],
		Close:       tcpStateMap["07"],
		CloseWait:   tcpStateMap["08"],
		LastAck:     tcpStateMap["09"],
		Listen:      tcpStateMap["0A"],
		Closing:     tcpStateMap["0B"],
	}

	return stats, nil
}

func udpStatsFromProc(rootFs string, pid int, file string) (info.UdpStat, error) {
	var err error
	var udpStats info.UdpStat

	udpStatsFile := path.Join(rootFs, "proc", strconv.Itoa(pid), file)

	r, err := os.Open(udpStatsFile)
	if err != nil {
		return udpStats, fmt.Errorf("failure opening %s: %v", udpStatsFile, err)
	}

	udpStats, err = scanUdpStats(r)
	if err != nil {
		return udpStats, fmt.Errorf("couldn't read udp stats: %v", err)
	}

	return udpStats, nil
}

func scanUdpStats(r io.Reader) (info.UdpStat, error) {
	var stats info.UdpStat

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	if b := scanner.Scan(); !b {
		return stats, scanner.Err()
	}

	listening := uint64(0)
	dropped := uint64(0)
	rxQueued := uint64(0)
	txQueued := uint64(0)

	for scanner.Scan() {
		line := scanner.Text()

		listening++

		fs := strings.Fields(line)
		if len(fs) != 13 {
			continue
		}

		rx, tx := uint64(0), uint64(0)
		fmt.Sscanf(fs[4], "%X:%X", &rx, &tx)
		rxQueued += rx
		txQueued += tx

		d, err := strconv.Atoi(string(fs[12]))
		if err != nil {
			continue
		}
		dropped += uint64(d)
	}

	stats = info.UdpStat{
		Listen:   listening,
		Dropped:  dropped,
		RxQueued: rxQueued,
		TxQueued: txQueued,
	}

	return stats, nil
}
