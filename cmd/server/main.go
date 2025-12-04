package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mw "github.com/piheta/apicore/middleware"
	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/features/img"
	"github.com/piheta/seq.re/internal/features/ip"
	"github.com/piheta/seq.re/internal/features/link"
	"github.com/piheta/seq.re/internal/features/paste"
	"github.com/piheta/seq.re/internal/features/secret"
	"github.com/piheta/seq.re/internal/features/seqre"
	"github.com/piheta/seq.re/internal/features/web"
	localmw "github.com/piheta/seq.re/internal/middleware"
	"github.com/piheta/seq.re/internal/shared"

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

	imageService.StartCleanupWorker(1 * time.Hour)

	ipHandler := ip.NewIPHandler(ipService)
	linkHandler := link.NewLinkHandler(linkService)
	secretHandler := secret.NewSecretHandler(secretService)
	imageHandler := img.NewImageHandler(imageService)
	pasteHandler := paste.NewPasteHandler(pasteService)
	seqreHandler := seqre.NewSeqreHandler(version, commit, date)
	webHandler := web.NewWebHandler()

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Web UI routes
	mux.Handle("GET /", mw.Public(webHandler.ServeIndex))
	mux.Handle("GET /tab/url", mw.Public(webHandler.ServeURLTab))
	mux.Handle("GET /tab/image", mw.Public(webHandler.ServeImageTab))
	mux.Handle("GET /tab/secret", mw.Public(webHandler.ServeSecretTab))
	mux.Handle("GET /tab/code", mw.Public(webHandler.ServeCodeTab))
	mux.Handle("GET /tab/ip", mw.Public(webHandler.ServeIPTab))
	mux.Handle("GET /web/detect-ip", mw.Public(webHandler.DetectIP))

	// API routes
	mux.Handle("GET /api/ip", mw.Public(ipHandler.GetPublicIP))
	mux.Handle("GET /api/version", mw.Public(seqreHandler.GetVersion))

	mux.Handle("POST /api/links", mw.Public(linkHandler.CreateLink))
	mux.Handle("GET /api/links/{short}", localmw.RateLimit(2, 5, mw.Public(linkHandler.GetLinkByShort)))

	mux.Handle("POST /api/secrets", mw.Public(secretHandler.CreateSecret))
	mux.Handle("GET /api/secrets/{short}", localmw.RateLimit(2, 5, mw.Public(secretHandler.GetSecretByShort)))

	mux.Handle("POST /api/images", mw.Public(imageHandler.CreateImage))
	mux.Handle("GET /i/{short}", localmw.RateLimit(2, 5, mw.Public(imageHandler.GetImageByShort)))

	mux.Handle("POST /api/pastes", mw.Public(pasteHandler.CreatePaste))
	mux.Handle("GET /p/{short}", localmw.RateLimit(2, 5, mw.Public(pasteHandler.GetPasteByShort)))

	// Apply rate limiting to redirect endpoint: 2 requests per second with burst of 5
	// This must be last to not conflict with other routes
	mux.Handle("GET /{short}", localmw.RateLimit(2, 5, mw.Public(linkHandler.RedirectByShort)))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mw.SecurityHeaders(mw.RequestLogger(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
