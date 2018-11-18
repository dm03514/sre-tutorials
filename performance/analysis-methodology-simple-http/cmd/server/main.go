package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"html"
	"net/http"
	"net/http/pprof"
	"time"
	"log"
	"database/sql"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/lib/pq"
)

func errToStatus(err error) string {
	if err != nil {
		return "error"
	}
	return "success"
}

var (
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_seconds",
		Help: "Distribution of request lengths",
	}, []string{"path"})

	findByAgeLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "find_by_age_seconds",
		Help: "Distribution of find by age durations, status=success|error",
	}, []string{"status"})

	findByAgeResultCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "find_by_age_results_count",
		Help: "# of results returned from find by age",
	}, []string{"status"})
)

func init() {
	prometheus.MustRegister(requestLatency)
	prometheus.MustRegister(findByAgeLatency)
	prometheus.MustRegister(findByAgeResultCount)
}

type Postgres struct {
	db *sql.DB
}

type PeopleResponse struct {
	People []Person
}

type Person struct {
	Address string
	FullName string
	Age int
}

type Payload struct {
	Age int
}

type Handler struct {
	Postgres *Postgres
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	defer func() {
		diff := time.Since(start)
		// fmt.Fprintf(w, "Took: %s\n", diff)
		requestLatency.WithLabelValues(html.EscapeString(r.URL.Path)).Observe(diff.Seconds())
	}()

	payload := Payload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		msg := fmt.Sprintf("received: %q.  Expected message of format %+v",
			err, Payload{})
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	findStart := time.Now()
	people, err := h.Postgres.FindByAge(payload.Age)

	findByAgeLatency.WithLabelValues(
		errToStatus(err),
	).Observe(
		time.Since(findStart).Seconds(),
	)

	findByAgeResultCount.WithLabelValues(errToStatus(err)).Set(float64(len(people)))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := PeopleResponse{
		People: people,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&resp)
}


func (p *Postgres) FindByAge(age int) ([]Person, error) {
	people := []Person{}

	q := `SELECT address, full_name, age FROM people WHERE age = $1`

	rows, err := p.db.Query(q, age)
	if err != nil {
		return people, nil
	}
	defer rows.Close()
	for rows.Next() {
		person := Person{}

		if err := rows.Scan(&person.Address, &person.FullName, &person.Address); err != nil {
			return people, err
		}

		people = append(people, person)
	}

	if err := rows.Err(); err != nil {
		return people, err
	}

	return people, nil
}


func NewPostges(dbConnectionString string) (*Postgres, error) {
	db, err := sql.Open("postgres", dbConnectionString)
	// db.SetMaxOpenConns(8)
	// db.SetMaxIdleConns(8)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &Postgres{
		db: db,
	}, nil
}

func AttachProfiler(router *http.ServeMux) {
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

	// Manually add support for paths linked to by index page at /debug/pprof/
	router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handle("/debug/pprof/block", pprof.Handler("block"))
}

func main() {
	dbConnectionString := flag.String("db-connection-string", "", "")
	flag.Parse()

	postgres, err := NewPostges(*dbConnectionString)
	if err != nil {
		panic(err)
	}

	h := &Handler{
		Postgres: postgres,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	AttachProfiler(mux)
	mux.Handle("/", h)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("starting_server: %q\n", s.Addr)
	log.Fatal(s.ListenAndServe())
}
