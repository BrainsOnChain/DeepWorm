package src

/*
#cgo CFLAGS: -I../nematoduino
#cgo LDFLAGS: -L.. -lnematoduino
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
	"math"
	"math/rand"
	"sync"

	"go.uber.org/zap"
)

const (
	magnification = 1_000_000
)

type Worm struct {
	cworm     *C.Worm
	mu        *sync.Mutex
	positions []position
}

func NewWorm() *Worm {
	return &Worm{
		cworm: C.Worm_Worm(),
		mu:    &sync.Mutex{},
	}
}

func (w *Worm) Run(fetcher *priceFetcher, mu *sync.Mutex) {
	defer C.Worm_destroy(w.cworm) // Ensure proper cleanup
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
		zap.S().Infow("price change", "change", intChange)

		w.mu.Lock()
		for i := 0; i < intChange; i++ {
			// 80% change of chemotaxis and 20% of nose touch for each cycle
			if rand.Intn(100) < 50 {
				C.Worm_chemotaxis(w.cworm)
			} else {
				C.Worm_noseTouch(w.cworm)
			}

			angle, magnitude := movement(
				float64(C.Worm_getLeftMuscle(w.cworm)),
				float64(C.Worm_getRightMuscle(w.cworm)),
			)

			p.update(angle, magnitude)
			w.positions = append(w.positions, p)
		}
		w.mu.Unlock()
	}
}

func (w *Worm) Positions() []position {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.positions
}

// -----------------------------------------------------------------------------
// Position and movement functions

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
