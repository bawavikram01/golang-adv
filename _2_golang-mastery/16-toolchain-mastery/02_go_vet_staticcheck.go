//go:build ignore
//go:build ignore

// =============================================================================
// GO TOOLCHAIN 2: go vet, staticcheck & CODE ANALYSIS
// =============================================================================
//
// go vet: official analyzer — catches bugs the compiler misses
// staticcheck: the gold standard third-party analyzer
// go fmt / goimports: code formatting
//
// These tools catch bugs that tests miss. Run them on EVERY commit.
//
// RUN: go run 02_go_vet_staticcheck.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== CODE ANALYSIS TOOLS ===")
	fmt.Println()
	goVet()
	goFmtAndImports()
	staticcheckTool()
	goVetAnalyzers()
	shadowDetection()
	customAnalyzers()
}

// =============================================================================
// PART 1: go vet — The Official Bug Finder
// =============================================================================
func goVet() {
	fmt.Println("--- go vet ---")
	// ─── What is go vet? ───
	// A static analysis tool bundled with Go.
	// Finds bugs that compile but are almost certainly wrong.
	// It does NOT find ALL bugs — it has ZERO false positives philosophy.
	// If vet reports something, it's a real bug.
	//
	// ─── Running go vet ───
	// go vet ./...                 # vet all packages
	// go vet main.go              # vet single file
	// go vet -json ./...          # JSON output (for CI)
	// go vet -v ./...             # verbose (show analyzed packages)
	//
	// ─── What go vet catches ───
	//
	// 1. PRINTF FORMAT BUGS:
	//    fmt.Printf("%d", "string")      // %d with string arg
	//    fmt.Printf("%s %s", one)        // wrong number of args
	//    log.Printf("%w", err)           // %w only works in fmt.Errorf
	//
	// 2. UNREACHABLE CODE:
	//    return
	//    fmt.Println("never runs")       // unreachable
	//
	// 3. INVALID STRUCT TAGS:
	//    type X struct {
	//        Name string `json:name`     // missing quotes: should be `json:"name"`
	//    }
	//
	// 4. COPYING MUTEX:
	//    var mu sync.Mutex
	//    mu2 := mu                        // copies lock state! BUG
	//
	// 5. SELF-ASSIGNMENT:
	//    x = x                            // certainly a mistake
	//
	// 6. UNUSABLE TEST:
	//    func TestFoo(t *Testing.T)       // wrong signature, won't run
	//
	// 7. BOOLEAN EXPRESSION BUGS:
	//    if x == 1 || x == 2 || x == 1   // duplicate condition
	//
	// 8. SHIFT ERRORS:
	//    var x int8 = 1 << 7              // overflows int8
	//
	// 9. INCORRECT BUILD TAGS:
	//    //go:build linus                  // typo: "linus" not "linux"
	fmt.Println("  go vet ./... → find bugs in all packages")
	fmt.Println("  Zero false positives: if vet says it, it's a bug")
	fmt.Println("  Catches: printf bugs, mutex copy, struct tags, unreachable code")
	fmt.Println()
}

// =============================================================================
// PART 2: go fmt & goimports
// =============================================================================
func goFmtAndImports() {
	fmt.Println("--- go fmt & goimports ---")
	// ─── go fmt: the formatter ───
	// go fmt ./...          # format all files
	// gofmt -w .            # same thing, lower level
	// gofmt -d .            # show diff without writing
	// gofmt -s -w .         # simplify + format
	//
	// go fmt is NON-NEGOTIABLE in Go. Every Go file should be formatted.
	// There's one style. No arguments. No configuration.
	// "gofmt's style is no one's favorite, yet gofmt is everyone's favorite."
	//
	// ─── What go fmt does ───
	// - Tabs for indentation (not spaces)
	// - Aligns struct fields
	// - Normalizes spacing
	// - Removes unnecessary semicolons
	// - Standardizes blank lines
	//
	// ─── goimports: go fmt + import management ───
	// go install golang.org/x/tools/cmd/goimports@latest
	// goimports -w .
	//
	// goimports = go fmt PLUS:
	// - Adds missing imports automatically
	// - Removes unused imports
	// - Groups imports (stdlib, third-party, local)
	// - Sorts imports within groups
	//
	// ─── gofumpt: stricter formatter ───
	// go install mvdan.cc/gofumpt@latest
	// gofumpt -w .
	//
	// gofumpt = go fmt + extra rules:
	// - No empty lines at start/end of functions
	// - Grouped var declarations use a single var block
	// - More consistent style
	//
	// ─── CI setup ───
	// test -z "$(gofmt -d . | head -1)" || (echo "Not formatted" && exit 1)
	fmt.Println("  go fmt ./... → format all code (mandatory)")
	fmt.Println("  goimports → fmt + auto-manage imports")
	fmt.Println("  gofumpt → stricter formatting")
	fmt.Println("  ONE style, zero arguments, no config")
	fmt.Println()
}

// =============================================================================
// PART 3: staticcheck — The Gold Standard
// =============================================================================
func staticcheckTool() {
	fmt.Println("--- staticcheck ---")
	// ─── Install ───
	// go install honnef.co/go/tools/cmd/staticcheck@latest
	//
	// ─── Run ───
	// staticcheck ./...              # analyze all packages
	// staticcheck -checks "all" ./... # enable ALL checks
	// staticcheck -explain SA1029    # explain a specific check
	//
	// ─── Check categories ───
	//
	// SA: staticcheck — real bugs and issues
	//   SA1000  Invalid regexp
	//   SA1006  Printf with func instead of call result
	//   SA1012  nil context.Context passed
	//   SA1029  Inappropriate key type in context.WithValue
	//   SA2000  sync.WaitGroup.Add called in goroutine (race)
	//   SA2001  Empty critical section (Lock then immediately Unlock)
	//   SA4000  Boolean expression always true/false
	//   SA4006  Assigned to but never used
	//   SA4010  Result of append not used
	//   SA5000  Nil pointer dereference
	//   SA5007  Infinite recursive call
	//   SA6000  Using regexp.Match in a loop (compile once!)
	//   SA9001  Defers in for loop
	//
	// S: simple — code simplifications
	//   S1000   Single-case select → just use the channel op
	//   S1002   Omit bool comparison: if x == true → if x
	//   S1017   Replace with strings.TrimPrefix
	//   S1025   Don't use fmt.Sprintf("%s", x) when x is a string
	//
	// ST: style suggestions
	//   ST1000  Package comment missing
	//   ST1003  Poorly named: MixedCaps not underscores
	//   ST1005  Error strings should not be capitalized
	//   ST1006  Poorly named receiver (use short, consistent names)
	//   ST1008  Error should be returned as last value
	//   ST1012  Poorly named error variable (use ErrXxx)
	//
	// QF: quick fixes
	//   QF1001  Apply De Morgan's law
	//   QF1003  Convert if/else chain to switch
	//
	// ─── Configuration (.staticcheck.conf) ───
	// checks = ["all", "-ST1000", "-ST1003"]
	// Or in code: //lint:ignore SA1029 reason
	//
	// ─── Why staticcheck over go vet? ───
	// go vet: conservative, zero false positives, limited scope
	// staticcheck: MORE checks, deeper analysis
	// USE BOTH. They complement each other.
	fmt.Println("  staticcheck ./... → deep code analysis")
	fmt.Println("  SA*: bugs, S*: simplifications, ST*: style")
	fmt.Println("  Catches: nil dereference, infinite recursion, regex in loops")
	fmt.Println("  Use alongside go vet, not instead of")
	fmt.Println()
}

// =============================================================================
// PART 4: go vet Analyzers In Depth
// =============================================================================
func goVetAnalyzers() {
	fmt.Println("--- go vet ANALYZERS ---")
	// go vet runs "analyzers". Each checks for specific bug patterns.
	//
	// ─── Key analyzers ───
	// assign       — useless assignments (x = x)
	// atomic       — common mistakes with sync/atomic
	// bools        — boolean expression mistakes
	// buildtag     — malformed build tags
	// composites   — unkeyed struct literal (fragile)
	// copylocks    — copying mutex, WaitGroup, etc.
	// directive    — malformed tool directives (//go:build)
	// errorsas     — wrong type passed to errors.As
	// httpresponse — unchecked return after http.Error
	// ifaceassert  — impossible interface assertion
	// loopclosure  — references to loop variable in closure (pre-1.22)
	// lostcancel   — context cancellation function never called
	// nilfunc      — useless nil comparison with function
	// printf       — printf format string bugs
	// shift        — shifts that exceed the width of the integer
	// sigchanyzer  — unbuffered os.Signal channel
	// slog         — bad keys in log/slog calls
	// stdmethods   — wrong signature for common methods (String, Error)
	// structtag    — malformed struct tags
	// tests        — wrong test/benchmark/example function signatures
	// unmarshal    — non-pointer passed to json.Unmarshal
	// unreachable  — unreachable code
	// unsafeptr    — misuse of unsafe.Pointer
	// unusedresult — unused result of certain function calls
	//
	// ─── Run specific analyzer ───
	// go vet -copylocks ./...    # only check for mutex copy
	// go vet -printf ./...       # only check printf formats
	//
	// ─── lostcancel example ───
	// ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	// // missing: defer cancel()
	// // vet reports: "the cancel function should be called, not discarded"
	//
	// ─── composites example ───
	// image.Point{5, 10}              // vet warns: unkeyed fields
	// image.Point{X: 5, Y: 10}       // OK: keyed fields
	fmt.Println("  go vet -help → list all analyzers")
	fmt.Println("  Key: copylocks, printf, lostcancel, errorsas, unmarshal")
	fmt.Println("  Every report is a real bug — fix all of them")
	fmt.Println()
}

// =============================================================================
// PART 5: Shadow Variable Detection
// =============================================================================
func shadowDetection() {
	fmt.Println("--- SHADOW DETECTION ---")
	// Variable shadowing: inner scope re-declares an outer variable.
	// The compiler allows it. It's often a bug.
	//
	// err := doFirst()
	// if err == nil {
	//     err := doSecond()  // BUG: shadows outer err! Uses :=
	//     // outer err is still nil even if doSecond fails
	// }
	//
	// ─── Detect with go vet shadow analyzer ───
	// go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	// go vet -vettool=$(which shadow) ./...
	//
	// ─── Common shadowing patterns ───
	//
	// 1. := in if block:
	//    x := 1
	//    if true {
	//        x := 2  // shadows! use = not :=
	//    }
	//    // x is still 1
	//
	// 2. Named return values:
	//    func foo() (err error) {
	//        if err := bar(); err != nil {
	//            return  // returns nil! inner err shadows named return
	//        }
	//    }
	//
	// 3. Loop variable (fixed in Go 1.22):
	//    for _, v := range items {
	//        go func() { use(v) }()  // pre-1.22: uses last v
	//    }
	fmt.Println("  Shadow = inner := hides outer variable (silent bug)")
	fmt.Println("  go vet -vettool=$(which shadow) → detect shadows")
	fmt.Println("  Most common: := inside if/for when = was intended")
	fmt.Println()
}

// =============================================================================
// PART 6: Custom / Additional Analyzers
// =============================================================================
func customAnalyzers() {
	fmt.Println("--- CUSTOM ANALYZERS ---")
	// ─── golang.org/x/tools analyzers ───
	// Beyond what's in go vet, the Go team provides extra analyzers:
	// nilness     — finds guaranteed nil pointer dereferences
	// sortslice   — checks sort.Slice comparator
	// fieldalignment — finds struct padding waste
	//
	// ─── fieldalignment example ───
	// type Bad struct {
	//     a bool       // 1 byte + 7 padding
	//     b int64      // 8 bytes
	//     c bool       // 1 byte + 7 padding
	// }  // Total: 24 bytes
	//
	// type Good struct {
	//     b int64      // 8 bytes
	//     a bool       // 1 byte
	//     c bool       // 1 byte + 6 padding
	// }  // Total: 16 bytes (33% less!)
	//
	// ─── Writing your own analyzer ───
	// Analyzers use golang.org/x/tools/go/analysis framework:
	//
	// var Analyzer = &analysis.Analyzer{
	//     Name: "mycheck",
	//     Doc:  "checks for my specific pattern",
	//     Run:  run,
	// }
	//
	// func run(pass *analysis.Pass) (interface{}, error) {
	//     for _, file := range pass.Files {
	//         ast.Inspect(file, func(n ast.Node) bool {
	//             pass.Reportf(n.Pos(), "found bad pattern")
	//             return true
	//         })
	//     }
	//     return nil, nil
	// }
	//
	// ─── govulncheck: vulnerability scanner ───
	// go install golang.org/x/vuln/cmd/govulncheck@latest
	// govulncheck ./...
	// Checks code against Go vulnerability database.
	// Only reports vulns in code you ACTUALLY CALL (not just import).
	fmt.Println("  fieldalignment → save memory by reordering struct fields")
	fmt.Println("  govulncheck → find known vulnerabilities in deps")
	fmt.Println("  Custom analyzers: golang.org/x/tools/go/analysis")
	fmt.Println()
}
