package decrypt

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	decrypterPoolQueued = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "decrypter_pool_queued_operations_total",
		Help: "How many operations are pending for the pool",
	})
)

func init() {
	prometheus.MustRegister(decrypterPoolQueued)
}

// Pool handles scheduling decryption to respect hard bounds such as Number of CPU
// cores.
type Pool struct {
	numWorkers int
	inputChan  chan workerRequest
}

func (p *Pool) IsMatch(b Bcrypter) bool {
	work := workerRequest{
		bcrypter:   b,
		outputChan: make(chan bool),
	}
	decrypterPoolQueued.Inc()
	p.inputChan <- work
	decrypterPoolQueued.Dec()

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
