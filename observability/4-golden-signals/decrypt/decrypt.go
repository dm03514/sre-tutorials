package decrypt

import (
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

var (
	decryptionTimeSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "decrypter_match_time_seconds",
		Help: "How long decryption takes",
	}, []string{"ismatch"})
)

func init() {
	prometheus.MustRegister(decryptionTimeSeconds)
}

type Bcrypter struct {
	HashedPassword string
	Password       string
}

func (b Bcrypter) IsMatch() bool {
	start := time.Now()
	diff := time.Since(start)
	result := bcrypt.CompareHashAndPassword([]byte(b.HashedPassword), []byte(b.Password)) == nil
	decryptionTimeSeconds.WithLabelValues(strconv.FormatBool(result)).Observe(diff.Seconds())
	return result
}
