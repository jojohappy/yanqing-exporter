package http

import (
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yanqing-exporter/metrics"
	"github.com/yanqing-exporter/storage"
)

func RegisterHandler(mux *http.ServeMux, memoryStorage storage.Storage, prometheusEndpoint string) error {
	// handler healthz
	mux.HandleFunc("/healthz", handlerHealthz)

	// TODO: handler api
	err := registerApiHandler(mux, memoryStorage)
	if err != nil {
		glog.Fatalf("Failed to register API handlers: %v", err)
	}

	// handler prometheus
	err = registerPrometheusHandler(mux, memoryStorage, prometheusEndpoint)
	if err != nil {
		glog.Fatalf("Failed to register prometheus handlers: %v", err)
	}
	return nil
}

func registerPrometheusHandler(mux *http.ServeMux, memoryStorage storage.Storage, prometheusEndpoint string) error {
	r := prometheus.NewRegistry()
	r.MustRegister(
		metrics.NewCollector(memoryStorage),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(os.Getpid(), ""),
	)
	mux.Handle(prometheusEndpoint, promhttp.HandlerFor(r, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	return nil
}
