# Worm

Implementation of the worm's brain and movements. It is based on
[nematoduino](https://github.com/nategri/nematoduino) for the underlying
simulation and translates nematoduino motor outputs to worm movements.

This project uses Go and CGO to interface with the C code that runs the worm's
brain. To build the C code, you need to have CMake and a C++ compiler installed.

### Movements
The worm responds to movements in the price of the asset it is tracking ($worm).

## Prerequisites

- CMake
- C++ compiler (gcc or clang)
- Go 1.22.5 or higher

## Build & Run

```bash
$ make build-worm-c

$ go run .
```

## License

This project is covered under the GNU Public License v2.
