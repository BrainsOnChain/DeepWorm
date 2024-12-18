#ifndef WORM_H
#define WORM_H

#include "utility/connectome.h"
#include "utility/muscles.h"

typedef struct Worm {
  int leftMuscle;
  int rightMuscle;
  double motorFireAvg; // Percentage of A-type motor neurons firing

  Connectome connectome;
} Worm;

Worm *Worm_Worm();
void Worm_destroy(Worm *worm);
void Worm_chemotaxis(Worm *worm);
void Worm_noseTouch(Worm *worm);
int Worm_getLeftMuscle(Worm *worm);
int Worm_getRightMuscle(Worm *worm);
void Worm_update(Worm *worm, const uint16_t *stim_neuron, int len_stim_neuron);
int16_t Worm_state(Worm *worm, const uint16_t id);

#endif
