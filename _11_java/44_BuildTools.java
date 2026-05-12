/*
 * ============================================================
 *  CHAPTER 44: BUILD TOOLS & JAVA MODULES
 * ============================================================
 *  Professional Java projects use build tools to:
 *    → Manage dependencies (download libraries automatically)
 *    → Compile code
 *    → Run tests
 *    → Package into JARs
 *    → Deploy
 *
 *  TWO MAIN BUILD TOOLS:
 *    Maven  — XML-based (pom.xml), convention over configuration
 *    Gradle — Groovy/Kotlin DSL, faster, more flexible
 *
 *  JAVA MODULES (Java 9+):
 *    Strong encapsulation at the package level.
 *    module-info.java defines what's exported.
 * ============================================================
 */

public class Chapter44_BuildTools {

    public static void main(String[] args) {

        // =========================================
        // 1. MAVEN
        // =========================================
        System.out.println("=== MAVEN ===\n");

        System.out.println("  Directory Structure:");
        System.out.println("  my-project/");
        System.out.println("  ├── pom.xml              ← build config");
        System.out.println("  ├── src/");
        System.out.println("  │   ├── main/");
        System.out.println("  │   │   ├── java/        ← source code");
        System.out.println("  │   │   └── resources/   ← config files");
        System.out.println("  │   └── test/");
        System.out.println("  │       ├── java/        ← test code");
        System.out.println("  │       └── resources/   ← test configs");
        System.out.println("  └── target/              ← build output");

        System.out.println("\n  pom.xml Example:");
        System.out.println("  ┌────────────────────────────────────────┐");
        System.out.println("  │ <project>                              │");
        System.out.println("  │   <groupId>com.myapp</groupId>        │");
        System.out.println("  │   <artifactId>my-project</artifactId> │");
        System.out.println("  │   <version>1.0.0</version>            │");
        System.out.println("  │   <dependencies>                      │");
        System.out.println("  │     <dependency>                      │");
        System.out.println("  │       <groupId>junit</groupId>        │");
        System.out.println("  │       <artifactId>junit</artifactId>  │");
        System.out.println("  │       <version>5.10.0</version>       │");
        System.out.println("  │       <scope>test</scope>             │");
        System.out.println("  │     </dependency>                     │");
        System.out.println("  │   </dependencies>                     │");
        System.out.println("  │ </project>                            │");
        System.out.println("  └────────────────────────────────────────┘");

        System.out.println("\n  Maven Commands:");
        System.out.println("  mvn compile        → compile source code");
        System.out.println("  mvn test           → run tests");
        System.out.println("  mvn package        → build JAR/WAR");
        System.out.println("  mvn install        → install to local repo");
        System.out.println("  mvn clean          → delete target/");
        System.out.println("  mvn clean package  → clean + build");
        System.out.println("  mvn dependency:tree → show dependency tree");

        System.out.println("\n  Maven Lifecycle Phases:");
        System.out.println("  validate → compile → test → package → verify → install → deploy");

        System.out.println("\n  Dependency Scopes:");
        System.out.println("  compile  → available everywhere (default)");
        System.out.println("  test     → only in test code");
        System.out.println("  provided → compile only, not packaged (e.g., servlet API)");
        System.out.println("  runtime  → not for compile, only runtime");

        // =========================================
        // 2. GRADLE
        // =========================================
        System.out.println("\n=== GRADLE ===\n");

        System.out.println("  Directory Structure: (same as Maven)");
        System.out.println("  my-project/");
        System.out.println("  ├── build.gradle         ← build config");
        System.out.println("  ├── settings.gradle      ← project name");
        System.out.println("  ├── gradlew / gradlew.bat ← wrapper scripts");
        System.out.println("  └── src/main/java/       ← same layout");

        System.out.println("\n  build.gradle Example:");
        System.out.println("  ┌────────────────────────────────────────┐");
        System.out.println("  │ plugins {                              │");
        System.out.println("  │     id 'java'                         │");
        System.out.println("  │ }                                     │");
        System.out.println("  │                                       │");
        System.out.println("  │ repositories {                        │");
        System.out.println("  │     mavenCentral()                    │");
        System.out.println("  │ }                                     │");
        System.out.println("  │                                       │");
        System.out.println("  │ dependencies {                        │");
        System.out.println("  │   testImplementation 'junit:5.10.0'  │");
        System.out.println("  │   implementation 'com.google:gson'   │");
        System.out.println("  │ }                                     │");
        System.out.println("  └────────────────────────────────────────┘");

        System.out.println("\n  Gradle Commands:");
        System.out.println("  ./gradlew build       → compile + test + package");
        System.out.println("  ./gradlew test        → run tests");
        System.out.println("  ./gradlew clean       → delete build/");
        System.out.println("  ./gradlew run         → run main class");
        System.out.println("  ./gradlew dependencies → show dependency tree");

        System.out.println("\n  Maven vs Gradle:");
        System.out.println("  Feature      Maven          Gradle");
        System.out.println("  Config       XML (verbose)  Groovy/Kotlin (concise)");
        System.out.println("  Speed        Slower         Faster (incremental)");
        System.out.println("  Flexibility  Convention     Very flexible");
        System.out.println("  Android      No             Yes (official)");
        System.out.println("  Learning     Easier         Steeper curve");
        System.out.println("  Adoption     Most projects  Growing fast");

        // =========================================
        // 3. JAR FILES
        // =========================================
        System.out.println("\n=== JAR FILES ===\n");

        System.out.println("  JAR = Java ARchive (ZIP with .class files)");
        System.out.println();
        System.out.println("  Create JAR:");
        System.out.println("  jar cf myapp.jar -C out/ .");
        System.out.println();
        System.out.println("  Create executable JAR:");
        System.out.println("  jar cfe myapp.jar com.myapp.Main -C out/ .");
        System.out.println();
        System.out.println("  Run JAR:");
        System.out.println("  java -jar myapp.jar");
        System.out.println();
        System.out.println("  MANIFEST.MF (inside JAR):");
        System.out.println("  Main-Class: com.myapp.Main");
        System.out.println("  Class-Path: lib/gson.jar lib/commons.jar");

        // =========================================
        // 4. JAVA MODULES (Java 9+)
        // =========================================
        System.out.println("\n=== JAVA MODULES (Java 9+) ===\n");

        System.out.println("  Module = group of packages with explicit dependencies");
        System.out.println();
        System.out.println("  module-info.java:");
        System.out.println("  ┌──────────────────────────────────────┐");
        System.out.println("  │ module com.myapp {                   │");
        System.out.println("  │   requires java.sql;      // needs  │");
        System.out.println("  │   requires java.logging;             │");
        System.out.println("  │   exports com.myapp.api;  // exposes│");
        System.out.println("  │   opens com.myapp.model to gson;    │");
        System.out.println("  │ }                                    │");
        System.out.println("  └──────────────────────────────────────┘");

        System.out.println("\n  Module Keywords:");
        System.out.println("  requires   → declare dependency on another module");
        System.out.println("  exports    → make package visible to other modules");
        System.out.println("  opens      → allow reflection access (for frameworks)");
        System.out.println("  provides   → provide service implementation");
        System.out.println("  uses       → consume a service");

        System.out.println("\n  Benefits of Modules:");
        System.out.println("  1. Strong encapsulation (internal packages hidden)");
        System.out.println("  2. Reliable configuration (missing deps caught early)");
        System.out.println("  3. Smaller runtime (jlink custom JRE)");
        System.out.println("  4. Better security (no access to internals)");

        // =========================================
        // 5. CREATING YOUR FIRST MAVEN PROJECT
        // =========================================
        System.out.println("\n=== QUICK START: MAVEN ===\n");
        System.out.println("  # Generate project from archetype:");
        System.out.println("  mvn archetype:generate \\");
        System.out.println("    -DgroupId=com.myapp \\");
        System.out.println("    -DartifactId=my-project \\");
        System.out.println("    -DarchetypeArtifactId=maven-archetype-quickstart");
        System.out.println();
        System.out.println("  # Or use Spring Initializr for web apps:");
        System.out.println("  https://start.spring.io");

        System.out.println("\n=== QUICK START: GRADLE ===\n");
        System.out.println("  # Generate project:");
        System.out.println("  gradle init --type java-application");
        System.out.println();
        System.out.println("  # Or with wrapper:");
        System.out.println("  gradle wrapper");
        System.out.println("  ./gradlew build");

        // =========================================
        // 6. USEFUL DEPENDENCIES
        // =========================================
        System.out.println("\n=== COMMON DEPENDENCIES ===\n");
        System.out.println("  Testing:   JUnit 5, Mockito, AssertJ");
        System.out.println("  Logging:   SLF4J + Logback, Log4j2");
        System.out.println("  JSON:      Jackson, Gson");
        System.out.println("  HTTP:      OkHttp, Apache HttpClient");
        System.out.println("  Database:  HikariCP, JDBC drivers");
        System.out.println("  Utils:     Guava, Apache Commons");
        System.out.println("  Web:       Spring Boot, Micronaut, Quarkus");
        System.out.println("  ORM:       Hibernate, MyBatis, jOOQ");

        System.out.println("\n✓ Build Tools & Modules Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Create a Maven project, add JUnit 5 + Gson dependencies, write tests.
 * 2. Create a Gradle project with the same setup. Compare build times.
 * 3. Create an executable JAR from one of your earlier chapters.
 * 4. Create a module-info.java for a multi-package project.
 *
 * NEXT: Chapter 45 — FINAL BOSS: Real-World Project
 */
