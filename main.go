package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/golang/glog"

	"github.com/yanqing-exporter/collector"
	"github.com/yanqing-exporter/collector/cadvisor"
	yanqinghttp "github.com/yanqing-exporter/http"
	"github.com/yanqing-exporter/storage"
)

var listenIp = flag.String("listen_ip", "", "IP to listen on, defaults all")
var listenPort = flag.Int("port", 9187, "listen port")
var cadvisorListenPort = flag.Int("cadvisor_port", 8080, "listen port for cadvisor")
var prometheusEndpoint = flag.String("prometheus_endpoint", "/metrics", "Endpoint to expose Prometheus metrics on")
var maxStatsLength = flag.Int("max_stats_length", 5, "maximal length of stats to store")

func main() {
	flag.Parse()
	hostIp, err := parseHostIp(os.Getenv("HOST_IP"))
	if nil != err {
		glog.Errorf("Failed to parse hostip: %v", err)
		os.Exit(1)
	}

	glog.Infof("host ip is %s", hostIp)

	memoryStorage := storage.New(*maxStatsLength)

	cadvisorClient, err := cadvisor.New(hostIp, *cadvisorListenPort)
	if nil != err {
		glog.Errorf("Failed to create cadvisor rest client: %v", err)
		os.Exit(1)
	}

	metricsCollector, err := collector.NewCollector(memoryStorage, cadvisorClient)
	if nil != err {
		os.Exit(1)
	}

	mux := http.NewServeMux()
	err = yanqinghttp.RegisterHandler(mux, memoryStorage, *prometheusEndpoint)
	if err != nil {
		glog.Fatalf("Failed to register HTTP handlers: %v", err)
		os.Exit(1)
	}

	metricsCollector.Start()

	server := fmt.Sprintf("%s:%d", *listenIp, *listenPort)
	glog.Fatal(http.ListenAndServe(server, mux))
}

func parseHostIp(s string) (net.IP, error) {
	ip := net.ParseIP(s)

	if ip := ip.To4(); nil != ip {
		return ip, nil
	}
	return nil, fmt.Errorf("Can not parse host ip %s", s)
}
