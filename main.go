package main

import (
	"decoration/handlers"
	"decoration/store"
	"log"
	"net/http"
)

func main() {
	phaseStore := store.NewMemoryStore()
	phaseHandler := handlers.NewPhaseHandler(phaseStore)

	mux := http.NewServeMux()
	mux.Handle("/api/phases/", phaseHandler)

	addr := ":8080"
	log.Printf("server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
