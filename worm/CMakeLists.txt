cmake_minimum_required(VERSION 3.5)

project(deepworm VERSION 1.0.0 DESCRIPTION "Deepworm" LANGUAGES C CXX)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

add_subdirectory(Nematoduino_Library)

add_executable(deepworm src/main.cpp)

target_link_libraries(deepworm PRIVATE nematoduino)
