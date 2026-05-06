# Makefile for C++ Learning
# Usage:
#   make lesson1     → compile and run hello.cpp
#   make lesson2     → compile and run types.cpp
#   make all         → compile all lessons
#   make clean       → remove all compiled binaries
#
# To compile any single file:
#   make run FILE=path/to/file.cpp

CXX = g++
CXXFLAGS = -std=c++20 -Wall -Wextra -g -fsanitize=address,undefined
# Note: not using -Werror so intentional demonstrations of compiler warnings work
# When writing your own code, ADD -Werror to catch bugs early!
# -std=c++20    : Use C++20 standard
# -Wall -Wextra : Maximum warnings
# -Werror       : Treat warnings as errors (forces clean code)
# -g            : Debug symbols (for gdb)
# -fsanitize=address,undefined : Catch memory bugs and UB at runtime

.PHONY: all clean lesson1 lesson2 lesson3 lesson4 lesson5 lesson6 lesson7

# Compile and run a specific file
run:
	$(CXX) $(CXXFLAGS) $(FILE) -o /tmp/cpp_out && /tmp/cpp_out

# Individual lessons
lesson1:
	$(CXX) $(CXXFLAGS) 01-foundations/01-compilation/hello.cpp -o /tmp/lesson1 && /tmp/lesson1

lesson1b:
	$(CXX) $(CXXFLAGS) 01-foundations/01-compilation/main.cpp 01-foundations/01-compilation/math.cpp -o /tmp/lesson1b && /tmp/lesson1b

lesson2:
	$(CXX) $(CXXFLAGS) 01-foundations/02-types/types.cpp -o /tmp/lesson2 && /tmp/lesson2

lesson2b:
	$(CXX) $(CXXFLAGS) 01-foundations/02-types/conversions.cpp -o /tmp/lesson2b && /tmp/lesson2b

lesson3:
	$(CXX) $(CXXFLAGS) 01-foundations/03-operators/operators.cpp -o /tmp/lesson3 && /tmp/lesson3

lesson4:
	$(CXX) $(CXXFLAGS) 01-foundations/04-control-flow/control_flow.cpp -o /tmp/lesson4 && /tmp/lesson4

lesson5:
	$(CXX) $(CXXFLAGS) 01-foundations/05-functions/functions.cpp -o /tmp/lesson5 && /tmp/lesson5

lesson6:
	$(CXX) $(CXXFLAGS) 01-foundations/06-arrays-strings/arrays_strings.cpp -o /tmp/lesson6 && /tmp/lesson6

lesson7:
	$(CXX) $(CXXFLAGS) 01-foundations/07-pointers-references/pointers.cpp -o /tmp/lesson7 && /tmp/lesson7

all: lesson1 lesson1b lesson2 lesson2b lesson3 lesson4 lesson5 lesson6 lesson7

clean:
	rm -f /tmp/lesson* /tmp/cpp_out
	find . -name "*.o" -delete
	find . -type f -executable -not -name "*.cpp" -not -name "*.h" -not -name "Makefile" -not -name "*.md" -delete
