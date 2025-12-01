package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piheta/apicore/middleware"
	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/features/address"
	"github.com/piheta/seq.re/internal/features/link"
	localmw "github.com/piheta/seq.re/internal/middleware"
	"github.com/piheta/seq.re/internal/shared"
	"golang.org/x/time/rate"

	_ "github.com/piheta/seq.re/docs"
)

func init() {
	shared.InitValidator()
	config.InitEnv()
	if err := config.ConnectDB(); err != nil {
		log.Fatal(err)
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		slog.Info("Shutting down...")
		_ = config.Close()
		os.Exit(0)
	}()
}

// @Title Seq.re
func main() {
	mux := http.NewServeMux()

	linkRepo := link.NewLinkRepo(config.DB)

	addressService := address.NewAddressService()
	linkService := link.NewLinkService(linkRepo)

	addressHandler := address.NewAddressHandler(addressService)
	linkHandler := link.NewLinkHandler(linkService)

	mux.Handle("GET /ip", middleware.Public(addressHandler.GetPublicIP))
	mux.Handle("POST /link", middleware.Public(linkHandler.CreateLink))

	// Apply rate limiting to redirect endpoint: 2 requests per second with burst of 5
	rateLimitedRedirect := localmw.RateLimit(rate.Limit(2), 5)(middleware.Public(linkHandler.RedirectByShort))
	mux.Handle("GET /{short}", rateLimitedRedirect)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      middleware.RequestLogger(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
