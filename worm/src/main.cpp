#include <cmath>
#include <iostream>
#include <unistd.h>

#include "SDL2/SDL.h"
#include "SDL2/SDL2_gfxPrimitives.h"

#include "Worm.hpp"

int main() {
  if (SDL_Init(SDL_INIT_VIDEO) < 0) {
    throw std::runtime_error("SDL initialization failed");
  }

  auto window =
      SDL_CreateWindow("Path Drawing", SDL_WINDOWPOS_CENTERED,
                       SDL_WINDOWPOS_CENTERED, 1600, 1200, SDL_WINDOW_SHOWN);

  if (!window) {
    SDL_Quit();
    throw std::runtime_error("Window creation failed");
  }

  auto renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED);

  if (!renderer) {
    SDL_DestroyWindow(window);
    SDL_Quit();
    throw std::runtime_error("Renderer creation failed");
  }

  std::cout << "Initial screen clear..." << std::endl;
  SDL_SetRenderDrawColor(renderer, 0, 0, 0, 255);
  SDL_RenderClear(renderer);
  SDL_RenderPresent(renderer);

  // Small delay to ensure window is ready
  SDL_Delay(100);

  std::cout << "Initialization complete" << std::endl;

  Worm worm;

  double pos_x = 0;
  double pos_y = 0;

  int direction = 0;

  while (true) {
    SDL_Event event;
    while (SDL_PollEvent(&event)) {
      if (event.type == SDL_QUIT ||
          (event.type == SDL_KEYDOWN && event.key.keysym.sym == SDLK_ESCAPE)) {
        return 0;
      }
    }

    worm.chemotaxis();

    std::cout << "Left: " << worm.getLeftMuscle()
              << ", Right: " << worm.getRightMuscle() << std::endl;

    direction += -(worm.getRightMuscle() - worm.getLeftMuscle()) / 5;
    direction = (direction + 360) % 360;
    double distance = (worm.getRightMuscle() + worm.getLeftMuscle()) / 100.0;

    auto new_pos_x = pos_x + sin(direction * 3.14 / 180) * distance;
    auto new_pos_y = pos_y + cos(direction * 3.14 / 180) * distance;

    filledCircleRGBA(renderer, new_pos_x + 400, new_pos_y + 300, 2, 255, 0, 0,
                     255);

    SDL_RenderPresent(renderer);

    pos_x = new_pos_x;
    pos_y = new_pos_y;

    std::cout << "D: " << direction << ", X: " << pos_x << ", Y: " << pos_y
              << std::endl;

    usleep(16000);
  }

  SDL_DestroyRenderer(renderer);
  SDL_DestroyWindow(window);
  SDL_Quit();

  return 0;
}
