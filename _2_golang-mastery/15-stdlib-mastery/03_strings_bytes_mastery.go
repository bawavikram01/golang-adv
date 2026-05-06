//go:build ignore

// =============================================================================
// LESSON 15.3: strings, bytes, strconv — Text Processing Mastery
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - strings package: every function that matters
// - bytes package: the same API but for []byte
// - strings.Builder: the ONLY way to build strings in hot paths
// - strings.Replacer: fast multi-string replacement
// - strconv: type conversions without fmt (faster, no alloc)
// - unicode/utf8: working with runes correctly
// - When to use string vs []byte vs strings.Builder
//
// THE KEY INSIGHT:
// Strings in Go are immutable byte slices. Every concatenation allocates.
// The strings and bytes packages provide the same API for different types.
// Knowing WHICH function to use (and when to use []byte instead) is the
// difference between 0 allocations and thousands.
//
// RUN: go run 03_strings_bytes_mastery.go
// =============================================================================

package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

func main() {
	fmt.Println("=== STRINGS, BYTES, STRCONV MASTERY ===")
	fmt.Println()

	stringInternals()
	stringsPackage()
	stringsBuilder()
	bytesPackage()
	strconvMastery()
	utf8Mastery()
	performancePatterns()
}

// =============================================================================
// PART 1: String Internals — What a string really is
// =============================================================================
func stringInternals() {
	fmt.Println("--- STRING INTERNALS ---")

	// A Go string is a (pointer, length) pair — an immutable view of bytes.
	//
	// type stringHeader struct {
	//     Data unsafe.Pointer  // points to byte array
	//     Len  int
	// }
	//
	// KEY PROPERTIES:
	// - Immutable: you cannot modify a string's bytes in place
	// - UTF-8 encoded (by convention, not enforced)
	// - len(s) returns BYTE count, not character/rune count
	// - Indexing s[i] returns a BYTE, not a rune
	// - Substring s[i:j] is O(1) — shares underlying bytes (no copy!)
	// - Comparison is byte-by-byte

	s := "Hello, 世界!" // 13 bytes, 9 runes

	fmt.Printf("  String: %q\n", s)
	fmt.Printf("  len() = %d bytes (NOT rune count)\n", len(s))
	fmt.Printf("  utf8.RuneCountInString() = %d runes\n", utf8.RuneCountInString(s))
	fmt.Printf("  s[0] = %d (byte 'H', not rune)\n", s[0])

	// Range over string iterates RUNES (not bytes!)
	fmt.Print("  Range runes: ")
	for i, r := range s {
		fmt.Printf("[%d:%c] ", i, r) // i is byte offset, r is rune
	}
	fmt.Println()

	// Substring: O(1) — shares underlying memory
	sub := s[7:] // "世界!"
	fmt.Printf("  Substring s[7:]: %q\n", sub)

	// string ↔ []byte conversion COPIES data (Go 1.22+ may optimize some cases)
	b := []byte(s)  // allocates + copies
	s2 := string(b) // allocates + copies
	fmt.Printf("  []byte roundtrip: %q\n", s2)

	fmt.Println()
}

// =============================================================================
// PART 2: strings Package — The Essential Functions
// =============================================================================
func stringsPackage() {
	fmt.Println("--- STRINGS PACKAGE ---")

	// ─── SEARCHING ───
	fmt.Println("  Searching:")
	fmt.Printf("    Contains(%q, %q) = %v\n", "seafood", "foo", strings.Contains("seafood", "foo"))
	fmt.Printf("    ContainsAny(%q, %q) = %v\n", "hello", "aeiou", strings.ContainsAny("hello", "aeiou"))
	fmt.Printf("    HasPrefix(%q, %q) = %v\n", "hello world", "hello", strings.HasPrefix("hello world", "hello"))
	fmt.Printf("    HasSuffix(%q, %q) = %v\n", "image.png", ".png", strings.HasSuffix("image.png", ".png"))
	fmt.Printf("    Index(%q, %q) = %d\n", "go gopher", "go", strings.Index("go gopher", "go"))
	fmt.Printf("    LastIndex(%q, %q) = %d\n", "go gopher go", "go", strings.LastIndex("go gopher go", "go"))
	fmt.Printf("    Count(%q, %q) = %d\n", "banana", "a", strings.Count("banana", "a"))

	// ─── SPLITTING ───
	fmt.Println("  Splitting:")
	fmt.Printf("    Split: %q\n", strings.Split("a,b,,c", ","))          // ["a" "b" "" "c"] — keeps empty
	fmt.Printf("    Fields: %q\n", strings.Fields("  foo  bar  baz "))   // ["foo" "bar" "baz"] — splits on whitespace, no empties
	fmt.Printf("    SplitN: %q\n", strings.SplitN("a,b,c,d", ",", 3))    // ["a" "b" "c,d"] — limit splits
	fmt.Printf("    SplitAfter: %q\n", strings.SplitAfter("a,b,c", ",")) // ["a," "b," "c"] — keeps delimiter

	// ─── FieldsFunc: custom split logic ───
	f := strings.FieldsFunc("foo1bar2baz", func(r rune) bool {
		return unicode.IsDigit(r)
	})
	fmt.Printf("    FieldsFunc (digits): %q\n", f) // ["foo" "bar" "baz"]

	// ─── JOINING ───
	fmt.Println("  Joining:")
	parts := []string{"host=localhost", "port=5432", "db=mydb"}
	fmt.Printf("    Join: %q\n", strings.Join(parts, " "))

	// ─── TRANSFORMING ───
	fmt.Println("  Transforming:")
	fmt.Printf("    ToUpper: %q\n", strings.ToUpper("hello"))
	fmt.Printf("    ToLower: %q\n", strings.ToLower("HELLO"))
	fmt.Printf("    Title (deprecated): use golang.org/x/text/cases instead\n")
	fmt.Printf("    Repeat: %q\n", strings.Repeat("ab", 3))                 // "ababab"
	fmt.Printf("    Replace: %q\n", strings.Replace("aaa", "a", "b", 2))    // "bba" (limit 2)
	fmt.Printf("    ReplaceAll: %q\n", strings.ReplaceAll("aaa", "a", "b")) // "bbb"

	// ─── TRIMMING ───
	fmt.Println("  Trimming:")
	fmt.Printf("    TrimSpace: %q\n", strings.TrimSpace("  hello  "))
	fmt.Printf("    Trim: %q\n", strings.Trim("***hello***", "*"))
	fmt.Printf("    TrimLeft: %q\n", strings.TrimLeft("***hello***", "*"))
	fmt.Printf("    TrimRight: %q\n", strings.TrimRight("***hello***", "*"))
	fmt.Printf("    TrimPrefix: %q\n", strings.TrimPrefix("hello world", "hello "))
	fmt.Printf("    TrimSuffix: %q\n", strings.TrimSuffix("file.go", ".go"))

	// ─── CUT: Split into two at first occurrence (Go 1.18+) ───
	before, after, found := strings.Cut("user:password:extra", ":")
	fmt.Printf("    Cut: before=%q, after=%q, found=%v\n", before, after, found)
	// "user", "password:extra", true — much cleaner than Index+slice

	// CutPrefix, CutSuffix (Go 1.20+)
	rest, ok := strings.CutPrefix("/api/v1/users", "/api/v1")
	fmt.Printf("    CutPrefix: rest=%q, ok=%v\n", rest, ok)

	// ─── MAPPING ───
	mapped := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		return unicode.ToLower(r)
	}, "Hello World GO")
	fmt.Printf("  Map (spaces→_, lower): %q\n", mapped)

	// ─── strings.Replacer: fast multi-string replacement ───
	// Pre-compiled replacer — much faster than calling ReplaceAll multiple times
	r := strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
		`"`, "&quot;",
	)
	escaped := r.Replace(`<script>alert("xss")</script>`)
	fmt.Printf("  Replacer (HTML escape): %q\n", escaped)

	// strings.NewReader: turns a string into an io.Reader
	// (covered in 01_io_mastery.go — key bridge between strings and io)

	fmt.Println()
}

// =============================================================================
// PART 3: strings.Builder — Zero-Waste String Construction
// =============================================================================
func stringsBuilder() {
	fmt.Println("--- STRINGS.BUILDER ---")

	// WHY: String concatenation with + allocates a NEW string every time.
	// "a" + "b" + "c" = 3 allocations (each intermediate result is a new string).
	//
	// strings.Builder batches writes into an internal []byte,
	// then converts to string with ZERO copy at the end.
	//
	// RULES:
	// 1. Don't copy a Builder after writing to it (compile error in Go 1.22+)
	// 2. Call Grow(n) if you know the final size (avoids realloc)
	// 3. Call Reset() to reuse

	// ─── Basic usage ───
	var b strings.Builder
	b.Grow(64) // pre-allocate capacity

	b.WriteString("SELECT ")
	b.WriteString("id, name, email")
	b.WriteString(" FROM users")
	b.WriteString(" WHERE active = true")

	fmt.Printf("  Builder: %q\n", b.String())
	fmt.Printf("  Len=%d, Cap=%d\n", b.Len(), b.Cap())

	// ─── Building with formatting ───
	b.Reset()
	for i := 0; i < 5; i++ {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "item_%d", i) // Builder is an io.Writer!
	}
	fmt.Printf("  Formatted: %q\n", b.String())

	// ─── Builder vs bytes.Buffer ───
	// Builder is ONLY for building strings (String() method, no ReadFrom, etc.)
	// bytes.Buffer is a full read/write buffer (implements io.Reader + io.Writer)
	//
	// Builder advantage: String() returns without copy (unsafe trick internally)
	// Buffer advantage: more methods (Read, ReadFrom, WriteTo, Next, etc.)
	//
	// RULE: Building a string → strings.Builder
	//        I/O buffer       → bytes.Buffer

	fmt.Println()
}

// =============================================================================
// PART 4: bytes Package — strings API for []byte
// =============================================================================
func bytesPackage() {
	fmt.Println("--- BYTES PACKAGE ---")

	// The bytes package mirrors strings but works on []byte.
	// Use when you're already working with []byte (network data, file data)
	// to avoid string↔[]byte conversions.

	data := []byte("Hello, World! Hello, Go!")

	fmt.Printf("  Contains: %v\n", bytes.Contains(data, []byte("Go")))
	fmt.Printf("  Count: %d\n", bytes.Count(data, []byte("Hello")))
	fmt.Printf("  HasPrefix: %v\n", bytes.HasPrefix(data, []byte("Hello")))

	// Split
	csv := []byte("a,b,c")
	parts := bytes.Split(csv, []byte(","))
	fmt.Printf("  Split: %d parts\n", len(parts))

	// Join
	joined := bytes.Join(parts, []byte(" | "))
	fmt.Printf("  Join: %q\n", joined)

	// Trim
	fmt.Printf("  TrimSpace: %q\n", bytes.TrimSpace([]byte("  hello  ")))

	// ─── bytes.Buffer: the read/write Swiss Army knife ───
	var buf bytes.Buffer

	buf.WriteString("Hello")
	buf.WriteByte(' ')
	buf.Write([]byte("World"))

	fmt.Printf("  Buffer write: %q\n", buf.String())

	// Read from buffer
	p := make([]byte, 5)
	buf.Read(p)
	fmt.Printf("  Buffer read: %q (remaining: %q)\n", p, buf.String())

	// ─── bytes.NewReader: read-only Reader over []byte ───
	// bytes.NewReader is cheaper than bytes.NewBuffer for read-only access
	// because it doesn't copy and supports ReadAt/Seek
	reader := bytes.NewReader([]byte("seek and read"))
	fmt.Printf("  bytes.NewReader: Len=%d, supports Seek+ReadAt\n", reader.Len())

	// ─── WHEN TO USE WHAT ───
	// bytes.NewReader: read-only access to []byte (supports Seek)
	// bytes.NewBuffer: read+write buffer (no Seek)
	// strings.NewReader: read-only access to string

	fmt.Println()
}

// =============================================================================
// PART 5: strconv — Type Conversion Without fmt
// =============================================================================
func strconvMastery() {
	fmt.Println("--- STRCONV MASTERY ---")

	// fmt.Sprintf is convenient but SLOW (reflection, allocations).
	// strconv functions are 2-10x faster with zero reflection.
	// USE strconv in hot paths (HTTP handlers, serialization, logging).

	// ─── int ↔ string ───
	s := strconv.Itoa(42)      // int → string (fast)
	n, _ := strconv.Atoi("42") // string → int (fast)
	fmt.Printf("  Itoa(42) = %q, Atoi(\"42\") = %d\n", s, n)

	// ─── ParseInt: with base and bit size ───
	n64, _ := strconv.ParseInt("FF", 16, 64)         // hex string → int64
	fmt.Printf("  ParseInt(\"FF\", 16) = %d\n", n64) // 255

	n64_2, _ := strconv.ParseInt("-128", 10, 8) // 8-bit range check
	fmt.Printf("  ParseInt(\"-128\", 10, 8) = %d\n", n64_2)

	// ─── FormatInt: int64 → string with base ───
	hex := strconv.FormatInt(255, 16)
	bin := strconv.FormatInt(42, 2)
	fmt.Printf("  FormatInt(255, 16) = %q, FormatInt(42, 2) = %q\n", hex, bin)

	// ─── float ↔ string ───
	f, _ := strconv.ParseFloat("3.14159", 64)
	fmt.Printf("  ParseFloat(\"3.14159\") = %f\n", f)

	fs := strconv.FormatFloat(3.14159, 'f', 2, 64) // format, precision, bitSize
	fmt.Printf("  FormatFloat(3.14159, 'f', 2) = %q\n", fs)

	// Format verbs: 'f' = decimal, 'e' = scientific, 'g' = shortest
	fmt.Printf("  'f': %s, 'e': %s, 'g': %s\n",
		strconv.FormatFloat(0.00123, 'f', -1, 64),
		strconv.FormatFloat(0.00123, 'e', -1, 64),
		strconv.FormatFloat(0.00123, 'g', -1, 64),
	)

	// ─── bool ↔ string ───
	bs, _ := strconv.ParseBool("true")
	fmt.Printf("  ParseBool(\"true\") = %v\n", bs)
	// Accepts: "1", "t", "T", "TRUE", "true", "True" → true
	// Accepts: "0", "f", "F", "FALSE", "false", "False" → false

	fmt.Printf("  FormatBool(true) = %q\n", strconv.FormatBool(true))

	// ─── Quote/Unquote: Go string literal handling ───
	quoted := strconv.Quote("hello\tworld\n") // → `"hello\tworld\n"`
	fmt.Printf("  Quote: %s\n", quoted)

	unquoted, _ := strconv.Unquote(`"hello\tworld\n"`) // → "hello\tworld\n"
	fmt.Printf("  Unquote: %q\n", unquoted)

	// ─── AppendXxx: append to []byte without allocation ───
	// THE FASTEST WAY to convert types in hot loops
	buf := make([]byte, 0, 64)
	buf = append(buf, "id="...)
	buf = strconv.AppendInt(buf, 42, 10)
	buf = append(buf, "&name="...)
	buf = strconv.AppendQuote(buf, "Vikram")
	buf = append(buf, "&pi="...)
	buf = strconv.AppendFloat(buf, 3.14, 'f', 2, 64)
	fmt.Printf("  AppendXxx: %q\n", string(buf))
	// This builds the string with ZERO intermediate allocations!

	fmt.Println()
}

// =============================================================================
// PART 6: unicode/utf8 — Working with Runes Correctly
// =============================================================================
func utf8Mastery() {
	fmt.Println("--- UTF-8 MASTERY ---")

	// Go strings are byte sequences. Most of the time they contain valid UTF-8,
	// but the language doesn't enforce it.
	//
	// A "rune" is an alias for int32 — represents a Unicode code point.
	// UTF-8 encodes runes using 1-4 bytes:
	//   0xxxxxxx                          → 1 byte  (ASCII, 0-127)
	//   110xxxxx 10xxxxxx                 → 2 bytes (128-2047)
	//   1110xxxx 10xxxxxx 10xxxxxx        → 3 bytes (2048-65535, incl. CJK)
	//   11110xxx 10xxxxxx 10xxxxxx 10xxxxxx → 4 bytes (65536+, emoji)

	s := "Go🚀世界"

	fmt.Printf("  String: %q\n", s)
	fmt.Printf("  len() = %d bytes\n", len(s))
	fmt.Printf("  RuneCountInString = %d runes\n", utf8.RuneCountInString(s))

	// ─── Decode rune by rune ───
	fmt.Print("  Runes: ")
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		fmt.Printf("[%c:%d bytes] ", r, size)
		i += size
	}
	fmt.Println()

	// ─── Valid UTF-8 check ───
	fmt.Printf("  Valid UTF-8 (%q): %v\n", s, utf8.ValidString(s))
	fmt.Printf("  Valid UTF-8 (bad): %v\n", utf8.Valid([]byte{0xff, 0xfe}))

	// ─── Rune encoding ───
	buf := make([]byte, 4)
	n := utf8.EncodeRune(buf, '🚀')
	fmt.Printf("  EncodeRune('🚀'): %d bytes → %v\n", n, buf[:n])

	// ─── Common patterns ───

	// Safe string truncation (don't cut in the middle of a rune!)
	long := "Hello, 世界! 🌍"
	truncated := safeStringTruncate(long, 10)
	fmt.Printf("  Safe truncate (%d bytes): %q\n", 10, truncated)

	// Count visible characters (for display width, you'd need runewidth package)
	fmt.Printf("  Rune count of %q: %d\n", "café", utf8.RuneCountInString("café"))

	fmt.Println()
}

// safeStringTruncate truncates to at most maxBytes without breaking UTF-8.
func safeStringTruncate(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	// Walk backwards from maxBytes to find a valid rune boundary
	for maxBytes > 0 && !utf8.RuneStart(s[maxBytes]) {
		maxBytes--
	}
	return s[:maxBytes]
}

// =============================================================================
// PART 7: Performance Patterns
// =============================================================================
func performancePatterns() {
	fmt.Println("--- PERFORMANCE PATTERNS ---")

	// ─── PATTERN 1: Avoid repeated string concatenation ───
	// BAD:  s += "a" + "b" + "c"  (O(n²) allocations)
	// GOOD: strings.Builder with Grow
	var b strings.Builder
	b.Grow(26)
	for c := 'a'; c <= 'z'; c++ {
		b.WriteRune(c)
	}
	fmt.Printf("  Builder (26 chars): %q\n", b.String())

	// ─── PATTERN 2: strings.Join vs Builder ───
	// Use strings.Join when you have a []string already
	// Use strings.Builder when building from mixed types/logic
	items := []string{"one", "two", "three"}
	fmt.Printf("  Join: %q\n", strings.Join(items, ", "))
	// Join is faster because it pre-calculates total length

	// ─── PATTERN 3: Pre-compiled Replacer for repeated use ───
	// strings.NewReplacer builds an efficient replacement structure once
	sanitizer := strings.NewReplacer(
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
		"\x00", "",
	)
	cleaned := sanitizer.Replace("line1\nline2\ttab\x00null")
	fmt.Printf("  Replacer: %q\n", cleaned)

	// ─── PATTERN 4: Avoid string↔[]byte when possible ───
	// strings.Contains(s, sub)   → no allocation
	// bytes.Contains(b, sub)     → no allocation
	// strings.Contains(string(b), sub) → ALLOCATES! (converts []byte to string)
	//
	// The compiler CAN optimize some cases (like comparisons) to avoid copies,
	// but don't rely on it for complex expressions.
	fmt.Println("  Avoid string([]byte) in hot paths — use bytes.Contains instead")

	// ─── PATTERN 5: strconv.AppendXxx in tight loops ───
	// Fastest way to build formatted output without allocations
	result := make([]byte, 0, 128)
	for i := 0; i < 5; i++ {
		if i > 0 {
			result = append(result, ',')
		}
		result = strconv.AppendInt(result, int64(i*10), 10)
	}
	fmt.Printf("  AppendInt loop: %q\n", string(result))

	// ─── PATTERN 6: strings.Cut vs Index+Slice ───
	// Old way:
	header := "Content-Type: application/json"
	idx := strings.Index(header, ": ")
	if idx >= 0 {
		_ = header[:idx]   // key
		_ = header[idx+2:] // value
	}
	// New way (Go 1.18+):
	key, value, _ := strings.Cut(header, ": ")
	fmt.Printf("  Cut: key=%q, value=%q\n", key, value)

	// ─── PATTERN 7: EqualFold for case-insensitive comparison ───
	// MUCH faster than ToLower(a) == ToLower(b) (no allocation!)
	fmt.Printf("  EqualFold: %v\n", strings.EqualFold("Go", "go"))

	fmt.Println()
}
