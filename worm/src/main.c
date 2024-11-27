#include <math.h>
#include <stdio.h>
#include <unistd.h>

#include "SDL2/SDL.h"
#include "SDL2/SDL2_gfxPrimitives.h"

#include "Worm.h"

int main() {
  if (SDL_Init(SDL_INIT_VIDEO) < 0) {
    return -1;
  }

  SDL_Window *window =
      SDL_CreateWindow("Path Drawing", SDL_WINDOWPOS_CENTERED,
                       SDL_WINDOWPOS_CENTERED, 1600, 1200, SDL_WINDOW_SHOWN);

  if (!window) {
    SDL_Quit();
    return -2;
  }

  SDL_Renderer *renderer =
      SDL_CreateRenderer(window, -1, SDL_RENDERER_ACCELERATED);

  if (!renderer) {
    SDL_DestroyWindow(window);
    SDL_Quit();
    return -3;
  }

  printf("Initial screen clear...\n");
  SDL_SetRenderDrawColor(renderer, 0, 0, 0, 255);
  SDL_RenderClear(renderer);
  SDL_RenderPresent(renderer);

  // Small delay to ensure window is ready
  SDL_Delay(100);

  printf("Initialization complete\n");

  Worm *worm = Worm_Worm();

  double pos_x = 0;
  double pos_y = 0;

  int direction = 0;

  while (1) {
    SDL_Event event;
    while (SDL_PollEvent(&event)) {
      if (event.type == SDL_QUIT ||
          (event.type == SDL_KEYDOWN && event.key.keysym.sym == SDLK_ESCAPE)) {
        return 0;
      }
    }

    Worm_chemotaxis(worm);

    printf("Left: %d, Right: %d\n", Worm_getLeftMuscle(worm),
           Worm_getRightMuscle(worm));

    direction += -(Worm_getRightMuscle(worm) - Worm_getLeftMuscle(worm)) / 5.0;
    direction = (direction + 360) % 360;
    double distance =
        (Worm_getRightMuscle(worm) + Worm_getLeftMuscle(worm)) / 100.0;

    double new_pos_x = pos_x + sin(direction * 3.14 / 180) * distance;
    double new_pos_y = pos_y + cos(direction * 3.14 / 180) * distance;

    filledCircleRGBA(renderer, new_pos_x + 400, new_pos_y + 300, 2, 255, 0, 0,
                     255);

    SDL_RenderPresent(renderer);

    pos_x = new_pos_x;
    pos_y = new_pos_y;

    printf("D: %d, X: %lf, Y: %lf\n", direction, pos_x, pos_y);

    usleep(16000);
  }

  Worm_destroy(worm);

  SDL_DestroyRenderer(renderer);
  SDL_DestroyWindow(window);
  SDL_Quit();

  return 0;
}
