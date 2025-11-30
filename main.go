package main

import (
	"log"
	"net/http"
	"time"

	"github.com/piheta/apicore/middleware"
	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/features/address"
	"github.com/piheta/seq.re/internal/features/link"
	"github.com/piheta/seq.re/internal/shared"
)

func main() {
	mux := http.NewServeMux()

	shared.InitValidator()
	config.InitEnv()
	if err := config.ConnectDB(); err != nil {
		log.Fatal(err)
	}

	linkRepo := link.NewLinkRepo(config.DB)

	addressService := address.NewAddressService()
	linkService := link.NewLinkService(linkRepo)

	addressHandler := address.NewAddressHandler(addressService)
	linkHandler := link.NewLinkHandler(linkService)

	mux.Handle("GET /api/ip", middleware.Public(addressHandler.GetPublicIP))
	mux.Handle("POST /api/link", middleware.Public(linkHandler.CreateLink))
	mux.Handle("GET /{short}", middleware.Public(linkHandler.RedirectByShort))

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
