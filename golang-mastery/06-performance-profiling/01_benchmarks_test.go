// =============================================================================
// LESSON 6: PERFORMANCE PROFILING & OPTIMIZATION
// =============================================================================
//
// Go has world-class built-in profiling tools. This lesson covers:
//   - CPU profiling (where is time spent?)
//   - Memory profiling (where are allocations happening?)
//   - Benchmarking (measuring performance changes)
//   - Tracing (understanding goroutine scheduling)
//   - pprof (production profiling via HTTP)
//
// KEY TOOLS:
//   go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
//   go tool pprof cpu.prof
//   go tool trace trace.out
//   go test -bench=. -count=10 | benchstat old.txt new.txt
// =============================================================================

package main

import (
	"fmt"
	"strings"
	"testing"
)

// =============================================================================
// PART 1: Writing Effective Benchmarks
// =============================================================================

// RULE 1: Use b.N — the framework controls iteration count
// RULE 2: Use b.ResetTimer() after expensive setup
// RULE 3: Prevent compiler optimization with global sinks
// RULE 4: Use b.ReportAllocs() to track allocations

// Global sink prevents compiler from optimizing away results
var globalSink interface{}

// --- String Concatenation Benchmark Comparison ---

func ConcatPlus(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s // O(n²) — copies entire string each iteration
	}
	return result
}

func ConcatBuilder(strs []string) string {
	var b strings.Builder
	for _, s := range strs {
		b.WriteString(s) // O(n) — amortized doubling
	}
	return b.String()
}

func ConcatJoin(strs []string) string {
	return strings.Join(strs, "") // O(n) — pre-calculates total length
}

func ConcatPrealloc(strs []string) string {
	total := 0
	for _, s := range strs {
		total += len(s)
	}
	var b strings.Builder
	b.Grow(total) // single allocation
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String()
}

// Benchmark functions (run with: go test -bench=BenchmarkConcat -benchmem)
func BenchmarkConcatPlus(b *testing.B) {
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "hello"
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		globalSink = ConcatPlus(strs)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "hello"
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		globalSink = ConcatBuilder(strs)
	}
}

func BenchmarkConcatJoin(b *testing.B) {
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "hello"
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		globalSink = ConcatJoin(strs)
	}
}

func BenchmarkConcatPrealloc(b *testing.B) {
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "hello"
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		globalSink = ConcatPrealloc(strs)
	}
}

// =============================================================================
// PART 2: Memory Optimization Patterns
// =============================================================================

// --- Slice pre-allocation ---
func BenchmarkSliceGrow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]int, 0) // unknown capacity
		for j := 0; j < 10000; j++ {
			s = append(s, j)
		}
		globalSink = s
	}
}

func BenchmarkSlicePrealloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s := make([]int, 0, 10000) // known capacity
		for j := 0; j < 10000; j++ {
			s = append(s, j)
		}
		globalSink = s
	}
}

// --- Map pre-sizing ---
func BenchmarkMapGrow(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[int]int)
		for j := 0; j < 10000; j++ {
			m[j] = j
		}
		globalSink = m
	}
}

func BenchmarkMapPresized(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[int]int, 10000)
		for j := 0; j < 10000; j++ {
			m[j] = j
		}
		globalSink = m
	}
}

// =============================================================================
// PART 3: Sub-benchmarks and Table-Driven Benchmarks
// =============================================================================

func BenchmarkSliceLookup(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		data := make([]int, size)
		for i := range data {
			data[i] = i
		}
		target := size - 1

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				for _, v := range data {
					if v == target {
						globalSink = v
						break
					}
				}
			}
		})
	}
}

func BenchmarkMapLookup(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		data := make(map[int]int, size)
		for i := 0; i < size; i++ {
			data[i] = i
		}
		target := size - 1

		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				globalSink = data[target]
			}
		})
	}
}

// =============================================================================
// PART 4: pprof HTTP Endpoint (for production profiling)
// =============================================================================
//
// Add this to your production server:
//
//   import _ "net/http/pprof"
//
//   go func() {
//       log.Println(http.ListenAndServe("localhost:6060", nil))
//   }()
//
// Then:
//   go tool pprof http://localhost:6060/debug/pprof/heap
//   go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
//   go tool pprof http://localhost:6060/debug/pprof/goroutine
//   curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out
//   go tool trace trace.out

// =============================================================================
// PART 5: Compiler Optimizations You Should Know
// =============================================================================

// 1. BCE (Bounds Check Elimination)
// The compiler can eliminate array bounds checks when it can prove safety.
func sumBCE(data []int) int {
	total := 0
	// Compiler sees len(data) in loop bound → eliminates bounds check
	for i := 0; i < len(data); i++ {
		total += data[i] // no bounds check needed
	}
	return total
}

// Force bounds check elimination with a hint:
func sumBCEHint(data []int) int {
	if len(data) < 4 {
		return 0
	}
	_ = data[3] // bounds check here, eliminated below
	return data[0] + data[1] + data[2] + data[3]
}

// 2. Dead Code Elimination
func deadCode() int {
	x := computeExpensive()
	if false { // compiler removes this entire block
		_ = x
	}
	return 42
}

func computeExpensive() int { return 1 }

// 3. Check what optimizations the compiler applies:
//    go build -gcflags="-d=ssa/check_bce/debug=1" .  # show bounds checks
//    go build -gcflags="-S" .                         # show assembly
//    go build -gcflags="-m" .                         # show escape analysis

func main() {
	fmt.Println("=== Performance Profiling & Optimization ===")
	fmt.Println()
	fmt.Println("This file is designed to be run as benchmarks:")
	fmt.Println()
	fmt.Println("  # Run all benchmarks with memory stats:")
	fmt.Println("  go test -bench=. -benchmem -run=^$ -v ./06-performance-profiling/")
	fmt.Println()
	fmt.Println("  # Generate CPU profile:")
	fmt.Println("  go test -bench=BenchmarkConcat -cpuprofile=cpu.prof ./06-performance-profiling/")
	fmt.Println("  go tool pprof -http=:8080 cpu.prof")
	fmt.Println()
	fmt.Println("  # Generate memory profile:")
	fmt.Println("  go test -bench=BenchmarkSlice -memprofile=mem.prof ./06-performance-profiling/")
	fmt.Println("  go tool pprof -http=:8080 mem.prof")
	fmt.Println()
	fmt.Println("  # Compare benchmarks (install: go install golang.org/x/perf/cmd/benchstat@latest):")
	fmt.Println("  go test -bench=. -count=10 > old.txt")
	fmt.Println("  # ... make changes ...")
	fmt.Println("  go test -bench=. -count=10 > new.txt")
	fmt.Println("  benchstat old.txt new.txt")
	fmt.Println()
	fmt.Println("  # Check bounds check elimination:")
	fmt.Println("  go build -gcflags='-d=ssa/check_bce/debug=1' .")

	_ = sumBCE([]int{1, 2, 3})
	_ = sumBCEHint([]int{1, 2, 3, 4})
	_ = deadCode()
}
