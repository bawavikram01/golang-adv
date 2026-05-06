//go:build ignore

// =============================================================================
// GO TOOLCHAIN 8: golangci-lint — The Meta-Linter
// =============================================================================
//
// golangci-lint runs 100+ linters in parallel, fast, with caching.
// It's the industry standard for Go code quality in CI.
// It replaces running go vet, staticcheck, errcheck, etc. individually.
//
// Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
//
// RUN: go run 08_golangci_lint.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== golangci-lint ===")
	fmt.Println()
	golangciBasics()
	configFile()
	essentialLinters()
	advancedLinters()
	ciIntegration()
	customRules()
}

// =============================================================================
// PART 1: Basics
// =============================================================================
func golangciBasics() {
	fmt.Println("--- golangci-lint BASICS ---")
	// ─── Running ───
	// golangci-lint run              # lint current package
	// golangci-lint run ./...        # lint all packages
	// golangci-lint run --fix        # auto-fix where possible
	// golangci-lint run --fast       # only fast linters
	// golangci-lint run --new-from-rev=HEAD~1  # only new issues
	//
	// ─── Why golangci-lint instead of individual linters? ───
	// 1. FAST: runs linters in parallel, reuses parsed AST
	// 2. CACHED: only re-lints changed code
	// 3. ONE CONFIG: .golangci.yml controls everything
	// 4. 100+ linters available
	// 5. IDE integration (VS Code, GoLand)
	// 6. CI-friendly (GitHub Actions, GitLab CI)
	//
	// ─── List available linters ───
	// golangci-lint linters           # list all linters
	// golangci-lint linters --enable-all  # show which are enabled
	//
	// ─── Default enabled linters ───
	// errcheck      — check error returns are used
	// gosimple      — simplify code (staticcheck S* rules)
	// govet         — go vet checks
	// ineffassign   — detect useless assignments
	// staticcheck   — staticcheck SA*/S* rules
	// unused        — detect unused code
	fmt.Println("  golangci-lint run ./... → lint everything")
	fmt.Println("  --fix → auto-fix issues")
	fmt.Println("  --new-from-rev=HEAD~1 → only new issues")
	fmt.Println("  100+ linters, parallel, cached")
	fmt.Println()
}

// =============================================================================
// PART 2: Configuration File
// =============================================================================
func configFile() {
	fmt.Println("--- CONFIGURATION ---")
	// Create .golangci.yml in project root:
	//
	// # .golangci.yml
	// run:
	//   timeout: 5m
	//   tests: true               # lint test files too
	//   go: '1.22'                # minimum Go version
	//
	// linters:
	//   enable:
	//     - errcheck              # check error returns
	//     - govet                 # go vet
	//     - staticcheck           # staticcheck
	//     - gosimple              # simplifications
	//     - unused                # unused code
	//     - ineffassign           # useless assignments
	//     - gocritic              # opinionated Go checks
	//     - revive                # fast, extensible linter
	//     - errname               # error naming conventions
	//     - errorlint             # error wrapping issues
	//     - exhaustive            # enum switch exhaustiveness
	//     - goconst               # repeated strings → const
	//     - gofumpt               # strict formatting
	//     - misspell              # spelling mistakes in comments
	//     - noctx                 # HTTP requests without context
	//     - prealloc              # suggest slice pre-allocation
	//     - unconvert             # unnecessary type conversions
	//     - unparam               # unused function parameters
	//     - wastedassign          # wasted assignments
	//
	// linters-settings:
	//   govet:
	//     enable-all: true        # all go vet analyzers
	//   errcheck:
	//     check-type-assertions: true  # check x.(Type) without ok
	//     check-blank: true       # check _ = errors
	//   gocritic:
	//     enabled-tags:
	//       - diagnostic
	//       - style
	//       - performance
	//   revive:
	//     rules:
	//       - name: exported
	//         disabled: true      # don't require comments on exports
	//   staticcheck:
	//     checks: ["all"]
	//
	// issues:
	//   max-issues-per-linter: 0  # no limit
	//   max-same-issues: 0        # no limit
	//   exclude-rules:
	//     - path: _test\.go       # test files
	//       linters:
	//         - errcheck           # OK to ignore errors in tests
	//         - goconst            # repeated strings OK in tests
	//
	// severity:
	//   default-severity: warning
	fmt.Println("  .golangci.yml → project-level configuration")
	fmt.Println("  enable linters explicitly (don't rely on defaults)")
	fmt.Println("  exclude-rules → relax rules for test files")
	fmt.Println()
}

// =============================================================================
// PART 3: Essential Linters Explained
// =============================================================================
func essentialLinters() {
	fmt.Println("--- ESSENTIAL LINTERS ---")
	// ─── errcheck: never ignore errors ───
	// Catches: result, _ := doSomething()
	// Also: file.Close() without checking error
	// Severity: HIGH — unchecked errors cause silent failures
	//
	// ─── govet: compiler-level checks ───
	// All the go vet analyzers (printf, copylocks, etc.)
	// Enable-all for maximum coverage.
	//
	// ─── staticcheck: deep analysis ───
	// SA* (bugs), S* (simplifications), ST* (style)
	// The gold standard for Go static analysis.
	//
	// ─── errname: error naming ───
	// Checks: error types end in "Error" (e.g., *NotFoundError)
	// Checks: sentinel errors start with "Err" (e.g., ErrNotFound)
	//
	// ─── errorlint: error wrapping ───
	// Checks: use errors.Is instead of ==
	// Checks: use errors.As instead of type assertion
	// Checks: use %w instead of %v in fmt.Errorf
	//
	// ─── noctx: HTTP + context ───
	// Checks: http.Get() → use http.NewRequestWithContext instead
	// All HTTP requests should carry a context for cancellation.
	//
	// ─── gocritic: opinionated best practices ───
	// 100+ checks in categories: diagnostic, style, performance
	// Examples:
	//   appendAssign:   x = append(y, ...)  // should be x = append(x, ...)
	//   hugeParam:      func(x [1000]int)   // pass pointer instead
	//   rangeValCopy:   for _, v := range bigStructSlice  // v is copy
	//   singleCaseSwitch: switch with 1 case → use if
	//
	// ─── exhaustive: enum switch completeness ───
	// If you switch on a type with known values, checks all cases covered.
	// type Color int
	// const (Red Color = iota; Green; Blue)
	// switch c {
	// case Red: ...
	// case Green: ...
	// }  // MISSING Blue! exhaustive catches this.
	//
	// ─── prealloc: slice optimization ───
	// Suggests: make([]T, 0, n) when n is known
	// Reduces GC pressure from slice growing.
	fmt.Println("  errcheck → never ignore errors")
	fmt.Println("  errorlint → proper error wrapping")
	fmt.Println("  noctx → HTTP with context")
	fmt.Println("  gocritic → 100+ best practice checks")
	fmt.Println("  exhaustive → complete switch cases")
	fmt.Println()
}

// =============================================================================
// PART 4: Advanced Linters
// =============================================================================
func advancedLinters() {
	fmt.Println("--- ADVANCED LINTERS ---")
	// ─── gocognit / gocyclo: complexity ───
	// Measures function complexity.
	// gocognit: cognitive complexity (human readability)
	// gocyclo: cyclomatic complexity (number of paths)
	// Set threshold: max-complexity: 30
	// If a function exceeds it → refactor.
	//
	// ─── dupl: copy-paste detection ───
	// Finds duplicate code blocks.
	// threshold: 100 (tokens)
	// Reduce duplication → extract functions.
	//
	// ─── funlen: function length ───
	// Flags functions over N lines.
	// Default: 60 lines
	// Long functions → hard to test, hard to read.
	//
	// ─── gosec: security checks ───
	// Finds security issues:
	//   G101: Hardcoded credentials
	//   G104: Unhandled errors
	//   G201: SQL injection (string concatenation in queries)
	//   G301: Poor file permissions
	//   G401: Use of weak crypto (MD5, SHA1 for security)
	//   G501: Import blocklist
	//
	// ─── bodyclose: HTTP response body ───
	// Detects: resp, _ := http.Get(url) without resp.Body.Close()
	// Leaked response bodies → connection leak → resource exhaustion.
	//
	// ─── sqlclosecheck: database rows ───
	// Detects: rows, _ := db.Query(...) without rows.Close()
	// Leaked rows → connection pool exhaustion.
	//
	// ─── contextcheck: context propagation ───
	// Checks that context.Context is properly passed through call chains.
	// Catches: using context.Background() when a context is available.
	//
	// ─── nilnil: nil interface returns ───
	// Detects: return nil, nil (when the first return is an interface)
	// This often indicates a bug.
	fmt.Println("  gosec → security vulnerabilities")
	fmt.Println("  bodyclose → HTTP response body leaks")
	fmt.Println("  sqlclosecheck → database connection leaks")
	fmt.Println("  gocognit → function complexity")
	fmt.Println()
}

// =============================================================================
// PART 5: CI Integration
// =============================================================================
func ciIntegration() {
	fmt.Println("--- CI INTEGRATION ---")
	// ─── GitHub Actions ───
	// # .github/workflows/lint.yml
	// name: Lint
	// on: [push, pull_request]
	// jobs:
	//   lint:
	//     runs-on: ubuntu-latest
	//     steps:
	//       - uses: actions/checkout@v4
	//       - uses: actions/setup-go@v5
	//         with:
	//           go-version: '1.22'
	//       - uses: golangci/golangci-lint-action@v4
	//         with:
	//           version: latest
	//           args: --timeout=5m
	//
	// ─── GitLab CI ───
	// lint:
	//   image: golangci/golangci-lint:latest
	//   script:
	//     - golangci-lint run --timeout=5m ./...
	//
	// ─── Only lint new issues (great for legacy codebases) ───
	// golangci-lint run --new-from-rev=origin/main
	// Only reports issues in code changed since main branch.
	// Lets you adopt linting without fixing everything at once.
	//
	// ─── Output formats ───
	// golangci-lint run --out-format=json        # JSON (for tools)
	// golangci-lint run --out-format=checkstyle  # for CI systems
	// golangci-lint run --out-format=github-actions  # GH annotations
	// golangci-lint run --out-format=tab          # human-readable
	//
	// ─── Pre-commit hook ───
	// # .pre-commit-config.yaml
	// repos:
	//   - repo: https://github.com/golangci/golangci-lint
	//     rev: v1.56.2
	//     hooks:
	//       - id: golangci-lint
	//
	// ─── Recommended CI pipeline ───
	// 1. go build ./...              # compile check
	// 2. golangci-lint run ./...     # lint
	// 3. go test -race -cover ./... # test + race + coverage
	// 4. govulncheck ./...          # vulnerability scan
	fmt.Println("  GitHub Actions: golangci/golangci-lint-action@v4")
	fmt.Println("  --new-from-rev=origin/main → only new issues")
	fmt.Println("  CI order: build → lint → test → vulncheck")
	fmt.Println()
}

// =============================================================================
// PART 6: Custom Rules & Nolint
// =============================================================================
func customRules() {
	fmt.Println("--- CUSTOM RULES & NOLINT ---")
	// ─── Suppress specific issues ───
	// //nolint:errcheck             // suppress errcheck on this line
	// //nolint:gocritic,staticcheck // suppress multiple
	// //nolint                       // suppress ALL (bad practice)
	//
	// ALWAYS include a reason:
	// //nolint:errcheck // best-effort cleanup, error not actionable
	//
	// ─── Suppress for whole file ───
	// Put at top of file:
	// //nolint:dupl // this file intentionally duplicates X for clarity
	//
	// ─── Enforce nolint reasons ───
	// # .golangci.yml
	// issues:
	//   exclude-use-default: false
	// linters:
	//   enable:
	//     - nolintlint            # checks nolint directives
	// linters-settings:
	//   nolintlint:
	//     require-explanation: true  # must have reason
	//     require-specific: true     # must name the linter
	//
	// ─── Custom revive rules ───
	// revive is extensible with custom rules:
	// linters-settings:
	//   revive:
	//     rules:
	//       - name: add-constant
	//         arguments:
	//           - maxLitCount: "3"  # flag magic numbers used > 3 times
	//       - name: function-result-limit
	//         arguments: [3]       # max 3 return values
	//       - name: argument-limit
	//         arguments: [5]       # max 5 function parameters
	//       - name: cognitive-complexity
	//         arguments: [20]      # max cognitive complexity
	//       - name: line-length-limit
	//         arguments: [120]     # max line length
	fmt.Println("  //nolint:linter // reason → suppress with explanation")
	fmt.Println("  nolintlint → enforce nolint has reasons")
	fmt.Println("  revive rules → custom code standards")
	fmt.Println()
}
