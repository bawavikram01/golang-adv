# Phase 1.2 — Maven Basics

---

## What is Maven & Why Do You Need It?

**Maven** is a **build tool** and **dependency manager** for Java.

Before Maven, if you wanted to use a library (like Spring, Jackson, Hibernate), you had to:
1. Find the `.jar` file on some website
2. Download it manually
3. Put it in your project folder
4. Also download all the libraries *that library* depends on (transitive dependencies)
5. Manage version conflicts yourself

Maven solves ALL of this with one XML file: **`pom.xml`** (Project Object Model).

### Analogy
Maven is like **npm for Java** (if you know Node.js) or like a **recipe card** for cooking.

The `pom.xml` says: *"I need flour (Spring), eggs (Jackson), butter (Hibernate) — version 3.x."*
Maven goes to the central pantry (Maven Central Repository), fetches everything, and puts it on your shelf (`~/.m2/repository`).

You never download a JAR by hand again.

---

## The 3 Things Maven Does

| # | What | How |
|---|------|-----|
| 1 | **Dependency Management** | You list what you need in `pom.xml`, Maven downloads it |
| 2 | **Build Lifecycle** | Compile → Test → Package → Install → Deploy — one command |
| 3 | **Project Structure** | Enforces a standard directory layout everyone recognizes |

---

## Maven Standard Project Structure

```
my-project/
├── pom.xml                          ← THE heart of Maven
├── src/
│   ├── main/
│   │   ├── java/com/learn/          ← Your application code
│   │   └── resources/               ← Config files (application.properties, etc.)
│   └── test/
│       ├── java/com/learn/          ← Your test code
│       └── resources/               ← Test-specific config files
└── target/                          ← Build output (Maven creates this)
    ├── classes/                     ← Compiled .class files
    └── my-project-1.0.0.jar        ← Final packaged JAR
```

**Every** Spring Boot project uses this exact layout. It's a universal Java convention.

---

## Anatomy of pom.xml

The `pom.xml` has 4 main sections:

### Section 1: Project Coordinates (Identity)
```xml
<groupId>com.learn</groupId>        <!-- Your organization (reversed domain) -->
<artifactId>maven-basics</artifactId> <!-- Your project name -->
<version>1.0.0</version>            <!-- Current version -->
<packaging>jar</packaging>          <!-- Output format: jar, war, pom -->
```

Together they form a **unique ID**: `com.learn:maven-basics:1.0.0`
This is how the entire Java world references your project.

**Packaging types:**
| Type | What | When |
|------|------|------|
| `jar` | Java Archive (default) | Libraries, Spring Boot apps |
| `war` | Web Archive | Deploy to external Tomcat |
| `pom` | Parent POM | Multi-module projects, no code |

### Section 2: Properties (Variables)
```xml
<properties>
    <maven.compiler.source>17</maven.compiler.source>
    <maven.compiler.target>17</maven.compiler.target>
    <gson.version>2.11.0</gson.version>
</properties>
```
Define values once, use everywhere with `${property.name}`.
Change a version in one place → applies everywhere.

### Section 3: Dependencies (Libraries You Need)
```xml
<dependencies>
    <dependency>
        <groupId>com.google.code.gson</groupId>  <!-- Who made it -->
        <artifactId>gson</artifactId>              <!-- Library name -->
        <version>${gson.version}</version>         <!-- Which version -->
        <scope>compile</scope>                     <!-- When it's available -->
    </dependency>
</dependencies>
```

#### Dependency Scopes — IMPORTANT
| Scope | Available at Compile? | Available at Runtime? | In Final JAR? | Use Case |
|-------|----------------------|----------------------|---------------|----------|
| `compile` (default) | ✅ | ✅ | ✅ | Most libraries (Spring, Jackson) |
| `test` | ❌ only in test/ | ❌ only in test/ | ❌ | JUnit, Mockito |
| `provided` | ✅ | ❌ | ❌ | Servlet API (server provides it) |
| `runtime` | ❌ | ✅ | ✅ | JDBC drivers |

### Section 4: Build (Plugins & Configuration)
```xml
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-compiler-plugin</artifactId>
            <configuration>
                <source>17</source>
                <target>17</target>
            </configuration>
        </plugin>
    </plugins>
</build>
```

Plugins extend Maven's capabilities: compile Java, run tests, package JARs, etc.

---

## Maven Build Lifecycle

When you run `mvn package`, Maven executes these phases **in order**:

```
mvn validate   → Check project is correct
mvn compile    → Compile src/main/java → target/classes
mvn test       → Run src/test/java tests
mvn package    → Create JAR/WAR in target/
mvn install    → Copy JAR to ~/.m2/repository (local)
mvn deploy     → Upload JAR to remote repository
```

**Each phase runs ALL previous phases.** So `mvn package` = validate + compile + test + package.

### Common Commands Cheat Sheet
| Command | What It Does |
|---------|-------------|
| `mvn clean` | Delete `target/` folder (fresh start) |
| `mvn compile` | Compile source code |
| `mvn test` | Compile + run tests |
| `mvn package` | Compile + test + create JAR |
| `mvn clean package` | Clean + compile + test + create JAR |
| `mvn install` | Package + install to local repo (~/.m2) |
| `mvn dependency:tree` | Show all dependencies (including transitive) |
| `mvn clean package -DskipTests` | Build without running tests |

---

## Transitive Dependencies

If you depend on Library A, and Library A depends on Library B, Maven **automatically** downloads B too. You don't list it.

```
Your Project
  └── spring-boot-starter-web (you declare this)
        ├── spring-web (transitive — auto-downloaded)
        ├── spring-webmvc (transitive)
        ├── jackson-databind (transitive)
        └── tomcat-embed-core (transitive)
```

This is why one `spring-boot-starter-web` dependency gives you an entire web framework.

Use `mvn dependency:tree` to see the full tree.

---

## Maven Repository Flow

```
Your pom.xml
     │
     ▼
Local Repository (~/.m2/repository)      ← Maven checks here FIRST
     │  (if not found)
     ▼
Maven Central (https://repo.maven.apache.org)  ← Downloads from here
     │
     ▼
Saved to ~/.m2/repository               ← Cached for next time
```

First build = slow (downloading). Subsequent builds = fast (cached locally).

---

## How Spring Boot Uses Maven

In Spring Boot, your `pom.xml` will look like this:

```xml
<!-- Spring Boot Parent — provides default versions for EVERYTHING -->
<parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.3.0</version>
</parent>

<dependencies>
    <!-- ONE starter = entire web framework (Tomcat + Spring MVC + Jackson) -->
    <dependency>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-web</artifactId>
        <!-- No version needed! Parent manages it -->
    </dependency>
</dependencies>
```

Spring Boot Starters = **curated bundles of dependencies** with tested, compatible versions.
The parent POM handles version management so you don't have conflicts.

---

## Files in This Module

```
phase1_maven_basics/
├── pom.xml                                    ← Project configuration
├── NOTES.md                                   ← This file
├── src/main/java/com/learn/
│   ├── App.java                               ← Main app using Gson dependency
│   └── User.java                              ← Simple data class
└── src/test/java/com/learn/
    └── UserTest.java                          ← JUnit 5 tests (scope=test)
```

---

## Key Takeaways

1. **pom.xml** = Single source of truth for your project's dependencies, build, and identity
2. **Dependencies** = Just declare them; Maven downloads, caches, and manages versions
3. **Scopes** = Control where a dependency is available (compile, test, provided, runtime)
4. **Lifecycle** = `clean → compile → test → package → install → deploy`
5. **Transitive deps** = Maven auto-downloads dependencies of your dependencies
6. **Spring Boot starters** = Pre-bundled dependency groups managed by a parent POM
