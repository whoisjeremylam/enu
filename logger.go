package main

import (
	"log"
	"net/http"
	"time"
)

func Logger(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fn(w, r)
		log.Printf(
			"%s,%s,%s,%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	}
}
