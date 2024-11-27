package main

/*
#cgo CFLAGS: -I./nematoduino
#cgo LDFLAGS: -L. -lnematoduino
#include "Worm.h"
#include <stdlib.h>

// Declare the C functions
extern Worm* Worm_new();
extern void Worm_delete(Worm* worm);
extern void Worm_chemotaxis(Worm* worm);
extern void Worm_noseTouch(Worm* worm);
extern int Worm_getLeftMuscle(Worm* worm);
extern int Worm_getRightMuscle(Worm* worm);
*/
import "C"
import "fmt"

func main() {
	// Create a new Worm instance
	worm := C.Worm_new()
	defer C.Worm_delete(worm) // Ensure proper cleanup

	// Trigger chemotaxis and noseTouch signals
	C.Worm_chemotaxis(worm)
	C.Worm_noseTouch(worm)

	// Get muscle activity
	left := int(C.Worm_getLeftMuscle(worm))
	right := int(C.Worm_getRightMuscle(worm))

	fmt.Printf("Left Muscle: %d, Right Muscle: %d\n", left, right)
}
