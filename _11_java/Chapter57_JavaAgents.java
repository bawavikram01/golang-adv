/*
 * ============================================================
 *  CHAPTER 57: JAVA AGENTS & INSTRUMENTATION
 * ============================================================
 *  Java Agents intercept class loading to transform bytecode
 *  AT RUNTIME. This is how profilers, code coverage tools,
 *  Spring AOP, Hibernate enhancement, and hot-reload work.
 *
 *  This is one of Java's most powerful and least understood
 *  features. It's pure black magic — and you need to know it.
 *
 *  TOPICS:
 *    1. What Is a Java Agent?
 *    2. premain() — Static Agent (before app starts)
 *    3. agentmain() — Dynamic Agent (attach to running JVM)
 *    4. ClassFileTransformer — Modify Class Bytes
 *    5. Instrumentation API
 *    6. MANIFEST.MF — Agent Packaging
 *    7. Attach API — Live Agent Injection
 *    8. Real-World Uses
 *    9. Building a Complete Agent
 * ============================================================
 *
 *  HOW JAVA AGENTS WORK:
 *  ──────────────────────────────────────────────────
 *
 *  WITHOUT AGENT:
 *    .class file → ClassLoader → defineClass() → Class in JVM
 *
 *  WITH AGENT:
 *    .class file → ClassLoader → TRANSFORMER → defineClass() → Class in JVM
 *                                    ↑
 *                             Your agent intercepts
 *                             the raw bytes here!
 *
 *  You can:
 *    - Add logging to every method (profiling)
 *    - Measure code coverage (JaCoCo)
 *    - Add @Transactional behavior (Spring)
 *    - Hot-swap classes (JRebel, Spring DevTools)
 *    - Monitor memory allocation
 *    - Enforce security policies
 *
 * ============================================================
 */

import java.lang.instrument.*;
import java.security.ProtectionDomain;
import java.util.*;

public class Chapter57_JavaAgents {

    // ========================================================
    // 1. PREMAIN — Static Agent
    // ========================================================
    /*
     * premain() is called BEFORE your application's main() method.
     * The JVM calls it when you use: java -javaagent:myagent.jar MyApp
     *
     * TWO SIGNATURES (JVM tries the first, falls back to second):
     *
     *   public static void premain(String agentArgs, Instrumentation inst)
     *   public static void premain(String agentArgs)
     *
     * agentArgs: string passed after '=' in -javaagent:agent.jar=THESE_ARGS
     * inst: the Instrumentation object — your tool for transforming classes
     *
     * EXAMPLE AGENT CLASS:
     */

    // This would be in a separate JAR (the agent)
    public static void premain(String agentArgs, Instrumentation inst) {
        System.out.println("[AGENT] premain called with args: " + agentArgs);
        System.out.println("[AGENT] Registering class transformer...");

        // Register a transformer
        inst.addTransformer(new LoggingTransformer(), false);

        // Print what's already loaded
        System.out.println("[AGENT] Already loaded classes: " + inst.getAllLoadedClasses().length);
    }

    // ========================================================
    // 2. AGENTMAIN — Dynamic Agent (Attach to Running JVM)
    // ========================================================
    /*
     * agentmain() is called when an agent is attached to an
     * ALREADY RUNNING JVM using the Attach API.
     *
     *   public static void agentmain(String agentArgs, Instrumentation inst)
     *   public static void agentmain(String agentArgs)
     *
     * HOW TO ATTACH:
     *   // In another JVM process:
     *   VirtualMachine vm = VirtualMachine.attach(pid);
     *   vm.loadAgent("/path/to/agent.jar", "optional-args");
     *   vm.detach();
     *
     * This is how:
     *   - VisualVM attaches profilers
     *   - IntelliJ debugger injects code
     *   - Arthas debugging tool works
     *   - JMX tools connect to running apps
     */

    public static void agentmain(String agentArgs, Instrumentation inst) {
        System.out.println("[AGENT] agentmain called — attached to running JVM!");
        System.out.println("[AGENT] Can retransform: " + inst.isRetransformClassesSupported());

        // Can retransform already-loaded classes!
        if (inst.isRetransformClassesSupported()) {
            inst.addTransformer(new LoggingTransformer(), true); // canRetransform=true
            try {
                // Retransform classes that are already loaded
                // inst.retransformClasses(SomeClass.class);
            } catch (Exception e) {
                e.printStackTrace();
            }
        }
    }

    // ========================================================
    // 3. ClassFileTransformer — The Core Interface
    // ========================================================

    static class LoggingTransformer implements ClassFileTransformer {

        @Override
        public byte[] transform(ClassLoader loader,
                                String className,
                                Class<?> classBeingRedefined,
                                ProtectionDomain protectionDomain,
                                byte[] classfileBuffer)
                throws IllegalClassFormatException {

            // className uses '/' separators, e.g., "com/example/MyClass"

            // Skip JDK/system classes
            if (className == null || className.startsWith("java/")
                    || className.startsWith("javax/")
                    || className.startsWith("sun/")
                    || className.startsWith("jdk/")) {
                return null; // null = don't transform
            }

            System.out.println("[TRANSFORMER] Loading: " + className
                + " (" + classfileBuffer.length + " bytes)");

            // Return null = use original bytes
            // Return modified bytes = use transformed class
            // throw IllegalClassFormatException = reject class

            // In a real agent, you'd use a bytecode library here:
            //   - ASM (low-level, fast)
            //   - ByteBuddy (high-level, easy)
            //   - Javassist (source-level API)

            return null; // no actual transformation in this demo
        }
    }

    // ========================================================
    // 4. INSTRUMENTATION API — What You Can Do
    // ========================================================
    /*
     * The Instrumentation interface provides:
     *
     * CLASS TRANSFORMATION:
     *   addTransformer(ClassFileTransformer, boolean canRetransform)
     *   removeTransformer(ClassFileTransformer)
     *   retransformClasses(Class<?>...)     — re-apply transformers
     *   redefineClasses(ClassDefinition...) — replace class bytes entirely
     *
     * INSPECTION:
     *   getAllLoadedClasses()          — every class in JVM
     *   getInitiatedClasses(loader)   — classes initiated by a ClassLoader
     *   getObjectSize(obj)            — shallow object size in bytes!
     *   isModifiableClass(clazz)      — can this class be retransformed?
     *   isRetransformClassesSupported()
     *   isRedefineClassesSupported()
     *
     * CAPABILITIES:
     *   isNativeMethodPrefixSupported()
     *   setNativeMethodPrefix(transformer, prefix)
     *
     * RESTRICTIONS on retransformation:
     *   You CAN change: method bodies
     *   You CANNOT change: class hierarchy, fields, method signatures
     *
     * getObjectSize() is especially useful — no other standard API
     * gives you the shallow size of an object!
     */

    // ========================================================
    // 5. MANIFEST.MF — Agent Packaging
    // ========================================================
    /*
     * To make a JAR a Java agent, add these to META-INF/MANIFEST.MF:
     *
     * FOR PREMAIN (static agent):
     *   Premain-Class: com.example.MyAgent
     *   Can-Retransform-Classes: true
     *   Can-Redefine-Classes: true
     *
     * FOR AGENTMAIN (dynamic agent):
     *   Agent-Class: com.example.MyAgent
     *   Can-Retransform-Classes: true
     *   Can-Redefine-Classes: true
     *
     * BOTH in one JAR:
     *   Premain-Class: com.example.MyAgent
     *   Agent-Class: com.example.MyAgent
     *   Can-Retransform-Classes: true
     *   Can-Redefine-Classes: true
     *
     * OPTIONAL:
     *   Boot-Class-Path: some-dependency.jar
     *     → Added to bootstrap classloader path
     *
     * BUILD COMMAND:
     *   jar cfm agent.jar MANIFEST.MF -C classes/ .
     *
     * RUN:
     *   java -javaagent:agent.jar=myargs com.example.MyApp
     *   java -javaagent:agent.jar com.example.MyApp  (no args)
     *
     * MULTIPLE AGENTS:
     *   java -javaagent:agent1.jar -javaagent:agent2.jar MyApp
     *   → premain() called for each, in order
     */

    // ========================================================
    // 6. BUILDING A COMPLETE AGENT — Step by Step
    // ========================================================
    /*
     * Here's a complete example: a method-timing agent.
     *
     * STEP 1: Create the agent class
     * ────────────────────────────────
     *   // TimingAgent.java
     *   import java.lang.instrument.*;
     *
     *   public class TimingAgent {
     *       public static void premain(String args, Instrumentation inst) {
     *           inst.addTransformer(new TimingTransformer());
     *       }
     *   }
     *
     * STEP 2: Create the transformer (using ASM or ByteBuddy)
     * ────────────────────────────────────────────────────────
     *   // With ByteBuddy (high-level API):
     *   import net.bytebuddy.agent.builder.AgentBuilder;
     *   import net.bytebuddy.asm.Advice;
     *
     *   public class TimingAgent {
     *       public static void premain(String args, Instrumentation inst) {
     *           new AgentBuilder.Default()
     *               .type(nameStartsWith("com.myapp"))
     *               .transform((builder, type, cl, module) ->
     *                   builder.visit(Advice.to(TimingAdvice.class)
     *                       .on(isMethod())))
     *               .installOn(inst);
     *       }
     *   }
     *
     *   class TimingAdvice {
     *       @Advice.OnMethodEnter
     *       static long enter() { return System.nanoTime(); }
     *
     *       @Advice.OnMethodExit
     *       static void exit(@Advice.Enter long start,
     *                        @Advice.Origin String method) {
     *           long elapsed = System.nanoTime() - start;
     *           System.out.println(method + " took " + elapsed + "ns");
     *       }
     *   }
     *
     * STEP 3: Create MANIFEST.MF
     * ──────────────────────────
     *   Manifest-Version: 1.0
     *   Premain-Class: TimingAgent
     *   Can-Retransform-Classes: true
     *
     * STEP 4: Package
     * ───────────────
     *   javac TimingAgent.java
     *   jar cfm timing-agent.jar MANIFEST.MF TimingAgent.class
     *
     * STEP 5: Run
     * ──────────
     *   java -javaagent:timing-agent.jar com.myapp.Main
     */

    // ========================================================
    // 7. REAL-WORLD AGENT EXAMPLES
    // ========================================================
    /*
     * JACOCO (Code Coverage):
     *   → Transforms class bytes to insert probes at branch points
     *   → On shutdown, writes coverage data
     *   → java -javaagent:jacocoagent.jar=output=file MyApp
     *
     * SPRING DEVTOOLS (Hot Reload):
     *   → Uses agent + custom ClassLoader
     *   → Watches file changes → reloads affected classes
     *   → agentmain() + retransformClasses()
     *
     * HIBERNATE (Entity Enhancement):
     *   → Transforms @Entity classes at load time
     *   → Adds lazy-loading proxies, dirty checking
     *   → Makes bytecode-level changes to POJOs
     *
     * JREBEL (Zero-Redeploy):
     *   → Most sophisticated agent usage
     *   → Intercepts ClassLoader, reloads classes from disk
     *   → Maintains object state across "reloads"
     *
     * MOCKITO (Test Mocking):
     *   → Uses ByteBuddy internally
     *   → Creates subclasses/proxies at runtime
     *   → mockito-agent for inline mocking of final classes
     *
     * OPENTELEMETRY (Distributed Tracing):
     *   → Auto-instruments HTTP clients, DB drivers, etc.
     *   → java -javaagent:opentelemetry-javaagent.jar MyApp
     *   → Zero code changes for tracing!
     */

    // ========================================================
    // MAIN — Demonstrates concepts without requiring agent mode
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 57: JAVA AGENTS & INSTRUMENTATION ===\n");

        // --- 1. Show Instrumentation isn't available in normal mode ---
        System.out.println("--- 1. Agent Mode vs Normal Mode ---\n");
        System.out.println("  This class has premain() and agentmain() methods.");
        System.out.println("  When run normally (java Chapter57_JavaAgents), they're NOT called.");
        System.out.println("  To activate: java -javaagent:thisjar.jar OtherApp");

        // --- 2. Demonstrate ClassLoader introspection ---
        System.out.println("\n--- 2. ClassLoader Introspection ---\n");

        ClassLoader cl = Chapter57_JavaAgents.class.getClassLoader();
        System.out.println("  Our ClassLoader: " + cl);

        // Count loaded classes (limited without Instrumentation)
        System.out.println("  To count ALL loaded classes, you need Instrumentation.getAllLoadedClasses()");
        System.out.println("  This is only available inside an agent.");

        // --- 3. Object sizing ---
        System.out.println("\n--- 3. Object Sizing ---\n");
        System.out.println("  Instrumentation.getObjectSize(obj) gives SHALLOW size:");
        System.out.println("    Object    → ~16 bytes (header only)");
        System.out.println("    Integer   → ~16 bytes (header + int field)");
        System.out.println("    String    → ~40 bytes (header + char[]/byte[] ref + hash + coder)");
        System.out.println("    int[100]  → ~416 bytes (header + 100 * 4 bytes)");
        System.out.println("  Note: Actual sizes depend on JVM, arch (32/64-bit), and");
        System.out.println("  compressed oops (-XX:+UseCompressedOops, default on).");

        // --- 4. Show what transformation looks like ---
        System.out.println("\n--- 4. Class Byte Inspection ---\n");

        // Every .class file starts with magic number 0xCAFEBABE
        String classFile = Chapter57_JavaAgents.class.getName().replace('.', '/') + ".class";
        try (var is = Chapter57_JavaAgents.class.getClassLoader().getResourceAsStream(classFile)) {
            if (is != null) {
                byte[] header = new byte[8];
                is.read(header);
                System.out.printf("  Class file magic: 0x%02X%02X%02X%02X%n",
                    header[0], header[1], header[2], header[3]);
                int minor = ((header[4] & 0xFF) << 8) | (header[5] & 0xFF);
                int major = ((header[6] & 0xFF) << 8) | (header[7] & 0xFF);
                System.out.println("  Class file version: " + major + "." + minor);
                System.out.println("  (Java 11 = 55.0, Java 17 = 61.0, Java 21 = 65.0)");
            }
        } catch (Exception e) {
            System.out.println("  Could not read class file: " + e.getMessage());
        }

        // --- 5. Build instructions ---
        System.out.println("\n--- 5. Build & Run an Agent ---\n");
        System.out.println("  1. Create agent class with premain(String, Instrumentation)");
        System.out.println("  2. Create MANIFEST.MF with Premain-Class: entry");
        System.out.println("  3. Package: jar cfm agent.jar MANIFEST.MF *.class");
        System.out.println("  4. Run: java -javaagent:agent.jar YourApp");
        System.out.println("  5. Multiple: java -javaagent:a.jar -javaagent:b.jar YourApp");

        // --- 6. Bytecode libraries ---
        System.out.println("\n--- 6. Bytecode Manipulation Libraries ---\n");
        System.out.println("  ┌───────────────┬────────────────────────────────────────┐");
        System.out.println("  │ Library        │ Level / Notes                          │");
        System.out.println("  ├───────────────┼────────────────────────────────────────┤");
        System.out.println("  │ ASM            │ Low-level, visitor pattern, FAST       │");
        System.out.println("  │ ByteBuddy      │ High-level, fluent API, recommended   │");
        System.out.println("  │ Javassist      │ Source-level API, write Java strings   │");
        System.out.println("  │ cglib          │ Subclass proxying (used by Spring)     │");
        System.out.println("  │ Byte Buddy     │ Also great for creating proxies        │");
        System.out.println("  └───────────────┴────────────────────────────────────────┘");

        // --- 7. Attach API ---
        System.out.println("\n--- 7. Attach API (Dynamic Agent Loading) ---\n");
        System.out.println("  // Requires tools.jar or jdk.attach module");
        System.out.println("  import com.sun.tools.attach.VirtualMachine;");
        System.out.println("  ");
        System.out.println("  // List running JVMs");
        System.out.println("  VirtualMachine.list().forEach(vmd -> ");
        System.out.println("      System.out.println(vmd.id() + \": \" + vmd.displayName()));");
        System.out.println("  ");
        System.out.println("  // Attach to a running JVM");
        System.out.println("  VirtualMachine vm = VirtualMachine.attach(\"12345\");  // PID");
        System.out.println("  vm.loadAgent(\"/path/to/agent.jar\", \"args\");");
        System.out.println("  vm.detach();");
        System.out.println("  ");
        System.out.println("  // The target JVM's agentmain() is called!");

        System.out.println("\n✓ Java Agents & Instrumentation Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Build a simple "class loading logger" agent that prints every
 *    class name as it's loaded (use premain + ClassFileTransformer).
 * 2. Create an agent JAR with proper MANIFEST.MF, package it, and
 *    run it against a sample application.
 * 3. Use ByteBuddy to build an agent that measures method execution
 *    time for all methods in a specific package.
 * 4. Write a dynamic agent that can be attached to a running JVM
 *    and prints thread dumps.
 *
 * NEXT: Chapter 58 — Unsafe & Off-Heap Memory
 */
