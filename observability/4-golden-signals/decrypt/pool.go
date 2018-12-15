package decrypt

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	poolRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "decrypter_pool_requests_total",
		Help: "Total number of requests made ot the pool",
	})
)

func init() {
	prometheus.MustRegister(
		poolRequestsTotal,
	)
}

// Pool handles scheduling decryption to respect hard bounds such as Number of CPU
// cores.
type Pool struct {
	numWorkers int
	inputChan  chan workerRequest
}

func (p *Pool) IsMatch(b Bcrypter) bool {
	poolRequestsTotal.Inc()

	work := workerRequest{
		bcrypter:   b,
		outputChan: make(chan bool),
	}

	p.inputChan <- work

	return <-work.outputChan
}

type workerRequest struct {
	bcrypter   Bcrypter
	outputChan chan bool
}

func NewPool(numWorkers int) *Pool {
	inputChan := make(chan workerRequest)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for work := range inputChan {
				work.outputChan <- work.bcrypter.IsMatch()
			}
		}()
	}

	return &Pool{
		numWorkers: numWorkers,
		inputChan:  inputChan,
	}
}
