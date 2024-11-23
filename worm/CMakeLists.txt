cmake_minimum_required(VERSION 3.5)

project(deepworm VERSION 1.0.0 DESCRIPTION "Deepworm" LANGUAGES C CXX)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Include SDL2 and SDL2_gfx headers
include_directories(/opt/homebrew/include)

# Find the SDL2 and SDL2_gfx libraries
find_library(SDL2_LIBRARY NAMES SDL2 PATHS /opt/homebrew/lib)
find_library(SDL2_GFX_LIBRARY NAMES SDL2_gfx PATHS /opt/homebrew/lib)

add_subdirectory(Nematoduino_Library)

add_executable(deepworm src/main.cpp)

# Link the libraries explicitly
target_link_libraries(deepworm PRIVATE nematoduino ${SDL2_LIBRARY} ${SDL2_GFX_LIBRARY})
