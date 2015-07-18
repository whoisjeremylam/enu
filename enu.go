package main

import (
	"log"
	"net/http"
)

func main() {
	router := NewRouter()

	log.Println("Enu API server started!!")
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
