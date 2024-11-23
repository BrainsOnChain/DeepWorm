#include <cmath>
#include <iostream>
#include <unistd.h>
#include <vector>
#include "SDL2/SDL.h"
#include "SDL2/SDL2_gfxPrimitives.h"
#include "Worm.hpp"

// Predefined gradient grid
int grid[9][9] = {
    { -4, -3, -2, -1, -1, -2, -3, -4, -4 },
    { -4, -3, -2, -1,  1, -2, -3, -4, -4 },
    { -4, -3, -2,  0,  2, -2, -3, -4, -4 },
    { -4, -3,  0,  3,  3,  2, -3, -4, -4 },
    { -4, -3, -2,  3,  4,  3, -2, -4, -4 },
    { -4, -3, -2,  3,  3,  3, -2, -3, -4 },
    { -4, -3, -2, -2, -2, -2, -2, -3, -4 },
    { -4, -3, -3, -3,  4, -3, -3, -3, -4 },
    { -4, -4, -4, -4, -4, -4, -4, -4, -4 }
};

// Worm and apple positions
int worm_x = 2, worm_y = 2;
int apple_x = 7, apple_y = 6;

void updateWormPosition(Worm& worm, int& worm_x, int& worm_y) {
  int left = worm.getLeftMuscle();
  int right = worm.getRightMuscle();

  // debug output
  std::cout << "Left muscle: " << left << ", Right muscle: " << right << std::endl;

  // Simulate simple movement based on muscle activity
  if (left > right) {
    worm_x = std::max(0, worm_x - 1); // Move left
  }
  else if (right > left) {
    worm_x = std::min(8, worm_x + 1); // Move right
  }
  else {
    worm_y = std::min(8, worm_y + 1); // Move downward
  }
}

void renderGrid(SDL_Renderer* renderer, Worm& worm) {
  // Render the grid
  for (int y = 0; y < 9; ++y) {
    for (int x = 0; x < 9; ++x) {
      if (x == worm_x && y == worm_y) {
        // Render worm as a green circle
        filledCircleRGBA(renderer, x * 50 + 25, y * 50 + 25, 20, 0, 255, 0, 255);
      }
      else if (x == apple_x && y == apple_y) {
        // Render apple as a red circle
        filledCircleRGBA(renderer, x * 50 + 25, y * 50 + 25, 20, 255, 0, 0, 255);
      }
      else {
        // Render empty cells
        rectangleRGBA(renderer, x * 50, y * 50, x * 50 + 50, y * 50 + 50, 255, 255, 255, 255);
      }
    }
  }
}

int main() {
  if (SDL_Init(SDL_INIT_VIDEO) < 0) {
    throw std::runtime_error("SDL initialization failed");
  }

  auto window = SDL_CreateWindow("Worm Pathfinding", SDL_WINDOWPOS_CENTERED, SDL_WINDOWPOS_CENTERED, 450, 450, SDL_WINDOW_SHOWN);

  if (!window) {
    SDL_Quit();
    throw std::runtime_error("Window creation failed");
  }

  auto renderer = SDL_CreateRenderer(window, -1, SDL_RENDERER_SOFTWARE | SDL_RENDERER_PRESENTVSYNC);

  if (!renderer) {
    SDL_DestroyWindow(window);
    SDL_Quit();
    throw std::runtime_error("Renderer creation failed");
  }

  SDL_SetRenderDrawColor(renderer, 0, 0, 0, 255);
  SDL_RenderClear(renderer);
  SDL_RenderPresent(renderer);

  Worm worm;
  bool running = true;

  while (running) {
    SDL_Event event;
    while (SDL_PollEvent(&event)) {
      if (event.type == SDL_QUIT || (event.type == SDL_KEYDOWN && event.key.keysym.sym == SDLK_ESCAPE)) {
        running = false;
      }
    }

    // Send signal based on the grid value
    int signal = grid[worm_y][worm_x];
    worm.chemotaxis(signal);

    // Trigger noseTouch if near the apple
    if (worm_x == apple_x && worm_y == apple_y) {
      worm.noseTouch(10);
    }
    else {
      worm.noseTouch(0);
    }

    // Update worm's position
    updateWormPosition(worm, worm_x, worm_y);

    // Render the grid and the worm
    SDL_SetRenderDrawColor(renderer, 0, 0, 0, 255);
    SDL_RenderClear(renderer);

    renderGrid(renderer, worm);

    SDL_RenderPresent(renderer);

    // Break if the worm reaches the apple
    if (worm_x == apple_x && worm_y == apple_y) {
      std::cout << "The worm reached the apple!" << std::endl;
      running = false;
    }

    usleep(200000);
  }

  SDL_DestroyRenderer(renderer);
  SDL_DestroyWindow(window);
  SDL_Quit();

  return 0;
}
