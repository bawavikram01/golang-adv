/*
 * ============================================================
 *  CHAPTER 32: REFLECTION
 * ============================================================
 *  Reflection = ability to inspect and modify classes, methods,
 *  fields, constructors AT RUNTIME.
 *
 *  Power: frameworks use it (Spring, Hibernate, JUnit, Jackson)
 *  Danger: bypasses access control, slower, breaks encapsulation
 *
 *  Key classes (java.lang.reflect):
 *    Class<T>      → represents a class
 *    Field         → represents a field
 *    Method        → represents a method
 *    Constructor   → represents a constructor
 *    Modifier      → decode access modifiers
 * ============================================================
 */

import java.lang.reflect.*;
import java.util.*;

public class Chapter32_Reflection {

    // Sample classes to reflect on
    static class Person {
        public String name;
        private int age;
        protected String email;

        public Person() { this.name = "Unknown"; this.age = 0; }
        public Person(String name, int age) { this.name = name; this.age = age; }
        private Person(String name) { this.name = name; this.age = -1; }

        public String getName() { return name; }
        public int getAge() { return age; }
        private void setAge(int age) { this.age = age; }

        public String greet(String greeting) {
            return greeting + ", " + name + "! Age: " + age;
        }

        private String secret() { return "I am a secret method!"; }

        @Override
        public String toString() { return "Person{name='" + name + "', age=" + age + "}"; }
    }

    interface Printable { void print(); }
    interface Serializable {}

    static abstract class Shape implements Printable, Serializable {
        abstract double area();
    }

    static class Circle extends Shape {
        private double radius;
        Circle(double radius) { this.radius = radius; }
        @Override public double area() { return Math.PI * radius * radius; }
        @Override public void print() { System.out.println("Circle r=" + radius); }
    }

    public static void main(String[] args) throws Exception {

        // --- 1. Getting Class Objects ---
        System.out.println("=== GETTING CLASS OBJECTS ===\n");

        // Three ways to get a Class object
        Class<?> c1 = Person.class;                           // from class literal
        Class<?> c2 = new Person().getClass();                // from instance
        Class<?> c3 = Class.forName("Chapter32_Reflection$Person"); // from name (inner class syntax)

        System.out.println("  c1: " + c1.getName());
        System.out.println("  c2: " + c2.getSimpleName());
        System.out.println("  c3: " + c3.getCanonicalName());
        System.out.println("  Are same? " + (c1 == c2 && c2 == c3));

        // --- 2. Inspecting Class Structure ---
        System.out.println("\n=== CLASS INSPECTION ===\n");

        Class<?> clazz = Circle.class;
        System.out.println("  Name: " + clazz.getSimpleName());
        System.out.println("  Superclass: " + clazz.getSuperclass().getSimpleName());
        System.out.println("  Interfaces: " + Arrays.toString(
            clazz.getSuperclass().getInterfaces()));
        System.out.println("  Is abstract? " + Modifier.isAbstract(Shape.class.getModifiers()));
        System.out.println("  Package: " + clazz.getPackageName());

        // --- 3. Inspecting Fields ---
        System.out.println("\n=== FIELDS ===\n");

        Class<?> personClass = Person.class;

        // getFields() → only PUBLIC fields (including inherited)
        System.out.println("  Public fields:");
        for (Field f : personClass.getFields()) {
            System.out.println("    " + f.getType().getSimpleName() + " " + f.getName());
        }

        // getDeclaredFields() → ALL fields of this class (not inherited)
        System.out.println("  All declared fields:");
        for (Field f : personClass.getDeclaredFields()) {
            String modifiers = Modifier.toString(f.getModifiers());
            System.out.println("    " + modifiers + " " + f.getType().getSimpleName() + " " + f.getName());
        }

        // --- 4. Reading/Writing Fields ---
        System.out.println("\n=== FIELD ACCESS ===\n");

        Person person = new Person("Alice", 30);

        // Public field — direct access
        Field nameField = personClass.getField("name");
        System.out.println("  Name: " + nameField.get(person));
        nameField.set(person, "Bob");
        System.out.println("  After set: " + nameField.get(person));

        // Private field — need setAccessible(true)
        Field ageField = personClass.getDeclaredField("age");
        ageField.setAccessible(true);  // bypass private!
        System.out.println("  Age (private): " + ageField.get(person));
        ageField.set(person, 25);
        System.out.println("  After set: " + ageField.get(person));

        // --- 5. Inspecting Methods ---
        System.out.println("\n=== METHODS ===\n");

        // getMethods() → all public methods (including Object's)
        System.out.println("  Public methods (declared in Person):");
        for (Method m : personClass.getDeclaredMethods()) {
            String modifiers = Modifier.toString(m.getModifiers());
            String params = Arrays.toString(
                Arrays.stream(m.getParameterTypes())
                    .map(Class::getSimpleName)
                    .toArray(String[]::new)
            );
            System.out.println("    " + modifiers + " " + m.getReturnType().getSimpleName()
                + " " + m.getName() + params);
        }

        // --- 6. Invoking Methods ---
        System.out.println("\n=== METHOD INVOCATION ===\n");

        // Public method with parameter
        Method greetMethod = personClass.getMethod("greet", String.class);
        String result = (String) greetMethod.invoke(person, "Hello");
        System.out.println("  greet() result: " + result);

        // Private method
        Method secretMethod = personClass.getDeclaredMethod("secret");
        secretMethod.setAccessible(true);
        String secretResult = (String) secretMethod.invoke(person);
        System.out.println("  secret() result: " + secretResult);

        // Private setter
        Method setAgeMethod = personClass.getDeclaredMethod("setAge", int.class);
        setAgeMethod.setAccessible(true);
        setAgeMethod.invoke(person, 99);
        System.out.println("  After setAge(99): " + person);

        // --- 7. Constructors ---
        System.out.println("\n=== CONSTRUCTORS ===\n");

        // Get all constructors
        System.out.println("  Constructors:");
        for (Constructor<?> c : personClass.getDeclaredConstructors()) {
            String modifiers = Modifier.toString(c.getModifiers());
            String params = Arrays.toString(
                Arrays.stream(c.getParameterTypes())
                    .map(Class::getSimpleName)
                    .toArray(String[]::new)
            );
            System.out.println("    " + modifiers + " Person" + params);
        }

        // Create instance with public constructor
        Constructor<?> pubCtor = personClass.getConstructor(String.class, int.class);
        Person p1 = (Person) pubCtor.newInstance("Charlie", 40);
        System.out.println("  Created: " + p1);

        // Create instance with private constructor
        Constructor<?> privCtor = personClass.getDeclaredConstructor(String.class);
        privCtor.setAccessible(true);
        Person p2 = (Person) privCtor.newInstance("Secret");
        System.out.println("  Private ctor: " + p2);

        // --- 8. Working with Arrays ---
        System.out.println("\n=== ARRAYS VIA REFLECTION ===\n");

        // Create array reflectively
        Object arr = Array.newInstance(int.class, 5);
        Array.set(arr, 0, 10);
        Array.set(arr, 1, 20);
        Array.set(arr, 2, 30);
        System.out.println("  Array[0]: " + Array.get(arr, 0));
        System.out.println("  Length: " + Array.getLength(arr));
        System.out.println("  Component type: " + arr.getClass().getComponentType());

        // --- 9. Generics and Type Info ---
        System.out.println("\n=== GENERIC TYPE INFO ===\n");

        // Generics are erased at runtime, but field declarations retain info
        class Container {
            public List<String> names;
            public Map<Integer, List<String>> data;
        }

        Field namesField = Container.class.getDeclaredField("names");
        Type genericType = namesField.getGenericType();
        if (genericType instanceof ParameterizedType) {
            ParameterizedType pt = (ParameterizedType) genericType;
            System.out.println("  Raw type: " + pt.getRawType());
            System.out.println("  Type args: " + Arrays.toString(pt.getActualTypeArguments()));
        }

        // --- 10. Practical: Simple JSON Serializer ---
        System.out.println("\n=== MINI JSON SERIALIZER ===\n");

        Person jsonPerson = new Person("Alice", 30);
        String json = toJson(jsonPerson);
        System.out.println("  " + json);

        // --- 11. Practical: Object Copier ---
        System.out.println("\n=== OBJECT COPIER ===\n");
        Person original = new Person("Dave", 50);
        Person copy = shallowCopy(original);
        System.out.println("  Original: " + original);
        System.out.println("  Copy: " + copy);
        System.out.println("  Same object? " + (original == copy));

        // --- Warnings ---
        System.out.println("\n=== REFLECTION CAVEATS ===");
        System.out.println("  1. SLOWER than direct access (no JIT optimization)");
        System.out.println("  2. Bypasses access control → breaks encapsulation");
        System.out.println("  3. No compile-time type checking → runtime errors");
        System.out.println("  4. Can break singleton patterns");
        System.out.println("  5. May not work with Java modules (Java 9+)");
        System.out.println("  6. Use ONLY when truly needed (frameworks, libraries)");

        System.out.println("\n✓ Reflection Complete!");
    }

    // Simple JSON serializer using reflection
    static String toJson(Object obj) throws Exception {
        StringBuilder sb = new StringBuilder("{");
        Field[] fields = obj.getClass().getDeclaredFields();

        for (int i = 0; i < fields.length; i++) {
            fields[i].setAccessible(true);
            String name = fields[i].getName();
            Object value = fields[i].get(obj);

            sb.append("\"").append(name).append("\":");
            if (value instanceof String) {
                sb.append("\"").append(value).append("\"");
            } else {
                sb.append(value);
            }
            if (i < fields.length - 1) sb.append(",");
        }
        sb.append("}");
        return sb.toString();
    }

    // Shallow copy using reflection
    @SuppressWarnings("unchecked")
    static <T> T shallowCopy(T obj) throws Exception {
        Class<?> clazz = obj.getClass();
        Constructor<?> ctor = clazz.getDeclaredConstructor();
        ctor.setAccessible(true);
        T copy = (T) ctor.newInstance();

        for (Field f : clazz.getDeclaredFields()) {
            f.setAccessible(true);
            f.set(copy, f.get(obj));
        }
        return copy;
    }
}

/*
 * EXERCISES:
 * 1. Write a method that prints all methods of any class, grouped by public/private.
 * 2. Create a dependency injection container: @Inject on fields, auto-wire.
 * 3. Build a simple ORM: read @Entity/@Column and generate SQL CREATE TABLE.
 * 4. Create a class that can't be instantiated via reflection (throw in constructor).
 *
 * NEXT: Chapter 33 — JVM Internals
 */
