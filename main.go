package main

import (
	"log"
	"net/http"
	"time"

	"github.com/piheta/apicore/middleware"
	"github.com/piheta/seq.re/internal/features/address"
)

func main() {
	mux := http.NewServeMux()

	addressService := address.NewAddressService()
	addressHandler := address.NewAddressHandler(addressService)

	mux.Handle("GET /api/ip", middleware.Public(addressHandler.GetPublicIP))

	server := &http.Server{
		Addr:         ":8081",
		Handler:      middleware.RequestLogger(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
