//go:build ignore

// =============================================================================
// GO TOOLCHAIN 4: pprof — PROFILING Go Programs Like a God
// =============================================================================
//
// pprof is Go's built-in profiler. It tells you EXACTLY where your program
// spends time and memory. No guessing. Data-driven optimization.
//
// Three ways to use pprof:
// 1. From benchmarks: go test -bench=. -cpuprofile=cpu.out
// 2. From running server: import _ "net/http/pprof"
// 3. Programmatically: runtime/pprof package
//
// RUN: go run 04_pprof_profiling.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== pprof PROFILING ===")
	fmt.Println()
	cpuProfiling()
	memoryProfiling()
	httpPprof()
	pprofCommands()
	flameGraphs()
	goroutineProfile()
	blockAndMutexProfile()
	profilingWorkflow()
}

// =============================================================================
// PART 1: CPU Profiling
// =============================================================================
func cpuProfiling() {
	fmt.Println("--- CPU PROFILING ---")
	// ─── From benchmarks (EASIEST) ───
	// go test -bench=BenchmarkHot -cpuprofile=cpu.out ./...
	// go tool pprof cpu.out
	//
	// ─── Programmatic ───
	// import "runtime/pprof"
	//
	// f, _ := os.Create("cpu.out")
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	// // ... run code ...
	//
	// ─── What CPU profiling does ───
	// Samples the call stack 100 times per second (default).
	// Each sample records which function is currently executing.
	// After profiling: "function X appeared in 40% of samples"
	// = function X uses ~40% of CPU time.
	//
	// ─── Interpreting CPU profiles ───
	// flat: time spent IN the function itself (not callees)
	// cum:  time spent in the function INCLUDING its callees
	//
	// Example:
	//   func A() {         ← flat = 0, cum = 100ms
	//       B()            ← flat = 30ms, cum = 80ms
	//       doWork(20ms)   ← flat = 20ms, cum = 20ms
	//   }
	//   func B() {
	//       C()            ← flat = 50ms, cum = 50ms
	//       doWork(30ms)   ← flat = 30ms
	//   }
	//
	// flat% = what to optimize FIRST (most time in this specific function)
	// cum%  = hottest path (optimize the chain)
	//
	// ─── Sampling rate ───
	// runtime.SetCPUProfileRate(500)  // samples per second (default 100)
	// Higher rate = more detail, more overhead
	fmt.Println("  go test -bench=. -cpuprofile=cpu.out")
	fmt.Println("  go tool pprof cpu.out → interactive analysis")
	fmt.Println("  flat = time in function, cum = including callees")
	fmt.Println()
}

// =============================================================================
// PART 2: Memory (Heap) Profiling
// =============================================================================
func memoryProfiling() {
	fmt.Println("--- MEMORY PROFILING ---")
	// ─── From benchmarks ───
	// go test -bench=. -memprofile=mem.out ./...
	// go tool pprof mem.out
	//
	// ─── Programmatic ───
	// f, _ := os.Create("mem.out")
	// runtime.GC()  // get up-to-date stats
	// pprof.WriteHeapProfile(f)
	//
	// ─── What memory profiling shows ───
	// - alloc_objects: number of allocations (total)
	// - alloc_space: bytes allocated (total, including freed)
	// - inuse_objects: currently live objects
	// - inuse_space: currently live bytes (what's on the heap NOW)
	//
	// ─── Choosing the right view ───
	// go tool pprof -alloc_space mem.out   → total bytes allocated
	// go tool pprof -alloc_objects mem.out → total allocation count
	// go tool pprof -inuse_space mem.out   → current heap usage
	// go tool pprof -inuse_objects mem.out → current live objects
	//
	// Use -alloc_space to find code that allocates a LOT (GC pressure).
	// Use -inuse_space to find memory leaks.
	//
	// ─── Memory profiling rate ───
	// runtime.MemProfileRate = 1  // record every allocation
	// Default: 1 sample per 512KB allocated
	// Set to 1 for precision (slower, more overhead)
	//
	// ─── Common memory problems ───
	// 1. String concatenation in loops → use strings.Builder
	// 2. Slice growing → pre-allocate with make([]T, 0, estimatedSize)
	// 3. Closures capturing variables → escape to heap
	// 4. Interface boxing → value types become heap-allocated
	// 5. Map growth → maps never shrink in Go
	fmt.Println("  go test -bench=. -memprofile=mem.out")
	fmt.Println("  -alloc_space → total allocations (GC pressure)")
	fmt.Println("  -inuse_space → current heap (memory leaks)")
	fmt.Println()
}

// =============================================================================
// PART 3: HTTP pprof — Profile Running Servers
// =============================================================================
func httpPprof() {
	fmt.Println("--- HTTP pprof ---")
	// ─── Enable with one import ───
	// import _ "net/http/pprof"
	//
	// // If you already have an http server:
	// // pprof handlers are registered on DefaultServeMux automatically
	// http.ListenAndServe(":6060", nil)
	//
	// // For a separate debug server (RECOMMENDED in production):
	// go func() {
	//     log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	//
	// ─── Available endpoints ───
	// http://localhost:6060/debug/pprof/           → index page
	// http://localhost:6060/debug/pprof/profile    → CPU profile (30s default)
	// http://localhost:6060/debug/pprof/heap       → heap profile
	// http://localhost:6060/debug/pprof/goroutine  → goroutine stacks
	// http://localhost:6060/debug/pprof/allocs     → allocation profile
	// http://localhost:6060/debug/pprof/block      → blocking profile
	// http://localhost:6060/debug/pprof/mutex      → mutex contention
	// http://localhost:6060/debug/pprof/threadcreate → OS thread creation
	// http://localhost:6060/debug/pprof/trace      → execution trace
	//
	// ─── Grab profiles from CLI ───
	// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
	// go tool pprof http://localhost:6060/debug/pprof/heap
	// go tool pprof http://localhost:6060/debug/pprof/goroutine
	//
	// ─── SECURITY WARNING ───
	// NEVER expose pprof on a public port!
	// It leaks internal state and can consume resources.
	// Always bind to localhost or put behind auth.
	// In production: use a separate internal port or admin API.
	fmt.Println("  import _ \"net/http/pprof\" → enable on running server")
	fmt.Println("  /debug/pprof/profile → CPU, /debug/pprof/heap → memory")
	fmt.Println("  NEVER expose on public port (security risk)")
	fmt.Println()
}

// =============================================================================
// PART 4: pprof Interactive Commands
// =============================================================================
func pprofCommands() {
	fmt.Println("--- pprof COMMANDS ---")
	// ─── Enter interactive mode ───
	// go tool pprof cpu.out
	// (pprof) _
	//
	// ─── Essential commands ───
	// top              → top functions by flat time
	// top 20           → top 20 functions
	// top -cum         → top by cumulative time
	// list funcName    → show source with per-line cost
	// peek funcName    → show callers and callees
	// web              → open flame graph in browser (needs graphviz)
	// svg              → save flame graph as SVG
	// png              → save as PNG
	// tree             → show call tree
	// traces           → show full call traces
	// disasm funcName  → show assembly with per-instruction cost
	// focus=funcName   → only show paths through funcName
	// ignore=funcName  → hide funcName from output
	// tags             → show available sample tags
	// quit / exit      → leave pprof
	//
	// ─── Non-interactive mode ───
	// go tool pprof -top cpu.out
	// go tool pprof -list=funcName cpu.out
	// go tool pprof -svg cpu.out > profile.svg
	// go tool pprof -http=:8080 cpu.out   → web UI (best!)
	//
	// ─── Web UI (BEST way to use pprof) ───
	// go tool pprof -http=:8080 cpu.out
	// Opens browser with:
	// - Interactive flame graph
	// - Graph view (call graph with sizes)
	// - Top view (sortable table)
	// - Source view (annotated code)
	// - Disassembly view
	// Navigate between views. Click nodes to focus.
	//
	// ─── Comparing profiles ───
	// go tool pprof -base=old.out new.out
	// Shows the DIFFERENCE between two profiles.
	// "This function uses 200ms MORE in the new version"
	// Essential for regression detection.
	fmt.Println("  go tool pprof -http=:8080 cpu.out → web UI (best!)")
	fmt.Println("  top, list, web, peek → key commands")
	fmt.Println("  -base=old.out new.out → compare profiles")
	fmt.Println()
}

// =============================================================================
// PART 5: Flame Graphs
// =============================================================================
func flameGraphs() {
	fmt.Println("--- FLAME GRAPHS ---")
	// ─── What is a flame graph? ───
	// Visual representation of profiling data.
	// X-axis: % of total samples (wider = more CPU)
	// Y-axis: call stack depth (bottom = entry point, top = leaf)
	//
	//   ┌──────────────────────────────────────────────────────────────┐
	//   │                        main.main                            │
	//   ├──────────────────────────────┬───────────────────────────────┤
	//   │     handler.ServeHTTP        │        db.Query              │
	//   ├──────────────┬───────────────┤───────────┬───────────────────┤
	//   │  json.Marshal│  http.Write   │ sql.Exec  │  net.Read        │
	//   └──────────────┴───────────────┴───────────┴───────────────────┘
	//
	// Wide boxes = hot functions (optimize these!)
	// Narrow boxes = rarely hit
	//
	// ─── Generate flame graph ───
	// go tool pprof -http=:8080 cpu.out
	// Click "Flame Graph" view in the web UI
	//
	// ─── Or with the `web` command ───
	// (pprof) web
	// Requires graphviz: sudo apt install graphviz
	//
	// ─── Reading flame graphs ───
	// 1. Look for WIDE bars → most CPU time
	// 2. Look for TALL stacks → deep call chains (maybe simplify?)
	// 3. Click to zoom into a subtree
	// 4. Compare before/after optimization
	//
	// ─── Differential flame graphs ───
	// go tool pprof -http=:8080 -diff_base=old.out new.out
	// RED = slower than before
	// GREEN = faster than before
	// BLUE = unchanged
	fmt.Println("  Flame graph: wider = more CPU time")
	fmt.Println("  go tool pprof -http=:8080 → click Flame Graph")
	fmt.Println("  -diff_base=old.out → differential flame graph")
	fmt.Println()
}

// =============================================================================
// PART 6: Goroutine Profiling
// =============================================================================
func goroutineProfile() {
	fmt.Println("--- GOROUTINE PROFILING ---")
	// ─── Why profile goroutines? ───
	// Goroutine leaks = goroutines that never terminate
	// Each leaked goroutine holds memory (min 2KB stack + any heap refs)
	// Thousands of leaked goroutines = OOM
	//
	// ─── Get goroutine dump ───
	// From HTTP pprof:
	// curl http://localhost:6060/debug/pprof/goroutine?debug=1
	// → one line per goroutine group
	//
	// curl http://localhost:6060/debug/pprof/goroutine?debug=2
	// → FULL stack trace for every goroutine
	//
	// With pprof tool:
	// go tool pprof http://localhost:6060/debug/pprof/goroutine
	// (pprof) top  → which functions have the most goroutines stuck
	// (pprof) tree → call tree of blocked goroutines
	//
	// ─── Detecting leaks ───
	// Monitor runtime.NumGoroutine() over time.
	// If it grows without bound → you have a leak.
	//
	// In tests, use go.uber.org/goleak:
	// func TestMain(m *testing.M) {
	//     goleak.VerifyTestMain(m)
	// }
	//
	// ─── Common leak patterns ───
	// 1. Channel send with no receiver
	// 2. Infinite loop with no exit condition
	// 3. Blocked on mutex that's never unlocked
	// 4. HTTP request with no timeout (context!)
	// 5. Goroutine waiting on channel from cancelled context
	//
	// ─── SIGQUIT trick ───
	// Send SIGQUIT to a running Go program:
	// kill -QUIT <pid>
	// It dumps ALL goroutine stacks to stderr and exits.
	// Works even if pprof isn't enabled!
	fmt.Println("  /debug/pprof/goroutine?debug=2 → full stacks")
	fmt.Println("  runtime.NumGoroutine() → monitor for leaks")
	fmt.Println("  goleak.VerifyTestMain → catch leaks in tests")
	fmt.Println("  kill -QUIT <pid> → dump all goroutine stacks")
	fmt.Println()
}

// =============================================================================
// PART 7: Block & Mutex Profiling
// =============================================================================
func blockAndMutexProfile() {
	fmt.Println("--- BLOCK & MUTEX PROFILING ---")
	// ─── Block profiling ───
	// Shows where goroutines BLOCK waiting:
	// - Channel sends/receives
	// - sync.Mutex.Lock()
	// - sync.WaitGroup.Wait()
	// - select statement
	// - time.Sleep, time.After
	//
	// Enable:
	// runtime.SetBlockProfileRate(1)  // record every block event
	//   rate N: record 1 sample per N nanoseconds of blocking
	//   rate 0: disable (default)
	//   rate 1: record everything
	//
	// From benchmarks:
	// go test -bench=. -blockprofile=block.out
	// go tool pprof block.out
	//
	// From HTTP:
	// curl http://localhost:6060/debug/pprof/block
	//
	// ─── Mutex profiling ───
	// Shows mutex CONTENTION: which mutexes are fought over.
	//
	// Enable:
	// runtime.SetMutexProfileFraction(1)  // record every contention
	//   fraction N: record 1/N of contention events
	//   fraction 0: disable (default)
	//
	// From benchmarks:
	// go test -bench=. -mutexprofile=mutex.out
	// go tool pprof mutex.out
	//
	// From HTTP:
	// curl http://localhost:6060/debug/pprof/mutex
	//
	// ─── When to use ───
	// Block profile: "why is my program slow?" → goroutines waiting
	// Mutex profile: "which lock is the bottleneck?" → contention
	// CPU profile:   "which code burns CPU?" → computation
	// Memory profile: "what allocates?" → heap pressure
	fmt.Println("  Block: where goroutines wait (channels, locks)")
	fmt.Println("  Mutex: which locks have contention")
	fmt.Println("  runtime.SetBlockProfileRate(1) to enable")
	fmt.Println("  runtime.SetMutexProfileFraction(1) to enable")
	fmt.Println()
}

// =============================================================================
// PART 8: Profiling Workflow — How to Actually Optimize
// =============================================================================
func profilingWorkflow() {
	fmt.Println("--- PROFILING WORKFLOW ---")
	// The optimization process:
	//
	// 1. BENCHMARK FIRST
	//    Write a benchmark for the hot path.
	//    go test -bench=BenchmarkHot -benchmem -count=5 > before.txt
	//
	// 2. PROFILE
	//    go test -bench=BenchmarkHot -cpuprofile=cpu.out -memprofile=mem.out
	//    go tool pprof -http=:8080 cpu.out
	//    → Find the hot function in flame graph
	//
	// 3. ANALYZE
	//    (pprof) list HotFunction
	//    → See per-line CPU/memory cost
	//    → Identify the expensive operation
	//
	// 4. OPTIMIZE
	//    Make ONE change at a time.
	//    Common fixes:
	//    - Pre-allocate slices (make([]T, 0, n))
	//    - Use strings.Builder instead of + concatenation
	//    - Avoid interface boxing in hot paths
	//    - Cache regex compilation (var re = regexp.MustCompile(...))
	//    - Use sync.Pool for frequently allocated objects
	//    - Reduce allocations (reuse buffers)
	//
	// 5. VERIFY
	//    go test -bench=BenchmarkHot -benchmem -count=5 > after.txt
	//    benchstat before.txt after.txt
	//    → Confirm improvement with statistical significance
	//
	// 6. REPEAT
	//    Profile again. The bottleneck may have shifted.
	//    Stop when performance meets requirements.
	//
	// ─── RULES ───
	// - NEVER optimize without profiling first
	// - NEVER optimize without benchmarks to prove improvement
	// - Optimize the #1 bottleneck, then re-profile
	// - "Premature optimization is the root of all evil" — Knuth
	// - But: optimization WITH data is engineering
	fmt.Println("  1. Benchmark → 2. Profile → 3. Analyze")
	fmt.Println("  4. Optimize ONE thing → 5. Verify with benchstat")
	fmt.Println("  NEVER optimize without profiling data")
	fmt.Println()
}
