package decrypt

import (
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/bcrypt"
)

var (
	requestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "decrypter_worker_requests_total",
		Help: "Total # of requests made to a bcrypt worker",
	})
)

func init() {
	prometheus.MustRegister(requestsTotal)
}

type Bcrypter struct {
	HashedPassword string
	Password       string
}

func (b Bcrypter) IsMatch() bool {
	requestsTotal.Inc()
	result := bcrypt.CompareHashAndPassword([]byte(b.HashedPassword), []byte(b.Password)) == nil
	return result
}
