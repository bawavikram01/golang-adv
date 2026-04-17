//go:build ignore

// =============================================================================
// LESSON 0.9: PACKAGES & MODULES — Organizing Go Code
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Package declaration and naming conventions
// - Exported vs unexported (public/private via capitalization)
// - import paths and aliases
// - Go modules: go.mod, go.sum, versioning
// - Module commands: go get, go mod tidy, go mod vendor
// - Internal packages, replace directives, workspace mode
// - init() functions and package initialization order
// - Best practices for package design
//
// NOTE: This file demonstrates concepts with comments and small examples.
// Module/build commands are shown as comments (not executable in a single file).
//
// RUN: go run 09_packages_modules.go
// =============================================================================

package main

import (
	"fmt"
)

func main() {
	fmt.Println("=== PACKAGES & MODULES ===")
	fmt.Println()

	packageBasics()
	exportRules()
	importPatterns()
	moduleSystem()
	goModFile()
	moduleCommands()
	versioningRules()
	internalPackages()
	initFunctions()
	packageDesign()
}

// =============================================================================
// PART 1: Package Basics
// =============================================================================
func packageBasics() {
	fmt.Println("--- PACKAGE BASICS ---")

	// ─── Every Go file starts with a package declaration ───
	// package main        // executable program
	// package http        // library package
	// package http_test   // external test package
	//
	// RULES:
	// 1. All files in a directory must have the same package name
	//    (exception: _test.go files can use package foo_test)
	// 2. Package name should match directory name
	//    myproject/server/ → package server
	// 3. Package main is special: it's the entry point (has func main())

	// ─── Package naming conventions ───
	// - Short, concise, lowercase, single word
	// - Good: http, fmt, json, sync, os
	// - Bad:  httpServer, string_utils, myHelpers
	// - No underscores, no camelCase, no plural
	// - Package name is part of the identifier:
	//   http.Client (not httputil.HTTPClient)
	//   json.Marshal (not json.JSONMarshal)

	fmt.Println("  package main = executable, all others = libraries")
	fmt.Println("  Names: short, lowercase, no underscores")
	fmt.Println("  All files in a directory = same package")
	fmt.Println()
}

// =============================================================================
// PART 2: Exported vs Unexported
// =============================================================================

// Exported: starts with uppercase → visible outside package
type Server struct {
	Addr string // exported field
	port int    // unexported field (only visible within this package)
}

// Exported method
func (s *Server) Start() {
	fmt.Printf("  Starting server on %s:%d\n", s.Addr, s.port)
}

// Unexported function (only visible in this package)
func helperFunction() string {
	return "I'm private"
}

func exportRules() {
	fmt.Println("--- EXPORTED VS UNEXPORTED ---")

	// THE RULE: Capitalization = visibility
	// Uppercase first letter → exported (public)
	// Lowercase first letter → unexported (private to package)
	//
	// This applies to EVERYTHING:
	// - Functions:  fmt.Println (exported), helper() (unexported)
	// - Types:      http.Server (exported), internalState (unexported)
	// - Fields:     s.Addr (exported), s.port (unexported)
	// - Methods:    s.Start() (exported), s.validate() (unexported)
	// - Variables:  MaxRetries (exported), defaultTimeout (unexported)
	// - Constants:  StatusOK (exported), bufferSize (unexported)

	s := &Server{Addr: "localhost", port: 8080}
	s.Start()
	fmt.Printf("  helperFunction: %s\n", helperFunction())

	// ─── Struct field visibility in JSON ───
	// Unexported fields are invisible to encoding/json
	// type User struct {
	//     Name     string `json:"name"`     // included in JSON
	//     password string `json:"-"`        // invisible to json anyway
	// }

	// ─── GOTCHA: embedding and visibility ───
	// If you embed an unexported type, its exported methods
	// become exported on the outer type. Surprises people.

	fmt.Println()
}

// =============================================================================
// PART 3: Import Patterns
// =============================================================================
func importPatterns() {
	fmt.Println("--- IMPORT PATTERNS ---")

	// ─── Standard import ───
	// import "fmt"
	// import "net/http"
	//
	// ─── Grouped imports (preferred style) ───
	// import (
	//     "fmt"           // stdlib
	//     "net/http"
	//
	//     "github.com/pkg/errors"   // third-party
	//
	//     "myproject/internal/db"   // local
	// )
	// Convention: stdlib, blank line, third-party, blank line, local

	// ─── Import alias ───
	// import (
	//     "crypto/rand"        // rand.Read()
	//     mrand "math/rand"    // mrand.Intn()
	// )
	// Used when two packages have the same name

	// ─── Blank import (side effects only) ───
	// import _ "image/png"          // registers PNG decoder
	// import _ "net/http/pprof"     // registers pprof handlers
	// import _ "github.com/lib/pq"  // registers postgres driver
	//
	// The _ import only runs init() functions, doesn't make
	// the package's names available.

	// ─── Dot import (avoid in production) ───
	// import . "fmt"
	// Println("no prefix needed")  // but nobody knows where it came from!
	// Only acceptable in test files for DSL-style testing

	fmt.Println("  Group: stdlib → third-party → local")
	fmt.Println("  Alias: resolve name conflicts")
	fmt.Println("  Blank import _: side effects only (init)")
	fmt.Println("  Dot import .: avoid except tests")
	fmt.Println()
}

// =============================================================================
// PART 4: Go Module System
// =============================================================================
func moduleSystem() {
	fmt.Println("--- MODULE SYSTEM ---")

	// A MODULE is a collection of packages versioned together.
	// Defined by go.mod at the root of the project.

	// ─── Create a new module ───
	// go mod init github.com/user/myproject
	//
	// This creates go.mod:
	// module github.com/user/myproject
	// go 1.22

	// ─── Module path conventions ───
	// github.com/user/repo        → public on GitHub
	// mycompany.com/team/service  → private company module
	// golang.org/x/tools          → Go extended stdlib
	//
	// For personal/learning: any path works
	// go mod init myproject       → fine for local projects

	// ─── Module vs Package ───
	// Module: the whole repo (has go.mod)
	//   github.com/gorilla/mux
	//
	// Package: a directory within a module
	//   github.com/gorilla/mux           (root package)
	//   github.com/gorilla/mux/internal  (internal package)

	fmt.Println("  Module = versioned collection of packages (go.mod)")
	fmt.Println("  go mod init <module-path>")
	fmt.Println("  Module path = import path prefix for all packages")
	fmt.Println()
}

// =============================================================================
// PART 5: go.mod and go.sum Files
// =============================================================================
func goModFile() {
	fmt.Println("--- go.mod & go.sum ---")

	// ─── go.mod structure ───
	// module github.com/user/myapp
	//
	// go 1.22
	//
	// require (
	//     github.com/gin-gonic/gin v1.9.1
	//     go.uber.org/zap v1.26.0
	// )
	//
	// require (
	//     // indirect dependencies (auto-managed by go mod tidy)
	//     github.com/json-iterator/go v1.1.12 // indirect
	// )

	// ─── go.sum: integrity checksums ───
	// Contains SHA-256 hashes for every dependency.
	// Ensures reproducible builds: same code, every time.
	// ALWAYS commit go.sum to version control.
	//
	// go.sum entries look like:
	// github.com/gin-gonic/gin v1.9.1 h1:4idEAncQnU5cB7...
	// github.com/gin-gonic/gin v1.9.1/go.mod h1:ReTOfc82...

	// ─── replace directive ───
	// Override where a module comes from:
	// replace github.com/user/lib => ../local-lib
	// replace github.com/user/lib => github.com/fork/lib v1.0.0
	//
	// Use cases:
	// - Local development with a fork
	// - Monorepo with local packages
	// - Testing fixes before publishing

	// ─── exclude directive ───
	// exclude github.com/user/lib v1.2.3  // skip a broken version

	fmt.Println("  go.mod: module path, Go version, dependencies")
	fmt.Println("  go.sum: checksums for reproducible builds")
	fmt.Println("  replace: override module source (local dev, forks)")
	fmt.Println("  ALWAYS commit both go.mod and go.sum")
	fmt.Println()
}

// =============================================================================
// PART 6: Module Commands
// =============================================================================
func moduleCommands() {
	fmt.Println("--- MODULE COMMANDS ---")

	// ─── Essential commands ───
	// go mod init <path>     Create new module
	// go mod tidy            Add missing, remove unused deps
	// go get <pkg>@<ver>     Add/update a dependency
	// go get -u <pkg>        Update to latest minor/patch
	// go get -u ./...        Update all dependencies
	// go mod download        Download deps to cache
	// go mod verify          Verify checksums match
	// go mod vendor          Copy deps into vendor/
	// go mod graph           Print dependency graph
	// go mod why <pkg>       Why is this dep needed?

	// ─── go get version queries ───
	// go get github.com/pkg/errors@v0.9.1    specific version
	// go get github.com/pkg/errors@latest    latest release
	// go get github.com/pkg/errors@master    specific branch
	// go get github.com/pkg/errors@abc1234   specific commit

	// ─── Workspace mode (Go 1.18+) ───
	// For working with multiple modules simultaneously:
	// go work init ./module-a ./module-b
	//
	// Creates go.work:
	// go 1.22
	// use (
	//     ./module-a
	//     ./module-b
	// )
	// DON'T commit go.work (personal development aid)

	fmt.Println("  go mod tidy     — sync deps with imports")
	fmt.Println("  go get pkg@ver  — add/update dependency")
	fmt.Println("  go mod vendor   — copy deps locally")
	fmt.Println("  go work init    — multi-module workspace")
	fmt.Println()
}

// =============================================================================
// PART 7: Versioning Rules (Semantic Import Versioning)
// =============================================================================
func versioningRules() {
	fmt.Println("--- VERSIONING ---")

	// Go uses Semantic Versioning: vMAJOR.MINOR.PATCH
	// v1.2.3 → Major=1, Minor=2, Patch=3

	// ─── The big rule: v2+ changes the import path ───
	// v0.x.x and v1.x.x:
	//   import "github.com/user/lib"
	//
	// v2.x.x:
	//   import "github.com/user/lib/v2"   // /v2 suffix!
	//
	// v3.x.x:
	//   import "github.com/user/lib/v3"
	//
	// WHY? Different major versions can coexist in the same program.
	// Your code can import both v1 and v2 simultaneously.

	// ─── Pre-v1: no stability guarantees ───
	// v0.x.x releases can break APIs freely.
	// Treat v0 modules as unstable.

	// ─── Minimum Version Selection (MVS) ───
	// Go picks the MINIMUM version that satisfies all requirements.
	// NOT the latest. This is different from npm/pip/cargo.
	//
	// If A requires lib >= v1.3.0 and B requires lib >= v1.5.0:
	// Go selects v1.5.0 (minimum that works for both)
	// NOT v1.9.0 (latest available)
	//
	// This makes builds more reproducible and predictable.

	fmt.Println("  Semantic versioning: vMAJOR.MINOR.PATCH")
	fmt.Println("  v2+ requires /v2 suffix in import path")
	fmt.Println("  Minimum Version Selection (not latest)")
	fmt.Println()
}

// =============================================================================
// PART 8: Internal Packages
// =============================================================================
func internalPackages() {
	fmt.Println("--- INTERNAL PACKAGES ---")

	// The `internal` directory restricts package visibility.
	//
	// myproject/
	// ├── internal/
	// │   ├── auth/       ← only importable by myproject and its children
	// │   └── db/         ← only importable by myproject and its children
	// ├── cmd/
	// │   └── server/     ← can import internal/*
	// └── pkg/
	//     └── api/        ← can import internal/*
	//
	// An external project CANNOT import myproject/internal/auth
	// The compiler enforces this.

	// ─── Common project layouts ───
	// myproject/
	// ├── cmd/            ← entry points (main packages)
	// │   ├── server/
	// │   │   └── main.go
	// │   └── cli/
	// │       └── main.go
	// ├── internal/       ← private packages
	// │   ├── handlers/
	// │   ├── models/
	// │   └── repository/
	// ├── pkg/            ← public reusable packages (optional)
	// │   └── client/
	// ├── go.mod
	// └── go.sum

	// ─── When to use internal/ ───
	// Always start with internal/. Move to pkg/ only when you
	// intentionally want external consumers to import it.
	// It's easier to make things public later than to take access away.

	fmt.Println("  internal/ → private, compiler-enforced")
	fmt.Println("  cmd/ → main packages (entry points)")
	fmt.Println("  pkg/ → public reusable library code")
	fmt.Println("  Default to internal/, expose deliberately")
	fmt.Println()
}

// =============================================================================
// PART 9: init() Functions
// =============================================================================

// init() runs automatically when the package is imported.
// It runs BEFORE main().
// Multiple init() functions can exist in a file (run in order).
// Multiple files with init() in a package: alphabetical file order.
func init() {
	// Runs before main()
	// Used for:
	// - Registering drivers
	// - Setting default values
	// - Validating configuration
}

func initFunctions() {
	fmt.Println("--- init() FUNCTIONS ---")

	// ─── Package initialization order ───
	// 1. All imported packages are initialized first (recursively)
	// 2. Package-level variables are initialized (in declaration order)
	// 3. init() functions run (in source file order within a package)
	// 4. main() runs
	//
	// For a dependency graph: A imports B imports C
	// C's init → B's init → A's init → main()

	// ─── Real-world init() uses ───
	// Database drivers:
	//   import _ "github.com/lib/pq"
	//   // pq's init() calls sql.Register("postgres", &Driver{})
	//
	// Image format decoders:
	//   import _ "image/png"
	//   // png's init() calls image.RegisterFormat("png", ...)

	// ─── AVOID: complex logic in init() ───
	// - Makes testing harder (runs before TestMain)
	// - Hidden side effects
	// - Can't return errors
	// - Order depends on import graph (fragile)
	//
	// Prefer explicit initialization:
	//   func SetupDB(dsn string) (*DB, error) { ... }
	// Over:
	//   func init() { db = connect(os.Getenv("DSN")) }

	fmt.Println("  init() runs before main(), after package vars")
	fmt.Println("  Used for driver registration via blank imports")
	fmt.Println("  Avoid complex logic in init() — prefer explicit setup")
	fmt.Println()
}

// =============================================================================
// PART 10: Package Design Principles
// =============================================================================
func packageDesign() {
	fmt.Println("--- PACKAGE DESIGN PRINCIPLES ---")

	// ─── 1. Name for what it provides, not what it contains ───
	// Good: package http    (provides HTTP functionality)
	// Bad:  package utils   (what does it provide? everything?)
	// Bad:  package common  (dumping ground)
	// Bad:  package helpers (not a coherent concept)
	fmt.Println("  1. Name for purpose, not contents (no 'utils')")

	// ─── 2. Small surface area ───
	// Export the minimum. Unexport everything else.
	// You can always export later; un-exporting breaks users.
	fmt.Println("  2. Export minimum API surface")

	// ─── 3. One package per concept ───
	// Don't put auth, db, and cache in one package.
	// Each package should have a single, clear responsibility.
	fmt.Println("  3. One concept per package")

	// ─── 4. Avoid circular imports ───
	// Go forbids circular imports (A→B→A): compile error.
	// Fix: extract shared types into a third package
	// Or: use interfaces to break the dependency
	fmt.Println("  4. No circular imports (use interfaces to break cycles)")

	// ─── 5. Package-level documentation ───
	// doc.go (or package comment in any file):
	// // Package json implements encoding and decoding of JSON
	// // as defined in RFC 7159.
	// package json
	fmt.Println("  5. Document the package (// Package foo ...)")

	// ─── 6. Avoid package-level state ───
	// Global variables make testing hard and create hidden coupling.
	// Prefer dependency injection.
	// Bad:  var db *sql.DB (package level)
	// Good: type UserRepo struct { db *sql.DB }
	fmt.Println("  6. Avoid global state — use dependency injection")

	fmt.Println()
}
