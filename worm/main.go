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

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	// Create a new Worm instance
	worm := C.Worm_Worm()
	defer C.Worm_destroy(worm) // Ensure proper cleanup

	// Initialize position and tracking
	p := position{x: 0, y: 0, direction: 0}
	var points plotter.XYs

	// Run the simulation
	for i := 0; i < 1000; i++ {
		C.Worm_chemotaxis(worm)
		C.Worm_noseTouch(worm)

		left := float64(C.Worm_getLeftMuscle(worm))
		right := float64(C.Worm_getRightMuscle(worm))
		angle, magnitude := movement(left, right)

		p.update(angle, magnitude)
		fmt.Printf("%+v\n", p)

		// Append the current position to points
		points = append(points, plotter.XY{X: p.x, Y: p.y})
	}

	// Plot the points
	if err := plotPath(points); err != nil {
		fmt.Println("Error plotting:", err)
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

// plotPath generates a plot of the worm's movement
func plotPath(points plotter.XYs) error {
	p := plot.New()
	p.Title.Text = "Worm Movement"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	line, err := plotter.NewLine(points)
	if err != nil {
		return err
	}
	p.Add(line)

	// Save the plot to a file
	return p.Save(6*vg.Inch, 6*vg.Inch, "worm_movement.png")
}
