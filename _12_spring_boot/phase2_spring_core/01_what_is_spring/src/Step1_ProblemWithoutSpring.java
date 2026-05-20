/**
 * PHASE 2.1 — THE PROBLEM WITHOUT SPRING
 *
 * This shows a realistic application with multiple layers.
 * Without a framework, YOU must wire everything manually.
 * As the app grows, this becomes unmaintainable.
 */

import java.util.*;

// ============================================================
// LAYER 1: "Database" (simulated)
// ============================================================
class DatabaseConfig {
    private String url;
    private String username;

    public DatabaseConfig(String url, String username) {
        this.url = url;
        this.username = username;
        System.out.println("  [1] DatabaseConfig created: " + url);
    }

    public String getUrl() { return url; }
    public String getUsername() { return username; }
}

class DataSource {
    private DatabaseConfig config;

    public DataSource(DatabaseConfig config) {
        this.config = config;
        System.out.println("  [2] DataSource created (connects to: " + config.getUrl() + ")");
    }

    public void execute(String sql) {
        System.out.println("      [DB] " + sql);
    }
}


// ============================================================
// LAYER 2: Repositories (data access)
// ============================================================
class UserRepository {
    private DataSource dataSource;

    public UserRepository(DataSource dataSource) {
        this.dataSource = dataSource;
        System.out.println("  [3] UserRepository created");
    }

    public String findByEmail(String email) {
        dataSource.execute("SELECT * FROM users WHERE email = '" + email + "'");
        return "User(" + email + ")";
    }

    public void save(String user) {
        dataSource.execute("INSERT INTO users VALUES ('" + user + "')");
    }
}

class OrderRepository {
    private DataSource dataSource;

    public OrderRepository(DataSource dataSource) {
        this.dataSource = dataSource;
        System.out.println("  [4] OrderRepository created");
    }

    public void save(String orderId, String userEmail) {
        dataSource.execute("INSERT INTO orders VALUES ('" + orderId + "', '" + userEmail + "')");
    }

    public List<String> findByUser(String userEmail) {
        dataSource.execute("SELECT * FROM orders WHERE user = '" + userEmail + "'");
        return List.of("ORD-001", "ORD-002");
    }
}


// ============================================================
// LAYER 3: Services (business logic)
// ============================================================
class EmailService {
    private String smtpHost;

    public EmailService(String smtpHost) {
        this.smtpHost = smtpHost;
        System.out.println("  [5] EmailService created (smtp: " + smtpHost + ")");
    }

    public void send(String to, String subject) {
        System.out.println("      [EMAIL] To: " + to + " | Subject: " + subject);
    }
}

class UserServiceManual {
    private UserRepository userRepository;
    private EmailService emailService;

    public UserServiceManual(UserRepository userRepository, EmailService emailService) {
        this.userRepository = userRepository;
        this.emailService = emailService;
        System.out.println("  [6] UserService created");
    }

    public void register(String email) {
        userRepository.save(email);
        emailService.send(email, "Welcome!");
        System.out.println("      [SERVICE] User registered: " + email);
    }

    public String getUser(String email) {
        return userRepository.findByEmail(email);
    }
}

class OrderServiceManual {
    private OrderRepository orderRepository;
    private UserServiceManual userService;
    private EmailService emailService;

    public OrderServiceManual(OrderRepository orderRepository,
                              UserServiceManual userService,
                              EmailService emailService) {
        this.orderRepository = orderRepository;
        this.userService = userService;
        this.emailService = emailService;
        System.out.println("  [7] OrderService created");
    }

    public void placeOrder(String orderId, String userEmail) {
        String user = userService.getUser(userEmail);
        orderRepository.save(orderId, userEmail);
        emailService.send(userEmail, "Order " + orderId + " confirmed!");
        System.out.println("      [SERVICE] Order placed: " + orderId + " for " + user);
    }
}


// ============================================================
// LAYER 4: Controllers (entry points)
// ============================================================
class UserControllerManual {
    private UserServiceManual userService;

    public UserControllerManual(UserServiceManual userService) {
        this.userService = userService;
        System.out.println("  [8] UserController created");
    }

    public void handleRegister(String email) {
        System.out.println("\n  → POST /users/register (email=" + email + ")");
        userService.register(email);
    }
}

class OrderControllerManual {
    private OrderServiceManual orderService;

    public OrderControllerManual(OrderServiceManual orderService) {
        this.orderService = orderService;
        System.out.println("  [9] OrderController created");
    }

    public void handlePlaceOrder(String orderId, String userEmail) {
        System.out.println("\n  → POST /orders (id=" + orderId + ", user=" + userEmail + ")");
        orderService.placeOrder(orderId, userEmail);
    }
}


// ============================================================
// MAIN — Look at how much wiring YOU have to do!
// ============================================================
public class Step1_ProblemWithoutSpring {
    public static void main(String[] args) {

        System.out.println("=== THE PROBLEM: MANUAL WIRING ===");
        System.out.println("  You must create EVERY object in the RIGHT ORDER.\n");

        // ---- You must know the exact order of creation ----
        // If you get the order wrong → NullPointerException
        // If a constructor changes → you fix it HERE manually

        System.out.println("--- Creating objects (9 steps!) ---\n");

        // Step 1: Config (depends on nothing)
        DatabaseConfig dbConfig = new DatabaseConfig("jdbc:mysql://localhost:3306/shop", "admin");

        // Step 2: DataSource (depends on config)
        DataSource dataSource = new DataSource(dbConfig);

        // Step 3-4: Repositories (depend on DataSource)
        UserRepository userRepo = new UserRepository(dataSource);
        OrderRepository orderRepo = new OrderRepository(dataSource);

        // Step 5: EmailService (depends on config)
        EmailService emailService = new EmailService("smtp.gmail.com");

        // Step 6: UserService (depends on UserRepo + EmailService)
        UserServiceManual userService = new UserServiceManual(userRepo, emailService);

        // Step 7: OrderService (depends on OrderRepo + UserService + EmailService)
        OrderServiceManual orderService = new OrderServiceManual(orderRepo, userService, emailService);

        // Step 8-9: Controllers (depend on Services)
        UserControllerManual userController = new UserControllerManual(userService);
        OrderControllerManual orderController = new OrderControllerManual(orderService);


        System.out.println("\n--- Application ready! Handling requests ---");

        userController.handleRegister("alice@example.com");
        orderController.handlePlaceOrder("ORD-100", "alice@example.com");


        System.out.println("\n\n=== PROBLEMS WITH THIS APPROACH ===");
        System.out.println("  1. YOU must know creation order (9 objects here, imagine 200)");
        System.out.println("  2. If UserService adds a new dependency → change main() manually");
        System.out.println("  3. DataSource must be shared (singleton) — YOU must ensure this");
        System.out.println("  4. No lifecycle management (who closes the DataSource on shutdown?)");
        System.out.println("  5. Testing is hard (can't easily swap EmailService with a mock)");
        System.out.println("  6. Config values are hardcoded (no profiles, no externalization)");
        System.out.println("  7. No cross-cutting concerns (where do you add logging? transactions?)");

        System.out.println("\n=== SPRING'S PROMISE ===");
        System.out.println("  You write classes + annotate them.");
        System.out.println("  Spring figures out the order, creates everything, wires everything.");
        System.out.println("  Next: We'll build our own mini-Spring to see HOW it works.");
    }
}
