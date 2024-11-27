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
import "fmt"

func main() {
	// Create a new Worm instance
	worm := C.Worm_Worm()
	defer C.Worm_destroy(worm) // Ensure proper cleanup

	// Trigger chemotaxis and noseTouch signals
	C.Worm_chemotaxis(worm)
	C.Worm_noseTouch(worm)

	// Get muscle activity
	left := int(C.Worm_getLeftMuscle(worm))
	right := int(C.Worm_getRightMuscle(worm))

	fmt.Printf("Left Muscle: %d, Right Muscle: %d\n", left, right)
}
