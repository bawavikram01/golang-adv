/*
 * ============================================================
 *  CHAPTER 35: DESIGN PATTERNS — STRUCTURAL
 * ============================================================
 *  Structural patterns deal with composing classes/objects
 *  into larger structures while keeping them flexible.
 *
 *  STRUCTURAL PATTERNS:
 *    1. Adapter    — convert interface to another
 *    2. Decorator  — add behavior dynamically
 *    3. Proxy      — controlled access
 *    4. Facade     — simplified interface
 *    5. Composite  — tree structures
 *    6. Bridge     — separate abstraction from implementation
 * ============================================================
 */

import java.util.*;

public class Chapter35_StructuralPatterns {

    // ========================================================
    // 1. ADAPTER — make incompatible interfaces work together
    // ========================================================
    // "Plug converter: US plug into EU socket"

    // Target interface (what the client expects)
    interface MediaPlayer {
        void play(String filename);
    }

    // Adaptee (legacy class with different interface)
    static class LegacyAudioPlayer {
        void playMp3(String file) { System.out.println("    Playing MP3: " + file); }
    }

    static class LegacyVideoPlayer {
        void playMp4(String file) { System.out.println("    Playing MP4: " + file); }
    }

    // Adapter
    static class MediaAdapter implements MediaPlayer {
        private LegacyAudioPlayer audio = new LegacyAudioPlayer();
        private LegacyVideoPlayer video = new LegacyVideoPlayer();

        @Override
        public void play(String filename) {
            if (filename.endsWith(".mp3")) audio.playMp3(filename);
            else if (filename.endsWith(".mp4")) video.playMp4(filename);
            else System.out.println("    Unsupported format: " + filename);
        }
    }

    // ========================================================
    // 2. DECORATOR — add responsibilities dynamically
    // ========================================================
    // "Wrapping a gift: each layer adds something"

    interface Coffee {
        double cost();
        String description();
    }

    static class BasicCoffee implements Coffee {
        @Override public double cost() { return 2.0; }
        @Override public String description() { return "Basic Coffee"; }
    }

    // Base decorator
    static abstract class CoffeeDecorator implements Coffee {
        protected Coffee wrapped;
        CoffeeDecorator(Coffee coffee) { this.wrapped = coffee; }
    }

    static class MilkDecorator extends CoffeeDecorator {
        MilkDecorator(Coffee coffee) { super(coffee); }
        @Override public double cost() { return wrapped.cost() + 0.5; }
        @Override public String description() { return wrapped.description() + " + Milk"; }
    }

    static class SugarDecorator extends CoffeeDecorator {
        SugarDecorator(Coffee coffee) { super(coffee); }
        @Override public double cost() { return wrapped.cost() + 0.3; }
        @Override public String description() { return wrapped.description() + " + Sugar"; }
    }

    static class WhipDecorator extends CoffeeDecorator {
        WhipDecorator(Coffee coffee) { super(coffee); }
        @Override public double cost() { return wrapped.cost() + 0.7; }
        @Override public String description() { return wrapped.description() + " + Whip"; }
    }

    // ========================================================
    // 3. PROXY — controlled access to an object
    // ========================================================
    // "Bouncer at a club: checks before allowing access"

    interface Image {
        void display();
    }

    // Real object (expensive to create)
    static class RealImage implements Image {
        private String filename;
        RealImage(String filename) {
            this.filename = filename;
            loadFromDisk();
        }
        private void loadFromDisk() {
            System.out.println("    Loading image: " + filename + " (slow!)");
        }
        @Override
        public void display() {
            System.out.println("    Displaying: " + filename);
        }
    }

    // Proxy (lazy loading)
    static class ImageProxy implements Image {
        private String filename;
        private RealImage realImage;

        ImageProxy(String filename) {
            this.filename = filename;  // no loading yet!
        }

        @Override
        public void display() {
            if (realImage == null) {
                realImage = new RealImage(filename);  // load on first use
            }
            realImage.display();
        }
    }

    // Protection Proxy
    interface Document {
        void read();
        void write(String content);
    }

    static class RealDocument implements Document {
        private String content = "Original content";
        @Override public void read() { System.out.println("    Content: " + content); }
        @Override public void write(String content) {
            this.content = content;
            System.out.println("    Written: " + content);
        }
    }

    static class ProtectedDocument implements Document {
        private RealDocument doc = new RealDocument();
        private String userRole;

        ProtectedDocument(String role) { this.userRole = role; }

        @Override public void read() { doc.read(); }  // anyone can read
        @Override public void write(String content) {
            if ("admin".equals(userRole)) {
                doc.write(content);
            } else {
                System.out.println("    ACCESS DENIED: " + userRole + " cannot write!");
            }
        }
    }

    // ========================================================
    // 4. FACADE — simplified interface to complex subsystem
    // ========================================================
    // "Remote control: one button for complex TV/receiver/speakers"

    // Complex subsystem classes
    static class CPU {
        void start() { System.out.println("    CPU started"); }
        void shutdown() { System.out.println("    CPU stopped"); }
    }
    static class Memory {
        void load() { System.out.println("    Memory loaded"); }
        void clear() { System.out.println("    Memory cleared"); }
    }
    static class HardDrive {
        void read() { System.out.println("    HDD read boot sector"); }
    }

    // Facade
    static class Computer {
        private CPU cpu = new CPU();
        private Memory memory = new Memory();
        private HardDrive hdd = new HardDrive();

        void start() {
            System.out.println("    --- Computer starting ---");
            cpu.start();
            memory.load();
            hdd.read();
            System.out.println("    --- Computer ready! ---");
        }

        void shutdown() {
            System.out.println("    --- Computer shutting down ---");
            memory.clear();
            cpu.shutdown();
            System.out.println("    --- Computer off ---");
        }
    }

    // ========================================================
    // 5. COMPOSITE — treat individual and group uniformly
    // ========================================================
    // "File system: files and folders treated the same way"

    interface FileSystemItem {
        String getName();
        long getSize();
        void print(String indent);
    }

    static class File implements FileSystemItem {
        private String name;
        private long size;

        File(String name, long size) { this.name = name; this.size = size; }

        @Override public String getName() { return name; }
        @Override public long getSize() { return size; }
        @Override public void print(String indent) {
            System.out.println(indent + "📄 " + name + " (" + size + " bytes)");
        }
    }

    static class Folder implements FileSystemItem {
        private String name;
        private List<FileSystemItem> children = new ArrayList<>();

        Folder(String name) { this.name = name; }

        void add(FileSystemItem item) { children.add(item); }

        @Override public String getName() { return name; }
        @Override public long getSize() {
            return children.stream().mapToLong(FileSystemItem::getSize).sum();
        }
        @Override public void print(String indent) {
            System.out.println(indent + "📁 " + name + " (" + getSize() + " bytes)");
            for (FileSystemItem child : children) {
                child.print(indent + "  ");
            }
        }
    }

    // ========================================================
    // 6. BRIDGE — separate abstraction from implementation
    // ========================================================
    // "Remote control (abstraction) works with any TV (implementation)"

    // Implementation interface
    interface MessageSender {
        void sendMessage(String message);
    }

    static class EmailSender implements MessageSender {
        @Override public void sendMessage(String msg) { System.out.println("    Email: " + msg); }
    }
    static class SlackSender implements MessageSender {
        @Override public void sendMessage(String msg) { System.out.println("    Slack: " + msg); }
    }

    // Abstraction
    static abstract class Alert {
        protected MessageSender sender;
        Alert(MessageSender sender) { this.sender = sender; }
        abstract void send(String message);
    }

    static class UrgentAlert extends Alert {
        UrgentAlert(MessageSender sender) { super(sender); }
        @Override void send(String message) {
            sender.sendMessage("🚨 URGENT: " + message);
        }
    }

    static class InfoAlert extends Alert {
        InfoAlert(MessageSender sender) { super(sender); }
        @Override void send(String message) {
            sender.sendMessage("ℹ️ INFO: " + message);
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Adapter ---
        System.out.println("=== ADAPTER ===\n");
        MediaPlayer player = new MediaAdapter();
        player.play("song.mp3");
        player.play("video.mp4");
        player.play("doc.pdf");

        // --- 2. Decorator ---
        System.out.println("\n=== DECORATOR ===\n");
        Coffee coffee = new BasicCoffee();
        System.out.println("  " + coffee.description() + " = $" + coffee.cost());

        coffee = new MilkDecorator(coffee);
        System.out.println("  " + coffee.description() + " = $" + coffee.cost());

        coffee = new SugarDecorator(coffee);
        System.out.println("  " + coffee.description() + " = $" + coffee.cost());

        coffee = new WhipDecorator(coffee);
        System.out.println("  " + coffee.description() + " = $" + coffee.cost());

        // Java uses Decorator pattern: BufferedReader wraps FileReader
        // InputStream → BufferedInputStream → DataInputStream

        // --- 3. Proxy ---
        System.out.println("\n=== PROXY (Lazy Loading) ===\n");
        Image img1 = new ImageProxy("photo1.jpg");
        Image img2 = new ImageProxy("photo2.jpg");
        System.out.println("  Images created (not loaded yet)");

        img1.display();  // loads on first call
        img1.display();  // already loaded, no reload

        System.out.println("\n  --- Protection Proxy ---");
        Document adminDoc = new ProtectedDocument("admin");
        Document userDoc = new ProtectedDocument("viewer");
        adminDoc.write("New content");
        userDoc.write("Hack attempt");  // denied!

        // --- 4. Facade ---
        System.out.println("\n=== FACADE ===\n");
        Computer computer = new Computer();
        computer.start();
        computer.shutdown();

        // --- 5. Composite ---
        System.out.println("\n=== COMPOSITE ===\n");
        Folder root = new Folder("project");
        Folder src = new Folder("src");
        Folder test = new Folder("test");

        src.add(new File("Main.java", 1200));
        src.add(new File("Utils.java", 800));
        test.add(new File("MainTest.java", 600));

        root.add(src);
        root.add(test);
        root.add(new File("README.md", 300));

        root.print("  ");

        // --- 6. Bridge ---
        System.out.println("\n=== BRIDGE ===\n");
        Alert urgentEmail = new UrgentAlert(new EmailSender());
        Alert infoSlack = new InfoAlert(new SlackSender());
        Alert urgentSlack = new UrgentAlert(new SlackSender());

        urgentEmail.send("Server down!");
        infoSlack.send("Build completed");
        urgentSlack.send("Database overloaded!");

        // --- Summary ---
        System.out.println("\n=== WHEN TO USE ===");
        System.out.println("  Adapter:   Legacy code integration, 3rd party libs");
        System.out.println("  Decorator: Add features without subclassing (I/O streams)");
        System.out.println("  Proxy:     Lazy loading, access control, caching, logging");
        System.out.println("  Facade:    Simplify complex APIs (JDBC, JPA)");
        System.out.println("  Composite: Trees, menus, file systems, UI components");
        System.out.println("  Bridge:    Multiple dimensions of variation");

        System.out.println("\n✓ Structural Patterns Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Adapter: Create XMLToJSON adapter.
 * 2. Decorator: Create EncryptedOutputStream decorator.
 * 3. Proxy: Create CachingProxy for a database query.
 * 4. Composite: Build an organization hierarchy (CEO→VPs→Managers→Employees).
 *
 * NEXT: Chapter 36 — Design Patterns: Behavioral
 */
