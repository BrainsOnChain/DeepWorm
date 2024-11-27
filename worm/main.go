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
	"fmt"
	"math"
)

func main() {
	// Create a new Worm instance
	worm := C.Worm_Worm()
	defer C.Worm_destroy(worm) // Ensure proper cleanup

	// Create a worm PriceFetcher instance
	wormPriceFetcher := newPriceFetcher("Worm", wormAddr)
	go wormPriceFetcher.fetchPrice()

	runPriceWorm(wormPriceFetcher, worm)
}

func runPriceWorm(fetcher *priceFetcher, worm *C.Worm) {
	// Initialize position and tracking
	p := position{x: 0, y: 0, direction: 0}

	// Fetch the price of worm and compare to previous price
	currentPrice := <-fetcher.priceChan
	for price := range fetcher.priceChan {
		priceChange := math.Abs(price.priceUSD - currentPrice.priceUSD)
		currentPrice = price

		if priceChange == 0 { // no change in price no worm movement
			continue
		}

		// Magnify the price change to simulate worm movement
		intChange := int(priceChange * 1_000_000)
		fmt.Println("Price change:", intChange)

		for i := 0; i < intChange; i++ {
			C.Worm_chemotaxis(worm)
			C.Worm_noseTouch(worm)

			angle, magnitude := movement(
				float64(C.Worm_getLeftMuscle(worm)),
				float64(C.Worm_getRightMuscle(worm)),
			)

			p.update(angle, magnitude)
			fmt.Printf("%+v\n", p)
		}
	}
}

type position struct {
	x, y      float64
	direction float64
}

func (p *position) update(angle, magnitude float64) {
	// Update the direction
	p.direction += angle
	if p.direction < 0 {
		p.direction += 360
	} else if p.direction >= 360 {
		p.direction -= 360
	}

	// Update the position based on the direction
	p.x += magnitude * math.Cos(p.direction*math.Pi/180) // Convert to radians
	p.y += magnitude * math.Sin(p.direction*math.Pi/180)
}

// movement outputs the movement in the form of angle and magnitude based on the
// left and right muscle activity.
func movement(left, right float64) (float64, float64) {
	// Calculate the angle and magnitude
	angle := (right - left) / 2
	magnitude := (right + left) / 2

	return angle, magnitude
}
