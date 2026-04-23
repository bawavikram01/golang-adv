//go:build ignore

// =============================================================================
// GO TOOLCHAIN 10: BUILD SYSTEM, LINKING & BINARY ANALYSIS
// =============================================================================
//
// The final piece: how Go turns source into binaries, how the linker works,
// how to analyze binaries, Makefiles, Docker builds, and CI pipelines.
// This is the production engineering layer.
//
// RUN: go run 10_build_and_link.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== BUILD SYSTEM & BINARY ANALYSIS ===")
	fmt.Println()
	compilationPipeline()
	linkerDeepDive()
	binaryAnalysis()
	makefilePatterns()
	dockerBuilds()
	ciPipeline()
	releaseAutomation()
}

// =============================================================================
// PART 1: Go Compilation Pipeline
// =============================================================================
func compilationPipeline() {
	fmt.Println("--- COMPILATION PIPELINE ---")
	// ─── What happens during `go build` ───
	//
	// Source (.go) → Parse → Type Check → SSA → Machine Code → Link → Binary
	//
	// Step by step:
	// 1. PARSE: Source → AST (Abstract Syntax Tree)
	//    - Lexer: tokens  - Parser: tree structure
	//    - Each file independently parsed
	//
	// 2. TYPE CHECK: AST → Type-checked AST
	//    - Resolve names, check types
	//    - Constant folding, dead code elimination
	//    - This catches compile errors
	//
	// 3. SSA: AST → SSA (Static Single Assignment form)
	//    - Intermediate representation for optimization
	//    - View it: GOSSAFUNC=myFunc go build
	//      → Opens browser with SSA visualization!
	//
	// 4. MACHINE CODE: SSA → Object file (.o)
	//    - Platform-specific code generation
	//    - Register allocation, instruction scheduling
	//    - See assembly: go build -gcflags="-S"
	//
	// 5. LINK: Object files → Single binary
	//    - Combines all packages
	//    - Resolves symbols, writes executable
	//    - Adds runtime (GC, scheduler, etc.)
	//
	// ─── Compilation is per-PACKAGE ───
	// Each package compiles independently and is cached.
	// Changing one file → only recompile that package + dependents.
	// This is why Go builds are fast.
	//
	// ─── View SSA ───
	// GOSSAFUNC=main go build
	// Opens ssa.html in browser showing optimization passes.
	// See how the compiler transforms your code step by step.
	fmt.Println("  Source → Parse → TypeCheck → SSA → MachineCode → Link")
	fmt.Println("  GOSSAFUNC=main go build → visualize SSA passes")
	fmt.Println("  go build -gcflags=\"-S\" → see generated assembly")
	fmt.Println()
}

// =============================================================================
// PART 2: Linker Deep Dive
// =============================================================================
func linkerDeepDive() {
	fmt.Println("--- LINKER DEEP DIVE ---")
	// ─── What the linker does ───
	// 1. Combines all package object files
	// 2. Resolves symbol references
	// 3. Adds the Go runtime
	// 4. Writes the final executable
	//
	// ─── Static vs dynamic linking ───
	// CGO_ENABLED=0 → fully static binary (default for cross-compile)
	//   No external dependencies. Copy binary anywhere and it runs.
	//
	// CGO_ENABLED=1 → may link against libc dynamically
	//   Smaller binary, but needs compatible libc on target.
	//   Default on Linux when build host = target.
	//
	// Force static with CGO:
	// CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static"'
	// Requires: static libc (apt install musl-tools, then CC=musl-gcc)
	//
	// ─── -ldflags reference ───
	// -s          Strip symbol table (can't debug, ~15% smaller)
	// -w          Strip DWARF debug info (~15% smaller)
	// -X pkg.Var=value  Set string variable at link time
	// -linkmode external  Use external linker (for static linking with CGO)
	// -extldflags "..."   Pass flags to external linker
	// -buildid=""         Remove build ID (reproducible builds)
	//
	// ─── Version injection pattern ───
	// // main.go
	// var (
	//     version   = "dev"
	//     commit    = "none"
	//     buildTime = "unknown"
	// )
	//
	// // Makefile
	// VERSION := $(shell git describe --tags --always --dirty)
	// COMMIT  := $(shell git rev-parse HEAD)
	// TIME    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
	// LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(TIME)
	// go build -ldflags "$(LDFLAGS)" -o myapp
	fmt.Println("  CGO_ENABLED=0 → fully static binary")
	fmt.Println("  -ldflags \"-s -w\" → strip debug info (~30% smaller)")
	fmt.Println("  -X main.version=... → inject version at link time")
	fmt.Println()
}

// =============================================================================
// PART 3: Binary Analysis
// =============================================================================
func binaryAnalysis() {
	fmt.Println("--- BINARY ANALYSIS ---")
	// ─── Inspect build info ───
	// go version -m ./myapp
	// Output:
	//   ./myapp: go1.22.0
	//     path    mymodule
	//     mod     mymodule  (devel)
	//     dep     github.com/gin-gonic/gin  v1.9.1  h1:4idE...
	//     build   -compiler=gc
	//     build   CGO_ENABLED=0
	//     build   GOOS=linux
	//     build   GOARCH=amd64
	//     build   vcs=git
	//     build   vcs.revision=abc123...
	//     build   vcs.time=2024-01-15T10:30:00Z
	//     build   vcs.modified=false
	//
	// This shows: Go version, all dependencies, build settings, git info.
	// Works on ANY Go binary (even third-party).
	//
	// ─── Binary size analysis ───
	// go build -o myapp
	// ls -lh myapp              # total size
	//
	// # What's in the binary?
	// go tool nm myapp | head -50   # list symbols
	// go tool nm myapp | grep 'T main\.'  # your functions
	//
	// # Size by package:
	// go install github.com/nicholasgasior/gokb@latest
	// gokb myapp                    # size breakdown by package
	//
	// # Or use bloaty (Google tool):
	// bloaty myapp                  # size per section
	//
	// ─── Reduce binary size ───
	// 1. -ldflags "-s -w"           Strip debug info (~30% reduction)
	// 2. CGO_ENABLED=0              Don't link libc
	// 3. Use -trimpath              Remove local paths
	// 4. UPX compression:
	//    upx --best myapp           # compress binary (~70% smaller)
	//    BUT: slower startup, some AV false positives
	// 5. Use smaller dependencies (fewer deps = smaller binary)
	// 6. go build -tags netgo       Pure Go net (no cgo net)
	//
	// ─── Typical sizes ───
	// Hello world:    ~1.5 MB
	// HTTP server:    ~7 MB
	// With stripped:  ~5 MB
	// With UPX:       ~2 MB
	fmt.Println("  go version -m binary → inspect build info")
	fmt.Println("  go tool nm binary → list symbols")
	fmt.Println("  -ldflags \"-s -w\" + UPX → minimize binary size")
	fmt.Println()
}

// =============================================================================
// PART 4: Makefile Patterns
// =============================================================================
func makefilePatterns() {
	fmt.Println("--- MAKEFILE PATTERNS ---")
	// # Makefile for Go projects
	//
	// .PHONY: all build test lint clean run docker help
	//
	// # Variables
	// BINARY := myapp
	// VERSION := $(shell git describe --tags --always --dirty)
	// COMMIT := $(shell git rev-parse --short HEAD)
	// BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
	// LDFLAGS := -s -w \
	//     -X main.version=$(VERSION) \
	//     -X main.commit=$(COMMIT) \
	//     -X main.buildTime=$(BUILD_TIME)
	//
	// ## help: Show this help message
	// help:
	// 	@grep -E '^## ' Makefile | sed 's/## //'
	//
	// ## build: Build the binary
	// build:
	// 	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/server
	//
	// ## test: Run all tests
	// test:
	// 	go test -race -cover -coverprofile=coverage.out ./...
	//
	// ## lint: Run linters
	// lint:
	// 	golangci-lint run --timeout=5m ./...
	//
	// ## fmt: Format code
	// fmt:
	// 	gofumpt -w .
	// 	goimports -w .
	//
	// ## generate: Run code generation
	// generate:
	// 	go generate ./...
	//
	// ## clean: Remove build artifacts
	// clean:
	// 	rm -f $(BINARY) coverage.out
	//
	// ## run: Build and run
	// run: build
	// 	./$(BINARY)
	//
	// ## docker: Build Docker image
	// docker:
	// 	docker build -t $(BINARY):$(VERSION) .
	//
	// ## coverage: Open coverage in browser
	// coverage: test
	// 	go tool cover -html=coverage.out
	//
	// ## check: Run all checks (format, lint, test)
	// check: fmt lint test
	//
	// ## vuln: Run vulnerability check
	// vuln:
	// 	govulncheck ./...
	//
	// ## all: Full CI pipeline locally
	// all: fmt generate lint test build
	fmt.Println("  make build/test/lint/clean/run → standard targets")
	fmt.Println("  Inject version via LDFLAGS")
	fmt.Println("  make check → fmt + lint + test (pre-commit)")
	fmt.Println()
}

// =============================================================================
// PART 5: Docker Builds
// =============================================================================
func dockerBuilds() {
	fmt.Println("--- DOCKER BUILDS ---")
	// ─── Multi-stage build (the standard pattern) ───
	//
	// # Dockerfile
	// # Stage 1: Build
	// FROM golang:1.22-alpine AS builder
	// WORKDIR /app
	// COPY go.mod go.sum ./
	// RUN go mod download            # cache deps layer
	// COPY . .
	// RUN CGO_ENABLED=0 GOOS=linux \
	//     go build -trimpath -ldflags "-s -w" \
	//     -o /app/myapp ./cmd/server
	//
	// # Stage 2: Run
	// FROM alpine:3.19
	// RUN apk --no-cache add ca-certificates tzdata
	// COPY --from=builder /app/myapp /usr/local/bin/myapp
	// RUN addgroup -S app && adduser -S app -G app
	// USER app
	// EXPOSE 8080
	// ENTRYPOINT ["myapp"]
	//
	// Result: ~15-20 MB image (vs ~1 GB with golang base)
	//
	// ─── Scratch image (smallest possible) ───
	// FROM scratch
	// COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
	// COPY --from=builder /app/myapp /myapp
	// ENTRYPOINT ["/myapp"]
	//
	// Result: ~7 MB (just the binary + CA certs)
	// No shell, no package manager, minimal attack surface.
	//
	// ─── Distroless (Google, recommended) ───
	// FROM gcr.io/distroless/static-debian12
	// COPY --from=builder /app/myapp /myapp
	// ENTRYPOINT ["/myapp"]
	//
	// No shell, no package manager, but includes CA certs and tzdata.
	//
	// ─── Docker cache optimization ───
	// COPY go.mod go.sum first, then RUN go mod download.
	// This layer is cached until go.mod/go.sum change.
	// Source changes only rebuild the final COPY + build layer.
	//
	// ─── Security best practices ───
	// 1. Run as non-root user (USER app)
	// 2. Use specific base image tags (not :latest)
	// 3. Scan with: docker scout cves myapp:latest
	// 4. Don't copy .git, secrets, or test files (use .dockerignore)
	fmt.Println("  Multi-stage: golang → build → alpine/scratch → run")
	fmt.Println("  Result: ~15 MB image with alpine, ~7 MB with scratch")
	fmt.Println("  Cache: COPY go.mod/sum first, then go mod download")
	fmt.Println("  Security: non-root user, specific tags, scan")
	fmt.Println()
}

// =============================================================================
// PART 6: CI Pipeline
// =============================================================================
func ciPipeline() {
	fmt.Println("--- CI PIPELINE ---")
	// ─── Recommended CI stages ───
	//
	// 1. SETUP
	//    - Checkout code
	//    - Setup Go (actions/setup-go)
	//    - Cache module downloads
	//
	// 2. BUILD CHECK
	//    go build ./...
	//
	// 3. LINT
	//    golangci-lint run ./...
	//
	// 4. TEST
	//    go test -race -cover -coverprofile=coverage.out ./...
	//
	// 5. VULNERABILITY SCAN
	//    govulncheck ./...
	//
	// 6. BUILD BINARY
	//    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o app
	//
	// 7. BUILD & PUSH DOCKER (on main/tags)
	//    docker build -t registry/app:$TAG .
	//    docker push registry/app:$TAG
	//
	// ─── GitHub Actions example ───
	// name: CI
	// on: [push, pull_request]
	// jobs:
	//   ci:
	//     runs-on: ubuntu-latest
	//     steps:
	//       - uses: actions/checkout@v4
	//       - uses: actions/setup-go@v5
	//         with:
	//           go-version: '1.22'
	//           cache: true
	//       - run: go build ./...
	//       - uses: golangci/golangci-lint-action@v4
	//       - run: go test -race -coverprofile=coverage.out ./...
	//       - run: go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck ./...
	//
	// ─── Branch strategy ───
	// PR: build + lint + test
	// main: build + lint + test + docker build + push
	// tags: build + lint + test + release binaries (goreleaser)
	fmt.Println("  CI: build → lint → test → vulncheck → docker")
	fmt.Println("  actions/setup-go + golangci-lint-action")
	fmt.Println("  Always: -race and -cover in CI")
	fmt.Println()
}

// =============================================================================
// PART 7: Release Automation (goreleaser)
// =============================================================================
func releaseAutomation() {
	fmt.Println("--- RELEASE AUTOMATION ---")
	// ─── goreleaser: the standard for Go releases ───
	// go install github.com/goreleaser/goreleaser@latest
	//
	// goreleaser init              # create .goreleaser.yml
	// goreleaser release --snapshot --clean  # test locally
	// goreleaser release           # create release (needs git tag)
	//
	// ─── What goreleaser does ───
	// 1. Cross-compiles for all platforms
	// 2. Creates archives (.tar.gz, .zip)
	// 3. Generates checksums
	// 4. Creates GitHub/GitLab release
	// 5. Builds Docker images
	// 6. Generates changelog
	// 7. Publishes to Homebrew taps
	//
	// ─── .goreleaser.yml example ───
	// project_name: myapp
	// builds:
	//   - env: [CGO_ENABLED=0]
	//     goos: [linux, darwin, windows]
	//     goarch: [amd64, arm64]
	//     ldflags:
	//       - -s -w
	//       - -X main.version={{.Version}}
	//       - -X main.commit={{.Commit}}
	// archives:
	//   - format_overrides:
	//       - goos: windows
	//         format: zip
	// dockers:
	//   - image_templates:
	//       - "ghcr.io/user/myapp:{{.Version}}"
	//       - "ghcr.io/user/myapp:latest"
	// changelog:
	//   sort: asc
	//   filters:
	//     exclude: ["^docs:", "^test:", "^chore:"]
	//
	// ─── Release workflow ───
	// git tag v1.2.3
	// git push origin v1.2.3
	// # CI runs goreleaser automatically on tag push
	//
	// ─── GitHub Actions for goreleaser ───
	// on:
	//   push:
	//     tags: ['v*']
	// jobs:
	//   release:
	//     runs-on: ubuntu-latest
	//     steps:
	//       - uses: actions/checkout@v4
	//         with: {fetch-depth: 0}
	//       - uses: actions/setup-go@v5
	//         with: {go-version: '1.22'}
	//       - uses: goreleaser/goreleaser-action@v5
	//         with:
	//           version: latest
	//           args: release --clean
	//         env:
	//           GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
	fmt.Println("  goreleaser → cross-compile, release, docker, homebrew")
	fmt.Println("  git tag v1.2.3 → CI builds and releases")
	fmt.Println("  .goreleaser.yml → configure platforms and outputs")
	fmt.Println()
}
