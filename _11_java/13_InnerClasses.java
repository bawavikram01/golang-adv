/*
 * ============================================================
 *  CHAPTER 13: INNER CLASSES
 * ============================================================
 *
 *  Inner class = a class defined INSIDE another class.
 *
 *  FOUR TYPES:
 *  ───────────
 *  1. Static Nested Class  — static, belongs to outer class
 *  2. Member Inner Class   — non-static, belongs to outer instance
 *  3. Local Inner Class    — defined inside a method
 *  4. Anonymous Inner Class — unnamed class, defined inline
 *
 *  WHY USE INNER CLASSES?
 *  - Logical grouping of classes used in one place
 *  - Increased encapsulation (inner can access outer's private members)
 *  - More readable and maintainable code
 *  - Implement callbacks and event handlers
 *
 * ============================================================
 */

public class Chapter13_InnerClasses {

    // =====================================================
    //  1. STATIC NESTED CLASS
    // =====================================================

    // A static class inside another class
    // - CANNOT access outer class's instance members
    // - CAN access outer class's static members
    // - Created WITHOUT an instance of outer class

    private static String outerStaticField = "Outer static field";
    private String outerInstanceField = "Outer instance field";

    static class StaticNested {
        void display() {
            System.out.println("Static nested class");
            System.out.println("Can access: " + outerStaticField);
            // System.out.println(outerInstanceField); // ERROR! Can't access instance
        }
    }

    // Practical use: LinkedList.Node, Map.Entry
    static class LinkedList {
        // Node doesn't need access to LinkedList's instance
        static class Node {
            int data;
            Node next;

            Node(int data) {
                this.data = data;
                this.next = null;
            }
        }

        Node head;

        void add(int data) {
            Node newNode = new Node(data);
            newNode.next = head;
            head = newNode;
        }

        void display() {
            Node current = head;
            while (current != null) {
                System.out.print(current.data + " → ");
                current = current.next;
            }
            System.out.println("null");
        }
    }

    // =====================================================
    //  2. MEMBER INNER CLASS (Non-static)
    // =====================================================

    // - CAN access ALL outer class members (including private!)
    // - Needs an instance of outer class to be created
    // - Cannot have static members (except static final constants)

    class MemberInner {
        void display() {
            System.out.println("Member inner class");
            System.out.println("Can access static: " + outerStaticField);
            System.out.println("Can access instance: " + outerInstanceField);
        }
    }

    // Practical: Iterator pattern
    static class MyCollection {
        private int[] items = {10, 20, 30, 40, 50};

        // Inner class can access private 'items'
        class MyIterator {
            private int index = 0;

            boolean hasNext() {
                return index < items.length;
            }

            int next() {
                return items[index++];
            }
        }

        MyIterator iterator() {
            return new MyIterator();
        }
    }

    // =====================================================
    //  3. LOCAL INNER CLASS
    // =====================================================

    // Defined inside a METHOD — scope limited to that method
    // Can access outer class members AND local variables (if effectively final)

    static void demonstrateLocalInner() {
        final String localVar = "I'm a local variable"; // effectively final

        class LocalInner {
            void display() {
                System.out.println("Local inner class");
                System.out.println("Can access: " + localVar);
                System.out.println("Can access: " + outerStaticField);
            }
        }

        LocalInner local = new LocalInner();
        local.display();
    }

    // =====================================================
    //  4. ANONYMOUS INNER CLASS
    // =====================================================

    // No name, defined and instantiated in a single expression
    // Used to override methods on the fly

    interface Greeter {
        void greet(String name);
    }

    static abstract class Formatter {
        abstract String format(String text);
    }

    // =====================================================
    //  MAIN
    // =====================================================

    public static void main(String[] args) {

        // --- 1. Static Nested Class ---
        System.out.println("=== STATIC NESTED CLASS ===\n");

        // Created WITHOUT outer class instance
        StaticNested nested = new StaticNested();
        nested.display();

        // Practical: LinkedList with static Node
        System.out.println();
        LinkedList list = new LinkedList();
        list.add(30);
        list.add(20);
        list.add(10);
        list.display();

        // --- 2. Member Inner Class ---
        System.out.println("\n=== MEMBER INNER CLASS ===\n");

        // MUST create outer instance first!
        Chapter13_InnerClasses outer = new Chapter13_InnerClasses();
        MemberInner inner = outer.new MemberInner(); // syntax: outer.new Inner()
        inner.display();

        // Practical: Iterator
        System.out.println();
        MyCollection collection = new MyCollection();
        MyCollection.MyIterator it = collection.iterator();
        System.out.print("Iterating: ");
        while (it.hasNext()) {
            System.out.print(it.next() + " ");
        }
        System.out.println();

        // --- 3. Local Inner Class ---
        System.out.println("\n=== LOCAL INNER CLASS ===\n");
        demonstrateLocalInner();

        // --- 4. Anonymous Inner Class ---
        System.out.println("\n=== ANONYMOUS INNER CLASS ===\n");

        // Anonymous implementation of an interface
        Greeter politeGreeter = new Greeter() {
            @Override
            public void greet(String name) {
                System.out.println("Good day, " + name + "! How do you do?");
            }
        };
        politeGreeter.greet("Vikram");

        Greeter casualGreeter = new Greeter() {
            @Override
            public void greet(String name) {
                System.out.println("Hey " + name + "! What's up?");
            }
        };
        casualGreeter.greet("Vikram");

        // Anonymous implementation of abstract class
        Formatter upperFormatter = new Formatter() {
            @Override
            String format(String text) {
                return text.toUpperCase();
            }
        };

        Formatter reverseFormatter = new Formatter() {
            @Override
            String format(String text) {
                return new StringBuilder(text).reverse().toString();
            }
        };

        System.out.println("Upper: " + upperFormatter.format("hello world"));
        System.out.println("Reverse: " + reverseFormatter.format("hello world"));

        // Anonymous with Runnable (threading preview)
        Runnable task = new Runnable() {
            @Override
            public void run() {
                System.out.println("Running anonymous Runnable!");
            }
        };
        task.run();

        // Lambda equivalent (Java 8+) — more concise
        Greeter lambdaGreeter = name -> System.out.println("Lambda says hi to " + name);
        lambdaGreeter.greet("Vikram");

        // --- 5. Comparison ---
        System.out.println("\n=== COMPARISON ===\n");
        System.out.println("┌──────────────────┬───────────┬────────────┬─────────────┬──────────────┐");
        System.out.println("│ Feature          │ Static    │ Member     │ Local       │ Anonymous    │");
        System.out.println("│                  │ Nested    │ Inner      │ Inner       │ Inner        │");
        System.out.println("├──────────────────┼───────────┼────────────┼─────────────┼──────────────┤");
        System.out.println("│ Access outer     │ static    │ all        │ all +       │ all +        │");
        System.out.println("│ members?         │ only      │ members    │ final local │ final local  │");
        System.out.println("│ Needs outer      │ No        │ Yes        │ Yes         │ Yes          │");
        System.out.println("│ instance?        │           │            │             │              │");
        System.out.println("│ Can have static  │ Yes       │ No         │ No          │ No           │");
        System.out.println("│ members?         │           │            │             │              │");
        System.out.println("│ Reusable?        │ Yes       │ Yes        │ No          │ No           │");
        System.out.println("└──────────────────┴───────────┴────────────┴─────────────┴──────────────┘");
    }
}

/*
 * ============================================================
 *  EXERCISES
 * ============================================================
 *
 *  1. Create a `Stack` class with a static nested `Node` class.
 *     Implement push(), pop(), peek(), isEmpty().
 *
 *  2. Create a `Button` class with a member inner class
 *     `ClickHandler`. Simulate click events.
 *
 *  3. Use an anonymous inner class to sort a String array
 *     by length (implement Comparator<String>).
 *
 *  4. Convert the anonymous inner classes above to lambda
 *     expressions wherever possible.
 *
 * ============================================================
 *  WHAT'S NEXT: Chapter 14 — Enums
 * ============================================================
 */
