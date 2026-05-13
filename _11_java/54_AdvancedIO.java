/*
 * ============================================================
 *  CHAPTER 54: ADVANCED I/O
 * ============================================================
 *  Beyond basic file reading/writing. Async channels, memory-
 *  mapped files, file watching, process execution, and low-
 *  level I/O that god-level Java demands.
 *
 *  TOPICS:
 *    1. AsynchronousFileChannel — Non-blocking file I/O
 *    2. AsynchronousSocketChannel — Async networking
 *    3. Memory-Mapped Files (MappedByteBuffer)
 *    4. WatchService — File System Events
 *    5. ProcessBuilder — Running External Processes
 *    6. Pipe — Inter-thread Communication
 *    7. FileChannel — Transfers & Locking
 *    8. ScatteringByteChannel / GatheringByteChannel
 *    9. Custom FileSystemProvider (concept)
 * ============================================================
 */

import java.io.*;
import java.nio.*;
import java.nio.channels.*;
import java.nio.file.*;
import java.nio.file.attribute.*;
import java.util.*;
import java.util.concurrent.*;

public class Chapter54_AdvancedIO {

    // Temp directory for demos
    static Path tempDir;

    public static void main(String[] args) throws Exception {

        System.out.println("=== CHAPTER 54: ADVANCED I/O ===\n");

        tempDir = Files.createTempDirectory("ch54_demo");
        System.out.println("  Temp dir: " + tempDir + "\n");

        try {
            demo1_AsyncFileChannel();
            demo2_MemoryMappedFile();
            demo3_WatchService();
            demo4_ProcessBuilder();
            demo5_Pipe();
            demo6_FileChannelTransferLock();
            demo7_ScatterGather();
            demo8_FileVisitor();
        } finally {
            // Cleanup temp directory
            Files.walk(tempDir)
                .sorted(Comparator.reverseOrder())
                .forEach(p -> { try { Files.deleteIfExists(p); } catch (Exception e) {} });
        }

        System.out.println("\n✓ Advanced I/O Complete!");
    }

    // ========================================================
    // 1. ASYNCHRONOUS FILE CHANNEL
    // ========================================================
    static void demo1_AsyncFileChannel() throws Exception {
        System.out.println("--- 1. AsynchronousFileChannel ---\n");

        Path file = tempDir.resolve("async_test.txt");

        // Write asynchronously
        try (AsynchronousFileChannel afc = AsynchronousFileChannel.open(
                file, StandardOpenOption.CREATE, StandardOpenOption.WRITE)) {

            ByteBuffer writeBuffer = ByteBuffer.wrap("Hello Async World!\nLine 2\nLine 3".getBytes());

            // Method 1: Future-based
            Future<Integer> writeFuture = afc.write(writeBuffer, 0);
            int bytesWritten = writeFuture.get(); // blocks for result
            System.out.println("  Written " + bytesWritten + " bytes (Future)");
        }

        // Read asynchronously with CompletionHandler
        try (AsynchronousFileChannel afc = AsynchronousFileChannel.open(
                file, StandardOpenOption.READ)) {

            ByteBuffer readBuffer = ByteBuffer.allocate(1024);

            // Method 2: CompletionHandler (callback)
            CountDownLatch latch = new CountDownLatch(1);
            afc.read(readBuffer, 0, null, new CompletionHandler<Integer, Void>() {
                @Override
                public void completed(Integer bytesRead, Void attachment) {
                    readBuffer.flip();
                    byte[] data = new byte[readBuffer.remaining()];
                    readBuffer.get(data);
                    System.out.println("  Read " + bytesRead + " bytes (CompletionHandler)");
                    System.out.println("  Content: " + new String(data).replace("\n", "\\n"));
                    latch.countDown();
                }

                @Override
                public void failed(Throwable exc, Void attachment) {
                    System.out.println("  Read failed: " + exc.getMessage());
                    latch.countDown();
                }
            });

            latch.await(5, TimeUnit.SECONDS);
        }
        System.out.println();
    }

    // ========================================================
    // 2. MEMORY-MAPPED FILES
    // ========================================================
    static void demo2_MemoryMappedFile() throws Exception {
        System.out.println("--- 2. Memory-Mapped Files ---\n");

        /*
         * MappedByteBuffer maps a file directly into memory.
         * The OS handles reading/writing — you just access memory.
         *
         * Benefits:
         *   - VERY fast for large files (OS manages caching)
         *   - No explicit read/write calls
         *   - Multiple processes can share the mapping
         *
         * Modes:
         *   READ_ONLY    — only read
         *   READ_WRITE   — read and write (changes go to file)
         *   PRIVATE       — copy-on-write (changes not saved to file)
         */

        Path file = tempDir.resolve("mmap_test.dat");

        // Create a memory-mapped file for writing
        int SIZE = 1024;
        try (FileChannel fc = FileChannel.open(file,
                StandardOpenOption.CREATE, StandardOpenOption.READ, StandardOpenOption.WRITE)) {

            MappedByteBuffer mmap = fc.map(FileChannel.MapMode.READ_WRITE, 0, SIZE);

            // Write directly to memory (automatically persisted to file)
            mmap.putInt(42);
            mmap.putDouble(3.14159);
            mmap.put("Hello MMap!".getBytes());

            mmap.force(); // flush to disk

            System.out.println("  Written to memory-mapped file");
        }

        // Read back
        try (FileChannel fc = FileChannel.open(file, StandardOpenOption.READ)) {
            MappedByteBuffer mmap = fc.map(FileChannel.MapMode.READ_ONLY, 0, SIZE);

            int intVal = mmap.getInt();
            double doubleVal = mmap.getDouble();
            byte[] strBytes = new byte[11];
            mmap.get(strBytes);

            System.out.println("  Read back: int=" + intVal
                + ", double=" + doubleVal
                + ", str=" + new String(strBytes));
        }
        System.out.println();
    }

    // ========================================================
    // 3. WATCHSERVICE — File System Events
    // ========================================================
    static void demo3_WatchService() throws Exception {
        System.out.println("--- 3. WatchService ---\n");

        /*
         * WatchService monitors directories for changes:
         *   ENTRY_CREATE — file/dir created
         *   ENTRY_MODIFY — file modified
         *   ENTRY_DELETE — file/dir deleted
         *
         * Uses OS-level notifications (inotify on Linux, FSEvents on Mac)
         * Much more efficient than polling!
         */

        Path watchDir = tempDir.resolve("watched");
        Files.createDirectories(watchDir);

        WatchService watcher = FileSystems.getDefault().newWatchService();
        watchDir.register(watcher,
            StandardWatchEventKinds.ENTRY_CREATE,
            StandardWatchEventKinds.ENTRY_MODIFY,
            StandardWatchEventKinds.ENTRY_DELETE);

        // Create events in a separate thread
        Thread eventThread = new Thread(() -> {
            try {
                Thread.sleep(100);
                Files.writeString(watchDir.resolve("test.txt"), "Hello!");
                Thread.sleep(100);
                Files.writeString(watchDir.resolve("test.txt"), "Modified!");
                Thread.sleep(100);
                Files.delete(watchDir.resolve("test.txt"));
            } catch (Exception e) { e.printStackTrace(); }
        });
        eventThread.start();

        // Watch for events (with timeout)
        System.out.println("  Watching " + watchDir + " for 2 seconds...");
        long deadline = System.currentTimeMillis() + 2000;
        while (System.currentTimeMillis() < deadline) {
            WatchKey key = watcher.poll(500, TimeUnit.MILLISECONDS);
            if (key != null) {
                for (WatchEvent<?> event : key.pollEvents()) {
                    WatchEvent.Kind<?> kind = event.kind();
                    if (kind == StandardWatchEventKinds.OVERFLOW) continue;
                    Path fileName = (Path) event.context();
                    System.out.println("    Event: " + kind.name() + " → " + fileName);
                }
                key.reset();
            }
        }

        eventThread.join();
        watcher.close();
        System.out.println();
    }

    // ========================================================
    // 4. PROCESSBUILDER — Running External Processes
    // ========================================================
    static void demo4_ProcessBuilder() throws Exception {
        System.out.println("--- 4. ProcessBuilder ---\n");

        // Basic command execution
        ProcessBuilder pb = new ProcessBuilder("echo", "Hello from subprocess!");
        pb.redirectErrorStream(true); // merge stderr into stdout

        Process process = pb.start();
        String output;
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(process.getInputStream()))) {
            output = reader.lines().collect(java.util.stream.Collectors.joining("\n"));
        }
        int exitCode = process.waitFor();
        System.out.println("  Output: " + output);
        System.out.println("  Exit code: " + exitCode);

        // Pipeline (Java 9+)
        // ProcessBuilder.startPipeline(List.of(pb1, pb2, pb3))
        // Equivalent to: command1 | command2 | command3

        // Environment variables
        ProcessBuilder pb2 = new ProcessBuilder("env");
        pb2.environment().put("MY_VAR", "my_value");
        pb2.redirectErrorStream(true);

        Process p2 = pb2.start();
        try (BufferedReader reader = new BufferedReader(
                new InputStreamReader(p2.getInputStream()))) {
            reader.lines()
                .filter(line -> line.startsWith("MY_VAR"))
                .forEach(line -> System.out.println("  Custom env: " + line));
        }
        p2.waitFor();

        // Redirect to file
        Path outputFile = tempDir.resolve("proc_output.txt");
        ProcessBuilder pb3 = new ProcessBuilder("echo", "Written to file");
        pb3.redirectOutput(outputFile.toFile());
        Process p3 = pb3.start();
        p3.waitFor();
        System.out.println("  File output: " + Files.readString(outputFile).trim());

        // Process info (Java 9+)
        ProcessHandle current = ProcessHandle.current();
        System.out.println("  Current PID: " + current.pid());
        current.info().command().ifPresent(cmd -> System.out.println("  Command: " + cmd));

        System.out.println();
    }

    // ========================================================
    // 5. PIPE — Inter-Thread I/O
    // ========================================================
    static void demo5_Pipe() throws Exception {
        System.out.println("--- 5. Pipe (Inter-Thread Communication) ---\n");

        /*
         * Pipe provides a pair of channels:
         *   SinkChannel (write end) → SourceChannel (read end)
         * Useful for communicating between threads via NIO channels.
         */

        Pipe pipe = Pipe.open();
        Pipe.SinkChannel sink = pipe.sink();
        Pipe.SourceChannel source = pipe.source();

        // Writer thread
        Thread writer = new Thread(() -> {
            try {
                String[] messages = {"Hello", "from", "pipe!"};
                for (String msg : messages) {
                    ByteBuffer buf = ByteBuffer.wrap((msg + "\n").getBytes());
                    sink.write(buf);
                    Thread.sleep(50);
                }
                sink.close();
            } catch (Exception e) { e.printStackTrace(); }
        });

        // Reader
        writer.start();
        ByteBuffer readBuf = ByteBuffer.allocate(256);
        StringBuilder sb = new StringBuilder();
        while (source.read(readBuf) > 0 || writer.isAlive()) {
            readBuf.flip();
            while (readBuf.hasRemaining()) {
                sb.append((char) readBuf.get());
            }
            readBuf.clear();
        }
        source.close();
        writer.join();

        System.out.println("  Received via pipe: " + sb.toString().replace("\n", " ").trim());
        System.out.println();
    }

    // ========================================================
    // 6. FILECHANNEL — Transfer & Locking
    // ========================================================
    static void demo6_FileChannelTransferLock() throws Exception {
        System.out.println("--- 6. FileChannel Transfer & Locking ---\n");

        // Zero-copy transfer between channels
        Path src = tempDir.resolve("transfer_src.txt");
        Path dst = tempDir.resolve("transfer_dst.txt");
        Files.writeString(src, "Data to transfer efficiently!");

        try (FileChannel srcChannel = FileChannel.open(src, StandardOpenOption.READ);
             FileChannel dstChannel = FileChannel.open(dst,
                 StandardOpenOption.CREATE, StandardOpenOption.WRITE)) {

            // transferTo uses OS-level zero-copy (sendfile on Linux)
            long transferred = srcChannel.transferTo(0, srcChannel.size(), dstChannel);
            System.out.println("  Transferred " + transferred + " bytes (zero-copy)");
        }
        System.out.println("  Destination: " + Files.readString(dst));

        // File locking
        Path lockFile = tempDir.resolve("locked.txt");
        Files.writeString(lockFile, "protected data");

        try (FileChannel fc = FileChannel.open(lockFile,
                StandardOpenOption.READ, StandardOpenOption.WRITE)) {

            // Exclusive lock (blocks other processes)
            FileLock lock = fc.tryLock();
            if (lock != null) {
                System.out.println("  Acquired lock: " + lock);
                System.out.println("  Shared? " + lock.isShared());
                // Do protected work...
                lock.release();
                System.out.println("  Lock released");
            }

            // Shared lock (allows other readers, blocks writers)
            FileLock sharedLock = fc.tryLock(0, Long.MAX_VALUE, true);
            if (sharedLock != null) {
                System.out.println("  Shared lock acquired");
                sharedLock.release();
            }
        }
        System.out.println();
    }

    // ========================================================
    // 7. SCATTER/GATHER
    // ========================================================
    static void demo7_ScatterGather() throws Exception {
        System.out.println("--- 7. Scatter/Gather ---\n");

        /*
         * Scattering Read: read from ONE channel into MULTIPLE buffers
         *   (e.g., read header into one buffer, body into another)
         *
         * Gathering Write: write from MULTIPLE buffers into ONE channel
         *   (e.g., combine header + body into one write)
         *
         * Useful for protocols with fixed-size headers.
         */

        Path file = tempDir.resolve("scatter_gather.dat");

        // Gathering Write — write from multiple buffers
        try (FileChannel fc = FileChannel.open(file,
                StandardOpenOption.CREATE, StandardOpenOption.WRITE)) {

            ByteBuffer header = ByteBuffer.wrap("HEADER:".getBytes());
            ByteBuffer body = ByteBuffer.wrap("This is the body content".getBytes());

            fc.write(new ByteBuffer[]{header, body}); // gathering write
            System.out.println("  Gathering write complete");
        }

        // Scattering Read — read into multiple buffers
        try (FileChannel fc = FileChannel.open(file, StandardOpenOption.READ)) {
            ByteBuffer headerBuf = ByteBuffer.allocate(7);  // "HEADER:" = 7 bytes
            ByteBuffer bodyBuf = ByteBuffer.allocate(100);

            fc.read(new ByteBuffer[]{headerBuf, bodyBuf}); // scattering read

            headerBuf.flip();
            bodyBuf.flip();

            System.out.println("  Header: " + new String(headerBuf.array(), 0, headerBuf.limit()));
            System.out.println("  Body: " + new String(bodyBuf.array(), 0, bodyBuf.limit()));
        }
        System.out.println();
    }

    // ========================================================
    // 8. FILEVISITOR — Walking File Trees
    // ========================================================
    static void demo8_FileVisitor() throws Exception {
        System.out.println("--- 8. FileVisitor (Advanced Tree Walk) ---\n");

        // Create a structure to walk
        Path root = tempDir.resolve("tree");
        Files.createDirectories(root.resolve("a/b"));
        Files.createDirectories(root.resolve("c"));
        Files.writeString(root.resolve("file1.txt"), "root file");
        Files.writeString(root.resolve("a/file2.txt"), "a file");
        Files.writeString(root.resolve("a/b/file3.txt"), "b file");
        Files.writeString(root.resolve("c/file4.txt"), "c file");

        // Custom FileVisitor
        List<String> visited = new ArrayList<>();
        Files.walkFileTree(root, new SimpleFileVisitor<Path>() {
            @Override
            public FileVisitResult preVisitDirectory(Path dir, BasicFileAttributes attrs) {
                visited.add("[DIR]  " + root.relativize(dir));
                return FileVisitResult.CONTINUE;
            }

            @Override
            public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) {
                visited.add("[FILE] " + root.relativize(file) + " (" + attrs.size() + " bytes)");
                return FileVisitResult.CONTINUE;
            }

            @Override
            public FileVisitResult visitFileFailed(Path file, IOException exc) {
                System.out.println("    Failed: " + file + " - " + exc.getMessage());
                return FileVisitResult.CONTINUE;
            }
        });

        visited.forEach(v -> System.out.println("    " + v));

        // Find files matching a pattern
        System.out.println("\n  Find *.txt files:");
        PathMatcher matcher = FileSystems.getDefault().getPathMatcher("glob:**.txt");
        Files.walk(root)
            .filter(Files::isRegularFile)
            .filter(matcher::matches)
            .forEach(p -> System.out.println("    " + root.relativize(p)));

        // File attributes
        System.out.println("\n  File attributes:");
        BasicFileAttributes attrs = Files.readAttributes(
            root.resolve("file1.txt"), BasicFileAttributes.class);
        System.out.println("    Size: " + attrs.size());
        System.out.println("    Created: " + attrs.creationTime());
        System.out.println("    Modified: " + attrs.lastModifiedTime());
        System.out.println("    Is regular file: " + attrs.isRegularFile());
    }
}

/*
 * EXERCISES:
 * 1. Build a directory synchronizer: watch source dir, copy new/modified
 *    files to target dir automatically.
 * 2. Create a simple HTTP server using AsynchronousServerSocketChannel.
 * 3. Write a log file tailer (like 'tail -f') using WatchService.
 * 4. Build a file search tool using FileVisitor that supports glob patterns,
 *    size filters, and date filters.
 *
 * NEXT: Chapter 55 — Advanced Enums & SPI (Final Chapter!)
 */
