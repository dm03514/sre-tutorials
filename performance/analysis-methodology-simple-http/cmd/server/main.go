package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
	"log"
	"database/sql"
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Postgres struct {
	db *sql.DB
}

type People []Person

type Person struct {
	Address string
	FullName string
	Age int
}

type Payload struct {
	Age int
}

func (p Payload) Bytes() ([]byte, error) {
	return json.Marshal(p)
}


type Handler struct {
	Postgres *Postgres
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload := Payload{}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		msg := fmt.Sprintf("received: %q.  Expected message of format %+v",
			err, Payload{})
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	result, err := h.Postgres.FindByAge(payload.Age)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&result)
}


func (p *Postgres) FindByAge(age int) (People, error) {
	var people People

	q := `
SELECT address, full_name, age FROM people WHERE age = ?
`

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

	http.Handle("/metrics", promhttp.Handler())

	s := &http.Server{
		Addr:           ":8080",
		Handler:        h,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Printf("starting_server: %q\n", s.Addr)
	log.Fatal(s.ListenAndServe())
}
