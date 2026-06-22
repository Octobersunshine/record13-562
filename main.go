package main

import (
	"context"
	"decoration/handlers"
	"decoration/scheduler"
	"decoration/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	phaseStore := store.NewMemoryStore()
	phaseHandler := handlers.NewPhaseHandler(phaseStore)
	projectHandler := handlers.NewProjectHandler(phaseStore)

	mux := http.NewServeMux()
	mux.Handle("/api/phases/", phaseHandler)
	mux.Handle("/api/projects/", projectHandler)

	alertScheduler := scheduler.NewAlertScheduler(phaseStore, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go alertScheduler.Start(ctx)

	addr := ":8080"
	log.Printf("server starting on %s", addr)

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")
		cancel()
		srv.Shutdown(context.Background())
	}()

	log.Fatal(srv.ListenAndServe())
}
