import java.lang.annotation.*;
import java.lang.reflect.*;

/**
 * PHASE 1.1 — ANNOTATIONS - Spring's Language
 * 
 * Theory
Annotations are metadata you attach to classes, methods, or fields using @. They don't execute code themselves — but frameworks read them and act accordingly.

Spring is annotation-driven. Almost everything in Spring is an annotation: @Component, @Autowired, @RestController, @GetMapping, @Transactional...

Without understanding annotations, Spring looks like a pile of @ symbols doing magic.

Analogy
Annotations are like sticky notes on a document. The document (code) works on its own, but the sticky notes tell the reviewer (framework) what to do with it. A sticky note saying "URGENT" tells the mail system to prioritize it. The letter's content doesn't change.

 * 
 * Spring is BUILT on annotations. Every @Component, @Autowired, @GetMapping
 * is an annotation. Understanding how they work removes all the "magic".
 */

// ============================================================
// STEP 1: Using Built-in Annotations
// ============================================================

class BuiltInAnnotations {

    @Override  // Tells compiler: "I'm overriding a parent method." Compile error if not.
    public String toString() {
        return "BuiltInAnnotations instance";
    }

    @Deprecated  // Tells: "Don't use this, it's old."
    public void oldMethod() {
        System.out.println("I'm deprecated!");
    }

    @SuppressWarnings("unchecked")  // Tells compiler: "I know this is risky, shut up."
    public void riskyMethod() { }
}


// ============================================================
// STEP 2: Creating Your Own Annotation
// ============================================================

// This is EXACTLY how Spring's annotations are defined.
// @interface = "I'm defining an annotation"

@Retention(RetentionPolicy.RUNTIME)  // Annotation survives until runtime (Spring needs this!)
@Target(ElementType.METHOD)          // Can only be placed on methods
@interface LogExecutionTime {
    // Annotations can have parameters (called "elements")
    String label() default ""; // optional label, default is empty
}

// Another custom annotation — for marking important methods
@Retention(RetentionPolicy.RUNTIME)
@Target(ElementType.TYPE)  // Can only be placed on classes
@interface Component {
    String value() default ""; // This is EXACTLY like Spring's @Component!
}


// ============================================================
// STEP 3: Using Custom Annotations
// ============================================================

@Component("userService")  // Just like Spring! Marking this class as a component.
class UserService {

    @LogExecutionTime(label = "find-user")
    public String findUser(Long id) {
        // Simulate some work
        try { Thread.sleep(50); } catch (Exception e) {}
        return "User-" + id;
    }

    @LogExecutionTime(label = "save-user")
    public void saveUser(String name) {
        try { Thread.sleep(100); } catch (Exception e) {}
        System.out.println("  Saved user: " + name);
    }

    public void deleteUser(Long id) {
        // This method has NO annotation — the framework will skip it
        System.out.println("  Deleted user: " + id);
    }
}


// ============================================================
// STEP 4: Reading Annotations at Runtime (This is what Spring does!)
// ============================================================

class AnnotationProcessor {

    // This method scans a class and processes its annotations.
    // Spring does EXACTLY this at startup for EVERY class it finds.
    public static void process(Object obj) throws Exception {
        Class<?> clazz = obj.getClass();

        // Check if class has @Component
        if (clazz.isAnnotationPresent(Component.class)) {
            Component comp = clazz.getAnnotation(Component.class);
            System.out.println("Found @Component on class: " + clazz.getSimpleName());
            System.out.println("  Component name: \"" + comp.value() + "\"");
        }

        System.out.println();

        // Scan all methods for @LogExecutionTime
        for (Method method : clazz.getDeclaredMethods()) {
            if (method.isAnnotationPresent(LogExecutionTime.class)) {
                LogExecutionTime annotation = method.getAnnotation(LogExecutionTime.class);

                System.out.println("Found @LogExecutionTime on method: " + method.getName());
                System.out.println("  Label: \"" + annotation.label() + "\"");

                // Measure execution time (like Spring AOP does!)
                long start = System.currentTimeMillis();
                method.invoke(obj, method.getParameterTypes()[0] == Long.class ? 1L : "Alice");
                long end = System.currentTimeMillis();

                System.out.println("  ⏱ Execution time: " + (end - start) + "ms");
                System.out.println();
            }
        }
    }
}


public class Step3_Annotations {
    public static void main(String[] args) throws Exception {

        System.out.println("=== BUILT-IN ANNOTATIONS ===");
        BuiltInAnnotations obj = new BuiltInAnnotations();
        System.out.println(obj.toString());    // @Override in action
        obj.oldMethod();                        // @Deprecated — IDE shows warning
        System.out.println();

        System.out.println("=== CUSTOM ANNOTATIONS + PROCESSING ===");
        System.out.println("(This is what Spring does at startup!)\n");

        UserService userService = new UserService();
        AnnotationProcessor.process(userService);

        System.out.println("=== KEY TAKEAWAY ===");
        System.out.println("Annotations are just LABELS. They do nothing alone.");
        System.out.println("A PROCESSOR (like Spring) reads them and acts on them.");
        System.out.println("@Component = Spring reads it and says 'I'll manage this class'");
        System.out.println("@Autowired = Spring reads it and says 'I'll inject a dependency here'");
    }
}
