/*
 * =============================================================
 * BEHAVIORAL PATTERN 4: TEMPLATE METHOD
 * =============================================================
 *
 * INTENT: Define the SKELETON of an algorithm in a method,
 *         deferring some steps to subclasses.
 *
 * ANALOGY: Making a beverage — boil water, brew, pour, add condiments.
 *          The PROCESS is the same; the STEPS differ (tea vs coffee).
 *
 * USE WHEN:
 *   - Multiple classes follow the SAME algorithm structure
 *   - Only some steps vary between implementations
 *   - You want to enforce a specific order of operations
 */

public class TemplateMethodPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Beverage Example
        // ═══════════════════════════════════════════════════════
        System.out.println("=== MAKING TEA ===");
        Beverage tea = new Tea();
        tea.prepare();  // template method controls the flow

        System.out.println("\n=== MAKING COFFEE ===");
        Beverage coffee = new Coffee();
        coffee.prepare();

        // ═══════════════════════════════════════════════════════
        // Real-world: Data Mining Pipeline
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== DATA MINING: CSV ===");
        DataMiner csvMiner = new CsvDataMiner();
        csvMiner.mine("sales_data.csv");

        System.out.println("\n=== DATA MINING: JSON ===");
        DataMiner jsonMiner = new JsonDataMiner();
        jsonMiner.mine("api_response.json");

        // ═══════════════════════════════════════════════════════
        // Game AI Template
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== GAME: Chess AI Turn ===");
        GameAI chessAI = new ChessAI();
        chessAI.takeTurn();

        System.out.println("\n=== GAME: Checkers AI Turn ===");
        GameAI checkersAI = new CheckersAI();
        checkersAI.takeTurn();
    }
}

// ═══════════════════════════════════════════════════════════════
// BEVERAGE TEMPLATE
// ═══════════════════════════════════════════════════════════════
abstract class Beverage {

    // TEMPLATE METHOD — defines the algorithm skeleton
    // `final` prevents subclasses from overriding the structure
    public final void prepare() {
        boilWater();
        brew();              // abstract — varies
        pourInCup();
        if (customerWantsCondiments()) {  // hook method
            addCondiments(); // abstract — varies
        }
    }

    // Common steps (concrete — same for all)
    private void boilWater() {
        System.out.println("  1. Boiling water...");
    }

    private void pourInCup() {
        System.out.println("  3. Pouring into cup...");
    }

    // Steps that VARY — subclasses MUST implement
    protected abstract void brew();
    protected abstract void addCondiments();

    // HOOK METHOD — subclasses CAN override, but don't have to
    protected boolean customerWantsCondiments() {
        return true;  // default: yes
    }
}

class Tea extends Beverage {
    @Override
    protected void brew() {
        System.out.println("  2. Steeping tea bag...");
    }

    @Override
    protected void addCondiments() {
        System.out.println("  4. Adding lemon...");
    }
}

class Coffee extends Beverage {
    @Override
    protected void brew() {
        System.out.println("  2. Dripping coffee through filter...");
    }

    @Override
    protected void addCondiments() {
        System.out.println("  4. Adding sugar and milk...");
    }
}

// ═══════════════════════════════════════════════════════════════
// DATA MINING PIPELINE
// ═══════════════════════════════════════════════════════════════
abstract class DataMiner {

    // Template method
    public final void mine(String path) {
        String rawData = openFile(path);    // varies
        String[] data = parseData(rawData); // varies
        String[] analyzed = analyzeData(data);  // common
        generateReport(analyzed);                // common
    }

    // Abstract steps — subclasses implement
    protected abstract String openFile(String path);
    protected abstract String[] parseData(String rawData);

    // Concrete steps — same for all miners
    protected String[] analyzeData(String[] data) {
        System.out.println("  3. Analyzing " + data.length + " records...");
        return data;
    }

    protected void generateReport(String[] analyzed) {
        System.out.println("  4. Report generated: " + analyzed.length + " insights found.");
    }
}

class CsvDataMiner extends DataMiner {
    @Override
    protected String openFile(String path) {
        System.out.println("  1. Opening CSV: " + path);
        return "csv,data,here";
    }

    @Override
    protected String[] parseData(String rawData) {
        System.out.println("  2. Parsing CSV rows (comma-separated)...");
        return rawData.split(",");
    }
}

class JsonDataMiner extends DataMiner {
    @Override
    protected String openFile(String path) {
        System.out.println("  1. Opening JSON: " + path);
        return "{\"a\":1,\"b\":2,\"c\":3}";
    }

    @Override
    protected String[] parseData(String rawData) {
        System.out.println("  2. Parsing JSON fields...");
        return new String[]{"a:1", "b:2", "c:3"};
    }
}

// ═══════════════════════════════════════════════════════════════
// GAME AI TEMPLATE
// ═══════════════════════════════════════════════════════════════
abstract class GameAI {

    // Template method
    public final void takeTurn() {
        collectResources();
        evaluateBoard();
        Move move = calculateBestMove();
        executeMove(move);
    }

    // Hook — optional override
    protected void collectResources() {
        System.out.println("  1. Collecting default resources...");
    }

    // Abstract — must override
    protected abstract void evaluateBoard();
    protected abstract Move calculateBestMove();

    protected void executeMove(Move move) {
        System.out.println("  4. Executing: " + move);
    }
}

record Move(String description) {
    @Override public String toString() { return description; }
}

class ChessAI extends GameAI {
    @Override
    protected void evaluateBoard() {
        System.out.println("  2. Evaluating chess board positions...");
    }

    @Override
    protected Move calculateBestMove() {
        System.out.println("  3. Running minimax with alpha-beta pruning...");
        return new Move("Knight to E5");
    }
}

class CheckersAI extends GameAI {
    @Override
    protected void collectResources() {
        System.out.println("  1. Counting remaining pieces...");
    }

    @Override
    protected void evaluateBoard() {
        System.out.println("  2. Evaluating checkers positions...");
    }

    @Override
    protected Move calculateBestMove() {
        System.out.println("  3. Calculating best jump sequence...");
        return new Move("Double jump from A3 to C5 to E7");
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Template Method = algorithm skeleton in parent, details in children.
 * ✦ Make the template method `final` → subclasses can't change structure.
 * ✦ Abstract methods = MANDATORY customization points.
 * ✦ Hook methods = OPTIONAL customization (provide default behavior).
 *
 * ✦ Template Method vs Strategy:
 *   - Template: uses INHERITANCE (override steps in subclass)
 *   - Strategy: uses COMPOSITION (inject behavior object)
 *   - Template: algorithm structure is FIXED
 *   - Strategy: entire algorithm is SWAPPED
 *
 * COMPILE & RUN:
 *   javac TemplateMethodPattern.java && java TemplateMethodPattern
 */
