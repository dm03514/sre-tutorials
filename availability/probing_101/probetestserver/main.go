package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stdout,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func durationFromString(ds string) (time.Duration, error) {
	if ds == "" {
		ds = "0s"
	}
	return time.ParseDuration(ds)
}

type artificialFailureHandler struct {
	next               http.Handler
	requestFailureRate int

	mu          sync.Mutex
	numRequests int
}

func (h *artificialFailureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Printf("artificialFailureHandler.ServeHTTP()")
	h.mu.Lock()
	h.numRequests += 1
	h.mu.Unlock()

	if h.requestFailureRate > 0 && h.numRequests%h.requestFailureRate == 0 {
		panic(fmt.Errorf(
			"injected error, requests count: %d w/ rate of %d", h.numRequests, h.requestFailureRate,
		))
	}

	h.next.ServeHTTP(w, r)
}

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Printf("handler.ServeHTTP()")
	_, err := ioutil.ReadAll(r.Body)
	logger.Printf("received request for %q", r.URL)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
}

type latencyHandler struct {
	next http.Handler
}

func (lh latencyHandler) RandDuration(max time.Duration, min time.Duration) time.Duration {
	if max.Nanoseconds() == 0 && min.Nanoseconds() == 0 {
		return 0
	}

	r := rand.Intn(
		int(max.Nanoseconds())-int(min.Nanoseconds())) + int(min.Nanoseconds())

	return time.Duration(r) * time.Nanosecond
}

func (lh latencyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Printf("latencyHandler.ServeHTTP()")

	minDuration, err := durationFromString(r.URL.Query().Get("minDuration"))
	if err != nil {
		panic(err)
	}

	maxDuration, err := durationFromString(r.URL.Query().Get("maxDuration"))
	if err != nil {
		panic(err)
	}
	if maxDuration.Nanoseconds() < minDuration.Nanoseconds() {
		panic(fmt.Errorf("maxDuration (%s) less than minDuration (%s)",
			maxDuration, minDuration))
	}

	d := lh.RandDuration(maxDuration, minDuration)
	logger.Printf("sleeping for %s", d)
	time.Sleep(d)

	lh.next.ServeHTTP(w, r)
}

func main() {
	var requestFailureRate = flag.Int("request-failure-rate", 0, "set to determine how many requests "+
		"should fail.  The default are 0 artificially failed requests, a rate of 1 will mean every request, a rate of two will mean "+
		"every other request, etc.")
	var addr = flag.String("addr", "127.0.0.1:5000", "addr/port for the tes")
	flag.Parse()

	h := &http.Server{
		Addr: *addr,
		Handler: &artificialFailureHandler{
			requestFailureRate: *requestFailureRate,
			next: latencyHandler{
				next: handler{},
			},
		},
	}

	logger.Printf("starting test server on: %q\n", *addr)

	if err := h.ListenAndServe(); err != nil {
		logger.Println(err)
	}

}
