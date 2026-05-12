/*
 * ============================================================
 *  CHAPTER 23: FILE I/O (java.io)
 * ============================================================
 *  Classic I/O for reading/writing files.
 *
 *  TWO TYPES OF STREAMS:
 *  1. Byte Streams  — raw bytes (images, binary)
 *     InputStream/OutputStream, FileInputStream/FileOutputStream
 *  2. Character Streams — text (uses encoding)
 *     Reader/Writer, FileReader/FileWriter, BufferedReader/BufferedWriter
 *
 *  ALWAYS use try-with-resources for I/O!
 * ============================================================
 */

import java.io.*;
import java.util.Scanner;

public class Chapter23_FileIO {

    public static void main(String[] args) {

        String dir = System.getProperty("user.dir");
        String testFile = dir + "/test_output.txt";
        String copyFile = dir + "/test_copy.txt";

        // --- 1. Writing text with BufferedWriter ---
        System.out.println("=== WRITING FILES ===\n");
        try (BufferedWriter writer = new BufferedWriter(new FileWriter(testFile))) {
            writer.write("Line 1: Hello, Java File I/O!");
            writer.newLine();
            writer.write("Line 2: BufferedWriter is efficient.");
            writer.newLine();
            writer.write("Line 3: Always close your resources!");
            System.out.println("Written to: " + testFile);
        } catch (IOException e) {
            System.out.println("Write error: " + e.getMessage());
        }

        // --- 2. Reading text with BufferedReader ---
        System.out.println("\n=== READING FILES ===\n");
        try (BufferedReader reader = new BufferedReader(new FileReader(testFile))) {
            String line;
            int lineNum = 1;
            while ((line = reader.readLine()) != null) {
                System.out.println(lineNum++ + ": " + line);
            }
        } catch (IOException e) {
            System.out.println("Read error: " + e.getMessage());
        }

        // --- 3. Appending to file ---
        System.out.println("\n=== APPENDING ===\n");
        try (FileWriter fw = new FileWriter(testFile, true)) { // true = append
            fw.write("\nLine 4: Appended text!");
            System.out.println("Appended successfully.");
        } catch (IOException e) {
            System.out.println("Append error: " + e.getMessage());
        }

        // --- 4. PrintWriter — most convenient for text ---
        System.out.println("\n=== PRINTWRITER ===\n");
        try (PrintWriter pw = new PrintWriter(new FileWriter(testFile))) {
            pw.println("PrintWriter line 1");
            pw.printf("Formatted: %s is %d years old%n", "Java", 29);
            pw.println("PrintWriter line 3");
            System.out.println("PrintWriter done.");
        } catch (IOException e) {
            System.out.println("Error: " + e.getMessage());
        }

        // --- 5. Byte Streams — binary copy ---
        System.out.println("\n=== BYTE STREAM COPY ===\n");
        try (FileInputStream fis = new FileInputStream(testFile);
             FileOutputStream fos = new FileOutputStream(copyFile)) {
            byte[] buffer = new byte[1024];
            int bytesRead;
            while ((bytesRead = fis.read(buffer)) != -1) {
                fos.write(buffer, 0, bytesRead);
            }
            System.out.println("File copied to: " + copyFile);
        } catch (IOException e) {
            System.out.println("Copy error: " + e.getMessage());
        }

        // --- 6. File class operations ---
        System.out.println("\n=== FILE CLASS ===\n");
        File file = new File(testFile);
        System.out.println("Exists: " + file.exists());
        System.out.println("Name: " + file.getName());
        System.out.println("Path: " + file.getAbsolutePath());
        System.out.println("Size: " + file.length() + " bytes");
        System.out.println("Readable: " + file.canRead());
        System.out.println("Writable: " + file.canWrite());
        System.out.println("Is file: " + file.isFile());
        System.out.println("Is dir: " + file.isDirectory());

        // List directory contents
        File dir2 = new File(dir);
        String[] files = dir2.list();
        if (files != null) {
            System.out.println("\nFiles in directory:");
            for (String f : files) {
                System.out.println("  " + f);
            }
        }

        // --- 7. Scanner for reading ---
        System.out.println("\n=== SCANNER READING ===\n");
        try (Scanner sc = new Scanner(new File(testFile))) {
            while (sc.hasNextLine()) {
                System.out.println("Scanner: " + sc.nextLine());
            }
        } catch (FileNotFoundException e) {
            System.out.println("File not found: " + e.getMessage());
        }

        // Cleanup test files
        new File(testFile).delete();
        new File(copyFile).delete();
        System.out.println("\nTest files cleaned up.");

        // --- Summary ---
        System.out.println("\n=== I/O SUMMARY ===");
        System.out.println("Text reading:  BufferedReader (fast) or Scanner (convenient)");
        System.out.println("Text writing:  BufferedWriter or PrintWriter");
        System.out.println("Binary:        FileInputStream / FileOutputStream");
        System.out.println("File info:     File class");
        System.out.println("ALWAYS:        Use try-with-resources!");
    }
}

/*
 * EXERCISES:
 * 1. Read a file and count words, lines, and characters.
 * 2. Write a program that copies a file and converts to uppercase.
 * 3. Merge multiple text files into one.
 * 4. Find and replace a word in a file.
 *
 * NEXT: Chapter 24 — NIO & NIO.2
 */
