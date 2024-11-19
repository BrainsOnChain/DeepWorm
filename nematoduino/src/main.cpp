#include <iostream>
#include <unistd.h>

#include "Worm.hpp"

int main() {
  Worm worm;

  while (true) {
    worm.chemotaxis();

    std::cout << "Left: " << worm.getLeftMuscle()
              << ", Right: " << worm.getRightMuscle() << std::endl;

    usleep(100000);
  }

  return 0;
}
