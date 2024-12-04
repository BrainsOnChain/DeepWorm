package main

/*
#cgo CFLAGS: -I./nematoduino
#cgo LDFLAGS: -L. -lnematoduino
#include "Worm.h"
#include <stdlib.h>

// Declare the C functions based on Worm.h
extern Worm* Worm_Worm();
extern void Worm_destroy(Worm* worm);
extern void Worm_chemotaxis(Worm* worm);
extern void Worm_noseTouch(Worm* worm);
extern int Worm_getLeftMuscle(Worm* worm);
extern int Worm_getRightMuscle(Worm* worm);
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"

	"github.com/go-chi/chi"

	_ "embed"
)

const (
	magnification = 1_000_000
)

var (
	wormPositions = []position{}

	// create a mutex to lock the wormPositions slice
	mutex = &sync.Mutex{}
)

// go:embed app/index.html
var html []byte

func main() {
	// Create a new Worm instance
	worm := C.Worm_Worm()
	defer C.Worm_destroy(worm) // Ensure proper cleanup

	// Create a worm PriceFetcher instance
	wormPriceFetcher := newPriceFetcher("Worm", wormAddr)
	go wormPriceFetcher.fetchPrice()

	// start a http server to serve the worm positions
	r := chi.NewRouter()
	r.Get("/worm", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		// Marshal the wormPositions slice to JSON
		json.NewEncoder(w).Encode(wormPositions)
	})

	// serve the html in /app directory
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(html)
	})

	go http.ListenAndServe(":8080", r)

	runPriceWorm(wormPriceFetcher, worm)
}

func runPriceWorm(fetcher *priceFetcher, worm *C.Worm) {
	var p position

	// Fetch the price of worm and compare to previous price
	currentPrice := <-fetcher.priceChan
	for price := range fetcher.priceChan {
		priceChange := math.Abs(price.priceUSD - currentPrice.priceUSD)
		currentPrice = price

		if priceChange == 0 { // no change in price no worm movement
			continue
		}

		// Magnify the price change to simulate worm movement
		intChange := int(priceChange * magnification)

		mutex.Lock()
		for i := 0; i < intChange; i++ {
			// 80% change of chemotaxis and 20% chance of nose touch for each
			// cycle
			if rand.Intn(100) < 80 {
				C.Worm_chemotaxis(worm)
			} else {
				C.Worm_noseTouch(worm)
			}

			angle, magnitude := movement(
				float64(C.Worm_getLeftMuscle(worm)),
				float64(C.Worm_getRightMuscle(worm)),
			)

			p.update(angle, magnitude)
			wormPositions = append(wormPositions, p)
			fmt.Printf("%+v\n", p)
		}
		mutex.Unlock()
	}
}

type position struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Direction float64 `json:"direction"`
}

func (p *position) update(angle, magnitude float64) {
	p.Direction += angle
	if p.Direction < 0 {
		p.Direction += 360
	} else if p.Direction >= 360 {
		p.Direction -= 360
	}

	// Update the position based on the Direction
	p.X += magnitude * math.Cos(p.Direction*math.Pi/180) // Convert to radians
	p.Y += magnitude * math.Sin(p.Direction*math.Pi/180)
}

// movement outputs the movement in the form of angle and magnitude based on the
// left and right muscle activity.
func movement(left, right float64) (float64, float64) {
	// Calculate the angle and magnitude
	angle := (right - left) / 2
	magnitude := (right + left) / 2

	return angle, magnitude
}
