all: lib

lib: lib.c tinyalloc.c
	zig build-exe lib.c tinyalloc.c  -target wasm32-freestanding -fno-entry --export=info --export=handle --export=alloc --export=free  --import-symbols --import-memory

clean:
	rm -f lib.wasm
