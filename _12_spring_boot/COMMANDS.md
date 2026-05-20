# Maven & Spring Boot Commands Reference

---

## Maven Build Lifecycle

```bash
mvn clean                    # Delete target/ folder (fresh start)
mvn compile                  # Compile src/main/java → target/classes
mvn test                     # Compile + run all tests
mvn package                  # Compile + test + create JAR in target/
mvn install                  # Package + copy JAR to ~/.m2/repository
mvn clean package            # Clean + full build (most common)
mvn clean package -DskipTests  # Build without running tests (faster)
```

---

## Run Spring Boot App

```bash
# Method 1: Run via Maven plugin (fastest for development)
mvn spring-boot:run

# Method 2: Build JAR then run (production-style)
mvn clean package -q -DskipTests
java -jar target/<artifact>-<version>.jar

# Method 3: Run with specific profile
java -jar target/app-1.0.0.jar --spring.profiles.active=dev
mvn spring-boot:run -Dspring-boot.run.profiles=dev
```

---

## Dependency Management

```bash
mvn dependency:tree                    # Full dependency tree (see transitive deps)
mvn dependency:tree -Dincludes=org.springframework  # Filter by group
mvn dependency:resolve                 # Download all deps without building
mvn dependency:copy-dependencies -DoutputDirectory=target/libs  # Copy JARs locally
mvn versions:display-dependency-updates  # Check for newer versions
```

---

## Debugging & Info

```bash
mvn help:effective-pom          # Show FULL resolved POM (parent + yours merged)
mvn help:effective-settings     # Show active Maven settings
mvn -X clean compile            # Debug mode (verbose output)
mvn -o package                  # Offline mode (no internet, use cached deps)
```

---

## Common Maven Flags

| Flag | Purpose |
|------|---------|
| `-q` | Quiet (less output) |
| `-DskipTests` | Skip test execution |
| `-Dmaven.test.skip=true` | Skip test compilation AND execution |
| `-pl module-name` | Build only a specific module |
| `-am` | Also build dependencies of specified module |
| `-T 4` | Parallel build (4 threads) |

---

## Spring Boot — DevTools (Hot Reload)

```bash
# Just run — DevTools auto-restarts on code change
mvn spring-boot:run
# (requires spring-boot-devtools dependency in pom.xml)
```

---

## Spring Boot — Actuator Endpoints

```bash
curl http://localhost:8080/actuator/health    # Health check
curl http://localhost:8080/actuator/info      # App info
curl http://localhost:8080/actuator/beans     # All beans in container
curl http://localhost:8080/actuator/env       # Environment variables
curl http://localhost:8080/actuator/metrics   # Metrics
```

---

## Passing Configuration to Spring Boot

```bash
# Via command line args
java -jar app.jar --server.port=9090 --spring.datasource.url=jdbc:h2:mem:test

# Via environment variables
export SERVER_PORT=9090
java -jar app.jar

# Via JVM system properties
java -Dserver.port=9090 -jar app.jar
```

---

## Spring Profiles

```bash
# Activate profile (uses application-dev.properties)
java -jar app.jar --spring.profiles.active=dev
java -jar app.jar --spring.profiles.active=prod
java -jar app.jar --spring.profiles.active=dev,local  # Multiple profiles

# Via environment variable
export SPRING_PROFILES_ACTIVE=prod
java -jar app.jar

# Via Maven plugin
mvn spring-boot:run -Dspring-boot.run.profiles=dev
```

---

## Daily Workflow (Quick Reference)

```bash
# Development cycle:
mvn spring-boot:run              # Start app (with auto-restart if DevTools)
# OR
mvn clean package -q -DskipTests && java -jar target/*.jar

# After changing pom.xml:
mvn clean compile                # Re-download deps, recompile

# Before committing:
mvn clean test                   # Ensure all tests pass

# Check what Spring pulled in:
mvn dependency:tree | grep spring
```

---

## Running This Course's Projects

### Phase 1.2 — Maven Basics
```bash
cd ~/Learn/_12_spring_boot/phase1_maven_basics
mvn clean compile                     # Compile
mvn test                              # Run tests
mvn package -q                        # Build JAR
mvn dependency:tree                   # Show dependency tree
mvn dependency:copy-dependencies -DoutputDirectory=target/libs -q
java -cp "target/classes:target/libs/*" com.learn.App
```

### Phase 2.1 — What is Spring
```bash
cd ~/Learn/_12_spring_boot/phase2_spring_core/01_what_is_spring/spring-demo
mvn clean package -q -DskipTests
java -jar target/spring-demo-1.0.0.jar
```

### Phase 2.2 — IoC
```bash
cd ~/Learn/_12_spring_boot/phase2_spring_core/02_ioc/spring-ioc-demo
mvn clean package -q -DskipTests
java -jar target/spring-ioc-demo-1.0.0.jar
```

### Phase 2.3 — Dependency Injection
```bash
cd ~/Learn/_12_spring_boot/phase2_spring_core/03_dependency_injection/spring-di-demo
mvn clean package -q -DskipTests
java -jar target/spring-di-demo-1.0.0.jar
```

### Rebuild ALL Maven projects at once
```bash
for dir in $(find ~/Learn/_12_spring_boot -name "pom.xml" -exec dirname {} \;); do
  echo "=== Building: $dir ==="
  (cd "$dir" && mvn clean package -q -DskipTests)
done
```

### Clean all compiled .class files
```bash
find ~/Learn/_12_spring_boot -type f -name "*.class" -delete
```
