# Worm

Implementation of the worm's brain and movements. It is based on [nematoduino](https://github.com/nategri/nematoduino) for the underlying simulation and translates nematoduino motor outputs to worm movements. 

## Prerequisites

- CMake
- C++ compiler (gcc or clang)

## Build

This project uses CMake to build the simulator.

```bash
# create a build dir
mkdir -p build && cd build

# cmake
cmake .. -DCMAKE_BUILD_TYPE=Release

# build
make
```

You should have a `deepworm` binary in the build directory.

## Usage

Run the deepworm binary using
```bash
./deepworm
```

It should output a series of logs that look something like this:
```
Left: 107, Right: 113
D: -44, X: -2691, Y: 7546
Left: 81, Right: 75
D: -45, X: -2746, Y: 7601
Left: 39, Right: 33
D: -46, X: -2771, Y: 7626
Left: 25, Right: 19
D: -47, X: -2787, Y: 7641
```

The `Left` and `Right` values correspond to the motor outputs from the neural simulation. `D` corresponds to the direction, measured in degrees with North set to 0. `X` and `Y` correspond to the worm's coordinates.

## License

This project is covered under the GNU Public License v2.
