package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jaguerra/go-comic-sniffer/sniffer"
)

func main() {
	port, portPresent := os.LookupEnv("PORT")
	if !portPresent {
		log.Fatal("PORT is not defined, exiting")
		os.Exit(-1)
	}
	log.Print("Starting server on port " + port + " ...")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/random", sniffer.Handler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}
