/*
 * =============================================================
 * LLD CASE STUDY 10: RATE LIMITER
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Limit API requests per client
 *   - Multiple algorithms: Token Bucket, Sliding Window
 *   - Configurable rate limits
 *   - Thread-safe
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (different rate limiting algorithms)
 *   - Factory (create limiter by type)
 *   - Decorator (add rate limiting to any service)
 *
 * THIS IS A VERY COMMON SYSTEM DESIGN + LLD INTERVIEW QUESTION.
 */

import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.*;

public class RateLimiter {

    public static void main(String[] args) throws InterruptedException {
        System.out.println("=== RATE LIMITER ===\n");

        // ═══════════════════════════════════════════════════════
        // Demo 1: Token Bucket
        // ═══════════════════════════════════════════════════════
        System.out.println("--- TOKEN BUCKET (5 tokens, refill 2/sec) ---");
        RateLimitStrategy tokenBucket = new TokenBucket(5, 2);

        for (int i = 1; i <= 8; i++) {
            boolean allowed = tokenBucket.allowRequest("user1");
            System.out.printf("  Request %d: %s%n", i, allowed ? "✓ ALLOWED" : "✗ DENIED");
        }

        System.out.println("  [waiting 2 seconds for token refill...]");
        Thread.sleep(2000);

        for (int i = 9; i <= 11; i++) {
            boolean allowed = tokenBucket.allowRequest("user1");
            System.out.printf("  Request %d: %s%n", i, allowed ? "✓ ALLOWED" : "✗ DENIED");
        }

        // ═══════════════════════════════════════════════════════
        // Demo 2: Sliding Window Counter
        // ═══════════════════════════════════════════════════════
        System.out.println("\n--- SLIDING WINDOW (3 requests per 2 seconds) ---");
        RateLimitStrategy slidingWindow = new SlidingWindowCounter(3, 2000);

        for (int i = 1; i <= 5; i++) {
            boolean allowed = slidingWindow.allowRequest("user2");
            System.out.printf("  Request %d: %s%n", i, allowed ? "✓ ALLOWED" : "✗ DENIED");
            Thread.sleep(300);
        }

        System.out.println("  [waiting for window to slide...]");
        Thread.sleep(2000);

        for (int i = 6; i <= 8; i++) {
            boolean allowed = slidingWindow.allowRequest("user2");
            System.out.printf("  Request %d: %s%n", i, allowed ? "✓ ALLOWED" : "✗ DENIED");
        }

        // ═══════════════════════════════════════════════════════
        // Demo 3: Rate-limited API service
        // ═══════════════════════════════════════════════════════
        System.out.println("\n--- RATE-LIMITED API SERVICE ---");
        ApiService rawService = new RealApiService();
        ApiService limitedService = new RateLimitedService(rawService,
                new TokenBucket(3, 1));

        for (int i = 1; i <= 5; i++) {
            String result = limitedService.handleRequest("user3", "/api/data");
            System.out.println("  " + result);
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// RATE LIMIT STRATEGY — Strategy Pattern
// ═══════════════════════════════════════════════════════════════
interface RateLimitStrategy {
    boolean allowRequest(String clientId);
}

// ═══════════════════════════════════════════════════════════════
// TOKEN BUCKET Algorithm
// ═══════════════════════════════════════════════════════════════
/*
 * HOW IT WORKS:
 *   - Bucket holds N tokens (capacity)
 *   - Each request consumes 1 token
 *   - Tokens refill at a fixed rate (e.g., 2/sec)
 *   - If bucket empty → request denied
 *
 * PROS: Allows bursts up to bucket capacity
 * CONS: If bucket large, can overwhelm backend during burst
 *
 *   Bucket: [●●●●●]  capacity=5, rate=2/sec
 *   Request → consume 1 → [●●●●○]
 *   Request → consume 1 → [●●●○○]
 *   ...after 1 sec → refill 2 → [●●●●●]
 */
class TokenBucket implements RateLimitStrategy {
    private final int capacity;
    private final int refillRate; // tokens per second
    private final Map<String, BucketState> buckets = new ConcurrentHashMap<>();

    public TokenBucket(int capacity, int refillRate) {
        this.capacity = capacity;
        this.refillRate = refillRate;
    }

    @Override
    public boolean allowRequest(String clientId) {
        BucketState bucket = buckets.computeIfAbsent(clientId,
                k -> new BucketState(capacity));
        return bucket.tryConsume(capacity, refillRate);
    }

    private static class BucketState {
        private double tokens;
        private long lastRefillTime;

        BucketState(int capacity) {
            this.tokens = capacity;
            this.lastRefillTime = System.nanoTime();
        }

        synchronized boolean tryConsume(int capacity, int refillRate) {
            refill(capacity, refillRate);
            if (tokens >= 1) {
                tokens -= 1;
                return true;
            }
            return false;
        }

        private void refill(int capacity, int refillRate) {
            long now = System.nanoTime();
            double elapsed = (now - lastRefillTime) / 1_000_000_000.0;
            double newTokens = elapsed * refillRate;
            tokens = Math.min(capacity, tokens + newTokens);
            lastRefillTime = now;
        }
    }
}

// ═══════════════════════════════════════════════════════════════
// SLIDING WINDOW COUNTER Algorithm
// ═══════════════════════════════════════════════════════════════
/*
 * HOW IT WORKS:
 *   - Track timestamps of recent requests
 *   - Window slides with current time
 *   - Count requests in window → if < limit → allow
 *
 * PROS: Smooth rate limiting, no bursts
 * CONS: More memory (stores timestamps)
 *
 *   Time: --|--req--req--req--|--req--     window = 2s, limit = 3
 *           ^                 ^
 *        window start      now
 *   Count in window = 3 → NEXT request DENIED
 */
class SlidingWindowCounter implements RateLimitStrategy {
    private final int maxRequests;
    private final long windowSizeMs;
    private final Map<String, Deque<Long>> requestLogs = new ConcurrentHashMap<>();

    public SlidingWindowCounter(int maxRequests, long windowSizeMs) {
        this.maxRequests = maxRequests;
        this.windowSizeMs = windowSizeMs;
    }

    @Override
    public synchronized boolean allowRequest(String clientId) {
        long now = System.currentTimeMillis();
        Deque<Long> timestamps = requestLogs.computeIfAbsent(clientId,
                k -> new LinkedList<>());

        // Remove expired entries
        while (!timestamps.isEmpty() && now - timestamps.peekFirst() > windowSizeMs) {
            timestamps.pollFirst();
        }

        if (timestamps.size() < maxRequests) {
            timestamps.addLast(now);
            return true;
        }
        return false;
    }
}

// ═══════════════════════════════════════════════════════════════
// API SERVICE — Decorator Pattern for rate limiting
// ═══════════════════════════════════════════════════════════════
interface ApiService {
    String handleRequest(String clientId, String endpoint);
}

class RealApiService implements ApiService {
    @Override
    public String handleRequest(String clientId, String endpoint) {
        return "200 OK — " + endpoint + " data for " + clientId;
    }
}

class RateLimitedService implements ApiService {
    private final ApiService wrapped;
    private final RateLimitStrategy limiter;

    public RateLimitedService(ApiService wrapped, RateLimitStrategy limiter) {
        this.wrapped = wrapped;
        this.limiter = limiter;
    }

    @Override
    public String handleRequest(String clientId, String endpoint) {
        if (!limiter.allowRequest(clientId)) {
            return "429 Too Many Requests — slow down, " + clientId + "!";
        }
        return wrapped.handleRequest(clientId, endpoint);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * 1. Token Bucket: allows bursts, simple, widely used (AWS, Stripe)
 * 2. Sliding Window: smooth limiting, more memory
 * 3. Strategy pattern → swap algorithms without changing clients
 * 4. Decorator pattern → add rate limiting to ANY service
 * 5. ConcurrentHashMap for thread-safe per-client state
 *
 * OTHER ALGORITHMS TO KNOW:
 *   - Fixed Window Counter (simpler but has boundary burst problem)
 *   - Leaky Bucket (process requests at fixed rate)
 *   - Sliding Window Log (hybrid)
 *
 * INTERVIEW TIPS:
 *   - Discuss distributed rate limiting (Redis + Lua scripts)
 *   - Mention HTTP 429 status code
 *   - Talk about rate limit headers (X-RateLimit-Remaining)
 *
 * COMPILE & RUN:
 *   javac RateLimiter.java && java RateLimiter
 */
