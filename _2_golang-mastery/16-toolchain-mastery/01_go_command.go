//go:build ignore

// =============================================================================
// GO TOOLCHAIN 1: THE `go` COMMAND — Your Swiss Army Knife
// =============================================================================
//
// The `go` command is the single entry point for:
// building, testing, profiling, formatting, vetting, managing dependencies,
// generating code, and more. Master it and you control Go itself.
//
// RUN: go run 01_go_command.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== THE go COMMAND — COMPLETE REFERENCE ===")
	fmt.Println()

	buildAndRun()
	moduleCommands()
	environmentAndConfig()
	crossCompilation()
	buildFlags()
	installAndBinaries()
	cleanAndCache()
	docAndGodoc()
	goEnvDeepDive()
}

// =============================================================================
// PART 1: go build & go run
// =============================================================================
func buildAndRun() {
	fmt.Println("--- go build & go run ---")

	// ─── go run: compile + execute in one step ───
	// go run main.go              # single file
	// go run .                    # current package
	// go run ./cmd/server         # specific package
	// go run -race main.go       # with race detector
	//
	// go run does NOT produce a binary. Compiles to temp dir and runs.
	// Use for development only.

	// ─── go build: compile to binary ───
	// go build                    # builds current package → binary in current dir
	// go build -o myapp           # custom output name
	// go build ./cmd/server       # build specific package
	// go build ./...              # build all packages (checks compilation)
	//
	// Default binary name = directory name (not package name)

	// ─── What go build actually does ───
	// 1. Resolves imports → dependency graph
	// 2. Compiles each package (parallel, cached)
	// 3. Links all packages into single binary
	// 4. Output: statically linked binary (usually)
	//
	// Go compiles FAST because:
	// - No header files (imports are compiled packages)
	// - Dependency graph is a DAG (no cycles)
	// - Unused imports = compile error (no wasted work)
	// - Compilation cache (recompiles only changed packages)

	// ─── go install: build + put in $GOPATH/bin ───
	// go install                          # current package
	// go install github.com/user/tool@latest  # install from remote
	//
	// After install: tool is in $(go env GOPATH)/bin/
	// Make sure $GOPATH/bin is in your $PATH

	fmt.Println("  go run   → compile + run (dev only)")
	fmt.Println("  go build → compile to binary")
	fmt.Println("  go build ./... → check all packages compile")
	fmt.Println("  go install → build + install to GOPATH/bin")
	fmt.Println()
}

// =============================================================================
// PART 2: Module Commands
// =============================================================================
func moduleCommands() {
	fmt.Println("--- MODULE COMMANDS ---")

	// ─── go mod init ───
	// go mod init github.com/user/project
	// Creates go.mod, starts a new module

	// ─── go mod tidy (USE THIS CONSTANTLY) ───
	// go mod tidy
	// - Adds missing dependencies
	// - Removes unused dependencies
	// - Updates go.sum
	// Run this after every import change.

	// ─── go get: add/update dependencies ───
	// go get github.com/pkg/errors          # latest
	// go get github.com/pkg/errors@v0.9.1   # specific version
	// go get github.com/pkg/errors@latest    # latest release
	// go get github.com/pkg/errors@master    # specific branch
	// go get github.com/pkg/errors@abc1234   # specific commit
	// go get -u github.com/pkg/errors        # update to latest minor
	// go get -u ./...                         # update ALL deps

	// ─── go mod vendor ───
	// go mod vendor
	// Copies all dependencies into vendor/ directory
	// go build -mod=vendor  # build using vendor only
	//
	// When to vendor:
	// - Reproducible builds without network
	// - CI/CD environments
	// - Corporate environments with restricted access

	// ─── go mod download ───
	// go mod download
	// Pre-downloads all dependencies to local cache
	// Useful in Docker builds (cache layer)

	// ─── go mod graph ───
	// go mod graph
	// Prints full dependency graph:
	// github.com/you/app github.com/gin-gonic/gin@v1.9.1
	// github.com/gin-gonic/gin@v1.9.1 github.com/json-iterator/go@v1.1.12

	// ─── go mod why ───
	// go mod why github.com/some/dep
	// Shows WHY a dependency exists (who imports it)
	// Great for "why is this in my go.sum?"

	// ─── go mod verify ───
	// go mod verify
	// Verifies dependencies haven't been tampered with
	// Checks hashes against go.sum

	// ─── go mod edit ───
	// go mod edit -require github.com/pkg/errors@v0.9.1
	// go mod edit -replace github.com/old=github.com/new@v1.0.0
	// go mod edit -replace github.com/lib=../local-lib
	// go mod edit -json  # output go.mod as JSON

	fmt.Println("  go mod tidy    → sync deps (run constantly)")
	fmt.Println("  go get pkg@ver → add/update dependency")
	fmt.Println("  go mod vendor  → copy deps locally")
	fmt.Println("  go mod graph   → print dependency tree")
	fmt.Println("  go mod why     → explain why dep exists")
	fmt.Println()
}

// =============================================================================
// PART 3: Environment & Configuration
// =============================================================================
func environmentAndConfig() {
	fmt.Println("--- ENVIRONMENT & CONFIG ---")

	// ─── Key environment variables ───
	//
	// GOROOT    — where Go is installed
	//             go env GOROOT → /usr/local/go
	//             Contains: stdlib source, compiler, tools
	//
	// GOPATH    — workspace for Go code
	//             go env GOPATH → ~/go
	//             Contains: bin/ (installed binaries), pkg/ (build cache)
	//
	// GOBIN     — where `go install` puts binaries
	//             Default: $GOPATH/bin
	//
	// GOMODCACHE — module download cache
	//             Default: $GOPATH/pkg/mod
	//             All downloaded modules live here
	//
	// GOPROXY   — module proxy URL
	//             Default: https://proxy.golang.org,direct
	//             Corporate: set to internal proxy
	//
	// GONOPROXY — modules to fetch directly (skip proxy)
	//             GONOPROXY=github.com/mycompany/*
	//
	// GONOSUMDB — modules to skip checksum DB
	//             GONOSUMDB=github.com/mycompany/*
	//
	// GOPRIVATE — shorthand for GONOPROXY + GONOSUMDB
	//             GOPRIVATE=github.com/mycompany/*
	//             USE THIS for private repos
	//
	// GOFLAGS   — default flags for go commands
	//             GOFLAGS="-mod=vendor" → always use vendor
	//
	// CGO_ENABLED — enable/disable cgo
	//             CGO_ENABLED=0 → pure Go, no C dependencies
	//             CGO_ENABLED=1 → allow C code (default on most platforms)

	// ─── Set environment permanently ───
	// go env -w GOPRIVATE=github.com/mycompany/*
	// go env -w GOBIN=/usr/local/bin
	// Stored in: $(go env GOENV) → usually ~/.config/go/env

	fmt.Println("  GOROOT     → Go installation")
	fmt.Println("  GOPATH     → workspace (~/go)")
	fmt.Println("  GOPRIVATE  → skip proxy for private repos")
	fmt.Println("  CGO_ENABLED=0 → pure Go binary")
	fmt.Println("  go env -w  → set vars permanently")
	fmt.Println()
}

// =============================================================================
// PART 4: Cross-Compilation
// =============================================================================
func crossCompilation() {
	fmt.Println("--- CROSS-COMPILATION ---")

	// Go has built-in cross-compilation. No extra tools needed.
	// Set GOOS and GOARCH before `go build`.

	// ─── Common targets ───
	// GOOS=linux   GOARCH=amd64  → Linux x86_64 (servers)
	// GOOS=linux   GOARCH=arm64  → Linux ARM (Raspberry Pi, AWS Graviton)
	// GOOS=darwin  GOARCH=amd64  → macOS Intel
	// GOOS=darwin  GOARCH=arm64  → macOS Apple Silicon (M1/M2/M3)
	// GOOS=windows GOARCH=amd64  → Windows x86_64
	// GOOS=js      GOARCH=wasm   → WebAssembly
	// GOOS=wasip1  GOARCH=wasm   → WASI (Go 1.21+)

	// ─── Build for Linux from macOS ───
	// GOOS=linux GOARCH=amd64 go build -o myapp-linux
	//
	// ─── Build for Windows from Linux ───
	// GOOS=windows GOARCH=amd64 go build -o myapp.exe
	//
	// ─── Build for all platforms ───
	// #!/bin/bash
	// platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")
	// for platform in "${platforms[@]}"; do
	//     GOOS=${platform%/*} GOARCH=${platform#*/} \
	//     go build -o "myapp-${GOOS}-${GOARCH}" .
	// done

	// ─── CGO and cross-compilation ───
	// CGO requires a C cross-compiler for the target platform.
	// Easiest solution: CGO_ENABLED=0 for pure Go cross-builds
	// GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o myapp

	// ─── List all supported platforms ───
	// go tool dist list
	// Shows all valid GOOS/GOARCH combinations (~50+)

	fmt.Println("  GOOS=linux GOARCH=amd64 go build → Linux binary")
	fmt.Println("  CGO_ENABLED=0 → no C compiler needed")
	fmt.Println("  go tool dist list → all supported platforms")
	fmt.Println()
}

// =============================================================================
// PART 5: Build Flags — The Power User Section
// =============================================================================
func buildFlags() {
	fmt.Println("--- BUILD FLAGS ---")

	// ─── -ldflags: Linker flags ───
	// Inject values at compile time (version, commit, build time)
	//
	// // In main.go:
	// var (
	//     version   string
	//     gitCommit string
	//     buildTime string
	// )
	//
	// // Build command:
	// go build -ldflags "-X main.version=1.2.3 \
	//                     -X main.gitCommit=$(git rev-parse HEAD) \
	//                     -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	//
	// This is how every Go binary gets its version info!

	// ─── -ldflags "-s -w": Strip debug info ───
	// -s → strip symbol table
	// -w → strip DWARF debugging info
	// Reduces binary size by ~30%
	// go build -ldflags "-s -w" -o myapp
	//
	// TRADEOFF: Can't use delve debugger or get good stack traces

	// ─── -trimpath: Reproducible builds ───
	// go build -trimpath
	// Removes local file paths from binary
	// Stack traces show: mypackage/file.go instead of /home/user/code/mypackage/file.go
	// Required for reproducible builds

	// ─── -tags: Build tags ───
	// go build -tags "integration,debug"
	// Only compiles files with matching //go:build tags
	//
	// //go:build integration
	// //go:build debug
	// //go:build linux && amd64

	// ─── -gcflags: Compiler flags ───
	// go build -gcflags="-m"             # escape analysis output
	// go build -gcflags="-m -m"          # verbose escape analysis
	// go build -gcflags="-S"             # assembly output
	// go build -gcflags="-N -l"          # disable optimizations (for debugging)
	// go build -gcflags="all=-m"         # escape analysis for ALL packages
	//
	// -m  → show optimization decisions
	// -S  → show generated assembly
	// -N  → disable optimizations
	// -l  → disable inlining
	// -B  → disable bounds checking (DANGEROUS)

	// ─── -race: Race detector ───
	// go build -race -o myapp
	// go run -race main.go
	// go test -race ./...
	// Adds ~10x CPU overhead, ~5x memory overhead
	// ALWAYS use in development and CI

	// ─── -cover: Code coverage instrumentation ───
	// go build -cover -o myapp
	// GOCOVERDIR=./coverage ./myapp
	// go tool covdata percent -i ./coverage

	// ─── Combining flags (production build) ───
	// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	// go build -trimpath \
	//   -ldflags "-s -w -X main.version=1.2.3" \
	//   -o myapp ./cmd/server

	fmt.Println("  -ldflags \"-X main.version=...\" → inject version")
	fmt.Println("  -ldflags \"-s -w\" → strip debug (30% smaller)")
	fmt.Println("  -trimpath → reproducible builds")
	fmt.Println("  -gcflags=\"-m\" → escape analysis")
	fmt.Println("  -race → race detector (always use in dev)")
	fmt.Println()
}

// =============================================================================
// PART 6: go install & Managing Binaries
// =============================================================================
func installAndBinaries() {
	fmt.Println("--- go install & BINARIES ---")

	// ─── Install tools from anywhere ───
	// go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	// go install golang.org/x/tools/gopls@latest
	// go install github.com/go-delve/delve/cmd/dlv@latest
	// go install golang.org/x/vuln/cmd/govulncheck@latest
	//
	// Binary goes to: $(go env GOPATH)/bin/
	// Make sure this is in your PATH:
	// export PATH=$PATH:$(go env GOPATH)/bin

	// ─── Essential tools to install ───
	// # Language server (used by VS Code, GoLand)
	// go install golang.org/x/tools/gopls@latest
	//
	// # Debugger
	// go install github.com/go-delve/delve/cmd/dlv@latest
	//
	// # Linter (comprehensive)
	// go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	//
	// # Vulnerability check
	// go install golang.org/x/vuln/cmd/govulncheck@latest
	//
	// # Mock generator
	// go install go.uber.org/mock/mockgen@latest

	fmt.Println("  go install tool@latest → install Go tools")
	fmt.Println("  Make sure $(go env GOPATH)/bin is in PATH")
	fmt.Println()
}

// =============================================================================
// PART 7: go clean & Cache Management
// =============================================================================
func cleanAndCache() {
	fmt.Println("--- CACHE MANAGEMENT ---")

	// ─── Build cache ───
	// go env GOCACHE → ~/.cache/go-build (Linux)
	// Stores compiled packages. Speeds up builds dramatically.
	// go clean -cache     # clear build cache
	// go clean -testcache # clear test result cache
	// go clean -modcache  # clear module download cache (re-downloads everything)
	// go clean -fuzzcache # clear fuzz test cache

	// ─── Cache is GOOD — rarely need to clear ───
	// Clear build cache when:
	// - Debugging weird compilation issues
	// - Build seems to use stale code (rare)
	//
	// Clear test cache when:
	// - Tests should re-run (external dependency changed)
	// - go test -count=1 also bypasses cache (per-run)
	//
	// Clear mod cache when:
	// - Corrupted downloads
	// - Reclaiming disk space
	// - Usually NOT needed

	// ─── Cache size check ───
	// du -sh $(go env GOCACHE)    # build cache size
	// du -sh $(go env GOMODCACHE) # module cache size

	fmt.Println("  Build cache: $(go env GOCACHE)")
	fmt.Println("  Module cache: $(go env GOMODCACHE)")
	fmt.Println("  go clean -cache → clear build cache")
	fmt.Println("  go test -count=1 → bypass test cache")
	fmt.Println()
}

// =============================================================================
// PART 8: go doc & Documentation
// =============================================================================
func docAndGodoc() {
	fmt.Println("--- DOCUMENTATION ---")

	// ─── go doc: read docs from terminal ───
	// go doc fmt                    # package overview
	// go doc fmt.Println            # specific function
	// go doc fmt.Stringer           # interface
	// go doc -all fmt               # everything in package
	// go doc -src fmt.Println       # show source code!
	// go doc net/http.Server        # qualified name
	// go doc net/http.Server.Shutdown  # method
	//
	// This reads the ACTUAL source comments. It's always up to date.

	// ─── pkgsite: modern docs server ───
	// go install golang.org/x/pkgsite/cmd/pkgsite@latest
	// pkgsite -http=:8080
	// Same as pkg.go.dev but for local packages

	// ─── Writing good docs ───
	//
	// // Package strings implements simple functions to manipulate
	// // UTF-8 encoded strings.
	// package strings
	//
	// // Split slices s into all substrings separated by sep and returns
	// // a slice of the substrings between those separators.
	// func Split(s, sep string) []string { ... }
	//
	// RULES:
	// 1. Start with the name: "Split slices s..."
	// 2. Complete sentences, proper grammar
	// 3. First paragraph = summary (shown in package list)
	// 4. Include examples (func ExampleSplit())

	fmt.Println("  go doc fmt.Println → read function docs")
	fmt.Println("  go doc -src → show source code")
	fmt.Println()
}

// =============================================================================
// PART 9: go env Deep Dive
// =============================================================================
func goEnvDeepDive() {
	fmt.Println("--- go env DEEP DIVE ---")

	// ─── View all env vars ───
	// go env          # all vars
	// go env GOROOT   # specific var
	// go env -json    # JSON format

	// ─── The essential ones explained ───
	// GOROOT       /usr/local/go        Go installation
	// GOPATH       ~/go                 Workspace
	// GOBIN        ~/go/bin             Binary install dir
	// GOCACHE      ~/.cache/go-build    Build cache
	// GOMODCACHE   ~/go/pkg/mod         Downloaded modules
	// GOENV        ~/.config/go/env     Persistent env settings
	// GOTOOLDIR    /usr/local/go/pkg/tool/linux_amd64  Internal tools
	// GOVERSION    go1.22.0             Go version
	// GOPROXY      https://proxy.golang.org,direct
	// GOPRIVATE    ""                   Private modules
	// CC           gcc                  C compiler (for cgo)

	// ─── Useful commands ───
	// go version          → Go version
	// go version -m myapp → show build info of compiled binary
	//                        (module versions, build flags, VCS info)
	//
	// go tool dist list   → all supported GOOS/GOARCH

	fmt.Println("  go env → view all settings")
	fmt.Println("  go env -json → JSON format")
	fmt.Println("  go version -m binary → inspect build info")
	fmt.Println()
}

