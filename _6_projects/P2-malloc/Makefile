CXX = g++
CXXFLAGS = -std=c++17 -Wall -Wextra -g
SRC_DIR = src
INCLUDE_DIR = include
BUILD_DIR = build
TEST_TARGET = test
LIB_TARGET = libmymalloc.so

SRCS = $(wildcard $(SRC_DIR)/*.cpp)
OBJS = $(patsubst $(SRC_DIR)/%.cpp, $(BUILD_DIR)/%.o, $(SRCS))

all: $(BUILD_DIR) $(TEST_TARGET)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build test binary
$(TEST_TARGET): $(OBJS) tests/test.cpp
	$(CXX) $(CXXFLAGS) -I$(INCLUDE_DIR) -o $@ $^

# Build as shared library (for LD_PRELOAD trick)
lib: $(SRCS)
	$(CXX) $(CXXFLAGS) -shared -fPIC -I$(INCLUDE_DIR) -o $(LIB_TARGET) $^ \
		-Wl,--wrap=malloc -Wl,--wrap=free -Wl,--wrap=realloc -Wl,--wrap=calloc

$(BUILD_DIR)/%.o: $(SRC_DIR)/%.cpp
	$(CXX) $(CXXFLAGS) -I$(INCLUDE_DIR) -c $< -o $@

bench: $(TEST_TARGET)
	@echo "Running benchmark..."
	@time ./$(TEST_TARGET)

clean:
	rm -rf $(BUILD_DIR) $(TEST_TARGET) $(LIB_TARGET)

.PHONY: all clean lib bench
