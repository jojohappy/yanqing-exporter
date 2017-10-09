package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/yanqing-exporter/storage"
)

const (
	containersApi = "containers"
	apiResource   = "/api/"
)

const (
	apiRequestType = iota + 1
	// apiRequestArgs
)

var apiRegexp = regexp.MustCompile(`/api/([^/]+)?(.*)`)

func registerApiHandler(mux *http.ServeMux, ms storage.Storage) error {
	mux.HandleFunc(apiResource, func(w http.ResponseWriter, r *http.Request) {
		err := handleRequest(ms, w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})
	return nil
}

func handleRequest(ms storage.Storage, w http.ResponseWriter, r *http.Request) error {
	start := time.Now()
	defer func() {
		glog.V(4).Infof("Request took %s", time.Since(start))
	}()

	request := r.URL.Path

	requestElements := apiRegexp.FindStringSubmatch(request)
	if len(requestElements) == 0 {
		return fmt.Errorf("malformed request %q", request)
	}
	requestType := requestElements[apiRequestType]
	// requestArgs := strings.Split(requestElements[apiRequestArgs], "/")

	if requestType == "" {
		requestTypes := []string{containersApi}
		sort.Strings(requestTypes)
		http.Error(w, fmt.Sprintf("Supported request types: %q", strings.Join(requestTypes, ",")), http.StatusBadRequest)
		return nil
	}

	switch requestType {
	case containersApi:
		glog.V(4).Infof("Api - Container")
		containerInfos := ms.GetAllContainerInfo()
		return writeResult(containerInfos, w)
	default:
		return fmt.Errorf("unknown request type %q", requestType)
	}
	return nil
}

func writeResult(res interface{}, w http.ResponseWriter) error {
	out, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("failed to marshall response %+v with error: %s", res, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
	return nil
}
