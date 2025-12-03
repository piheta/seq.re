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
	"github.com/piheta/seq.re/internal/features/img"
	"github.com/piheta/seq.re/internal/features/ip"
	"github.com/piheta/seq.re/internal/features/link"
	"github.com/piheta/seq.re/internal/features/paste"
	"github.com/piheta/seq.re/internal/features/secret"
	"github.com/piheta/seq.re/internal/features/seqre"
	localmw "github.com/piheta/seq.re/internal/middleware"
	"github.com/piheta/seq.re/internal/shared"
	"golang.org/x/time/rate"

	_ "github.com/piheta/seq.re/docs"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func init() {
	shared.InitValidator()
	config.InitEnv()
	if err := config.ConnectDB(config.GetDataPath() + "/badger"); err != nil {
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
	slog.With("version", version).With("commit", commit).With("date", date).Info("Starting seq.re server")

	mux := http.NewServeMux()

	linkRepo := link.NewLinkRepo(config.DB)
	secretRepo := secret.NewSecretRepo(config.DB)
	imageRepo := img.NewImageRepo(config.DB)
	pasteRepo := paste.NewPasteRepo(config.DB)

	ipService := ip.NewIPService()
	linkService := link.NewLinkService(linkRepo)
	secretService := secret.NewSecretService(secretRepo)
	imageService := img.NewImageService(imageRepo, config.GetDataPath()+"/imgs")
	pasteService := paste.NewPasteService(pasteRepo)

	ipHandler := ip.NewIPHandler(ipService)
	linkHandler := link.NewLinkHandler(linkService)
	secretHandler := secret.NewSecretHandler(secretService)
	imageHandler := img.NewImageHandler(imageService)
	pasteHandler := paste.NewPasteHandler(pasteService)
	seqreHandler := seqre.NewSeqreHandler(version, commit, date)

	mux.Handle("GET /api/ip", middleware.Public(ipHandler.GetPublicIP))
	mux.Handle("GET /api/version", middleware.Public(seqreHandler.GetVersion))

	mux.Handle("GET /api/links/{short}", middleware.Public(linkHandler.GetLinkByShort))
	mux.Handle("POST /api/links", middleware.Public(linkHandler.CreateLink))

	mux.Handle("GET /api/secrets/{short}", middleware.Public(secretHandler.GetSecretByShort))
	mux.Handle("POST /api/secrets", middleware.Public(secretHandler.CreateSecret))

	mux.Handle("POST /api/images", middleware.Public(imageHandler.CreateImage))
	mux.Handle("GET /i/{short}", middleware.Public(imageHandler.GetImageByShort))

	mux.Handle("POST /api/pastes", middleware.Public(pasteHandler.CreatePaste))
	mux.Handle("GET /p/{short}", middleware.Public(pasteHandler.GetPasteByShort))

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
