/*
 * ============================================================
 *  CHAPTER 24: NIO & NIO.2 (java.nio.file)
 * ============================================================
 *  Modern file I/O — more powerful than java.io.
 *  Key classes: Path, Files, Paths
 *
 *  Path  → represents a file/directory path
 *  Files → static utility methods for file operations
 * ============================================================
 */

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.List;
import java.util.stream.Stream;

public class Chapter24_NIO {

    public static void main(String[] args) throws IOException {

        // --- 1. Path ---
        System.out.println("=== PATH ===\n");
        Path path = Paths.get("test_nio.txt");
        Path absPath = path.toAbsolutePath();
        System.out.println("Path: " + path);
        System.out.println("Absolute: " + absPath);
        System.out.println("FileName: " + path.getFileName());
        System.out.println("Parent: " + absPath.getParent());

        // Path manipulation
        Path resolved = Paths.get("/home").resolve("user/file.txt");
        System.out.println("Resolved: " + resolved);

        Path relative = Paths.get("/a/b").relativize(Paths.get("/a/b/c/d"));
        System.out.println("Relativize: " + relative); // c/d

        // --- 2. Files — Write & Read ---
        System.out.println("\n=== FILES READ/WRITE ===\n");

        Path testFile = Paths.get("test_nio.txt");

        // Write (creates or overwrites)
        List<String> lines = List.of("Line 1: NIO is modern", "Line 2: Files class is powerful", "Line 3: Paths are flexible");
        Files.write(testFile, lines, StandardCharsets.UTF_8);
        System.out.println("Written " + lines.size() + " lines.");

        // Read all lines
        List<String> readLines = Files.readAllLines(testFile, StandardCharsets.UTF_8);
        readLines.forEach(l -> System.out.println("  " + l));

        // Read all bytes
        byte[] bytes = Files.readAllBytes(testFile);
        System.out.println("Total bytes: " + bytes.length);

        // Read as string
        String content = new String(Files.readAllBytes(testFile), StandardCharsets.UTF_8);
        System.out.println("Content: " + content.substring(0, Math.min(50, content.length())) + "...");

        // Read with Stream (lazy, good for large files)
        System.out.println("\nStream reading:");
        try (Stream<String> stream = Files.lines(testFile)) {
            stream.filter(l -> l.contains("NIO"))
                  .forEach(l -> System.out.println("  Found: " + l));
        }

        // Append
        Files.write(testFile, List.of("Line 4: Appended!"), StandardCharsets.UTF_8,
                StandardOpenOption.APPEND);

        // --- 3. File Operations ---
        System.out.println("\n=== FILE OPERATIONS ===\n");

        // Copy
        Path copyPath = Paths.get("test_copy_nio.txt");
        Files.copy(testFile, copyPath, StandardCopyOption.REPLACE_EXISTING);
        System.out.println("Copied to: " + copyPath);

        // Move/rename
        Path movedPath = Paths.get("test_moved_nio.txt");
        Files.move(copyPath, movedPath, StandardCopyOption.REPLACE_EXISTING);
        System.out.println("Moved to: " + movedPath);

        // File attributes
        System.out.println("\nAttributes of " + testFile + ":");
        System.out.println("  Size: " + Files.size(testFile));
        System.out.println("  Exists: " + Files.exists(testFile));
        System.out.println("  Readable: " + Files.isReadable(testFile));
        System.out.println("  Writable: " + Files.isWritable(testFile));
        System.out.println("  Directory: " + Files.isDirectory(testFile));
        System.out.println("  Regular file: " + Files.isRegularFile(testFile));

        // --- 4. Directory Operations ---
        System.out.println("\n=== DIRECTORY OPERATIONS ===\n");

        Path tempDir = Paths.get("temp_test_dir");
        Files.createDirectories(tempDir.resolve("sub1/sub2"));
        System.out.println("Created nested dirs: " + tempDir.resolve("sub1/sub2"));

        // Create temp file
        Path tempFile = Files.createTempFile("prefix_", ".tmp");
        System.out.println("Temp file: " + tempFile);

        // List directory contents
        System.out.println("\nCurrent directory contents:");
        try (Stream<Path> entries = Files.list(Paths.get("."))) {
            entries.filter(Files::isRegularFile)
                   .map(Path::getFileName)
                   .sorted()
                   .limit(10)
                   .forEach(p -> System.out.println("  " + p));
        }

        // Walk directory tree (recursive)
        System.out.println("\nWalking temp_test_dir:");
        try (Stream<Path> walk = Files.walk(tempDir)) {
            walk.forEach(p -> System.out.println("  " + p));
        }

        // Find files matching pattern
        System.out.println("\nFind .java files (first 5):");
        try (Stream<Path> found = Files.find(Paths.get("."), 1,
                (p, attr) -> p.toString().endsWith(".java"))) {
            found.limit(5).forEach(p -> System.out.println("  " + p.getFileName()));
        }

        // --- 5. Cleanup ---
        Files.deleteIfExists(movedPath);
        Files.deleteIfExists(testFile);
        Files.deleteIfExists(tempFile);
        // Delete nested directories
        try (Stream<Path> walk = Files.walk(tempDir)) {
            walk.sorted(java.util.Comparator.reverseOrder())
                .forEach(p -> { try { Files.delete(p); } catch (IOException e) {} });
        }

        System.out.println("\nCleanup complete.");
        System.out.println("\nNIO.2 is the preferred way for file operations in modern Java!");
    }
}

/*
 * EXERCISES:
 * 1. Find all .txt files recursively in a directory.
 * 2. Count total lines across all .java files in a directory.
 * 3. Copy an entire directory tree (recursive copy).
 * 4. Watch a directory for changes using WatchService.
 *
 * NEXT: Chapter 25 — Serialization
 */
