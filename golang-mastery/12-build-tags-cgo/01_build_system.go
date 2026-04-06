// =============================================================================
// LESSON 12: BUILD TAGS, CROSS-COMPILATION, CGo, AND LINKER TRICKS
// =============================================================================
//
// Go's build system has powerful features most developers never use:
//   - Build constraints (//go:build) for conditional compilation
//   - Cross-compilation (GOOS/GOARCH)
//   - CGo for calling C from Go
//   - Linker flags for embedding version info
//   - Go plugins (dynamically loaded code)
//   - //go:embed for embedding files at compile time
//
// =============================================================================

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"runtime"
)

// =============================================================================
// PART 1: Build Constraints (//go:build)
// =============================================================================
//
// Control which files are compiled based on OS, arch, Go version, or custom tags.
//
// SYNTAX (Go 1.17+):
//   //go:build linux && amd64
//   //go:build !windows
//   //go:build integration
//   //go:build go1.21
//   //go:build (linux || darwin) && amd64
//
// OLD SYNTAX (still works but prefer new):
//   // +build linux,amd64
//
// FILE NAMING CONVENTION (implicit build constraints):
//   file_linux.go          → only on linux
//   file_windows_amd64.go  → only on windows/amd64
//   file_test.go           → only during testing
//
// USAGE:
//   go build                           → default tags
//   go build -tags=integration         → include integration-tagged files
//   go build -tags="integration e2e"   → multiple tags
//
// EXAMPLE: Platform-specific code
//
// --- dns_linux.go ---
// //go:build linux
// package mypackage
// func resolveDNS() { /* use Linux-specific syscalls */ }
//
// --- dns_darwin.go ---
// //go:build darwin
// package mypackage
// func resolveDNS() { /* use macOS-specific APIs */ }
//
// --- dns_windows.go ---
// //go:build windows
// package mypackage
// func resolveDNS() { /* use Windows DNS APIs */ }

// =============================================================================
// PART 2: Cross-Compilation
// =============================================================================
//
// Go can cross-compile for any supported OS/arch WITHOUT external tools:
//
//   GOOS=linux   GOARCH=amd64 go build -o myapp-linux-amd64
//   GOOS=darwin  GOARCH=arm64 go build -o myapp-darwin-arm64
//   GOOS=windows GOARCH=amd64 go build -o myapp.exe
//
// List all supported targets:
//   go tool dist list
//
// Common targets:
//   linux/amd64, linux/arm64, linux/arm
//   darwin/amd64, darwin/arm64
//   windows/amd64, windows/arm64
//   freebsd/amd64
//   js/wasm, wasip1/wasm
//
// NOTE: Cross-compilation disables CGo by default (CGO_ENABLED=0).
// To enable CGo cross-compilation, you need a cross-compiler toolchain.

func showBuildInfo() {
	fmt.Println("=== Build Information ===")
	fmt.Printf("OS:      %s\n", runtime.GOOS)
	fmt.Printf("Arch:    %s\n", runtime.GOARCH)
	fmt.Printf("Go:      %s\n", runtime.Version())
	fmt.Printf("Compiler:%s\n", runtime.Compiler)
}

// =============================================================================
// PART 3: Linker Flags — Embed version info at build time
// =============================================================================
//
// Inject values into variables at link time with -ldflags:
//
//   go build -ldflags="-X main.version=1.2.3 -X main.commit=$(git rev-parse HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// Also useful:
//   -ldflags="-s -w"  → strip debug info & DWARF (smaller binary)

var (
	version   = "dev"      // overwritten by -ldflags
	commit    = "unknown"
	buildDate = "unknown"
)

func showVersion() {
	fmt.Println("\n=== Version Info (set via -ldflags) ===")
	fmt.Printf("Version:    %s\n", version)
	fmt.Printf("Commit:     %s\n", commit)
	fmt.Printf("Build Date: %s\n", buildDate)
	fmt.Println()
	fmt.Println("Build with:")
	fmt.Println(`  go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" .`)
}

// =============================================================================
// PART 4: //go:embed — Embed files into the binary
// =============================================================================
//
// Since Go 1.16, you can embed files directly into the binary.
// No more separate config files, templates, or static assets!

//go:embed templates/greeting.txt
var greetingFile string // single file as string

//go:embed templates/greeting.txt
var greetingBytes []byte // single file as []byte

//go:embed templates/*
var templateFS embed.FS // directory as filesystem

func demonstrateEmbed() {
	fmt.Println("\n=== go:embed ===")

	// Single file as string
	fmt.Printf("Greeting (string): %s", greetingFile)

	// Single file as bytes
	fmt.Printf("Greeting (bytes, len=%d): %s", len(greetingBytes), greetingBytes)

	// Walk embedded filesystem
	fmt.Println("Embedded files:")
	fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			data, _ := templateFS.ReadFile(path)
			fmt.Printf("  %s (%d bytes)\n", path, len(data))
		}
		return nil
	})
}

// =============================================================================
// PART 5: CGo — Calling C from Go
// =============================================================================
//
// CGo lets you call C code from Go. The C code can be inline or in separate files.
//
// IMPORTANT TRADEOFFS:
//   - Breaks cross-compilation (needs C compiler for target)
//   - CGo calls are ~100x slower than Go calls (goroutine stack switching)
//   - GC can't track C memory — you must free it manually
//   - Complicates build process
//
// EXAMPLE (uncomment to use — requires gcc):
//
// /*
// #include <stdlib.h>
// #include <string.h>
//
// int add(int a, int b) {
//     return a + b;
// }
//
// char* greet(const char* name) {
//     char* buf = malloc(256);
//     snprintf(buf, 256, "Hello from C, %s!", name);
//     return buf;
// }
// */
// import "C"
// import "unsafe"
//
// func callC() {
//     // Call C function
//     result := C.add(40, 2)
//     fmt.Printf("C.add(40, 2) = %d\n", result)
//
//     // Pass string to C
//     cName := C.CString("Vikram")  // Go string → C string (allocates!)
//     defer C.free(unsafe.Pointer(cName))  // MUST free C memory!
//
//     cGreeting := C.greet(cName)
//     defer C.free(unsafe.Pointer(cGreeting))
//
//     goGreeting := C.GoString(cGreeting)  // C string → Go string (copies)
//     fmt.Printf("C.greet() = %s\n", goGreeting)
// }
//
// RULES FOR CGO:
// 1. import "C" MUST be immediately after the C comment block (no blank lines)
// 2. C.CString allocates — YOU must call C.free
// 3. Never pass Go pointers to C that might be stored (GC can move them)
// 4. Use runtime.KeepAlive if C code accesses Go memory asynchronously
// 5. Consider using purego or FFI alternatives to avoid CGo overhead

// =============================================================================
// PART 6: Build Modes
// =============================================================================
//
// go build -buildmode=<mode>
//
// Modes:
//   default    — static binary (normal executable)
//   exe        — same as default
//   pie        — position-independent executable (more secure, required by some distros)
//   c-shared   — C shared library (.so/.dll) — lets C/Python/Ruby call Go code
//   c-archive  — C static library (.a) — link Go code into C programs
//   plugin     — Go plugin (.so) — dynamically loaded at runtime
//
// STATIC LINKING (no external dependencies):
//   CGO_ENABLED=0 go build -ldflags="-s -w" -o myapp
//
// SMALLEST BINARY:
//   CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o myapp
//   upx --best myapp  # further compress with UPX

// =============================================================================
// PART 7: Go Plugins (dynamic loading)
// =============================================================================
//
// Build a plugin:
//   go build -buildmode=plugin -o myplugin.so myplugin.go
//
// Load it:
//   p, err := plugin.Open("myplugin.so")
//   sym, err := p.Lookup("ProcessData")
//   processFn := sym.(func([]byte) []byte)
//   result := processFn(data)
//
// LIMITATIONS:
// - Linux/macOS only (no Windows)
// - Plugin must be built with same Go version
// - All dependencies must match exactly
// - Generally prefer interfaces + separate processes over plugins

// =============================================================================
// PART 8: Compiler Directives
// =============================================================================
//
// //go:noinline       — prevent function inlining
// //go:nosplit         — don't insert stack growth preamble (dangerous)
// //go:norace          — skip race detector for this function
// //go:noescape        — compiler hint: args don't escape (unsafe if wrong)
// //go:linkname        — access unexported symbols from other packages (DANGEROUS)
// //go:generate cmd    — run cmd during `go generate`
//
// EXAMPLE: //go:linkname (accessing runtime internals)
//
// //go:linkname nanotime runtime.nanotime
// func nanotime() int64
//
// WARNING: linkname bypasses all Go visibility rules and can break
// between Go versions. Only use for extreme debugging or compatibility.

func main() {
	showBuildInfo()
	showVersion()
	demonstrateEmbed()

	fmt.Println("\n=== BUILD COMMANDS CHEAT SHEET ===")
	fmt.Println()
	fmt.Println("# Cross-compile for Linux (from any OS):")
	fmt.Println("  GOOS=linux GOARCH=amd64 go build -o app-linux")
	fmt.Println()
	fmt.Println("# Smallest possible binary:")
	fmt.Println("  CGO_ENABLED=0 go build -ldflags='-s -w' -trimpath -o app")
	fmt.Println()
	fmt.Println("# Inject version info:")
	fmt.Println(`  go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD)"`)
	fmt.Println()
	fmt.Println("# Build with custom tags:")
	fmt.Println("  go build -tags=integration,debug")
	fmt.Println()
	fmt.Println("# Static binary (no libc dependency):")
	fmt.Println("  CGO_ENABLED=0 GOOS=linux go build -a -o app")
	fmt.Println()
	fmt.Println("# Build C shared library (for Python/Ruby/C interop):")
	fmt.Println("  go build -buildmode=c-shared -o libmycode.so")
	fmt.Println()
	fmt.Println("# Compile to WebAssembly:")
	fmt.Println("  GOOS=js GOARCH=wasm go build -o main.wasm")
}
