package main

import (
	"github.com/brainsonchain/deepworm/src"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	// Create a worm EventFetcher instance
	log.Info("creating worm event fetcher")
	wormEventFetcher := src.NewEventFetcher()
	go func() {
		errChan <- wormEventFetcher.Fetch()
	}()

	// -------------------------------------------------------------------------
	// Run the worm
	log.Info("creating worm")
	worm := src.NewWorm()
	go worm.Run(wormPriceFetcher, wormEventFetcher)

	// -------------------------------------------------------------------------
	// Create the server
	log.Info("creating server")
	go func() {
		errChan <- worm.StateServe(wormEventFetcher)
	}()

	// -------------------------------------------------------------------------
	// Catch errors from the goroutines
	for err := range errChan {
		if err != nil {
			log.Error(err.Error())
		}
	}
}
