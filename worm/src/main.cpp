#include <cmath>
#include <iostream>
#include <unistd.h>

#include "Worm.hpp"

int main() {
  Worm worm;

  int pos_x = 0;
  int pos_y = 0;

  int direction = 0;

  while (true) {
    worm.chemotaxis();

    std::cout << "Left: " << worm.getLeftMuscle()
              << ", Right: " << worm.getRightMuscle() << std::endl;

    direction += -(worm.getRightMuscle() - worm.getLeftMuscle()) / 5;
    direction = (direction + 360) % 360;
    int distance = (worm.getRightMuscle() + worm.getLeftMuscle()) / 2;

    pos_x += sin(direction * 3.14 / 180) * distance;
    pos_y += cos(direction * 3.14 / 180) * distance;

    std::cout << "D: " << direction << ", X: " << pos_x << ", Y: " << pos_y
              << std::endl;

    usleep(10000);
  }

  return 0;
}
