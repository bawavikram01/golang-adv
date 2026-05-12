/*
 * =============================================================
 * STRUCTURAL PATTERN 5: COMPOSITE
 * =============================================================
 *
 * INTENT: Compose objects into TREE structures to represent
 *         part-whole hierarchies. Treat individual objects
 *         and compositions UNIFORMLY.
 *
 * ANALOGY: File system — a Directory contains Files AND other
 *          Directories. Both respond to getSize(). You don't
 *          care whether it's a file or a folder of 1000 files.
 *
 * USE WHEN:
 *   - Tree/hierarchy structure (org charts, menus, file systems)
 *   - You want to treat a group the same as an individual
 *   - Operations should work recursively on the tree
 *
 * REAL EXAMPLES: java.awt.Container, React component tree,
 *                DOM tree, organization hierarchy
 */

import java.util.*;

public class CompositePattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // File System Example
        // ═══════════════════════════════════════════════════════
        System.out.println("=== FILE SYSTEM COMPOSITE ===");

        // Build the tree
        Directory root = new Directory("root");
        
        File readme = new File("README.md", 2);
        File gitignore = new File(".gitignore", 1);
        
        Directory src = new Directory("src");
        src.add(new File("Main.java", 5));
        src.add(new File("Utils.java", 3));
        
        Directory test = new Directory("test");
        test.add(new File("MainTest.java", 4));
        
        Directory docs = new Directory("docs");
        docs.add(new File("guide.pdf", 50));
        docs.add(new File("api.html", 10));
        
        root.add(readme);
        root.add(gitignore);
        root.add(src);
        root.add(test);
        root.add(docs);

        // Treat the ENTIRE tree uniformly
        root.display("");
        System.out.println("\nTotal size: " + root.getSize() + " KB");
        System.out.println("src/ size:  " + src.getSize() + " KB");

        // ═══════════════════════════════════════════════════════
        // Organization Hierarchy
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ORGANIZATION COMPOSITE ===");

        Manager ceo = new Manager("Alice (CEO)", 200000);

        Manager vpEng = new Manager("Bob (VP Eng)", 150000);
        vpEng.add(new Developer("Charlie", 100000));
        vpEng.add(new Developer("Diana", 95000));

        Manager vpSales = new Manager("Eve (VP Sales)", 140000);
        vpSales.add(new Developer("Frank", 80000));

        ceo.add(vpEng);
        ceo.add(vpSales);

        ceo.display("  ");
        System.out.println("\nTotal org salary: $" + ceo.getSalary());
        System.out.println("Engineering salary: $" + vpEng.getSalary());

        // ═══════════════════════════════════════════════════════
        // Menu System (UI)
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== MENU COMPOSITE ===");

        MenuGroup mainMenu = new MenuGroup("Main Menu");

        MenuGroup fileMenu = new MenuGroup("File");
        fileMenu.add(new MenuItem("New", "Ctrl+N"));
        fileMenu.add(new MenuItem("Open", "Ctrl+O"));
        fileMenu.add(new MenuItem("Save", "Ctrl+S"));

        MenuGroup editMenu = new MenuGroup("Edit");
        editMenu.add(new MenuItem("Undo", "Ctrl+Z"));
        editMenu.add(new MenuItem("Redo", "Ctrl+Y"));

        MenuGroup findSubmenu = new MenuGroup("Find");  // nested submenu
        findSubmenu.add(new MenuItem("Find", "Ctrl+F"));
        findSubmenu.add(new MenuItem("Replace", "Ctrl+H"));
        editMenu.add(findSubmenu);

        mainMenu.add(fileMenu);
        mainMenu.add(editMenu);
        mainMenu.add(new MenuItem("Quit", "Ctrl+Q"));

        mainMenu.render("");
    }
}

// ═══════════════════════════════════════════════════════════════
// FILE SYSTEM COMPOSITE
// ═══════════════════════════════════════════════════════════════

// Component — common interface for both leaves and composites
interface FileSystemItem {
    String getName();
    int getSize();   // KB
    void display(String indent);
}

// Leaf — a single file
class File implements FileSystemItem {
    private String name;
    private int size;

    public File(String name, int size) {
        this.name = name;
        this.size = size;
    }

    @Override public String getName() { return name; }
    @Override public int getSize()    { return size; }

    @Override
    public void display(String indent) {
        System.out.println(indent + "📄 " + name + " (" + size + " KB)");
    }
}

// Composite — contains other items (files AND directories)
class Directory implements FileSystemItem {
    private String name;
    private List<FileSystemItem> children = new ArrayList<>();

    public Directory(String name) { this.name = name; }

    public void add(FileSystemItem item)    { children.add(item); }
    public void remove(FileSystemItem item) { children.remove(item); }

    @Override public String getName() { return name; }

    @Override
    public int getSize() {
        // Recursively sum all children sizes
        return children.stream().mapToInt(FileSystemItem::getSize).sum();
    }

    @Override
    public void display(String indent) {
        System.out.println(indent + "📁 " + name + "/");
        for (FileSystemItem child : children) {
            child.display(indent + "  ");
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// ORGANIZATION HIERARCHY
// ═══════════════════════════════════════════════════════════════
interface Employee {
    String getName();
    int getSalary();
    void display(String indent);
}

class Developer implements Employee {
    private String name;
    private int salary;

    public Developer(String name, int salary) {
        this.name = name;
        this.salary = salary;
    }

    @Override public String getName()  { return name; }
    @Override public int getSalary()   { return salary; }

    @Override
    public void display(String indent) {
        System.out.println(indent + "👨‍💻 " + name + " ($" + salary + ")");
    }
}

class Manager implements Employee {
    private String name;
    private int salary;
    private List<Employee> reports = new ArrayList<>();

    public Manager(String name, int salary) {
        this.name = name;
        this.salary = salary;
    }

    public void add(Employee e)    { reports.add(e); }
    public void remove(Employee e) { reports.remove(e); }

    @Override public String getName() { return name; }

    @Override
    public int getSalary() {
        // Manager's salary + all reports' salaries
        return salary + reports.stream().mapToInt(Employee::getSalary).sum();
    }

    @Override
    public void display(String indent) {
        System.out.println(indent + "👔 " + name + " ($" + salary + ")");
        for (Employee e : reports) {
            e.display(indent + "  ");
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// MENU SYSTEM
// ═══════════════════════════════════════════════════════════════
interface MenuComponent {
    void render(String indent);
}

class MenuItem implements MenuComponent {
    private String label;
    private String shortcut;

    public MenuItem(String label, String shortcut) {
        this.label = label;
        this.shortcut = shortcut;
    }

    @Override
    public void render(String indent) {
        System.out.println(indent + "  " + label + " [" + shortcut + "]");
    }
}

class MenuGroup implements MenuComponent {
    private String title;
    private List<MenuComponent> items = new ArrayList<>();

    public MenuGroup(String title) { this.title = title; }

    public void add(MenuComponent item) { items.add(item); }

    @Override
    public void render(String indent) {
        System.out.println(indent + "▸ " + title);
        for (MenuComponent item : items) {
            item.render(indent + "  ");
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Composite = treat individual and group objects uniformly.
 * ✦ Component (interface) → Leaf + Composite (both implement it).
 * ✦ Composite CONTAINS a list of Components (can be leaves or composites).
 * ✦ Operations like getSize() work recursively through the tree.
 *
 * ✦ Composite vs Decorator:
 *   - Composite: groups objects into trees (1 → many)
 *   - Decorator: wraps one object with layers (1 → 1)
 *
 * ✦ Look for Composite when you see:
 *   - Tree structures (org chart, file system, menus, categories)
 *   - "Calculate total for a group" requirements
 *   - Part-whole relationships
 *
 * COMPILE & RUN:
 *   javac CompositePattern.java && java CompositePattern
 */
