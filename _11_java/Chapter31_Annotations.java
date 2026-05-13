/*
 * ============================================================
 *  CHAPTER 31: ANNOTATIONS
 * ============================================================
 *  Annotation = metadata about code. Does not change behavior
 *  by itself — but tools, compilers, and frameworks read them.
 *
 *  Built-in Annotations:
 *    @Override       → method overrides superclass method
 *    @Deprecated     → marks as obsolete
 *    @SuppressWarnings → suppress compiler warnings
 *    @FunctionalInterface → marks single-abstract-method interface
 *    @SafeVarargs    → suppresses heap pollution warning
 *
 *  Meta-Annotations (annotations ON annotations):
 *    @Target        → where annotation can be used
 *    @Retention     → how long annotation is kept
 *    @Documented    → include in javadoc
 *    @Inherited     → subclasses inherit the annotation
 *    @Repeatable    → can be applied multiple times
 *
 *  Retention Policies:
 *    SOURCE  → discarded by compiler (e.g., @Override)
 *    CLASS   → kept in .class file but not at runtime (default)
 *    RUNTIME → available via reflection at runtime
 * ============================================================
 */

import java.lang.annotation.*;
import java.lang.reflect.*;
import java.util.*;

public class Chapter31_Annotations {

    // === USING BUILT-IN ANNOTATIONS ===

    static class Animal {
        @Deprecated(since = "2.0")
        void makeSound() {
            System.out.println("  Generic sound");
        }

        void speak(String sound) {
            System.out.println("  " + sound);
        }
    }

    static class Dog extends Animal {
        @Override
        void speak(String sound) {
            System.out.println("  Dog says: " + sound);
        }
    }

    @FunctionalInterface
    interface Transformer<T> {
        T transform(T input);
        // Only ONE abstract method allowed
    }

    @SuppressWarnings("unchecked")
    static void uncheckedExample() {
        List rawList = new ArrayList();  // raw type warning suppressed
        rawList.add("hello");
    }

    // === CREATING CUSTOM ANNOTATIONS ===

    // Simple marker annotation
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.METHOD)
    @interface Test {
    }

    // Annotation with elements (like properties)
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.METHOD)
    @interface TestCase {
        String name() default "unnamed";
        int priority() default 0;
        String[] tags() default {};
    }

    // Annotation for fields
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.FIELD)
    @interface NotNull {
        String message() default "Field cannot be null";
    }

    // Annotation for classes
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.TYPE)
    @interface Entity {
        String table();
    }

    // Column annotation for fields
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.FIELD)
    @interface Column {
        String name() default "";
        boolean nullable() default true;
        int length() default 255;
    }

    // === REPEATABLE ANNOTATION ===
    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.METHOD)
    @Repeatable(Schedules.class)
    @interface Schedule {
        String cron();
    }

    @Retention(RetentionPolicy.RUNTIME)
    @Target(ElementType.METHOD)
    @interface Schedules {
        Schedule[] value();
    }

    // === USING CUSTOM ANNOTATIONS ===

    @Entity(table = "users")
    static class User {
        @Column(name = "user_id", nullable = false)
        @NotNull
        private String id;

        @Column(name = "user_name", length = 100)
        @NotNull(message = "Name is required")
        private String name;

        @Column(name = "email")
        private String email;

        User(String id, String name, String email) {
            this.id = id;
            this.name = name;
            this.email = email;
        }
    }

    // Test runner example
    static class MyTests {
        @Test
        void testAddition() { System.out.println("    ✓ testAddition passed"); }

        @Test
        void testSubtraction() { System.out.println("    ✓ testSubtraction passed"); }

        void notATest() { System.out.println("    This should NOT run"); }

        @TestCase(name = "Login Test", priority = 1, tags = {"smoke", "auth"})
        void testLogin() { System.out.println("    ✓ testLogin passed"); }

        @Schedule(cron = "0 0 * * *")
        @Schedule(cron = "0 12 * * *")
        void scheduledTask() { System.out.println("    Scheduled task ran"); }
    }

    // === ANNOTATION PROCESSOR (runtime via reflection) ===

    static void runTests(Object testInstance) throws Exception {
        Class<?> clazz = testInstance.getClass();

        for (Method method : clazz.getDeclaredMethods()) {
            // Check for @Test
            if (method.isAnnotationPresent(Test.class)) {
                method.setAccessible(true);
                method.invoke(testInstance);
            }

            // Check for @TestCase
            if (method.isAnnotationPresent(TestCase.class)) {
                TestCase tc = method.getAnnotation(TestCase.class);
                System.out.println("    Running: " + tc.name() +
                    " (priority=" + tc.priority() +
                    ", tags=" + Arrays.toString(tc.tags()) + ")");
                method.setAccessible(true);
                method.invoke(testInstance);
            }
        }
    }

    static void validateNotNull(Object obj) throws Exception {
        Class<?> clazz = obj.getClass();
        for (Field field : clazz.getDeclaredFields()) {
            if (field.isAnnotationPresent(NotNull.class)) {
                field.setAccessible(true);
                Object value = field.get(obj);
                if (value == null) {
                    NotNull ann = field.getAnnotation(NotNull.class);
                    throw new IllegalArgumentException(
                        field.getName() + ": " + ann.message()
                    );
                }
            }
        }
    }

    static void inspectEntity(Class<?> clazz) {
        if (clazz.isAnnotationPresent(Entity.class)) {
            Entity entity = clazz.getAnnotation(Entity.class);
            System.out.println("  Table: " + entity.table());

            for (Field field : clazz.getDeclaredFields()) {
                if (field.isAnnotationPresent(Column.class)) {
                    Column col = field.getAnnotation(Column.class);
                    String colName = col.name().isEmpty() ? field.getName() : col.name();
                    System.out.printf("    Column: %-15s nullable=%-5s length=%d%n",
                        colName, col.nullable(), col.length());
                }
            }
        }
    }

    public static void main(String[] args) throws Exception {

        // --- 1. Built-in Annotations ---
        System.out.println("=== BUILT-IN ANNOTATIONS ===\n");
        Dog dog = new Dog();
        dog.speak("Woof!");

        Transformer<String> upper = String::toUpperCase;
        System.out.println("  Transformed: " + upper.transform("hello"));

        // --- 2. Custom Test Runner ---
        System.out.println("\n=== CUSTOM TEST RUNNER ===\n");
        runTests(new MyTests());

        // --- 3. @NotNull Validation ---
        System.out.println("\n=== @NotNull VALIDATION ===\n");
        User validUser = new User("1", "Alice", "alice@test.com");
        validateNotNull(validUser);
        System.out.println("  Valid user passed validation");

        try {
            User invalidUser = new User("2", null, "bob@test.com");
            validateNotNull(invalidUser);
        } catch (IllegalArgumentException e) {
            System.out.println("  Validation failed: " + e.getMessage());
        }

        // --- 4. Entity Inspection ---
        System.out.println("\n=== ENTITY INSPECTION ===\n");
        inspectEntity(User.class);

        // --- 5. Repeatable Annotations ---
        System.out.println("\n=== REPEATABLE ANNOTATIONS ===\n");
        Method method = MyTests.class.getDeclaredMethod("scheduledTask");
        Schedule[] schedules = method.getAnnotationsByType(Schedule.class);
        for (Schedule s : schedules) {
            System.out.println("  Schedule: " + s.cron());
        }

        // --- 6. All Meta-Annotations Explained ---
        System.out.println("\n=== META-ANNOTATIONS ===");
        System.out.println("  @Target(ElementType.TYPE)       → class, interface, enum");
        System.out.println("  @Target(ElementType.FIELD)      → fields");
        System.out.println("  @Target(ElementType.METHOD)     → methods");
        System.out.println("  @Target(ElementType.PARAMETER)  → method parameters");
        System.out.println("  @Target(ElementType.CONSTRUCTOR)→ constructors");
        System.out.println("  @Target(ElementType.LOCAL_VARIABLE) → local vars (source only)");
        System.out.println("  @Target(ElementType.ANNOTATION_TYPE) → other annotations");
        System.out.println("  @Target(ElementType.PACKAGE)    → package-info.java");
        System.out.println("  @Target(ElementType.TYPE_USE)   → any type usage (Java 8+)");

        System.out.println("\n✓ Annotations Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Create @Range(min, max) for int fields. Write a validator.
 * 2. Create @Benchmark annotation that measures method execution time.
 * 3. Create @JsonField(name) to map Java fields to JSON keys. Write serializer.
 * 4. Create @Retry(times) that retries a method on failure.
 *
 * NEXT: Chapter 32 — Reflection
 */
