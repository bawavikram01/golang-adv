/*
 * =============================================================
 * BEHAVIORAL PATTERN 6: CHAIN OF RESPONSIBILITY
 * =============================================================
 *
 * INTENT: Pass a request along a CHAIN of handlers.
 *         Each handler either processes the request OR passes
 *         it to the NEXT handler in the chain.
 *
 * ANALOGY: Customer support escalation —
 *          Level 1 → Level 2 → Manager → Director.
 *          Each level either solves it or escalates.
 *
 * USE WHEN:
 *   - Multiple handlers could process a request
 *   - The handler isn't known in advance
 *   - You want to decouple senders from receivers
 *   - Middleware pipelines (auth → validate → log → handle)
 *
 * REAL EXAMPLES: Java servlet filters, Spring interceptors,
 *                logging frameworks, exception handling,
 *                ATM cash dispensing
 */

public class ChainOfResponsibilityPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Support Ticket Escalation
        // ═══════════════════════════════════════════════════════
        System.out.println("=== SUPPORT TICKET CHAIN ===");

        SupportHandler chain = buildSupportChain();

        chain.handle(new SupportTicket("Password reset", Priority.LOW));
        System.out.println();
        chain.handle(new SupportTicket("Server is slow", Priority.MEDIUM));
        System.out.println();
        chain.handle(new SupportTicket("Data breach detected!", Priority.CRITICAL));
        System.out.println();
        chain.handle(new SupportTicket("System completely down!", Priority.CRITICAL));

        // ═══════════════════════════════════════════════════════
        // HTTP Middleware Pipeline
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== HTTP MIDDLEWARE CHAIN ===");

        Middleware pipeline = new AuthMiddleware()
                .setNext(new RateLimitMiddleware(3))
                .setNext(new LoggingMiddleware())
                .setNext(new RequestHandlerMiddleware());

        HttpRequest req1 = new HttpRequest("/api/users", "valid-token", "192.168.1.1");
        pipeline.process(req1);

        System.out.println();
        HttpRequest req2 = new HttpRequest("/api/admin", null, "10.0.0.1");
        pipeline.process(req2);  // should be blocked by auth

        // ═══════════════════════════════════════════════════════
        // ATM Cash Dispenser
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ATM CASH DISPENSER ===");

        CashHandler dispenser = new CashHandler(2000)
                .setNextHandler(new CashHandler(500))
                .setNextHandler(new CashHandler(200))
                .setNextHandler(new CashHandler(100));

        System.out.println("Dispensing 3800:");
        dispenser.dispense(3800);

        System.out.println("\nDispensing 2700:");
        dispenser.dispense(2700);
    }

    static SupportHandler buildSupportChain() {
        SupportHandler l1 = new Level1Support();
        SupportHandler l2 = new Level2Support();
        SupportHandler manager = new ManagerSupport();
        SupportHandler director = new DirectorSupport();

        l1.setNext(l2);
        l2.setNext(manager);
        manager.setNext(director);

        return l1;  // return head of chain
    }
}

// ═══════════════════════════════════════════════════════════════
// SUPPORT TICKET CHAIN
// ═══════════════════════════════════════════════════════════════
enum Priority { LOW, MEDIUM, HIGH, CRITICAL }

class SupportTicket {
    private String issue;
    private Priority priority;

    public SupportTicket(String issue, Priority priority) {
        this.issue = issue;
        this.priority = priority;
    }

    public String getIssue()     { return issue; }
    public Priority getPriority() { return priority; }
}

// Abstract handler
abstract class SupportHandler {
    protected SupportHandler next;

    public SupportHandler setNext(SupportHandler next) {
        this.next = next;
        return next;  // enables chaining
    }

    public void handle(SupportTicket ticket) {
        if (canHandle(ticket)) {
            process(ticket);
        } else if (next != null) {
            System.out.println("  ↳ Escalating...");
            next.handle(ticket);
        } else {
            System.out.println("  ✗ No handler available for: " + ticket.getIssue());
        }
    }

    protected abstract boolean canHandle(SupportTicket ticket);
    protected abstract void process(SupportTicket ticket);
}

class Level1Support extends SupportHandler {
    @Override protected boolean canHandle(SupportTicket t) { return t.getPriority() == Priority.LOW; }
    @Override protected void process(SupportTicket t) {
        System.out.println("  👤 Level 1 resolved: " + t.getIssue());
    }
}

class Level2Support extends SupportHandler {
    @Override protected boolean canHandle(SupportTicket t) { return t.getPriority() == Priority.MEDIUM; }
    @Override protected void process(SupportTicket t) {
        System.out.println("  👥 Level 2 resolved: " + t.getIssue());
    }
}

class ManagerSupport extends SupportHandler {
    @Override protected boolean canHandle(SupportTicket t) { return t.getPriority() == Priority.HIGH; }
    @Override protected void process(SupportTicket t) {
        System.out.println("  👔 Manager resolved: " + t.getIssue());
    }
}

class DirectorSupport extends SupportHandler {
    @Override protected boolean canHandle(SupportTicket t) { return t.getPriority() == Priority.CRITICAL; }
    @Override protected void process(SupportTicket t) {
        System.out.println("  🏢 Director handling CRITICAL: " + t.getIssue());
    }
}

// ═══════════════════════════════════════════════════════════════
// HTTP MIDDLEWARE CHAIN
// ═══════════════════════════════════════════════════════════════
class HttpRequest {
    String path;
    String authToken;
    String ip;
    boolean authenticated = false;

    public HttpRequest(String path, String token, String ip) {
        this.path = path;
        this.authToken = token;
        this.ip = ip;
    }
}

abstract class Middleware {
    protected Middleware next;

    public Middleware setNext(Middleware next) {
        this.next = next;
        return next;
    }

    public void process(HttpRequest request) {
        if (check(request) && next != null) {
            next.process(request);
        }
    }

    protected abstract boolean check(HttpRequest request);
}

class AuthMiddleware extends Middleware {
    @Override
    protected boolean check(HttpRequest req) {
        if (req.authToken == null || req.authToken.isEmpty()) {
            System.out.println("  ✗ [Auth] Unauthorized — no token provided");
            return false;  // STOPS the chain
        }
        req.authenticated = true;
        System.out.println("  ✓ [Auth] Token validated");
        return true;  // continues
    }
}

class RateLimitMiddleware extends Middleware {
    private int maxRequests;
    private int count = 0;

    public RateLimitMiddleware(int max) { this.maxRequests = max; }

    @Override
    protected boolean check(HttpRequest req) {
        count++;
        if (count > maxRequests) {
            System.out.println("  ✗ [RateLimit] Too many requests (limit: " + maxRequests + ")");
            return false;
        }
        System.out.println("  ✓ [RateLimit] " + count + "/" + maxRequests);
        return true;
    }
}

class LoggingMiddleware extends Middleware {
    @Override
    protected boolean check(HttpRequest req) {
        System.out.println("  ✓ [Log] " + req.path + " from " + req.ip);
        return true;  // logging never blocks
    }
}

class RequestHandlerMiddleware extends Middleware {
    @Override
    protected boolean check(HttpRequest req) {
        System.out.println("  ✓ [Handler] Processing " + req.path + " → 200 OK");
        return true;
    }
}

// ═══════════════════════════════════════════════════════════════
// ATM CASH DISPENSER
// ═══════════════════════════════════════════════════════════════
class CashHandler {
    private int denomination;
    private CashHandler next;

    public CashHandler(int denomination) {
        this.denomination = denomination;
    }

    public CashHandler setNextHandler(CashHandler next) {
        this.next = next;
        return next;
    }

    public void dispense(int amount) {
        int notes = amount / denomination;
        int remainder = amount % denomination;

        if (notes > 0) {
            System.out.println("  💵 " + notes + " × ₹" + denomination);
        }

        if (remainder > 0 && next != null) {
            next.dispense(remainder);
        } else if (remainder > 0) {
            System.out.println("  ✗ Cannot dispense ₹" + remainder + " (no smaller denomination)");
        }
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Chain = linked list of handlers. Request flows until handled.
 * ✦ Each handler: process here OR pass to next.
 * ✦ Sender doesn't know which handler will process.
 *
 * ✦ Two flavors:
 *   1. FIRST-MATCH: stops at the first handler that can process
 *   2. PIPELINE: ALL handlers process in sequence (middleware)
 *
 * ✦ Very common in:
 *   - Servlet Filters / Spring Interceptors
 *   - Exception handling (catch blocks)
 *   - Event bubbling in UI
 *   - ATM cash dispensing
 *
 * COMPILE & RUN:
 *   javac ChainOfResponsibilityPattern.java && java ChainOfResponsibilityPattern
 */
