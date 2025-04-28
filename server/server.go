package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// SetupRoutes sets up the HTTP routes
func SetupRoutes(handlePutObject, handleGetObject func(http.ResponseWriter, *http.Request)) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/object/{id}", handlePutObject).Methods("PUT")
	r.HandleFunc("/object/{id}", handleGetObject).Methods("GET")
	return r
}

// Start starts the gateway server
func Start(router *mux.Router) error {
	server := &http.Server{
		Addr:         ":3000",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Println("Starting Object Storage Gateway on :3000")
	return server.ListenAndServe()
}
