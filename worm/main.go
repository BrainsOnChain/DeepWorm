package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/brainsonchain/deepworm/src"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mutex = &sync.Mutex{} // create a mutex to lock the wormPositions slice

func main() {

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
	wormPriceFetcher := src.NewPriceFetcher(src.WormAddr)
	go func() {
		errChan <- wormPriceFetcher.Fetch()
	}()

	// -------------------------------------------------------------------------
	// Run the worm
	log.Info("creating worm")
	worm := src.NewWorm()
	go worm.Run(wormPriceFetcher, mutex)

	// -------------------------------------------------------------------------
	// Create the server
	log.Info("creating server")
	s := newServer(log, worm)
	go func() {
		errChan <- s.start()
	}()

	// -------------------------------------------------------------------------
	// Catch errors from the goroutines
	for err := range errChan {
		if err != nil {
			log.Error(err.Error())
		}
	}
}

type server struct {
	log  *zap.Logger
	r    *chi.Mux
	worm *src.Worm
}

func newServer(l *zap.Logger, w *src.Worm) *server {
	s := &server{
		log:  l,
		r:    chi.NewRouter(),
		worm: w,
	}

	// Serve index.html in the /app directory
	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("app", "index.html")
		s.log.Info("serving index.html", zap.String("path", filePath))
		http.ServeFile(w, r, filePath)
	})

	s.r.Get("/worm", func(w http.ResponseWriter, r *http.Request) {
		wormPositions := s.worm.Positions()

		s.log.Info("serving worm positions")
		json.NewEncoder(w).Encode(wormPositions)
	})

	return s
}

func (s *server) start() error {
	return http.ListenAndServe(":8080", s.r)
}
