/*
 * =============================================================
 * STRUCTURAL PATTERN 4: PROXY
 * =============================================================
 *
 * INTENT: Provide a SURROGATE or placeholder for another object
 *         to control access to it.
 *
 * ANALOGY: A security guard at a building entrance — you don't
 *          go directly to the CEO. The guard checks credentials first.
 *
 * TYPES:
 *   1. PROTECTION PROXY — access control (authorization)
 *   2. VIRTUAL PROXY — lazy initialization (expensive objects)
 *   3. LOGGING PROXY — logging/auditing without touching real code
 *   4. CACHING PROXY — add cache layer transparently
 *
 * REAL EXAMPLES: Spring AOP, Hibernate lazy loading, Java Proxy,
 *                RMI stubs, VPN (network proxy)
 */

import java.util.*;

public class ProxyPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // 1. PROTECTION PROXY — Access Control
        // ═══════════════════════════════════════════════════════
        System.out.println("=== PROTECTION PROXY ===");

        Database adminDb = new DatabaseProxy(new RealDatabase(), "ADMIN");
        adminDb.read("users");    // allowed
        adminDb.write("users", "new data");  // allowed

        System.out.println();
        Database viewerDb = new DatabaseProxy(new RealDatabase(), "VIEWER");
        viewerDb.read("users");   // allowed
        viewerDb.write("users", "hack");  // DENIED!

        // ═══════════════════════════════════════════════════════
        // 2. VIRTUAL PROXY — Lazy Loading
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== VIRTUAL PROXY (Lazy Loading) ===");

        Image img1 = new ImageProxy("photo1.jpg");
        Image img2 = new ImageProxy("photo2.jpg");
        Image img3 = new ImageProxy("photo3.jpg");

        // Images NOT loaded yet — no memory used!
        System.out.println("Images created but NOT loaded yet.");

        // Only loaded when actually displayed
        img1.display();  // loads + displays
        img1.display();  // already loaded — just displays
        // img2 and img3 never loaded — memory saved!

        // ═══════════════════════════════════════════════════════
        // 3. CACHING PROXY
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== CACHING PROXY ===");

        WeatherService realService = new RealWeatherService();
        WeatherService cachedService = new CachingWeatherProxy(realService);

        System.out.println(cachedService.getWeather("London"));  // fetches
        System.out.println(cachedService.getWeather("Paris"));   // fetches
        System.out.println(cachedService.getWeather("London"));  // CACHED!
        System.out.println(cachedService.getWeather("Paris"));   // CACHED!

        // ═══════════════════════════════════════════════════════
        // 4. LOGGING PROXY
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== LOGGING PROXY ===");

        UserService realUserService = new RealUserService();
        UserService loggedService = new LoggingUserServiceProxy(realUserService);

        loggedService.createUser("alice");
        loggedService.getUser("alice");
        loggedService.deleteUser("alice");
    }
}

// ═══════════════════════════════════════════════════════════════
// 1. PROTECTION PROXY
// ═══════════════════════════════════════════════════════════════
interface Database {
    String read(String table);
    void write(String table, String data);
}

class RealDatabase implements Database {
    @Override
    public String read(String table) {
        System.out.println("  [DB] Reading from " + table);
        return "data from " + table;
    }

    @Override
    public void write(String table, String data) {
        System.out.println("  [DB] Writing to " + table + ": " + data);
    }
}

class DatabaseProxy implements Database {
    private Database realDb;
    private String userRole;

    public DatabaseProxy(Database realDb, String userRole) {
        this.realDb = realDb;
        this.userRole = userRole;
    }

    @Override
    public String read(String table) {
        // Everyone can read
        return realDb.read(table);
    }

    @Override
    public void write(String table, String data) {
        if (!"ADMIN".equals(userRole)) {
            System.out.println("  ✗ ACCESS DENIED: " + userRole + " cannot write!");
            return;
        }
        realDb.write(table, data);
    }
}

// ═══════════════════════════════════════════════════════════════
// 2. VIRTUAL PROXY — Lazy Loading
// ═══════════════════════════════════════════════════════════════
interface Image {
    void display();
}

class RealImage implements Image {
    private String filename;

    public RealImage(String filename) {
        this.filename = filename;
        loadFromDisk();  // EXPENSIVE operation
    }

    private void loadFromDisk() {
        System.out.println("  [Disk] Loading image: " + filename + " (slow!)");
    }

    @Override
    public void display() {
        System.out.println("  🖼️ Displaying: " + filename);
    }
}

class ImageProxy implements Image {
    private String filename;
    private RealImage realImage;  // null until needed

    public ImageProxy(String filename) {
        this.filename = filename;
        // NOT loading the image yet!
    }

    @Override
    public void display() {
        if (realImage == null) {
            realImage = new RealImage(filename);  // lazy load
        }
        realImage.display();
    }
}

// ═══════════════════════════════════════════════════════════════
// 3. CACHING PROXY
// ═══════════════════════════════════════════════════════════════
interface WeatherService {
    String getWeather(String city);
}

class RealWeatherService implements WeatherService {
    @Override
    public String getWeather(String city) {
        System.out.println("  [API] Fetching weather for " + city + "...");
        return "Sunny, 25°C in " + city;
    }
}

class CachingWeatherProxy implements WeatherService {
    private WeatherService realService;
    private Map<String, String> cache = new HashMap<>();

    public CachingWeatherProxy(WeatherService service) {
        this.realService = service;
    }

    @Override
    public String getWeather(String city) {
        if (cache.containsKey(city)) {
            System.out.println("  [CACHE HIT] " + city);
            return cache.get(city);
        }
        String result = realService.getWeather(city);
        cache.put(city, result);
        return result;
    }
}

// ═══════════════════════════════════════════════════════════════
// 4. LOGGING PROXY
// ═══════════════════════════════════════════════════════════════
interface UserService {
    void createUser(String name);
    String getUser(String name);
    void deleteUser(String name);
}

class RealUserService implements UserService {
    @Override public void createUser(String name) { System.out.println("  Created user: " + name); }
    @Override public String getUser(String name)  { System.out.println("  Got user: " + name); return name; }
    @Override public void deleteUser(String name) { System.out.println("  Deleted user: " + name); }
}

class LoggingUserServiceProxy implements UserService {
    private UserService realService;

    public LoggingUserServiceProxy(UserService service) { this.realService = service; }

    @Override
    public void createUser(String name) {
        System.out.println("  [LOG] createUser(\"" + name + "\") called");
        long start = System.nanoTime();
        realService.createUser(name);
        System.out.println("  [LOG] Completed in " + (System.nanoTime() - start) / 1000 + "μs");
    }

    @Override
    public String getUser(String name) {
        System.out.println("  [LOG] getUser(\"" + name + "\") called");
        return realService.getUser(name);
    }

    @Override
    public void deleteUser(String name) {
        System.out.println("  [LOG] deleteUser(\"" + name + "\") called");
        realService.deleteUser(name);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Proxy wraps a real object, SAME interface, adds control:
 *   - Protection → check permissions before delegating
 *   - Virtual → create expensive object only when needed
 *   - Caching → store results, avoid redundant calls
 *   - Logging → log calls without touching original code
 *
 * ✦ Proxy vs Decorator:
 *   - Proxy CONTROLS access to an object
 *   - Decorator ADDS behavior to an object
 *   - Proxy often creates the real object itself
 *   - Decorator always receives the wrapped object
 *
 * ✦ Hibernate lazy loading = Virtual Proxy
 * ✦ Spring AOP = Proxy (logging, security, transactions)
 *
 * COMPILE & RUN:
 *   javac ProxyPattern.java && java ProxyPattern
 */
