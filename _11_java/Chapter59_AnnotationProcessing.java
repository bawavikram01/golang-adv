/*
 * ============================================================
 *  CHAPTER 59: ANNOTATION PROCESSING (Compile-Time)
 * ============================================================
 *  Annotations at RUNTIME you already know (Chapter 35).
 *  This chapter is about COMPILE-TIME annotation processing —
 *  generating code BEFORE the program runs.
 *
 *  This is how Lombok, Dagger, MapStruct, AutoValue, and
 *  Immutables work. They read your annotations and GENERATE
 *  Java source files during compilation.
 *
 *  TOPICS:
 *    1. Runtime vs Compile-Time Processing
 *    2. javax.annotation.processing API
 *    3. AbstractProcessor
 *    4. Processing Rounds
 *    5. Filer — Generating Source Files
 *    6. Messager — Compiler Warnings/Errors
 *    7. Element API — Inspecting Code Structure
 *    8. Building a Complete Processor
 *    9. Service Registration (META-INF/services)
 *   10. Real-World Examples
 * ============================================================
 */

import java.lang.annotation.*;
import java.util.*;

public class Chapter59_AnnotationProcessing {

    // ========================================================
    // 1. RUNTIME vs COMPILE-TIME
    // ========================================================
    /*
     * RUNTIME (Chapter 35):
     *   @Retention(RetentionPolicy.RUNTIME)
     *   → Annotation exists in .class file and at runtime
     *   → Read via reflection: field.getAnnotation(MyAnno.class)
     *   → Cost: reflection overhead at runtime
     *   → Examples: @Override, @Transactional, @Autowired
     *
     * COMPILE-TIME (This chapter):
     *   @Retention(RetentionPolicy.SOURCE) or CLASS
     *   → Annotation processor runs during javac compilation
     *   → Generates NEW .java source files
     *   → ZERO runtime cost (code is generated at compile time)
     *   → Examples: @Data (Lombok), @Component (Dagger), @Mapper (MapStruct)
     *
     * WHEN TO USE WHICH:
     *   Compile-time: when you can determine everything at build time
     *     → Boilerplate generation (getters, builders, etc.)
     *     → Dependency injection wiring
     *     → Mapping code generation
     *
     *   Runtime: when behavior depends on runtime state
     *     → Aspect-oriented programming (@Transactional)
     *     → Framework configuration (@RequestMapping)
     *     → Conditional behavior
     */

    // ========================================================
    // 2. THE ANNOTATION PROCESSING PIPELINE
    // ========================================================
    /*
     * javac compilation with annotation processing:
     *
     *   ┌────────────────────────────────────────────────────────┐
     *   │ Round 1:                                               │
     *   │   1. javac parses all .java source files               │
     *   │   2. Finds annotations on elements                     │
     *   │   3. Calls matching processors                        │
     *   │   4. Processors may generate NEW .java files          │
     *   │                                                        │
     *   │ Round 2:                                               │
     *   │   1. javac parses NEWLY GENERATED .java files          │
     *   │   2. Finds annotations in new files                    │
     *   │   3. Calls processors again                           │
     *   │   4. Repeat until no new files generated               │
     *   │                                                        │
     *   │ Final Round:                                           │
     *   │   1. No new files generated                            │
     *   │   2. processingOver() returns true                     │
     *   │   3. Processors can do final validation                │
     *   │                                                        │
     *   │ Compilation:                                           │
     *   │   All source files (original + generated) compiled     │
     *   └────────────────────────────────────────────────────────┘
     */

    // ========================================================
    // 3. ANNOTATIONS FOR OUR PROCESSOR
    // ========================================================

    // Example: annotation that generates a Builder class
    @Retention(RetentionPolicy.SOURCE)
    @Target(ElementType.TYPE)
    @interface GenerateBuilder {
        // Processor will generate MyClassBuilder for each @GenerateBuilder class
    }

    // Example: annotation for toString generation
    @Retention(RetentionPolicy.SOURCE)
    @Target(ElementType.TYPE)
    @interface GenerateToString {
    }

    // ========================================================
    // 4. ABSTRACTPROCESSOR — The Base Class
    // ========================================================
    /*
     * Every annotation processor extends AbstractProcessor:
     *
     *   import javax.annotation.processing.*;
     *   import javax.lang.model.element.*;
     *   import javax.lang.model.SourceVersion;
     *   import javax.tools.Diagnostic;
     *   import java.io.Writer;
     *   import java.util.Set;
     *
     *   @SupportedAnnotationTypes("com.example.GenerateBuilder")
     *   @SupportedSourceVersion(SourceVersion.RELEASE_11)
     *   public class BuilderProcessor extends AbstractProcessor {
     *
     *       // Called once when processor is initialized
     *       @Override
     *       public synchronized void init(ProcessingEnvironment env) {
     *           super.init(env);
     *           // env gives you:
     *           //   env.getFiler()     — create source/class/resource files
     *           //   env.getMessager()  — emit warnings/errors
     *           //   env.getElementUtils()  — utility methods for elements
     *           //   env.getTypeUtils()     — utility methods for types
     *       }
     *
     *       // Called each round with annotated elements
     *       @Override
     *       public boolean process(Set<? extends TypeElement> annotations,
     *                              RoundEnvironment roundEnv) {
     *
     *           for (TypeElement annotation : annotations) {
     *               Set<? extends Element> elements =
     *                   roundEnv.getElementsAnnotatedWith(annotation);
     *
     *               for (Element element : elements) {
     *                   if (element.getKind() != ElementKind.CLASS) {
     *                       processingEnv.getMessager().printMessage(
     *                           Diagnostic.Kind.ERROR,
     *                           "@GenerateBuilder only on classes",
     *                           element);
     *                       continue;
     *                   }
     *
     *                   TypeElement classElement = (TypeElement) element;
     *                   generateBuilder(classElement);
     *               }
     *           }
     *
     *           return true;  // true = we claimed these annotations
     *       }
     *
     *       private void generateBuilder(TypeElement classElement) {
     *           String className = classElement.getSimpleName().toString();
     *           String builderName = className + "Builder";
     *           String packageName = processingEnv.getElementUtils()
     *               .getPackageOf(classElement).getQualifiedName().toString();
     *
     *           // Get all fields
     *           List<VariableElement> fields = new ArrayList<>();
     *           for (Element enclosed : classElement.getEnclosedElements()) {
     *               if (enclosed.getKind() == ElementKind.FIELD) {
     *                   fields.add((VariableElement) enclosed);
     *               }
     *           }
     *
     *           // Generate source code
     *           try {
     *               JavaFileObject file = processingEnv.getFiler()
     *                   .createSourceFile(packageName + "." + builderName);
     *
     *               try (Writer writer = file.openWriter()) {
     *                   writer.write("package " + packageName + ";\n\n");
     *                   writer.write("public class " + builderName + " {\n");
     *
     *                   // Fields
     *                   for (VariableElement field : fields) {
     *                       writer.write("    private " + field.asType()
     *                           + " " + field.getSimpleName() + ";\n");
     *                   }
     *
     *                   // Setter methods (return this for chaining)
     *                   for (VariableElement field : fields) {
     *                       String fieldName = field.getSimpleName().toString();
     *                       writer.write("\n    public " + builderName
     *                           + " " + fieldName + "("
     *                           + field.asType() + " " + fieldName + ") {\n");
     *                       writer.write("        this." + fieldName
     *                           + " = " + fieldName + ";\n");
     *                       writer.write("        return this;\n");
     *                       writer.write("    }\n");
     *                   }
     *
     *                   // build() method
     *                   writer.write("\n    public " + className + " build() {\n");
     *                   writer.write("        return new " + className + "(");
     *                   StringJoiner sj = new StringJoiner(", ");
     *                   for (VariableElement field : fields) {
     *                       sj.add(field.getSimpleName().toString());
     *                   }
     *                   writer.write(sj.toString());
     *                   writer.write(");\n");
     *                   writer.write("    }\n");
     *                   writer.write("}\n");
     *               }
     *
     *           } catch (Exception e) {
     *               processingEnv.getMessager().printMessage(
     *                   Diagnostic.Kind.ERROR,
     *                   "Failed to generate: " + e.getMessage());
     *           }
     *       }
     *   }
     */

    // ========================================================
    // 5. ELEMENT API — Inspecting Code Structure
    // ========================================================
    /*
     * Elements represent code constructs (NO runtime objects needed):
     *
     *   Element (base)
     *   ├── TypeElement        → class, interface, enum
     *   ├── VariableElement    → field, parameter, local var, enum constant
     *   ├── ExecutableElement   → method, constructor
     *   ├── PackageElement      → package
     *   └── TypeParameterElement → generic type parameter (<T>)
     *
     * USEFUL METHODS:
     *   element.getSimpleName()       → "MyClass"
     *   element.getKind()             → ElementKind.CLASS, FIELD, METHOD...
     *   element.getModifiers()        → Set<Modifier> (PUBLIC, FINAL, etc.)
     *   element.getAnnotation(X.class) → get annotation
     *   element.getEnclosedElements() → children (fields, methods of a class)
     *   element.getEnclosingElement() → parent (class of a field)
     *   element.asType()              → TypeMirror (the type)
     *
     * TypeMirror represents types at compile time:
     *   DeclaredType    → class/interface type
     *   PrimitiveType   → int, long, etc.
     *   ArrayType        → int[], String[]
     *   TypeVariable     → T (generic)
     *   WildcardType     → ? extends Number
     *   NoType           → void
     */

    // ========================================================
    // 6. REGISTERING A PROCESSOR
    // ========================================================
    /*
     * METHOD 1: META-INF/services (SPI — most common)
     *   Create file: META-INF/services/javax.annotation.processing.Processor
     *   Content: com.example.BuilderProcessor
     *
     *   javac automatically discovers and loads the processor.
     *
     * METHOD 2: Compiler flag
     *   javac -processor com.example.BuilderProcessor MyClass.java
     *
     * METHOD 3: Module system (Java 9+)
     *   module com.example.processor {
     *       requires java.compiler;
     *       provides javax.annotation.processing.Processor
     *           with com.example.BuilderProcessor;
     *   }
     *
     * PACKAGING:
     *   Processor goes in a SEPARATE JAR from your annotations.
     *   Annotations JAR → compile dependency
     *   Processor JAR → annotation processor dependency
     *
     *   Maven:
     *   <dependency>
     *       <groupId>com.example</groupId>
     *       <artifactId>my-processor</artifactId>
     *       <scope>provided</scope> <!-- not needed at runtime -->
     *   </dependency>
     *
     *   Or in annotationProcessorPaths for Maven compiler plugin.
     */

    // ========================================================
    // 7. COMPLETE PROJECT STRUCTURE
    // ========================================================
    /*
     * my-annotations/              (annotation definitions)
     *   src/main/java/
     *     com/example/
     *       GenerateBuilder.java    (@interface)
     *
     * my-processor/                 (processor)
     *   src/main/java/
     *     com/example/
     *       BuilderProcessor.java   (extends AbstractProcessor)
     *   src/main/resources/
     *     META-INF/services/
     *       javax.annotation.processing.Processor
     *         → com.example.BuilderProcessor
     *
     * my-app/                       (uses the annotation)
     *   src/main/java/
     *     com/example/
     *       @GenerateBuilder
     *       public class Person {
     *           String name;
     *           int age;
     *       }
     *
     *   After compilation, GENERATED:
     *   target/generated-sources/annotations/
     *     com/example/
     *       PersonBuilder.java      ← generated by processor!
     *
     *   Usage:
     *     Person p = new PersonBuilder()
     *         .name("Alice")
     *         .age(30)
     *         .build();
     */

    // ========================================================
    // 8. REAL-WORLD PROCESSORS
    // ========================================================
    /*
     * LOMBOK (@Data, @Builder, @Getter, @Setter):
     *   Actually a COMPILER PLUGIN (not standard annotation processing)
     *   Uses internal javac API to modify the AST directly
     *   Much more powerful but non-standard
     *   → Adds methods/fields to existing classes (processors can't)
     *
     * DAGGER (@Inject, @Component):
     *   Compile-time dependency injection
     *   Generates factory classes, component implementations
     *   ZERO runtime reflection cost
     *
     * MAPSTRUCT (@Mapper):
     *   Generates type-safe mapping code between DTOs
     *     @Mapper
     *     interface PersonMapper {
     *         PersonDTO toDTO(Person person);
     *     }
     *   → Generates PersonMapperImpl with field-by-field mapping
     *
     * AUTOVALUE (@AutoValue):
     *   Google's version of records (before Java 16)
     *   Generates equals, hashCode, toString, builder
     *
     * IMMUTABLES (@Value.Immutable):
     *   Generates immutable implementations with builders
     *   More powerful than AutoValue
     *
     * QUERYDSL (Q-classes):
     *   Generates type-safe query classes from JPA entities
     *   @Entity class Person → QPersonis generated
     *
     * JMH (@Benchmark):
     *   Generates benchmark harness code from annotated methods
     */

    // ========================================================
    // 9. KEY RULES & GOTCHAS
    // ========================================================
    /*
     * 1. Processors CAN generate new files, CANNOT modify existing files
     *    (Lombok is special — it's a compiler plugin, not a standard processor)
     *
     * 2. Don't create the same file twice — IOException
     *    Check if file already exists or track generated files
     *
     * 3. Processors run in ROUNDS — don't assume all types available in Round 1
     *    A type generated in Round 1 is available in Round 2
     *
     * 4. Return true from process() to CLAIM the annotation
     *    (other processors won't see it)
     *    Return false to let other processors also handle it
     *
     * 5. Use Messager for errors, NOT System.err or exceptions
     *    processingEnv.getMessager().printMessage(Diagnostic.Kind.ERROR, ...)
     *    Kind.ERROR stops compilation
     *    Kind.WARNING continues compilation
     *
     * 6. Test with compile-testing library:
     *    com.google.testing.compile:compile-testing
     *    Compilation result = Compiler.javac()
     *        .withProcessors(new MyProcessor())
     *        .compile(JavaFileObjects.forSourceString(...));
     *    assertThat(result).succeeded();
     *
     * 7. Generated code should be clean and readable
     *    Users will see it and debug through it
     *    Consider using JavaPoet library for code generation
     */

    // ========================================================
    // MAIN — Demonstrates concepts
    // ========================================================

    public static void main(String[] args) {

        System.out.println("=== CHAPTER 59: ANNOTATION PROCESSING ===\n");

        // --- Comparison ---
        System.out.println("--- Runtime vs Compile-Time ---\n");
        System.out.println("  ┌────────────────────┬───────────────────┬───────────────────┐");
        System.out.println("  │ Aspect              │ Runtime           │ Compile-Time      │");
        System.out.println("  ├────────────────────┼───────────────────┼───────────────────┤");
        System.out.println("  │ When processed      │ App running       │ javac compiling   │");
        System.out.println("  │ How                  │ Reflection        │ Processor API     │");
        System.out.println("  │ Performance cost     │ Reflection at RT  │ Zero at runtime   │");
        System.out.println("  │ Can modify classes   │ No (proxy only)   │ No (generate new) │");
        System.out.println("  │ Error detection      │ Runtime exception │ Compile error     │");
        System.out.println("  │ @Retention           │ RUNTIME           │ SOURCE or CLASS   │");
        System.out.println("  └────────────────────┴───────────────────┴───────────────────┘");

        // --- Simulated processing ---
        System.out.println("\n--- Simulated Annotation Processing ---\n");

        // Simulate what a processor does by reading class metadata
        Class<?> sampleClass = SampleEntity.class;
        System.out.println("  Inspecting: " + sampleClass.getSimpleName());

        // Get fields (at compile time, processor uses Element API)
        java.lang.reflect.Field[] fields = sampleClass.getDeclaredFields();
        System.out.println("  Fields found: " + fields.length);

        // Generate builder code (what a processor would output)
        System.out.println("\n  GENERATED CODE (PersonBuilder.java):");
        System.out.println("  ─────────────────────────────────────");
        String className = sampleClass.getSimpleName();
        String builderName = className + "Builder";
        System.out.println("  public class " + builderName + " {");
        for (java.lang.reflect.Field f : fields) {
            System.out.println("      private " + f.getType().getSimpleName()
                + " " + f.getName() + ";");
        }
        System.out.println();
        for (java.lang.reflect.Field f : fields) {
            System.out.println("      public " + builderName + " " + f.getName()
                + "(" + f.getType().getSimpleName() + " " + f.getName() + ") {");
            System.out.println("          this." + f.getName() + " = " + f.getName() + ";");
            System.out.println("          return this;");
            System.out.println("      }");
        }
        System.out.println("      public " + className + " build() { ... }");
        System.out.println("  }");

        // --- JavaPoet ---
        System.out.println("\n--- Code Generation Libraries ---\n");
        System.out.println("  JavaPoet (by Square) — fluent API for generating .java files:");
        System.out.println("    TypeSpec builder = TypeSpec.classBuilder(\"PersonBuilder\")");
        System.out.println("        .addModifiers(Modifier.PUBLIC)");
        System.out.println("        .addField(String.class, \"name\", Modifier.PRIVATE)");
        System.out.println("        .addMethod(MethodSpec.methodBuilder(\"name\")");
        System.out.println("            .addParameter(String.class, \"name\")");
        System.out.println("            .addStatement(\"this.name = name\")");
        System.out.println("            .addStatement(\"return this\")");
        System.out.println("            .returns(ClassName.get(\"\", \"PersonBuilder\"))");
        System.out.println("            .build())");
        System.out.println("        .build();");
        System.out.println("    JavaFile.builder(\"com.example\", builder).build().writeTo(filer);");

        // --- Build steps ---
        System.out.println("\n--- Build Your First Processor ---\n");
        System.out.println("  1. Create annotation: @GenerateBuilder (RetentionPolicy.SOURCE)");
        System.out.println("  2. Create processor extending AbstractProcessor");
        System.out.println("  3. Override process() — inspect elements, generate code");
        System.out.println("  4. Register in META-INF/services/javax.annotation.processing.Processor");
        System.out.println("  5. Package as JAR");
        System.out.println("  6. Compile target code: javac -processor ... or auto-discovery");

        System.out.println("\n✓ Annotation Processing Complete!");
    }

    // Sample class for the demo
    static class SampleEntity {
        String name;
        int age;
        String email;
    }
}

/*
 * EXERCISES:
 * 1. Build a @GenerateToString processor that generates a toString()
 *    helper method listing all fields and their values.
 * 2. Build a @GenerateEquals processor that generates equals/hashCode.
 * 3. Create a @Validate processor that emits compile errors if a class
 *    has mutable public fields (enforce encapsulation at compile time).
 * 4. Use JavaPoet to generate a type-safe configuration class from
 *    annotations on an interface.
 *
 * NEXT: Chapter 60 — JVM Diagnostics & Troubleshooting
 */
