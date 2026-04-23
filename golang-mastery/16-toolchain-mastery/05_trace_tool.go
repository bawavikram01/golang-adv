//go:build ignore

// =============================================================================
// GO TOOLCHAIN 5: go tool trace — Execution Tracing
// =============================================================================
//
// pprof shows WHERE time is spent. trace shows WHEN things happen.
// trace gives you a timeline: goroutine scheduling, GC pauses,
// syscalls, network I/O — all on a time axis.
//
// RUN: go run 05_trace_tool.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== EXECUTION TRACING ===")
	fmt.Println()
	traceBasics()
	traceFromTests()
	traceProgrammatic()
	traceWebUI()
	traceAnalysis()
	traceVsPprof()
}

// =============================================================================
// PART 1: Trace Basics
// =============================================================================
func traceBasics() {
	fmt.Println("--- TRACE BASICS ---")
	// ─── What does trace show? ───
	// A timeline of EVERYTHING the Go runtime does:
	// - Goroutine creation, blocking, unblocking, scheduling
	// - GC start/stop, STW (stop-the-world) pauses
	// - Syscalls (file I/O, network)
	// - Which goroutine runs on which processor (P)
	// - Channel operations, mutex contention
	// - Network I/O latency
	//
	// ─── When to use trace vs pprof ───
	// pprof: "Which function is slow?" (statistical, aggregate)
	// trace: "What happened at time X?" (exact, timeline)
	//
	// Use trace when:
	// - Debugging latency spikes (GC pauses? scheduling delays?)
	// - Understanding goroutine interaction
	// - Finding why parallel code isn't actually parallel
	// - Diagnosing GC impact
	// - Understanding request lifecycle
	//
	// ─── Trace overhead ───
	// ~25% CPU overhead, significant memory
	// Use for short periods (seconds to minutes)
	// NOT for continuous production monitoring
	fmt.Println("  trace = timeline of runtime events")
	fmt.Println("  pprof = statistical sampling (which function)")
	fmt.Println("  trace = deterministic recording (when/what happened)")
	fmt.Println()
}

// =============================================================================
// PART 2: Trace from Tests
// =============================================================================
func traceFromTests() {
	fmt.Println("--- TRACE FROM TESTS ---")
	// ─── Easiest way (from test/benchmark) ───
	// go test -trace=trace.out ./...
	// go tool trace trace.out
	//   → opens browser with interactive timeline
	//
	// ─── From benchmark ───
	// go test -bench=BenchmarkHot -trace=trace.out
	// go tool trace trace.out
	//
	// ─── From HTTP pprof ───
	// curl -o trace.out 'http://localhost:6060/debug/pprof/trace?seconds=5'
	// go tool trace trace.out
	//
	// The HTTP endpoint records a 5-second trace of your running server.
	// Great for diagnosing production latency.
	fmt.Println("  go test -trace=trace.out → collect trace")
	fmt.Println("  go tool trace trace.out → open timeline in browser")
	fmt.Println("  curl .../debug/pprof/trace?seconds=5 → production trace")
	fmt.Println()
}

// =============================================================================
// PART 3: Programmatic Tracing
// =============================================================================
func traceProgrammatic() {
	fmt.Println("--- PROGRAMMATIC TRACING ---")
	// import "runtime/trace"
	//
	// f, _ := os.Create("trace.out")
	// trace.Start(f)
	// defer trace.Stop()
	//
	// // Your code here — all goroutine events are recorded
	//
	// ─── User annotations (Go 1.11+) ───
	// Add custom regions and tasks to the trace:
	//
	// ctx, task := trace.NewTask(ctx, "processOrder")
	// defer task.End()
	//
	// trace.WithRegion(ctx, "validateInput", func() {
	//     // this region appears in the trace timeline
	//     validate(order)
	// })
	//
	// trace.WithRegion(ctx, "saveToDatabase", func() {
	//     db.Save(order)
	// })
	//
	// trace.Log(ctx, "orderID", order.ID)
	//
	// Tasks and regions appear in the trace UI as named spans.
	// Essential for understanding request flow in servers.
	//
	// ─── Regions in the trace viewer ───
	// User-defined tasks:
	//   processOrder ──┬── validateInput (2ms)
	//                  ├── saveToDatabase (15ms)
	//                  └── sendNotification (3ms)
	fmt.Println("  trace.Start/Stop → record trace programmatically")
	fmt.Println("  trace.NewTask → create named task span")
	fmt.Println("  trace.WithRegion → annotate code sections")
	fmt.Println("  trace.Log → add key-value to trace")
	fmt.Println()
}

// =============================================================================
// PART 4: Trace Web UI
// =============================================================================
func traceWebUI() {
	fmt.Println("--- TRACE WEB UI ---")
	// go tool trace trace.out
	// Opens browser with these views:
	//
	// ─── 1. Goroutine analysis ───
	// Shows all goroutines grouped by function.
	// For each group: count, execution time, scheduling wait.
	// Click to see individual goroutine timelines.
	//
	// ─── 2. Scheduler latency ───
	// How long goroutines wait to be scheduled.
	// If high → too many goroutines or GOMAXPROCS too low.
	//
	// ─── 3. Network/Sync blocking ───
	// Time goroutines spend blocked on:
	// - Network I/O
	// - Channel operations
	// - Mutex/sync primitives
	//
	// ─── 4. Syscall blocking ───
	// Time in system calls (file I/O, etc.)
	// High syscall time → consider async I/O or batching.
	//
	// ─── 5. GC pauses ───
	// Shows every GC cycle:
	// - STW pause duration
	// - Concurrent GC work
	// - Heap size at each GC
	//
	// ─── 6. View trace (main timeline) ───
	// Interactive timeline showing:
	// - Each processor (P) as a row
	// - Which goroutine runs when
	// - GC events
	// - Syscalls
	// - User regions/tasks
	//
	// Navigation:
	// W/S = zoom in/out
	// A/D = pan left/right
	// Click goroutine → see its full lifecycle
	// Click GC event → see pause details
	//
	// ─── 7. User tasks view ───
	// If you used trace.NewTask/WithRegion:
	// Shows task durations, regions, logs.
	// Filter by task type. Sort by duration.
	fmt.Println("  Views: goroutines, scheduler, network, GC, timeline")
	fmt.Println("  W/S zoom, A/D pan in timeline view")
	fmt.Println("  User tasks: trace.NewTask + WithRegion")
	fmt.Println()
}

// =============================================================================
// PART 5: Trace Analysis — What to Look For
// =============================================================================
func traceAnalysis() {
	fmt.Println("--- TRACE ANALYSIS ---")
	// ─── 1. GC impact ───
	// Look at STW (Stop-The-World) pauses.
	// Go 1.19+: STW should be < 1ms for most workloads.
	// If GC is frequent → reduce allocation rate.
	// If STW is long → check GOGC setting, reduce heap.
	//
	// ─── 2. Scheduling delays ───
	// Goroutine states:
	//   Running   → executing on a P
	//   Runnable  → ready but waiting for a P (scheduling delay)
	//   Waiting   → blocked (channel, mutex, syscall, timer)
	//
	// High runnable time → increase GOMAXPROCS or reduce goroutines.
	// High waiting time  → I/O bound or contention.
	//
	// ─── 3. Parallelism ───
	// In timeline: if only 1 P is active → your code isn't parallel.
	// Look for goroutines that could run simultaneously but don't.
	// Causes: shared mutex, sequential channel operations, single producer.
	//
	// ─── 4. Channel contention ───
	// Goroutines blocked on channel operations.
	// If many goroutines wait on the same channel → bottleneck.
	// Fix: buffer the channel, add more producers/consumers.
	//
	// ─── 5. Syscall impact ───
	// Goroutines in syscalls block OS threads.
	// Too many concurrent syscalls → thread explosion.
	// Fix: limit concurrent I/O (semaphore pattern).
	//
	// ─── 6. Long-running goroutines ───
	// Goroutines that run without yielding block other goroutines.
	// Since Go 1.14: runtime preempts at async safe points.
	// Tight loops with no function calls may still block.
	fmt.Println("  Check: GC pauses, scheduling delays, parallelism")
	fmt.Println("  High runnable time → increase GOMAXPROCS")
	fmt.Println("  High waiting time → I/O bound or contention")
	fmt.Println("  Only 1 P active → code isn't actually parallel")
	fmt.Println()
}

// =============================================================================
// PART 6: Trace vs pprof — When to Use Which
// =============================================================================
func traceVsPprof() {
	fmt.Println("--- TRACE vs PPROF ---")
	// ┌─────────────┬────────────────────┬────────────────────┐
	// │             │ pprof              │ trace              │
	// ├─────────────┼────────────────────┼────────────────────┤
	// │ Data type   │ Statistical sample │ Event recording    │
	// │ Shows       │ WHERE time spent   │ WHEN things happen │
	// │ Granularity │ Per function       │ Per nanosecond     │
	// │ Overhead    │ ~5%                │ ~25%               │
	// │ Duration    │ Minutes to hours   │ Seconds to minutes │
	// │ Best for    │ CPU/memory hotspot │ Latency/scheduling │
	// │ GC detail   │ Total GC time      │ Each GC event      │
	// │ Goroutines  │ Count by state     │ Full lifecycle      │
	// │ Output      │ Flame graph        │ Timeline            │
	// └─────────────┴────────────────────┴────────────────────┘
	//
	// ─── Decision guide ───
	// "My app is using too much CPU" → pprof CPU profile
	// "My app is using too much memory" → pprof heap profile
	// "My app has random latency spikes" → trace
	// "My goroutines seem stuck" → trace + goroutine profile
	// "GC is taking too long" → trace (see each GC)
	// "Parallel code runs serially" → trace (see P utilization)
	// "Which function to optimize?" → pprof
	// "Why did this request take 500ms?" → trace with user tasks
	fmt.Println("  pprof: WHERE (which function is hot)")
	fmt.Println("  trace: WHEN (timeline of events)")
	fmt.Println("  Latency spikes / scheduling → use trace")
	fmt.Println("  CPU/memory hotspot → use pprof")
	fmt.Println()
}
