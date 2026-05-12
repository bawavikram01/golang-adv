/*
 * ============================================================
 *  CHAPTER 36: DESIGN PATTERNS — BEHAVIORAL
 * ============================================================
 *  Behavioral patterns deal with communication between objects.
 *
 *  BEHAVIORAL PATTERNS:
 *    1. Observer     — event-driven notification
 *    2. Strategy     — interchangeable algorithms
 *    3. Command      — encapsulate actions as objects
 *    4. Iterator     — traverse collections uniformly
 *    5. Template Method — algorithm skeleton, steps overridden
 *    6. State        — behavior changes with state
 *    7. Chain of Responsibility — pass request along handlers
 * ============================================================
 */

import java.util.*;
import java.util.function.*;

public class Chapter36_BehavioralPatterns {

    // ========================================================
    // 1. OBSERVER — "subscribe to events"
    // ========================================================

    interface EventListener {
        void onEvent(String eventType, String data);
    }

    static class EventManager {
        private Map<String, List<EventListener>> listeners = new HashMap<>();

        void subscribe(String event, EventListener listener) {
            listeners.computeIfAbsent(event, k -> new ArrayList<>()).add(listener);
        }

        void unsubscribe(String event, EventListener listener) {
            List<EventListener> list = listeners.get(event);
            if (list != null) list.remove(listener);
        }

        void notify(String event, String data) {
            List<EventListener> list = listeners.get(event);
            if (list != null) {
                for (EventListener l : list) l.onEvent(event, data);
            }
        }
    }

    static class Store {
        EventManager events = new EventManager();
        private List<String> products = new ArrayList<>();

        void addProduct(String product) {
            products.add(product);
            events.notify("product_added", product);
        }

        void removeProduct(String product) {
            products.remove(product);
            events.notify("product_removed", product);
        }
    }

    // ========================================================
    // 2. STRATEGY — "swap algorithms at runtime"
    // ========================================================

    // Using functional interfaces for clean strategy pattern
    interface SortStrategy {
        void sort(int[] arr);
    }

    static class BubbleSort implements SortStrategy {
        @Override
        public void sort(int[] arr) {
            for (int i = 0; i < arr.length - 1; i++)
                for (int j = 0; j < arr.length - 1 - i; j++)
                    if (arr[j] > arr[j + 1]) {
                        int tmp = arr[j]; arr[j] = arr[j + 1]; arr[j + 1] = tmp;
                    }
            System.out.println("    Bubble sorted: " + Arrays.toString(arr));
        }
    }

    static class QuickSort implements SortStrategy {
        @Override
        public void sort(int[] arr) {
            Arrays.sort(arr);  // simplified
            System.out.println("    Quick sorted: " + Arrays.toString(arr));
        }
    }

    static class Sorter {
        private SortStrategy strategy;

        void setStrategy(SortStrategy strategy) { this.strategy = strategy; }
        void sort(int[] arr) { strategy.sort(arr); }
    }

    // ========================================================
    // 3. COMMAND — "encapsulate action as object"
    // ========================================================

    interface Command {
        void execute();
        void undo();
    }

    static class TextEditor {
        private StringBuilder content = new StringBuilder();

        void insert(String text) { content.append(text); }
        void deleteLast(int n) {
            if (n <= content.length()) content.delete(content.length() - n, content.length());
        }
        String getContent() { return content.toString(); }
    }

    static class InsertCommand implements Command {
        private TextEditor editor;
        private String text;

        InsertCommand(TextEditor editor, String text) {
            this.editor = editor;
            this.text = text;
        }

        @Override public void execute() { editor.insert(text); }
        @Override public void undo() { editor.deleteLast(text.length()); }
    }

    static class CommandHistory {
        private Deque<Command> history = new ArrayDeque<>();

        void execute(Command cmd) {
            cmd.execute();
            history.push(cmd);
        }

        void undo() {
            if (!history.isEmpty()) {
                history.pop().undo();
            }
        }
    }

    // ========================================================
    // 4. TEMPLATE METHOD — "skeleton algorithm, steps overridden"
    // ========================================================

    static abstract class DataMiner {
        // Template method — defines the algorithm skeleton
        final void mine(String source) {
            String data = extractData(source);
            String parsed = parseData(data);
            analyzeData(parsed);
            generateReport(parsed);
        }

        abstract String extractData(String source);
        abstract String parseData(String raw);

        void analyzeData(String data) {
            System.out.println("    Analyzing: " + data.length() + " chars");
        }

        void generateReport(String data) {
            System.out.println("    Report generated for: " + data);
        }
    }

    static class CSVMiner extends DataMiner {
        @Override String extractData(String src) {
            System.out.println("    Extracting CSV from: " + src);
            return "csv-raw-data";
        }
        @Override String parseData(String raw) {
            System.out.println("    Parsing CSV...");
            return "csv-parsed";
        }
    }

    static class JSONMiner extends DataMiner {
        @Override String extractData(String src) {
            System.out.println("    Extracting JSON from: " + src);
            return "json-raw-data";
        }
        @Override String parseData(String raw) {
            System.out.println("    Parsing JSON...");
            return "json-parsed";
        }
    }

    // ========================================================
    // 5. STATE — "behavior changes based on state"
    // ========================================================

    interface OrderState {
        void next(Order order);
        void prev(Order order);
        String status();
    }

    static class Order {
        private OrderState state;

        Order() { this.state = new NewState(); }
        void setState(OrderState state) { this.state = state; }
        void next() { state.next(this); }
        void prev() { state.prev(this); }
        String getStatus() { return state.status(); }
    }

    static class NewState implements OrderState {
        @Override public void next(Order o) { o.setState(new PaidState()); }
        @Override public void prev(Order o) { System.out.println("      Already at initial state"); }
        @Override public String status() { return "NEW"; }
    }

    static class PaidState implements OrderState {
        @Override public void next(Order o) { o.setState(new ShippedState()); }
        @Override public void prev(Order o) { o.setState(new NewState()); }
        @Override public String status() { return "PAID"; }
    }

    static class ShippedState implements OrderState {
        @Override public void next(Order o) { o.setState(new DeliveredState()); }
        @Override public void prev(Order o) { o.setState(new PaidState()); }
        @Override public String status() { return "SHIPPED"; }
    }

    static class DeliveredState implements OrderState {
        @Override public void next(Order o) { System.out.println("      Already delivered!"); }
        @Override public void prev(Order o) { System.out.println("      Cannot un-deliver!"); }
        @Override public String status() { return "DELIVERED"; }
    }

    // ========================================================
    // 6. CHAIN OF RESPONSIBILITY — "pass along until handled"
    // ========================================================

    static abstract class AuthHandler {
        private AuthHandler next;

        AuthHandler setNext(AuthHandler next) {
            this.next = next;
            return next;  // for chaining
        }

        boolean handle(Map<String, String> request) {
            if (!check(request)) return false;
            if (next != null) return next.handle(request);
            return true;
        }

        abstract boolean check(Map<String, String> request);
    }

    static class AuthenticationHandler extends AuthHandler {
        @Override boolean check(Map<String, String> req) {
            String user = req.get("user");
            if (user == null || user.isEmpty()) {
                System.out.println("    ✗ Authentication failed: no user");
                return false;
            }
            System.out.println("    ✓ Authenticated: " + user);
            return true;
        }
    }

    static class AuthorizationHandler extends AuthHandler {
        @Override boolean check(Map<String, String> req) {
            String role = req.get("role");
            if (!"admin".equals(role)) {
                System.out.println("    ✗ Authorization failed: role=" + role);
                return false;
            }
            System.out.println("    ✓ Authorized: " + role);
            return true;
        }
    }

    static class RateLimitHandler extends AuthHandler {
        private int requests = 0;
        @Override boolean check(Map<String, String> req) {
            requests++;
            if (requests > 5) {
                System.out.println("    ✗ Rate limit exceeded");
                return false;
            }
            System.out.println("    ✓ Rate limit ok (" + requests + "/5)");
            return true;
        }
    }

    // ========================================================
    // MAIN
    // ========================================================

    public static void main(String[] args) {

        // --- 1. Observer ---
        System.out.println("=== OBSERVER ===\n");
        Store store = new Store();

        EventListener emailAlert = (event, data) ->
            System.out.println("    📧 Email: " + event + " → " + data);
        EventListener logAlert = (event, data) ->
            System.out.println("    📝 Log: " + event + " → " + data);

        store.events.subscribe("product_added", emailAlert);
        store.events.subscribe("product_added", logAlert);
        store.events.subscribe("product_removed", logAlert);

        store.addProduct("Laptop");
        store.removeProduct("Laptop");

        // --- 2. Strategy ---
        System.out.println("\n=== STRATEGY ===\n");
        Sorter sorter = new Sorter();

        sorter.setStrategy(new BubbleSort());
        sorter.sort(new int[]{5, 2, 8, 1, 9});

        sorter.setStrategy(new QuickSort());
        sorter.sort(new int[]{5, 2, 8, 1, 9});

        // Lambda strategy (since Strategy is a functional interface)
        sorter.setStrategy(arr -> {
            Arrays.sort(arr);
            System.out.println("    Lambda sorted: " + Arrays.toString(arr));
        });
        sorter.sort(new int[]{3, 7, 1, 4});

        // --- 3. Command ---
        System.out.println("\n=== COMMAND (with Undo) ===\n");
        TextEditor editor = new TextEditor();
        CommandHistory history = new CommandHistory();

        history.execute(new InsertCommand(editor, "Hello"));
        System.out.println("    After insert: \"" + editor.getContent() + "\"");

        history.execute(new InsertCommand(editor, " World"));
        System.out.println("    After insert: \"" + editor.getContent() + "\"");

        history.undo();
        System.out.println("    After undo:   \"" + editor.getContent() + "\"");

        history.undo();
        System.out.println("    After undo:   \"" + editor.getContent() + "\"");

        // --- 4. Template Method ---
        System.out.println("\n=== TEMPLATE METHOD ===\n");
        DataMiner csvMiner = new CSVMiner();
        csvMiner.mine("data.csv");

        System.out.println();
        DataMiner jsonMiner = new JSONMiner();
        jsonMiner.mine("data.json");

        // --- 5. State ---
        System.out.println("\n=== STATE ===\n");
        Order order = new Order();
        System.out.println("    Status: " + order.getStatus());

        order.next();
        System.out.println("    Status: " + order.getStatus());

        order.next();
        System.out.println("    Status: " + order.getStatus());

        order.next();
        System.out.println("    Status: " + order.getStatus());

        order.next();  // already delivered

        // --- 6. Chain of Responsibility ---
        System.out.println("\n=== CHAIN OF RESPONSIBILITY ===\n");

        AuthHandler chain = new AuthenticationHandler();
        chain.setNext(new AuthorizationHandler())
             .setNext(new RateLimitHandler());

        System.out.println("  Request 1 (admin):");
        Map<String, String> req1 = Map.of("user", "alice", "role", "admin");
        chain.handle(req1);

        System.out.println("\n  Request 2 (viewer):");
        Map<String, String> req2 = Map.of("user", "bob", "role", "viewer");
        chain.handle(req2);

        System.out.println("\n  Request 3 (no user):");
        Map<String, String> req3 = Map.of("role", "admin");
        chain.handle(req3);

        // --- Summary ---
        System.out.println("\n=== WHEN TO USE ===");
        System.out.println("  Observer:  Event systems, pub/sub, UI updates");
        System.out.println("  Strategy:  Sorting, validation, pricing, auth");
        System.out.println("  Command:   Undo/redo, task queues, macros");
        System.out.println("  Template:  Frameworks, data processing pipelines");
        System.out.println("  State:     Order status, game states, workflows");
        System.out.println("  Chain:     Middleware, filters, request processing");

        System.out.println("\n✓ Behavioral Patterns Complete!");
    }
}

/*
 * EXERCISES:
 * 1. Observer: Build a stock price ticker with multiple subscribers.
 * 2. Strategy: Create PaymentProcessor with CreditCard, PayPal, Crypto strategies.
 * 3. Command: Build a remote control with macro (sequence of commands).
 * 4. State: Implement a TrafficLight with Red/Yellow/Green states.
 *
 * NEXT: Chapter 37 — SOLID Principles
 */
