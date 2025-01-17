package main

import (
	"sync"

	"github.com/brainsonchain/deepworm/src"
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
	worm.Run(wormPriceFetcher, mutex)
}
