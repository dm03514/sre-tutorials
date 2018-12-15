package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/dm03514/sre-tutorials/observability/4-golden-signals/decrypt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of requests received from external clients",
	})
	responsesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_responses_total",
		Help: "Total number of responses returned to the client",
	})
	externDep1ServiceRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "extern_dep1_service_requests_total",
		Help: "Total number of requests made to extern Dep1 Service",
	})
)

func init() {
	prometheus.MustRegister(
		requestsTotal,
		responsesTotal,
		externDep1ServiceRequestsTotal,
	)
}

func makeDep1Req() error {
	externDep1ServiceRequestsTotal.Inc()
	depResp, err := http.Get("http://dep1http:8080/pong")
	if err != nil {
		return err
	}

	depResp.Body.Close()
	// do something with response
	return nil
}

// returns true if the configuration file exists
func configExist() bool {
	// emulate reading to FS
	return true
}

type Handler struct {
	DecrypterPool *decrypt.Pool
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestsTotal.Inc()
	defer responsesTotal.Inc()

	var payload decrypt.Bcrypter
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		msg := fmt.Sprintf("received: %q.  Expected message of format %+v",
			err, decrypt.Bcrypter{})
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	r.Body.Close()

	configExist()

	if err := makeDep1Req(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	match := h.DecrypterPool.IsMatch(payload)
	resp := decrypt.HTTPResponse{
		Match: match,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
}

func main() {
	numDecryptPoolWorkers := flag.Int("num-decrypt-pool-workers", 4, "")
	flag.Parse()

	fmt.Printf("starting_server: :8080\n")
	http.Handle("/decrypt", Handler{
		DecrypterPool: decrypt.NewPool(*numDecryptPoolWorkers),
	})
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
