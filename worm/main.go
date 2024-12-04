package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// go:embed app/index.html
var html []byte

func main() {

	fmt.Println(html)

	// -------------------------------------------------------------------------
	// Initialize the logger
	logConfig := zap.NewDevelopmentConfig()
	logConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	log, err := logConfig.Build()
	if err != nil {
		log.Sugar().Fatalf("error creating logger: %v", err)
	}
	zap.ReplaceGlobals(log)
	log.Info("logger initialized")

	// -------------------------------------------------------------------------
	// Create an error channel to catch errors from the goroutines
	log.Info("creating error channel")
	errChan := make(chan error)

	// -------------------------------------------------------------------------
	// Create a worm PriceFetcher instance
	log.Info("creating worm price fetcher")
	wormPriceFetcher := newPriceFetcher(wormAddr)
	go func() {
		errChan <- wormPriceFetcher.fetchPrice()
	}()

	// -------------------------------------------------------------------------
	// Create the server
	log.Info("creating server")
	s := newServer(log)
	go func() {
		errChan <- s.start()
	}()

	// -------------------------------------------------------------------------
	// Run the worm
	go runWorm(wormPriceFetcher)

	// -------------------------------------------------------------------------
	// Catch errors from the goroutines
	for err := range errChan {
		if err != nil {
			log.Error(err.Error())
		}
	}
}

type server struct {
	log *zap.Logger
	r   *chi.Mux
}

func newServer(l *zap.Logger) *server {
	s := &server{
		log: l,
		r:   chi.NewRouter(),
	}

	// Serve index.html in the /app directory
	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("app", "index.html")
		s.log.Info("serving index.html", zap.String("path", filePath))
		http.ServeFile(w, r, filePath)
	})

	s.r.Get("/worm", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		s.log.Info("serving worm positions")
		json.NewEncoder(w).Encode(wormPositions)
	})

	return s
}

func (s *server) start() error {
	return http.ListenAndServe(":8080", s.r)
}
