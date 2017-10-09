package metrics

import (
	"regexp"
	"time"

	"github.com/google/cadvisor/metrics"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/yanqing-exporter/container/docker"
	"github.com/yanqing-exporter/storage"
)

const (
	ContainerKubernetesPrefix      = "io.kubernetes."
	ContainerAnnotationLabelPrefix = "annotation." + ContainerKubernetesPrefix
	ContainerLabelContainerName    = ContainerKubernetesPrefix + "container.name"
	ContainerLabelPodNamespace     = ContainerKubernetesPrefix + "pod.namespace"
	ContainerLabelPodName          = ContainerKubernetesPrefix + "pod.name"
	ContainerLabelApp              = "app"
)

var (
	yanqingScropedLastSeenDesc = prometheus.NewDesc("yanqing_scroped_last_seen", "yanqing_scroped_last_seen Last timstamp when scroped.", nil, nil)
	contaierLabelIgnore        = map[string]bool{
		ContainerKubernetesPrefix + "container.logpath": true,
		ContainerKubernetesPrefix + "sandbox.id":        true,
	}
)

type metricValue struct {
	value  float64
	labels []string
}

type metricValues []metricValue

type ContainerLabelsFunc func(*docker.ContainerInfo) map[string]string

type containerMetric struct {
	name        string
	help        string
	valueType   prometheus.ValueType
	extraLabels []string
	getValues   func(s *docker.ContainerStats) metricValues
}

func (cm *containerMetric) desc(baseLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(cm.name, cm.help, append(baseLabels, cm.extraLabels...), nil)
}

type yanqingCollector struct {
	containerMetrics    []containerMetric
	containerLabelsFunc ContainerLabelsFunc
	cacheStorage        storage.Storage
}

func NewCollector(memoryStorage storage.Storage) *yanqingCollector {
	return &yanqingCollector{
		containerLabelsFunc: DefaultLabels,
		containerMetrics: []containerMetric{
			{
				name:      "yanqing_last_seen",
				help:      "Last time was seen by the yanqing-exporter",
				valueType: prometheus.GaugeValue,
				getValues: func(s *docker.ContainerStats) metricValues {
					return metricValues{{value: float64(time.Now().Unix())}}
				},
			},
			{
				name:      "yq_container_timestamp",
				help:      "statistic timestamp",
				valueType: prometheus.GaugeValue,
				getValues: func(s *docker.ContainerStats) metricValues {
					return metricValues{{value: float64(s.Timestamp.Unix())}}
				},
			},
			{
				name:        "yq_container_network_tcp_usage_total",
				help:        "tcp connection usage statistic for container by yanqing-exporter",
				valueType:   prometheus.GaugeValue,
				extraLabels: []string{"tcp_state"},
				getValues: func(s *docker.ContainerStats) metricValues {
					return metricValues{
						{
							value:  float64(s.Tcp.Established),
							labels: []string{"established"},
						},
						{
							value:  float64(s.Tcp.SynSent),
							labels: []string{"synsent"},
						},
						{
							value:  float64(s.Tcp.SynRecv),
							labels: []string{"synrecv"},
						},
						{
							value:  float64(s.Tcp.FinWait1),
							labels: []string{"finwait1"},
						},
						{
							value:  float64(s.Tcp.FinWait2),
							labels: []string{"finwait2"},
						},
						{
							value:  float64(s.Tcp.TimeWait),
							labels: []string{"timewait"},
						},
						{
							value:  float64(s.Tcp.Close),
							labels: []string{"close"},
						},
						{
							value:  float64(s.Tcp.CloseWait),
							labels: []string{"closewait"},
						},
						{
							value:  float64(s.Tcp.LastAck),
							labels: []string{"lastack"},
						},
						{
							value:  float64(s.Tcp.Listen),
							labels: []string{"listen"},
						},
						{
							value:  float64(s.Tcp.Closing),
							labels: []string{"closing"},
						},
					}
				},
			}, {
				name:        "yq_container_network_udp_usage_total",
				help:        "udp connection usage statistic for container by yanqing-exporter",
				valueType:   prometheus.GaugeValue,
				extraLabels: []string{"udp_state"},
				getValues: func(s *docker.ContainerStats) metricValues {
					return metricValues{
						{
							value:  float64(s.Udp.Listen),
							labels: []string{"listen"},
						},
						{
							value:  float64(s.Udp.Dropped),
							labels: []string{"dropped"},
						},
						{
							value:  float64(s.Udp.RxQueued),
							labels: []string{"rxqueued"},
						},
						{
							value:  float64(s.Udp.TxQueued),
							labels: []string{"txqueued"},
						},
					}
				},
			},
		},
		cacheStorage: memoryStorage,
	}
}

func (y *yanqingCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, cm := range y.containerMetrics {
		ch <- cm.desc([]string{})
	}
	ch <- yanqingScropedLastSeenDesc
}

func (y *yanqingCollector) Collect(ch chan<- prometheus.Metric) {
	y.collectLastSeen(ch)
	y.collectContainerStats(ch)
}

func DefaultLabels(container *docker.ContainerInfo) map[string]string {
	var name, image, podName, namespace, containerName, appName string
	if len(container.Aliases) > 0 {
		name = container.Aliases[0]
	}
	image = container.Spec.Image
	if v, ok := container.Labels[ContainerLabelPodName]; ok {
		podName = v
	}
	if v, ok := container.Labels[ContainerLabelPodNamespace]; ok {
		namespace = v
	}
	if v, ok := container.Labels[ContainerLabelContainerName]; ok {
		containerName = v
	}
	if v, ok := container.Labels[ContainerLabelApp]; ok {
		appName = v
	}

	set := map[string]string{
		metrics.LabelID:                      container.Name,
		metrics.LabelName:                    name,
		metrics.LabelImage:                   image,
		"pod_name":                           podName,
		"namespace":                          namespace,
		"container_name":                     containerName,
		metrics.ContainerLabelPrefix + "app": appName,
	}
	return set
}

func (*yanqingCollector) collectLastSeen(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(yanqingScropedLastSeenDesc, prometheus.GaugeValue, float64(time.Now().Unix()))
}

func (y *yanqingCollector) collectContainerStats(ch chan<- prometheus.Metric) {
	containerInfos := y.cacheStorage.GetAllContainerInfo()
	for _, container := range containerInfos {
		labels, values := []string{}, []string{}
		for l, v := range y.containerLabelsFunc(container) {
			labels = append(labels, sanitizeLabelName(l))
			values = append(values, v)
		}
		l := len(container.Stats)
		if l > 0 {
			stats := container.Stats[l-1]
			for _, cm := range y.containerMetrics {
				desc := cm.desc(labels)
				for _, metricValue := range cm.getValues(stats) {
					ch <- prometheus.MustNewConstMetric(desc, cm.valueType, float64(metricValue.value), append(values, metricValue.labels...)...)
				}
			}
		}
	}
}

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}
