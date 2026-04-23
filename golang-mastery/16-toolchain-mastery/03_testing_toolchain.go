//go:build ignore

// =============================================================================
// GO TOOLCHAIN 3: TESTING TOOLCHAIN — go test Deep Dive
// =============================================================================
//
// go test is much more than "run my tests". It's a full testing platform:
// unit tests, benchmarks, examples, fuzzing, coverage, subtests, test caching.
//
// RUN: go run 03_testing_toolchain.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== TESTING TOOLCHAIN ===")
	fmt.Println()
	goTestBasics()
	testFlags()
	coverageTool()
	benchmarkTool()
	fuzzTool()
	subtestsAndTableDriven()
	testMainAndFixtures()
	testHelpers()
}

// =============================================================================
// PART 1: go test Basics
// =============================================================================
func goTestBasics() {
	fmt.Println("--- go test BASICS ---")
	// ─── Running tests ───
	// go test              # test current package
	// go test ./...        # test ALL packages recursively
	// go test ./pkg/...    # test packages under pkg/
	// go test -v ./...     # verbose (show each test name + output)
	// go test -run TestFoo # only run tests matching "TestFoo"
	// go test -run TestFoo/subcase  # run specific subtest
	// go test -count=1     # bypass test cache (force re-run)
	// go test -short       # skip long tests (check testing.Short())
	// go test -timeout 30s # override default 10m timeout
	// go test -parallel 4  # max parallel tests
	//
	// ─── Test file naming ───
	// foo.go      → source code
	// foo_test.go → tests for foo.go (same package)
	//
	// Test files are NEVER compiled into your binary.
	// They only exist during `go test`.
	//
	// ─── Test function signatures ───
	// func TestXxx(t *testing.T)      — unit test (Xxx starts with uppercase)
	// func BenchmarkXxx(b *testing.B) — benchmark
	// func FuzzXxx(f *testing.F)      — fuzz test (Go 1.18+)
	// func ExampleXxx()               — example (verified by go test!)
	// func TestMain(m *testing.M)     — setup/teardown for whole package
	//
	// ─── testing.T methods ───
	// t.Error(args...)     — log error, continue running
	// t.Errorf(fmt, ...)   — formatted error, continue
	// t.Fatal(args...)     — log error, STOP this test
	// t.Fatalf(fmt, ...)   — formatted error, STOP
	// t.Log(args...)       — log message (only shown with -v)
	// t.Logf(fmt, ...)     — formatted log
	// t.Skip(args...)      — skip this test
	// t.Skipf(fmt, ...)    — formatted skip
	// t.Helper()           — mark as helper (better error locations)
	// t.Parallel()         — allow parallel execution
	// t.Run(name, func)    — run subtest
	// t.Cleanup(func)      — register cleanup (runs after test, LIFO)
	// t.TempDir()          — create temp directory (auto-cleaned)
	// t.Setenv(k, v)       — set env var (auto-restored)
	fmt.Println("  go test ./... → test all packages")
	fmt.Println("  go test -v -run TestXxx → specific test, verbose")
	fmt.Println("  go test -count=1 → bypass cache")
	fmt.Println()
}

// =============================================================================
// PART 2: Test Flags — The Complete List
// =============================================================================
func testFlags() {
	fmt.Println("--- TEST FLAGS ---")
	// ─── Filtering ───
	// -run regexp        Run only matching test functions
	// -bench regexp      Run only matching benchmarks
	// -fuzz regexp       Run only matching fuzz targets
	// -skip regexp       Skip tests matching pattern (Go 1.20+)
	//
	// ─── Execution ───
	// -v                 Verbose output (show all test logs)
	// -count n           Run each test n times (0 = cache, 1 = no cache)
	// -timeout d         Timeout per test binary (default 10m)
	// -short             Set testing.Short() flag
	// -parallel n        Max parallel tests (default GOMAXPROCS)
	// -failfast          Stop on first failure
	// -shuffle on|off|N  Randomize test order (Go 1.17+)
	//
	// ─── Coverage ───
	// -cover             Enable coverage analysis
	// -coverprofile=c.out  Write coverage profile to file
	// -covermode=set|count|atomic  Coverage mode
	// -coverpkg=./...    Which packages to instrument
	//
	// ─── Race detector ───
	// -race              Enable race detector
	//
	// ─── Build ───
	// -tags "integration" Only include files with matching build tags
	// -ldflags            Linker flags (inject vars)
	// -gcflags            Compiler flags
	//
	// ─── Output ───
	// -json              Output results as JSON (for CI parsing)
	// -list regexp       List matching tests without running
	//
	// ─── Benchmark specific ───
	// -benchtime 5s      Run each benchmark for 5 seconds
	// -benchmem          Print memory allocation stats
	// -cpuprofile=cpu.out  Write CPU profile
	// -memprofile=mem.out  Write memory profile
	// -blockprofile=b.out  Write goroutine blocking profile
	// -mutexprofile=m.out  Write mutex contention profile
	//
	// ─── Common combos ───
	// go test -v -run TestLogin -count=1 ./internal/auth/
	// go test -race -cover -coverprofile=coverage.out ./...
	// go test -bench=. -benchmem -count=5 ./...
	// go test -fuzz FuzzParse -fuzztime 30s ./parser/
	// go test -json ./... | gotestfmt   (pretty JSON output)
	fmt.Println("  -run/-bench/-fuzz → filter which tests run")
	fmt.Println("  -race -cover → always use in CI")
	fmt.Println("  -benchtime -benchmem → benchmark details")
	fmt.Println("  -json → machine-readable output")
	fmt.Println()
}

// =============================================================================
// PART 3: Coverage Tool — Measure What's Tested
// =============================================================================
func coverageTool() {
	fmt.Println("--- COVERAGE ---")
	// ─── Generate coverage ───
	// go test -cover ./...
	//   → shows % coverage per package
	//
	// go test -coverprofile=coverage.out ./...
	//   → writes detailed coverage data
	//
	// ─── View coverage ───
	// go tool cover -func=coverage.out
	//   → shows coverage per function:
	//   github.com/you/app/handler.go:25:  CreateUser  85.7%
	//   github.com/you/app/handler.go:50:  DeleteUser  60.0%
	//   total:                              (statements)  72.4%
	//
	// go tool cover -html=coverage.out
	//   → opens browser with color-coded source
	//   GREEN = covered, RED = not covered
	//   This is the most useful view.
	//
	// go tool cover -html=coverage.out -o coverage.html
	//   → save to file instead of opening browser
	//
	// ─── Coverage modes ───
	// -covermode=set     → line covered: yes/no (default)
	// -covermode=count   → how many times each line ran
	// -covermode=atomic  → count, but safe for concurrent tests
	//                       (use with -race or parallel tests)
	//
	// ─── Cover specific packages ───
	// go test -coverpkg=./... ./...
	//   → instrument ALL packages, not just the one being tested
	//   Without -coverpkg: only the tested package is instrumented
	//   This gives you true integration coverage
	//
	// ─── Coverage in CI ───
	// go test -race -coverprofile=coverage.out -coverpkg=./... ./...
	// go tool cover -func=coverage.out | tail -1
	//   → total:  (statements)  78.5%
	//
	// Enforce minimum coverage:
	// COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}' | tr -d '%')
	// if (( $(echo "$COVERAGE < 70" | bc -l) )); then
	//     echo "Coverage $COVERAGE% is below 70%"
	//     exit 1
	// fi
	//
	// ─── Binary coverage (Go 1.20+) ───
	// Build instrumented binary:
	// go build -cover -o myapp ./cmd/server
	// GOCOVERDIR=./coverdata ./myapp   # run it
	// go tool covdata percent -i ./coverdata
	// go tool covdata textfmt -i ./coverdata -o coverage.out
	// Great for integration/E2E test coverage!
	fmt.Println("  go test -coverprofile=c.out → generate profile")
	fmt.Println("  go tool cover -html=c.out → visual coverage")
	fmt.Println("  go tool cover -func=c.out → per-function %")
	fmt.Println("  -coverpkg=./... → instrument all packages")
	fmt.Println()
}

// =============================================================================
// PART 4: Benchmark Tool
// =============================================================================
func benchmarkTool() {
	fmt.Println("--- BENCHMARKS ---")
	// ─── Writing benchmarks ───
	// func BenchmarkFoo(b *testing.B) {
	//     for i := 0; i < b.N; i++ {
	//         Foo()  // the code being benchmarked
	//     }
	// }
	// b.N is automatically adjusted to get stable results.
	// The framework increases b.N until the result is statistically stable.
	//
	// ─── Running benchmarks ───
	// go test -bench=.                # run ALL benchmarks
	// go test -bench=BenchmarkFoo     # specific benchmark
	// go test -bench=. -benchmem      # include memory stats
	// go test -bench=. -count=5       # run 5 times (for benchstat)
	// go test -bench=. -benchtime=5s  # run each for 5 seconds
	// go test -bench=. -benchtime=100x # run exactly 100 iterations
	//
	// ─── Output format ───
	// BenchmarkFoo-8   1000000   1234 ns/op   256 B/op   3 allocs/op
	//   ↑ name+cores   ↑ iters  ↑ time/op    ↑ bytes    ↑ allocations
	//
	// ─── benchstat: compare benchmarks ───
	// go install golang.org/x/perf/cmd/benchstat@latest
	//
	// # Before change:
	// go test -bench=. -count=10 > old.txt
	// # After change:
	// go test -bench=. -count=10 > new.txt
	// benchstat old.txt new.txt
	//
	// Output:
	// name     old time/op  new time/op  delta
	// Foo-8    1.23µs ± 2%  0.89µs ± 1%  -27.6% (p=0.000 n=10+10)
	//
	// ─── testing.B methods ───
	// b.ResetTimer()     Reset timer (after expensive setup)
	// b.StopTimer()      Pause timer
	// b.StartTimer()     Resume timer
	// b.ReportAllocs()   Same as -benchmem flag
	// b.SetBytes(n)      Report throughput (MB/s)
	// b.RunParallel()    Parallel benchmark
	// b.Run(name, func)  Sub-benchmarks
	//
	// ─── Parallel benchmark ───
	// func BenchmarkFoo(b *testing.B) {
	//     b.RunParallel(func(pb *testing.PB) {
	//         for pb.Next() {
	//             Foo()
	//         }
	//     })
	// }
	//
	// ─── Profile from benchmarks ───
	// go test -bench=. -cpuprofile=cpu.out -memprofile=mem.out
	// go tool pprof cpu.out
	// → This is the #1 way to profile Go code!
	fmt.Println("  go test -bench=. -benchmem → run all benchmarks")
	fmt.Println("  benchstat old.txt new.txt → compare results")
	fmt.Println("  -cpuprofile/-memprofile → profile from benchmarks")
	fmt.Println("  b.ResetTimer() → exclude setup from measurement")
	fmt.Println()
}

// =============================================================================
// PART 5: Fuzzing Tool (Go 1.18+)
// =============================================================================
func fuzzTool() {
	fmt.Println("--- FUZZING ---")
	// ─── What is fuzzing? ───
	// Fuzzing = automated testing with random inputs.
	// Go generates random inputs and tries to crash your code.
	// Finds edge cases you'd never think of.
	//
	// ─── Writing a fuzz test ───
	// func FuzzParseJSON(f *testing.F) {
	//     // Add seed corpus (known good inputs)
	//     f.Add([]byte(`{"name":"alice"}`))
	//     f.Add([]byte(`{}`))
	//     f.Add([]byte(`[]`))
	//
	//     f.Fuzz(func(t *testing.T, data []byte) {
	//         // This function is called with random data
	//         var result map[string]any
	//         err := json.Unmarshal(data, &result)
	//         if err != nil {
	//             return // invalid input, that's fine
	//         }
	//         // If it parsed, re-encode should work
	//         _, err = json.Marshal(result)
	//         if err != nil {
	//             t.Fatalf("marshal after unmarshal failed: %v", err)
	//         }
	//     })
	// }
	//
	// ─── Running fuzz tests ───
	// go test -fuzz FuzzParseJSON              # fuzz until stopped (Ctrl+C)
	// go test -fuzz FuzzParseJSON -fuzztime 30s  # fuzz for 30 seconds
	// go test -fuzz FuzzParseJSON -fuzztime 1000x # 1000 iterations
	// go test ./...                             # runs seed corpus only (no fuzzing)
	//
	// ─── Crash corpus ───
	// When fuzzing finds a crash, it saves the input to:
	// testdata/fuzz/FuzzParseJSON/<hash>
	//
	// This file is automatically used as a test case in future runs.
	// Commit it to version control! It's a regression test.
	//
	// ─── Supported seed types ───
	// f.Add() supports: string, []byte, int, int8-64, uint, uint8-64,
	// float32, float64, rune, bool
	// NOT: structs, slices, maps, pointers
	//
	// ─── Fuzzing strategies ───
	// 1. Roundtrip: encode → decode → compare
	// 2. Decode-only: don't crash on any input
	// 3. Differential: two implementations should agree
	// 4. Invariant: output always satisfies some property
	fmt.Println("  go test -fuzz FuzzXxx -fuzztime 30s → fuzz for 30s")
	fmt.Println("  Crashes saved to testdata/fuzz/ (commit them!)")
	fmt.Println("  Seed corpus: f.Add() + testdata/fuzz/")
	fmt.Println("  go test ./... runs seeds without fuzzing")
	fmt.Println()
}

// =============================================================================
// PART 6: Subtests & Table-Driven Tests
// =============================================================================
func subtestsAndTableDriven() {
	fmt.Println("--- SUBTESTS & TABLE-DRIVEN ---")
	// ─── Subtests with t.Run ───
	// func TestMath(t *testing.T) {
	//     t.Run("addition", func(t *testing.T) {
	//         if Add(1, 2) != 3 { t.Error("wrong") }
	//     })
	//     t.Run("subtraction", func(t *testing.T) {
	//         if Sub(5, 3) != 2 { t.Error("wrong") }
	//     })
	// }
	//
	// Run specific subtest:
	// go test -run TestMath/addition
	//
	// ─── Table-driven tests (THE Go pattern) ───
	// func TestAdd(t *testing.T) {
	//     tests := []struct {
	//         name     string
	//         a, b     int
	//         expected int
	//     }{
	//         {"positive", 1, 2, 3},
	//         {"negative", -1, -2, -3},
	//         {"zero", 0, 0, 0},
	//         {"mixed", -1, 1, 0},
	//     }
	//     for _, tt := range tests {
	//         t.Run(tt.name, func(t *testing.T) {
	//             got := Add(tt.a, tt.b)
	//             if got != tt.expected {
	//                 t.Errorf("Add(%d, %d) = %d, want %d",
	//                     tt.a, tt.b, got, tt.expected)
	//             }
	//         })
	//     }
	// }
	//
	// ─── Parallel subtests ───
	// for _, tt := range tests {
	//     t.Run(tt.name, func(t *testing.T) {
	//         t.Parallel()  // runs concurrently with other parallel subtests
	//         got := Slow(tt.input)
	//         if got != tt.expected { t.Errorf(...) }
	//     })
	// }
	fmt.Println("  t.Run(name, func) → subtests")
	fmt.Println("  go test -run TestFoo/subcase → run one subtest")
	fmt.Println("  Table-driven + t.Run = THE Go testing pattern")
	fmt.Println()
}

// =============================================================================
// PART 7: TestMain & Fixtures
// =============================================================================
func testMainAndFixtures() {
	fmt.Println("--- TestMain & FIXTURES ---")
	// ─── TestMain: package-level setup/teardown ───
	// func TestMain(m *testing.M) {
	//     // Setup (runs once before all tests)
	//     db := setupTestDB()
	//     defer db.Close()
	//
	//     // Run tests
	//     code := m.Run()
	//
	//     // Teardown (runs after all tests)
	//     cleanupTestDB(db)
	//
	//     os.Exit(code)
	// }
	//
	// ONE TestMain per package. If present, it controls test execution.
	//
	// ─── t.Cleanup: per-test teardown ───
	// func TestFoo(t *testing.T) {
	//     db := createTestDB(t)
	//     t.Cleanup(func() {
	//         db.Drop()  // runs after TestFoo finishes
	//     })
	//     // test code...
	// }
	// Cleanup runs in LIFO order (like defer).
	// Runs even if test fails or panics.
	//
	// ─── t.TempDir: auto-cleaned temp directory ───
	// func TestFileProcessing(t *testing.T) {
	//     dir := t.TempDir()  // auto-removed after test
	//     path := filepath.Join(dir, "test.txt")
	//     os.WriteFile(path, []byte("data"), 0644)
	//     // test...
	// }
	//
	// ─── t.Setenv: auto-restored env vars (Go 1.17+) ───
	// func TestConfig(t *testing.T) {
	//     t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	//     // env var restored after test
	// }
	//
	// ─── testdata/ directory ───
	// Files in testdata/ are ignored by go build but available to tests.
	// mypackage/
	//   parser.go
	//   parser_test.go
	//   testdata/
	//     input1.json      → accessible as "testdata/input1.json"
	//     expected1.json
	//     golden/          → golden files for snapshot testing
	fmt.Println("  TestMain → package-level setup/teardown")
	fmt.Println("  t.Cleanup → per-test cleanup (LIFO, always runs)")
	fmt.Println("  t.TempDir → auto-cleaned temp directory")
	fmt.Println("  testdata/ → test fixtures (ignored by build)")
	fmt.Println()
}

// =============================================================================
// PART 8: Test Helpers & Patterns
// =============================================================================
func testHelpers() {
	fmt.Println("--- TEST HELPERS ---")
	// ─── t.Helper() — better error locations ───
	// func assertEqual(t *testing.T, got, want int) {
	//     t.Helper()  // error points to CALLER, not this line
	//     if got != want {
	//         t.Errorf("got %d, want %d", got, want)
	//     }
	// }
	//
	// Without t.Helper(): error shows assertEqual line
	// With t.Helper(): error shows the test that called it
	//
	// ─── httptest: HTTP testing ───
	// func TestHandler(t *testing.T) {
	//     req := httptest.NewRequest("GET", "/api/users", nil)
	//     rec := httptest.NewRecorder()
	//     handler.ServeHTTP(rec, req)
	//     if rec.Code != 200 { t.Errorf("status %d", rec.Code) }
	// }
	//
	// // Test with a real server:
	// srv := httptest.NewServer(handler)
	// defer srv.Close()
	// resp, _ := http.Get(srv.URL + "/api/users")
	//
	// ─── iotest: IO testing ───
	// iotest.ErrReader(err)     → Reader that always errors
	// iotest.HalfReader(r)     → reads half the requested bytes
	// iotest.OneByteReader(r)  → reads one byte at a time
	// iotest.TimeoutReader(r)  → first read OK, second times out
	// Great for testing io.Reader implementations!
	//
	// ─── testing/fstest: filesystem testing (Go 1.16+) ───
	// fs := fstest.MapFS{
	//     "config.json": &fstest.MapFile{Data: []byte(`{"port":8080}`)},
	//     "data/users.csv": &fstest.MapFile{Data: []byte("alice,30")},
	// }
	// // Use fs wherever io/fs.FS is accepted
	//
	// ─── External test packages ───
	// parser_test.go with `package parser_test` (note: _test suffix)
	// Can only use exported API → tests the public interface
	// Use for integration-style tests that test from the outside
	//
	// parser_test.go with `package parser` (same package)
	// Can access unexported functions → for unit testing internals
	fmt.Println("  t.Helper() → error points to caller, not helper")
	fmt.Println("  httptest → test HTTP handlers without a server")
	fmt.Println("  iotest → test io.Reader edge cases")
	fmt.Println("  package foo_test → test only public API")
	fmt.Println()
}
