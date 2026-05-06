//go:build ignore

// =============================================================================
// GO TOOLCHAIN 7: DELVE (dlv) — The Go Debugger
// =============================================================================
//
// Delve is the standard debugger for Go. It understands goroutines,
// channels, interfaces, and Go's runtime. GDB works but Delve is
// purpose-built for Go.
//
// Install: go install github.com/go-delve/delve/cmd/dlv@latest
//
// RUN: go run 07_delve_debugger.go
// =============================================================================

package main

import "fmt"

func main() {
	fmt.Println("=== DELVE DEBUGGER ===")
	fmt.Println()
	delveBasics()
	breakpoints()
	steppingAndNavigation()
	inspectingState()
	goroutineDebugging()
	conditionalAndTracepoints()
	remoteDebugging()
	delveVsCodeIntegration()
}

// =============================================================================
// PART 1: Delve Basics
// =============================================================================
func delveBasics() {
	fmt.Println("--- DELVE BASICS ---")
	// ─── Launch modes ───
	// dlv debug                    # compile + debug current package
	// dlv debug ./cmd/server       # debug specific package
	// dlv debug -- -flag1 value    # pass args to program (after --)
	// dlv test                     # debug tests
	// dlv test -- -run TestFoo     # debug specific test
	// dlv exec ./myapp             # debug pre-compiled binary
	// dlv attach <pid>             # attach to running process
	// dlv core <binary> <corefile> # debug core dump
	// dlv connect <addr>           # connect to remote debugger
	//
	// ─── Build flags ───
	// dlv debug -gcflags="all=-N -l"   # disable optimizations
	//   -N: disable optimizations
	//   -l: disable inlining
	// Without these, debugger may skip lines or show <optimized out>.
	// Delve adds these automatically when you use `dlv debug`.
	//
	// ─── Basic commands ───
	// (dlv) help          # list all commands
	// (dlv) break main.go:25  # set breakpoint
	// (dlv) continue      # run until breakpoint
	// (dlv) next           # step over (next line)
	// (dlv) step           # step into function
	// (dlv) stepout        # step out of current function
	// (dlv) print x        # print variable
	// (dlv) quit           # exit debugger
	fmt.Println("  dlv debug → compile and debug")
	fmt.Println("  dlv test → debug tests")
	fmt.Println("  dlv attach <pid> → debug running process")
	fmt.Println()
}

// =============================================================================
// PART 2: Breakpoints
// =============================================================================
func breakpoints() {
	fmt.Println("--- BREAKPOINTS ---")
	// ─── Setting breakpoints ───
	// break main.go:25              # by file:line
	// break main.main               # by function name
	// break mypackage.MyFunc        # by package.function
	// break (*Server).HandleRequest # by method
	//
	// ─── Shorthand ───
	// b main.go:25                  # short for break
	// bp                            # list all breakpoints
	// clear 1                       # delete breakpoint #1
	// clearall                      # delete all breakpoints
	// toggle 1                      # enable/disable breakpoint
	//
	// ─── Conditional breakpoints ───
	// break main.go:25
	// condition 1 x > 100           # only trigger when x > 100
	// condition 1 name == "alice"   # string comparison
	// condition 1 err != nil        # only on error
	//
	// ─── Hit count ───
	// break main.go:25
	// condition 1 -hitcount 10      # trigger on 10th hit
	// condition 1 -hitcount >5      # trigger after 5th hit
	//
	// ─── On-hit commands ───
	// on 1 print x                  # auto-print x when breakpoint 1 hits
	// on 1 goroutine                # show goroutine info
	// on 1 stack                    # show stack trace
	//
	// ─── Breakpoint on panic ───
	// break runtime.gopanic         # break on any panic
	// This is extremely useful for catching panics before they unwind.
	fmt.Println("  break file:line or break package.Func")
	fmt.Println("  condition N expr → conditional breakpoint")
	fmt.Println("  break runtime.gopanic → catch panics")
	fmt.Println()
}

// =============================================================================
// PART 3: Stepping & Navigation
// =============================================================================
func steppingAndNavigation() {
	fmt.Println("--- STEPPING & NAVIGATION ---")
	// ─── Execution control ───
	// continue (c)    Run until next breakpoint
	// next (n)        Step over (execute line, skip into calls)
	// step (s)        Step into (enter function calls)
	// stepout (so)    Step out (run until current func returns)
	// restart (r)     Restart program from beginning
	//
	// ─── Step instruction ───
	// step-instruction (si)   Step one CPU instruction
	// Useful when debugging assembly or runtime code.
	//
	// ─── Reverse execution (experimental) ───
	// dlv debug --backend=rr   # uses Mozilla rr for recording
	// rev step                 # step backwards!
	// rev next                 # reverse next
	// rev continue             # run backwards to previous breakpoint
	// Requires: Linux + rr installed (mozilla/rr)
	// Records execution and replays it — can step backward in time!
	//
	// ─── Frame navigation ───
	// stack                     # show call stack
	// frame 3                   # switch to frame #3
	// up                        # go up one frame
	// down                      # go down one frame
	// deferred                  # show deferred calls in current frame
	fmt.Println("  next/step/stepout → basic stepping")
	fmt.Println("  stack → show call stack")
	fmt.Println("  frame N → switch to stack frame")
	fmt.Println("  rr backend → reverse debugging!")
	fmt.Println()
}

// =============================================================================
// PART 4: Inspecting State
// =============================================================================
func inspectingState() {
	fmt.Println("--- INSPECTING STATE ---")
	// ─── Print variables ───
	// print x                    # print variable x
	// print *ptr                 # dereference pointer
	// print mySlice[3]           # slice element
	// print myMap["key"]         # map value
	// print myStruct.Field       # struct field
	// print len(mySlice)         # built-in functions work
	//
	// ─── Local variables ───
	// locals                     # show all local vars
	// args                       # show function arguments
	//
	// ─── Evaluate expressions ───
	// print x + y                # arithmetic
	// print x > 5                # boolean expression
	// print fmt.Sprintf("%d", x) # call functions!
	// call myFunc(42)            # call a function
	//
	// ─── Set variables ───
	// set x = 42                 # change variable value
	// set myStruct.Field = "new" # change struct field
	//
	// ─── Type inspection ───
	// whatis x                   # show type of x
	// print reflect.TypeOf(x)    # runtime type (for interfaces)
	//
	// ─── Watchpoints (Go 1.22+ / Delve 1.22+) ───
	// watch x                    # break when x changes
	// watch -r x                 # break when x is read
	// watch -w x                 # break when x is written
	// Hardware-assisted — very efficient, limited number (usually 4).
	fmt.Println("  print x → inspect variable")
	fmt.Println("  locals/args → show all vars")
	fmt.Println("  set x = val → modify variable at runtime")
	fmt.Println("  watch x → break on variable change")
	fmt.Println()
}

// =============================================================================
// PART 5: Goroutine Debugging
// =============================================================================
func goroutineDebugging() {
	fmt.Println("--- GOROUTINE DEBUGGING ---")
	// ─── This is where Delve shines vs GDB ───
	// Delve understands Go's goroutine scheduler natively.
	//
	// goroutines                   # list all goroutines
	// goroutine                    # show current goroutine
	// goroutine 7                  # switch to goroutine 7
	// goroutine 7 bt              # backtrace of goroutine 7
	// goroutines -t                # list with stack traces
	// goroutines -g                # group by current function
	//
	// ─── Goroutine states ───
	// Running    Currently executing
	// Runnable   Ready to run, waiting for P
	// Waiting    Blocked (channel, mutex, syscall, timer)
	// Idle       Unused (in pool)
	//
	// ─── Filtering goroutines ───
	// goroutines -group user       # group by user function
	// goroutines -group go         # group by go statement location
	//
	// ─── Debug specific goroutine ───
	// break main.go:50
	// goroutine 7 continue         # continue only goroutine 7
	//
	// ─── Thread commands ───
	// threads                      # list OS threads
	// thread 3                     # switch to thread
	// Usually you debug goroutines, not threads.
	fmt.Println("  goroutines → list all goroutines")
	fmt.Println("  goroutine N → switch to goroutine N")
	fmt.Println("  goroutines -t → with stack traces")
	fmt.Println("  Delve understands Go scheduler natively")
	fmt.Println()
}

// =============================================================================
// PART 6: Conditional & Tracepoints
// =============================================================================
func conditionalAndTracepoints() {
	fmt.Println("--- CONDITIONAL & TRACEPOINTS ---")
	// ─── Tracepoint: log without stopping ───
	// break main.go:25
	// on 1 print fmt.Sprintf("x=%d, y=%d", x, y)
	// condition 1 -hitcount 0      # set hit count to never stop
	//   (or use trace instead of break)
	//
	// trace main.go:25             # print when hit, don't stop
	// trace main.handleRequest     # trace function entry
	//
	// Tracepoints are like adding fmt.Println without modifying code.
	//
	// ─── trace with message ───
	// trace main.go:25
	// on 1 print x, y, z           # print these vars when hit
	//
	// ─── Common debug scenarios ───
	//
	// "Why is this function called with nil?"
	// break main.processUser
	// condition 1 user == nil
	// continue
	//
	// "What's the 100th request look like?"
	// break handler.go:50
	// condition 1 -hitcount 100
	//
	// "Log all DB queries without modifying code"
	// trace db.Query
	// on 1 print query, args
	//
	// "Catch goroutine leak"
	// break runtime.newproc1       # break on goroutine creation
	// on 1 stack 5                 # show 5 frames of creator
	fmt.Println("  trace file:line → log without stopping")
	fmt.Println("  condition N expr → break only when expr is true")
	fmt.Println("  on N print vars → auto-print on breakpoint hit")
	fmt.Println()
}

// =============================================================================
// PART 7: Remote Debugging
// =============================================================================
func remoteDebugging() {
	fmt.Println("--- REMOTE DEBUGGING ---")
	// ─── Debug on remote server / container ───
	//
	// On the server:
	// dlv exec --headless --listen=:2345 --api-version=2 ./myapp
	//
	// Or attach to running process:
	// dlv attach --headless --listen=:2345 --api-version=2 <pid>
	//
	// From your machine:
	// dlv connect server:2345
	//
	// ─── Docker debugging ───
	// # Dockerfile
	// FROM golang:1.22
	// RUN go install github.com/go-delve/delve/cmd/dlv@latest
	// COPY . /app
	// WORKDIR /app
	// RUN go build -gcflags="all=-N -l" -o /app/myapp
	// EXPOSE 2345
	// CMD ["dlv", "exec", "--headless", "--listen=:2345",
	//      "--api-version=2", "--accept-multiclient", "/app/myapp"]
	//
	// docker run -p 2345:2345 myapp-debug
	// dlv connect localhost:2345
	//
	// ─── Kubernetes debugging ───
	// kubectl port-forward pod/myapp-debug 2345:2345
	// dlv connect localhost:2345
	//
	// ─── Security ───
	// NEVER expose dlv port publicly.
	// Always use port-forwarding or VPN.
	// Delve gives full access to process memory!
	fmt.Println("  dlv exec --headless --listen=:2345 → remote server")
	fmt.Println("  dlv connect host:2345 → connect from your machine")
	fmt.Println("  Works in Docker, K8s via port-forward")
	fmt.Println()
}

// =============================================================================
// PART 8: VS Code Integration
// =============================================================================
func delveVsCodeIntegration() {
	fmt.Println("--- VS CODE INTEGRATION ---")
	// VS Code uses Delve under the hood via the Go extension.
	//
	// ─── launch.json configuration ───
	// {
	//     "version": "0.2.0",
	//     "configurations": [
	//         {
	//             "name": "Launch",
	//             "type": "go",
	//             "request": "launch",
	//             "mode": "auto",
	//             "program": "${workspaceFolder}/cmd/server",
	//             "args": ["-port", "8080"],
	//             "env": {"DATABASE_URL": "postgres://..."}
	//         },
	//         {
	//             "name": "Test Current File",
	//             "type": "go",
	//             "request": "launch",
	//             "mode": "test",
	//             "program": "${file}"
	//         },
	//         {
	//             "name": "Attach Remote",
	//             "type": "go",
	//             "request": "attach",
	//             "mode": "remote",
	//             "remotePath": "/app",
	//             "port": 2345,
	//             "host": "localhost"
	//         }
	//     ]
	// }
	//
	// ─── VS Code debug features ───
	// - Click gutter to set breakpoints
	// - Hover variables to see values
	// - Variables panel shows locals/args
	// - Call stack panel shows goroutine stacks
	// - Debug console: evaluate expressions
	// - Conditional breakpoints: right-click breakpoint
	// - Logpoints: log without stopping (right-click → "Log Message")
	//
	// ─── Keyboard shortcuts ───
	// F5         Start/continue debugging
	// F9         Toggle breakpoint
	// F10        Step over
	// F11        Step into
	// Shift+F11  Step out
	// Shift+F5   Stop debugging
	// Ctrl+Shift+F5  Restart
	fmt.Println("  VS Code Go extension uses Delve automatically")
	fmt.Println("  F5=debug, F9=breakpoint, F10=step over, F11=step in")
	fmt.Println("  launch.json for configuration")
	fmt.Println()
}
