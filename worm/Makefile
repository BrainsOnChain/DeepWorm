.PHONY: confirm
_WARN := "\033[33m[%s]\033[0m %s\n"  # Yellow text for "printf"
_TITLE := "\033[32m[%s]\033[0m %s\n" # Green text for "printf"
_ERROR := "\033[31m[%s]\033[0m %s\n" # Red text for "printf"

CURRENT_BRANCH = $(shell git branch --show-current) 
COMMIT = $(shell git rev-parse --short=12 HEAD)

build-worm-c:
	cd nematoduino && \
	gcc -c Worm.c behaviors.c utility/*.c -I. -fPIC && \
	ar rcs libnematoduino.a *.o && \
	mv libnematoduino.a ..