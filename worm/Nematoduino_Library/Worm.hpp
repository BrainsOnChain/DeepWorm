#ifndef WORM_HPP
#define WORM_HPP

extern "C" {
#include "utility/connectome.h"
#include "utility/muscles.h"
};

class Worm {
  public:
    Worm();


    void chemotaxis();

    void noseTouch();
    void noseTouch(int i);

    int getLeftMuscle();
    int getRightMuscle();

  private:
    int _leftMuscle;
    int _rightMuscle;
    double _motorFireAvg; // Percentage of A-type motor neurons firing

    int debug_left_count;
    int debug_right_count;

    Connectome _connectome;
    void _update(const uint16_t*, int);
};

#endif
