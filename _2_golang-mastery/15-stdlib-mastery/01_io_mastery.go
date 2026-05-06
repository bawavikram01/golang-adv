//go:build ignore

// =============================================================================
// LESSON 15.1: io.Reader & io.Writer — The Backbone of All Go Programs
// =============================================================================
//
// WHAT YOU'LL LEARN:
// - Why Reader/Writer are the most important interfaces in Go
// - The full io interface hierarchy (Reader → ReadCloser → ReadWriteCloser...)
// - Composing readers/writers like Unix pipes
// - Building custom readers and writers
// - Zero-copy techniques with io.Copy, io.Pipe
// - Buffered I/O with bufio (Scanner, Reader, Writer)
// - Multi-reader, TeeReader, LimitReader, SectionReader
// - Real-world patterns: progress tracking, rate limiting, checksumming
//
// THE KEY INSIGHT:
// In Go, EVERYTHING is a stream. Files, network connections, HTTP bodies,
// compressed data, encrypted data, JSON encoders — they ALL implement
// io.Reader or io.Writer. By programming to these interfaces, you can
// compose them like Lego blocks. This is Go's most powerful abstraction.
//
// RUN: go run 01_io_mastery.go
// =============================================================================

package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
)

func main() {
	fmt.Println("=== IO MASTERY ===")
	fmt.Println()

	interfaceHierarchy()
	readerFundamentals()
	writerFundamentals()
	composingReaders()
	composingWriters()
	bufioMastery()
	customReaderWriter()
	ioPipePattern()
	realWorldPatterns()
}

// =============================================================================
// PART 1: The io Interface Hierarchy
// =============================================================================
func interfaceHierarchy() {
	fmt.Println("--- IO INTERFACE HIERARCHY ---")

	// THE TWO FUNDAMENTAL INTERFACES:
	//
	// type Reader interface {
	//     Read(p []byte) (n int, err error)
	// }
	//
	// type Writer interface {
	//     Write(p []byte) (n int, err error)
	// }
	//
	// That's it. Two methods. But they compose into everything.
	//
	// THE FULL HIERARCHY:
	// ────────────────────
	//
	// Reader ──────────────┐
	//                      ├── ReadCloser  (Reader + Closer)
	// Closer ──────────────┤
	//                      ├── WriteCloser (Writer + Closer)
	// Writer ──────────────┤
	//                      ├── ReadWriter  (Reader + Writer)
	// Reader ──────────────┤
	//                      └── ReadWriteCloser (Reader + Writer + Closer)
	//
	// Additional interfaces:
	// - ReaderAt:   ReadAt(p []byte, off int64) (n int, err error)  ← random access
	// - WriterAt:   WriteAt(p []byte, off int64) (n int, err error)
	// - ReaderFrom: ReadFrom(r Reader) (n int64, err error)         ← efficient copy
	// - WriterTo:   WriteTo(w Writer) (n int64, err error)          ← efficient copy
	// - Seeker:     Seek(offset int64, whence int) (int64, error)   ← reposition
	// - ByteReader: ReadByte() (byte, error)                        ← single byte
	// - ByteWriter: WriteByte(c byte) error
	// - RuneReader: ReadRune() (r rune, size int, err error)        ← Unicode
	// - StringWriter: WriteString(s string) (n int, err error)      ← avoid []byte copy
	//
	// WHO IMPLEMENTS Reader?
	// ──────────────────────
	// *os.File             — files on disk
	// *bytes.Buffer        — in-memory buffer
	// *bytes.Reader        — read-only []byte wrapper
	// *strings.Reader      — read-only string wrapper
	// net.Conn             — TCP/UDP connections
	// *http.Request.Body   — HTTP request body (ReadCloser)
	// *http.Response.Body  — HTTP response body (ReadCloser)
	// *gzip.Reader         — decompresses on read
	// *bufio.Reader        — buffered reading
	// io.LimitReader(r, n) — reads at most n bytes
	// io.TeeReader(r, w)   — mirrors reads to a writer
	// io.MultiReader(...)  — concatenates readers
	// cipher.StreamReader   — decrypts on read
	// base64.NewDecoder     — decodes on read
	//
	// THIS IS THE POWER: the same code that reads a file can read
	// a network connection, a decompressed stream, or an in-memory buffer.

	fmt.Println("  Reader: the source (files, network, buffers, decompressors)")
	fmt.Println("  Writer: the sink (files, network, compressors, hashers)")
	fmt.Println("  Everything in Go is a stream you can compose")
	fmt.Println()
}

// =============================================================================
// PART 2: Reader Fundamentals — The Read Contract
// =============================================================================
func readerFundamentals() {
	fmt.Println("--- READER FUNDAMENTALS ---")

	// THE READ CONTRACT (most misunderstood thing in Go):
	// ────────────────────────────────────────────────────
	// Read(p []byte) (n int, err error)
	//
	// Rules:
	// 1. Read reads UP TO len(p) bytes. It may return FEWER.
	//    Getting n < len(p) is NOT an error. It's normal.
	//
	// 2. err == nil means: "n bytes are valid, may have more data"
	//    err == io.EOF means: "n bytes are valid, no more data"
	//    n > 0 AND err == io.EOF is VALID (last chunk + EOF together)
	//
	// 3. NEVER ignore n when err != nil. The n bytes ARE valid.
	//
	// 4. Callers should process n bytes BEFORE checking err.

	// WRONG way to read (ignores partial reads):
	// buf := make([]byte, 1024)
	// n, err := r.Read(buf)
	// if err != nil { return err }  // WRONG: dropped n valid bytes!
	// process(buf[:n])

	// RIGHT way to read all data:
	r := strings.NewReader("Hello, io.Reader!")
	buf := make([]byte, 5) // small buffer to show multiple reads

	fmt.Print("  Reading 5 bytes at a time: ")
	for {
		n, err := r.Read(buf)
		if n > 0 {
			// ALWAYS process n bytes first, even if err != nil
			fmt.Print(string(buf[:n]))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}
	}
	fmt.Println()

	// io.ReadAll: reads everything into memory (careful with large streams!)
	r2 := strings.NewReader("ReadAll is convenient but dangerous for large inputs")
	data, err := io.ReadAll(r2)
	if err != nil {
		fmt.Printf("  error: %v\n", err)
		return
	}
	fmt.Printf("  io.ReadAll: %q (%d bytes)\n", data, len(data))

	// io.ReadFull: reads exactly len(buf) bytes or fails
	r3 := strings.NewReader("exactly 10")
	exact := make([]byte, 10)
	n, err := io.ReadFull(r3, exact)
	fmt.Printf("  io.ReadFull: %q (n=%d, err=%v)\n", exact[:n], n, err)

	// io.ReadFull with short stream → ErrUnexpectedEOF
	r4 := strings.NewReader("short")
	exact2 := make([]byte, 10)
	n, err = io.ReadFull(r4, exact2)
	fmt.Printf("  io.ReadFull (short): %q (n=%d, err=%v)\n", exact2[:n], n, err)

	fmt.Println()
}

// =============================================================================
// PART 3: Writer Fundamentals — The Write Contract
// =============================================================================
func writerFundamentals() {
	fmt.Println("--- WRITER FUNDAMENTALS ---")

	// THE WRITE CONTRACT:
	// ────────────────────
	// Write(p []byte) (n int, err error)
	//
	// Rules:
	// 1. Write MUST return err != nil if n < len(p)
	//    (Unlike Read, short writes ARE errors for writers)
	//
	// 2. Write must NOT modify the slice p, even temporarily
	//
	// 3. Write must not retain p (don't save the pointer for later)

	// bytes.Buffer: the Swiss Army knife of in-memory I/O
	var buf bytes.Buffer

	// Write raw bytes
	buf.Write([]byte("Hello"))

	// WriteString avoids []byte allocation
	buf.WriteString(", World")

	// WriteByte for single bytes
	buf.WriteByte('!')

	// WriteRune for Unicode
	buf.WriteRune(' ')
	buf.WriteRune('🚀')

	fmt.Printf("  Buffer content: %q\n", buf.String())
	fmt.Printf("  Buffer length: %d bytes\n", buf.Len())

	// fmt.Fprintf writes to ANY io.Writer
	buf.Reset()
	fmt.Fprintf(&buf, "Name: %s, Age: %d", "Vikram", 25)
	fmt.Printf("  Fprintf to buffer: %q\n", buf.String())

	// io.WriteString: uses StringWriter interface if available (avoids alloc)
	buf.Reset()
	io.WriteString(&buf, "efficient string write")
	fmt.Printf("  io.WriteString: %q\n", buf.String())

	// os.Stdout is an io.Writer — print to terminal
	fmt.Fprint(os.Stdout, "  Writing directly to stdout\n")

	// io.Discard: /dev/null writer — discards everything (useful for benchmarks)
	n, _ := io.WriteString(io.Discard, "this goes nowhere")
	fmt.Printf("  io.Discard: wrote %d bytes to nowhere\n", n)

	fmt.Println()
}

// =============================================================================
// PART 4: Composing Readers — The Unix Philosophy in Go
// =============================================================================
func composingReaders() {
	fmt.Println("--- COMPOSING READERS ---")

	// Like Unix pipes (cat file | grep foo | sort), Go readers compose.
	// Each reader wraps another, adding behavior.

	// ─── io.LimitReader: read at most N bytes ───
	// Prevents reading unbounded input (OOM protection)
	full := strings.NewReader("This is a long string that we want to limit")
	limited := io.LimitReader(full, 14)
	data, _ := io.ReadAll(limited)
	fmt.Printf("  LimitReader(14): %q\n", data)

	// ─── io.MultiReader: concatenate multiple readers ───
	// Like `cat file1 file2 file3`
	r1 := strings.NewReader("Hello ")
	r2 := strings.NewReader("from ")
	r3 := strings.NewReader("MultiReader!")
	multi := io.MultiReader(r1, r2, r3)
	data, _ = io.ReadAll(multi)
	fmt.Printf("  MultiReader: %q\n", data)

	// ─── io.TeeReader: mirror reads to a writer ───
	// Like the Unix `tee` command: read data AND write a copy elsewhere
	source := strings.NewReader("data flowing through tee")
	var mirror bytes.Buffer
	tee := io.TeeReader(source, &mirror)

	// Read from tee → data also written to mirror
	data, _ = io.ReadAll(tee)
	fmt.Printf("  TeeReader read: %q\n", data)
	fmt.Printf("  TeeReader mirror: %q\n", mirror.String())

	// ─── REAL USE CASE: Hash while reading ───
	// Read a file (or HTTP body) and compute SHA256 simultaneously
	content := strings.NewReader("compute hash while reading, zero extra passes")
	hasher := sha256.New()
	tee2 := io.TeeReader(content, hasher) // reads flow through hasher

	result, _ := io.ReadAll(tee2) // read all data
	hash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("  Hash-while-read: %q → SHA256: %s...\n", result, hash[:16])

	// ─── io.SectionReader: read a slice of a ReaderAt ───
	// Random access to a portion of data (like reading a file segment)
	fullData := strings.NewReader("HEADER|PAYLOAD_DATA|FOOTER")
	// Read bytes 7 through 19 (the PAYLOAD_DATA section)
	section := io.NewSectionReader(fullData, 7, 12)
	payload, _ := io.ReadAll(section)
	fmt.Printf("  SectionReader(7,12): %q\n", payload)

	fmt.Println()
}

// =============================================================================
// PART 5: Composing Writers — Layering output transforms
// =============================================================================
func composingWriters() {
	fmt.Println("--- COMPOSING WRITERS ---")

	// ─── io.MultiWriter: write to multiple destinations ───
	// Like tee: one write goes to all writers
	var buf1, buf2 bytes.Buffer
	multi := io.MultiWriter(&buf1, &buf2)
	fmt.Fprint(multi, "written to both")
	fmt.Printf("  MultiWriter: buf1=%q, buf2=%q\n", buf1.String(), buf2.String())

	// ─── REAL USE CASE: Write to file + compute hash simultaneously ───
	var output bytes.Buffer
	hasher := sha256.New()
	combo := io.MultiWriter(&output, hasher)

	fmt.Fprint(combo, "data written to output AND hashed")
	hash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("  Write+Hash: output=%d bytes, SHA256=%s...\n", output.Len(), hash[:16])

	// ─── Stacked writers: compress → output ───
	// This is how Go composes transforms:
	// raw data → gzip compressor → buffer
	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	gzWriter.Write([]byte("compress this data with gzip, it saves space!"))
	gzWriter.Close() // MUST close to flush and write gzip footer

	fmt.Printf("  Gzip: original ~46 bytes → compressed %d bytes\n", compressed.Len())

	// Decompress it back
	gzReader, _ := gzip.NewReader(&compressed)
	decompressed, _ := io.ReadAll(gzReader)
	gzReader.Close()
	fmt.Printf("  Gunzip: %q\n", decompressed)

	// ─── io.Copy: the most efficient way to move data ───
	// io.Copy checks for WriterTo/ReaderFrom interfaces for zero-copy
	src := strings.NewReader("efficiently copied data")
	var dst bytes.Buffer
	n, _ := io.Copy(&dst, src)
	fmt.Printf("  io.Copy: %d bytes → %q\n", n, dst.String())

	// io.CopyN: copy exactly N bytes
	src2 := strings.NewReader("copy only the first 10 bytes of this")
	var dst2 bytes.Buffer
	n, _ = io.CopyN(&dst2, src2, 10)
	fmt.Printf("  io.CopyN(10): %d bytes → %q\n", n, dst2.String())

	// ─── WHY io.Copy is special ───
	// io.Copy doesn't just loop Read/Write.
	// It checks:
	//   1. Does src implement WriterTo? → calls src.WriteTo(dst)
	//   2. Does dst implement ReaderFrom? → calls dst.ReadFrom(src)
	//   3. Fall back to allocating a 32KB buffer and looping Read/Write
	//
	// *os.File implements ReaderFrom using sendfile(2) syscall on Linux
	// → data flows kernel-to-kernel, never enters user space!
	// This is WHY io.Copy from file to socket is extremely fast.

	fmt.Println()
}

// =============================================================================
// PART 6: bufio — Buffered I/O for Performance
// =============================================================================
func bufioMastery() {
	fmt.Println("--- BUFIO MASTERY ---")

	// WHY BUFIO?
	// ───────────
	// Every Read() syscall is expensive (~100ns on Linux).
	// Reading one byte at a time from a file = one syscall per byte.
	// bufio.Reader reads 4KB chunks, serves individual reads from buffer.
	// Result: 4000x fewer syscalls for byte-at-a-time reads.

	// ─── bufio.Scanner: line-by-line (or token-by-token) reading ───
	// The most common pattern for processing text streams
	input := "line one\nline two\nline three\nline four"
	scanner := bufio.NewScanner(strings.NewReader(input))

	fmt.Print("  Scanner lines: ")
	for scanner.Scan() {
		fmt.Printf("[%s] ", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("scanner error: %v", err)
	}
	fmt.Println()

	// ─── Scanner with custom split function ───
	// bufio.ScanLines (default), ScanWords, ScanBytes, ScanRunes
	wordInput := "the quick brown fox"
	wordScanner := bufio.NewScanner(strings.NewReader(wordInput))
	wordScanner.Split(bufio.ScanWords)

	fmt.Print("  Scanner words: ")
	for wordScanner.Scan() {
		fmt.Printf("[%s] ", wordScanner.Text())
	}
	fmt.Println()

	// ─── Custom split function ───
	// Split on commas (CSV-like)
	csvInput := "field1,field2,field3,field4"
	csvScanner := bufio.NewScanner(strings.NewReader(csvInput))
	csvScanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// Find the next comma
		for i := 0; i < len(data); i++ {
			if data[i] == ',' {
				return i + 1, data[:i], nil
			}
		}
		// No comma found
		if atEOF && len(data) > 0 {
			return len(data), data, nil // last field
		}
		return 0, nil, nil // request more data
	})

	fmt.Print("  Custom split (comma): ")
	for csvScanner.Scan() {
		fmt.Printf("[%s] ", csvScanner.Text())
	}
	fmt.Println()

	// ─── Scanner max token size ───
	// Default max token size is 64KB (bufio.MaxScanTokenSize)
	// For large lines, increase the buffer:
	// scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max

	// ─── bufio.Reader: when you need more control than Scanner ───
	br := bufio.NewReader(strings.NewReader("peek ahead\nand read lines\n"))

	// Peek: look at bytes without consuming them
	peeked, _ := br.Peek(4)
	fmt.Printf("  Peek(4): %q (not consumed)\n", peeked)

	// ReadString: read until delimiter (includes delimiter)
	line, _ := br.ReadString('\n')
	fmt.Printf("  ReadString('\\n'): %q\n", line)

	// ReadBytes works like ReadString but returns []byte
	lineBytes, _ := br.ReadBytes('\n')
	fmt.Printf("  ReadBytes('\\n'): %q\n", lineBytes)

	// ─── bufio.Writer: batch small writes into larger ones ───
	var output bytes.Buffer
	bw := bufio.NewWriter(&output)

	// These three writes get buffered (no actual Write to output yet)
	bw.WriteString("hello ")
	bw.WriteString("buffered ")
	bw.WriteString("world")

	fmt.Printf("  Before Flush: output has %d bytes\n", output.Len()) // 0!
	bw.Flush()                                                        // NOW it writes everything in one batch
	fmt.Printf("  After Flush: output has %d bytes → %q\n", output.Len(), output.String())

	// ─── bufio.ReadWriter: combined buffered reader + writer ───
	// Used in net/http for connection handling
	fmt.Println("  bufio.ReadWriter: combine buffered Reader + Writer (used in net/http)")

	fmt.Println()
}

// =============================================================================
// PART 7: Custom Reader & Writer — Build Your Own
// =============================================================================

// ── UppercaseReader: transforms all bytes to uppercase ──
// Wraps any Reader, uppercases as you read through it.
// This is a DECORATOR pattern — same interface, added behavior.
type UppercaseReader struct {
	src io.Reader
}

func NewUppercaseReader(r io.Reader) *UppercaseReader {
	return &UppercaseReader{src: r}
}

func (u *UppercaseReader) Read(p []byte) (int, error) {
	n, err := u.src.Read(p)
	for i := 0; i < n; i++ {
		p[i] = byte(unicode.ToUpper(rune(p[i])))
	}
	return n, err
}

// ── CountingWriter: counts bytes written through it ──
// Wraps any Writer, tracks total bytes.
type CountingWriter struct {
	dst   io.Writer
	Count int64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{dst: w}
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.dst.Write(p)
	c.Count += int64(n)
	return n, err
}

// ── ProgressReader: reports read progress ──
// Useful for showing download progress bars.
type ProgressReader struct {
	src      io.Reader
	total    int64 // total expected bytes (0 if unknown)
	read     int64
	callback func(bytesRead, total int64)
}

func NewProgressReader(r io.Reader, total int64, cb func(int64, int64)) *ProgressReader {
	return &ProgressReader{src: r, total: total, callback: cb}
}

func (p *ProgressReader) Read(buf []byte) (int, error) {
	n, err := p.src.Read(buf)
	p.read += int64(n)
	if p.callback != nil {
		p.callback(p.read, p.total)
	}
	return n, err
}

func customReaderWriter() {
	fmt.Println("--- CUSTOM READER & WRITER ---")

	// UppercaseReader in action
	src := strings.NewReader("hello from a custom reader")
	upper := NewUppercaseReader(src)
	data, _ := io.ReadAll(upper)
	fmt.Printf("  UppercaseReader: %q\n", data)

	// Stack them: uppercase → limit to 10 bytes
	src2 := strings.NewReader("composing multiple custom readers is easy")
	stacked := io.LimitReader(NewUppercaseReader(src2), 10)
	data2, _ := io.ReadAll(stacked)
	fmt.Printf("  Stacked (upper→limit10): %q\n", data2)

	// CountingWriter in action
	var buf bytes.Buffer
	counter := NewCountingWriter(&buf)
	fmt.Fprint(counter, "track every byte")
	fmt.Fprint(counter, " that flows through")
	fmt.Printf("  CountingWriter: %d bytes written → %q\n", counter.Count, buf.String())

	// ProgressReader in action (simulating a download)
	bigData := strings.NewReader(strings.Repeat("x", 100))
	var lastPct int64
	progress := NewProgressReader(bigData, 100, func(read, total int64) {
		pct := read * 100 / total
		if pct != lastPct && pct%25 == 0 {
			fmt.Printf("    Progress: %d%%\n", pct)
			lastPct = pct
		}
	})
	io.Copy(io.Discard, progress) // consume all data

	fmt.Println()
}

// =============================================================================
// PART 8: io.Pipe — Connect a Writer to a Reader
// =============================================================================
func ioPipePattern() {
	fmt.Println("--- IO PIPE ---")

	// io.Pipe creates a synchronous in-memory pipe.
	// Write to the PipeWriter → appears in the PipeReader.
	// Like a Unix pipe connecting two processes.
	//
	// WHY: When you have a function that writes to a Writer,
	// but you need to read the output as a Reader.
	//
	// Example: gzip compress data and process it as a stream
	// (without buffering the entire compressed output in memory)

	pr, pw := io.Pipe()

	var wg sync.WaitGroup
	wg.Add(1)

	// Writer goroutine: compress data into the pipe
	go func() {
		defer wg.Done()
		gzw := gzip.NewWriter(pw)
		gzw.Write([]byte("data compressed through a pipe — no intermediate buffer!"))
		gzw.Close()
		pw.Close() // signals EOF to the reader side
	}()

	// Reader side: decompress from the pipe
	gzr, _ := gzip.NewReader(pr)
	result, _ := io.ReadAll(gzr)
	gzr.Close()
	pr.Close()

	wg.Wait()

	fmt.Printf("  Pipe result: %q\n", result)

	// WHEN TO USE io.Pipe:
	// ─────────────────────
	// - Streaming JSON encoding → HTTP request body
	// - Compressing data → uploading to S3
	// - Any time you need to connect a "push" API to a "pull" API
	// - When you can't buffer the entire output in memory
	//
	// CAUTION:
	// - Writer blocks until Reader reads (synchronous!)
	// - Always close PipeWriter when done (otherwise Reader hangs)
	// - Consider a buffer (bytes.Buffer) if data is small

	fmt.Println()
}

// =============================================================================
// PART 9: Real-World Patterns
// =============================================================================
func realWorldPatterns() {
	fmt.Println("--- REAL-WORLD PATTERNS ---")

	// ─── PATTERN 1: Read HTTP body safely (limit + drain) ───
	// In production, ALWAYS limit request body size to prevent OOM.
	//
	// func handler(w http.ResponseWriter, r *http.Request) {
	//     // Limit body to 1MB
	//     r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	//     body, err := io.ReadAll(r.Body)
	//     if err != nil {
	//         // MaxBytesReader returns *http.MaxBytesError if exceeded
	//         http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
	//         return
	//     }
	//     // process body...
	// }
	fmt.Println("  Pattern 1: http.MaxBytesReader to limit request body size")

	// ─── PATTERN 2: Stream processing without loading into memory ───
	// Process a large file line by line without loading it all.
	//
	// func processLargeFile(path string) error {
	//     f, err := os.Open(path)
	//     if err != nil { return err }
	//     defer f.Close()
	//
	//     scanner := bufio.NewScanner(f)
	//     scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1MB lines
	//     for scanner.Scan() {
	//         processLine(scanner.Text())
	//     }
	//     return scanner.Err()
	// }
	fmt.Println("  Pattern 2: bufio.Scanner for streaming line processing")

	// ─── PATTERN 3: Tee for logging request/response bodies ───
	//
	// func loggingMiddleware(next http.Handler) http.Handler {
	//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//         var bodyLog bytes.Buffer
	//         // Mirror the body to our log buffer while reading
	//         r.Body = io.NopCloser(io.TeeReader(r.Body, &bodyLog))
	//         next.ServeHTTP(w, r)
	//         log.Printf("Request body: %s", bodyLog.String())
	//     })
	// }
	fmt.Println("  Pattern 3: TeeReader to log HTTP body while consuming it")

	// ─── PATTERN 4: io.NopCloser — Adapter from Reader to ReadCloser ───
	// Many APIs expect ReadCloser (like http.Request.Body).
	// When you have a plain Reader, wrap it:
	body := io.NopCloser(strings.NewReader("now I'm a ReadCloser"))
	data, _ := io.ReadAll(body)
	body.Close() // does nothing, but satisfies the interface
	fmt.Printf("  Pattern 4: io.NopCloser: %q\n", data)

	// ─── PATTERN 5: Multi-destination write (file + hash + counter) ───
	var file bytes.Buffer
	hasher := sha256.New()
	counter := NewCountingWriter(io.MultiWriter(&file, hasher))

	fmt.Fprint(counter, "data written to file, hashed, and counted simultaneously")

	hash := hex.EncodeToString(hasher.Sum(nil))
	fmt.Printf("  Pattern 5: Multi-write: %d bytes, SHA256=%s...\n", counter.Count, hash[:16])

	// ─── SUMMARY OF KEY FUNCTIONS ───
	//
	// READING:
	//   io.ReadAll(r)           — read everything (careful: loads all into memory)
	//   io.ReadFull(r, buf)     — read exactly len(buf) bytes
	//   io.Copy(dst, src)       — stream src→dst efficiently (may use sendfile)
	//   io.CopyN(dst, src, n)   — copy exactly n bytes
	//   io.LimitReader(r, n)    — cap reads at n bytes
	//   io.TeeReader(r, w)      — mirror reads to writer
	//   io.MultiReader(r1,r2)   — concatenate readers
	//   io.SectionReader        — read a slice of ReaderAt
	//   io.NopCloser(r)         — Reader → ReadCloser adapter
	//
	// WRITING:
	//   io.MultiWriter(w1,w2)   — write to multiple destinations
	//   io.WriteString(w, s)    — efficient string write
	//   io.Discard              — /dev/null writer
	//   io.Pipe()               — connect writer to reader
	//
	// BUFFERED:
	//   bufio.NewReader(r)      — buffered reader (Peek, ReadString, etc.)
	//   bufio.NewWriter(w)      — buffered writer (Flush!)
	//   bufio.NewScanner(r)     — line/word/token scanner

	fmt.Println()
}
