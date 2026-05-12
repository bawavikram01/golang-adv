CC = gcc
CFLAGS = -Wall -Wextra -g -fsanitize=address
SOURCES = $(wildcard code/*.c)
TARGETS = $(SOURCES:code/%.c=build/%)

all: build_dir $(TARGETS)

build_dir:
	@mkdir -p build

build/%: code/%.c
	$(CC) $(CFLAGS) $< -o $@

run-%: build/%
	./build/$*

clean:
	rm -rf build

# Run all programs
run-all: $(TARGETS)
	@for prog in build/*; do \
		echo "\n========== Running $$prog =========="; \
		./$$prog; \
	done

# Check for memory leaks with Valgrind
valgrind-%: build/%
	valgrind --leak-check=full --show-leak-kinds=all ./build/$*

.PHONY: all clean run-all build_dir
