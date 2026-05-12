/*
 * =============================================================
 * LLD CASE STUDY 5: SPLITWISE (Expense Sharing)
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Users can add expenses (split equally, by exact amount, %)
 *   - Track who owes whom and how much
 *   - Simplify debts (minimize transactions)
 *   - Show balances for each user
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (split calculation)
 *   - Observer (expense notification)
 *   - Singleton (expense manager)
 *
 * Very popular LLD interview question.
 */

import java.util.*;
import java.util.stream.*;

public class Splitwise {

    public static void main(String[] args) {

        ExpenseManager manager = ExpenseManager.getInstance();

        // Register users
        User alice = new User("U1", "Alice");
        User bob = new User("U2", "Bob");
        User charlie = new User("U3", "Charlie");
        User diana = new User("U4", "Diana");

        manager.addUser(alice);
        manager.addUser(bob);
        manager.addUser(charlie);
        manager.addUser(diana);

        // ═══════════════════════════════════════════════════════
        // Expense 1: Alice pays $1000 for dinner, split EQUALLY
        // ═══════════════════════════════════════════════════════
        System.out.println("=== EXPENSE 1: Equal Split ===");
        manager.addExpense(new Expense(
                alice,
                1000.0,
                new EqualSplit(),
                List.of(alice, bob, charlie, diana),
                "Dinner"
        ));

        // ═══════════════════════════════════════════════════════
        // Expense 2: Bob pays $600, exact split
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== EXPENSE 2: Exact Split ===");
        Map<User, Double> exactAmounts = new LinkedHashMap<>();
        exactAmounts.put(alice, 100.0);
        exactAmounts.put(bob, 200.0);
        exactAmounts.put(charlie, 300.0);

        manager.addExpense(new Expense(
                bob,
                600.0,
                new ExactSplit(exactAmounts),
                List.of(alice, bob, charlie),
                "Movie tickets"
        ));

        // ═══════════════════════════════════════════════════════
        // Expense 3: Charlie pays $400, percentage split
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== EXPENSE 3: Percentage Split ===");
        Map<User, Double> percentages = new LinkedHashMap<>();
        percentages.put(alice, 40.0);
        percentages.put(bob, 30.0);
        percentages.put(charlie, 20.0);
        percentages.put(diana, 10.0);

        manager.addExpense(new Expense(
                charlie,
                400.0,
                new PercentageSplit(percentages),
                List.of(alice, bob, charlie, diana),
                "Groceries"
        ));

        // ═══════════════════════════════════════════════════════
        // Show all balances
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ALL BALANCES ===");
        manager.showBalances();

        // ═══════════════════════════════════════════════════════
        // Show individual balance
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ALICE'S BALANCES ===");
        manager.showBalance(alice);
    }
}

// ═══════════════════════════════════════════════════════════════
// USER
// ═══════════════════════════════════════════════════════════════
class User {
    private final String userId;
    private final String name;

    public User(String userId, String name) {
        this.userId = userId;
        this.name = name;
    }

    public String getUserId() { return userId; }
    public String getName()   { return name; }

    @Override
    public String toString() { return name; }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof User)) return false;
        return userId.equals(((User) o).userId);
    }

    @Override
    public int hashCode() { return userId.hashCode(); }
}

// ═══════════════════════════════════════════════════════════════
// SPLIT STRATEGY
// ═══════════════════════════════════════════════════════════════
interface SplitStrategy {
    Map<User, Double> calculateShares(double totalAmount, List<User> participants);
    boolean validate(double totalAmount, List<User> participants);
}

class EqualSplit implements SplitStrategy {
    @Override
    public Map<User, Double> calculateShares(double total, List<User> participants) {
        double share = Math.round(total * 100.0 / participants.size()) / 100.0;
        Map<User, Double> shares = new LinkedHashMap<>();
        for (User u : participants) {
            shares.put(u, share);
        }
        return shares;
    }

    @Override
    public boolean validate(double total, List<User> participants) {
        return !participants.isEmpty();
    }
}

class ExactSplit implements SplitStrategy {
    private Map<User, Double> exactAmounts;

    public ExactSplit(Map<User, Double> amounts) {
        this.exactAmounts = amounts;
    }

    @Override
    public Map<User, Double> calculateShares(double total, List<User> participants) {
        return new LinkedHashMap<>(exactAmounts);
    }

    @Override
    public boolean validate(double total, List<User> participants) {
        double sum = exactAmounts.values().stream().mapToDouble(Double::doubleValue).sum();
        return Math.abs(sum - total) < 0.01;
    }
}

class PercentageSplit implements SplitStrategy {
    private Map<User, Double> percentages;

    public PercentageSplit(Map<User, Double> percentages) {
        this.percentages = percentages;
    }

    @Override
    public Map<User, Double> calculateShares(double total, List<User> participants) {
        Map<User, Double> shares = new LinkedHashMap<>();
        for (Map.Entry<User, Double> entry : percentages.entrySet()) {
            shares.put(entry.getKey(), Math.round(total * entry.getValue()) / 100.0);
        }
        return shares;
    }

    @Override
    public boolean validate(double total, List<User> participants) {
        double sum = percentages.values().stream().mapToDouble(Double::doubleValue).sum();
        return Math.abs(sum - 100.0) < 0.01;
    }
}

// ═══════════════════════════════════════════════════════════════
// EXPENSE
// ═══════════════════════════════════════════════════════════════
class Expense {
    private final User paidBy;
    private final double amount;
    private final SplitStrategy splitStrategy;
    private final List<User> participants;
    private final String description;
    private final Map<User, Double> shares;

    public Expense(User paidBy, double amount, SplitStrategy strategy,
                   List<User> participants, String description) {
        this.paidBy = paidBy;
        this.amount = amount;
        this.splitStrategy = strategy;
        this.participants = participants;
        this.description = description;

        if (!strategy.validate(amount, participants)) {
            throw new IllegalArgumentException("Invalid split for expense: " + description);
        }

        this.shares = strategy.calculateShares(amount, participants);
    }

    public User getPaidBy()              { return paidBy; }
    public double getAmount()            { return amount; }
    public Map<User, Double> getShares() { return shares; }
    public String getDescription()       { return description; }
}

// ═══════════════════════════════════════════════════════════════
// EXPENSE MANAGER (Singleton)
// ═══════════════════════════════════════════════════════════════
class ExpenseManager {
    private static ExpenseManager instance;

    private final Map<String, User> users = new LinkedHashMap<>();
    private final List<Expense> expenses = new ArrayList<>();

    // balanceSheet[A][B] > 0 means A owes B that amount
    private final Map<String, Map<String, Double>> balanceSheet = new HashMap<>();

    private ExpenseManager() {}

    public static ExpenseManager getInstance() {
        if (instance == null) instance = new ExpenseManager();
        return instance;
    }

    public void addUser(User user) {
        users.put(user.getUserId(), user);
        balanceSheet.put(user.getUserId(), new HashMap<>());
    }

    public void addExpense(Expense expense) {
        expenses.add(expense);
        User payer = expense.getPaidBy();

        System.out.println("  " + payer.getName() + " paid $"
                + String.format("%.2f", expense.getAmount())
                + " for \"" + expense.getDescription() + "\"");

        for (Map.Entry<User, Double> entry : expense.getShares().entrySet()) {
            User participant = entry.getKey();
            double share = entry.getValue();

            if (participant.equals(payer)) continue;  // skip self

            // participant owes payer
            updateBalance(participant.getUserId(), payer.getUserId(), share);

            System.out.println("    " + participant.getName() + " owes "
                    + payer.getName() + ": $" + String.format("%.2f", share));
        }
    }

    private void updateBalance(String debtor, String creditor, double amount) {
        // Check if creditor already owes debtor (net off)
        Map<String, Double> creditorBalances = balanceSheet.get(creditor);
        double reverseDebt = creditorBalances.getOrDefault(debtor, 0.0);

        if (reverseDebt >= amount) {
            // Creditor owed debtor more — reduce that
            creditorBalances.put(debtor, reverseDebt - amount);
        } else {
            // Net the difference
            creditorBalances.put(debtor, 0.0);
            Map<String, Double> debtorBalances = balanceSheet.get(debtor);
            double existingDebt = debtorBalances.getOrDefault(creditor, 0.0);
            debtorBalances.put(creditor, existingDebt + amount - reverseDebt);
        }
    }

    public void showBalance(User user) {
        String uid = user.getUserId();
        Map<String, Double> owes = balanceSheet.get(uid);
        boolean hasDebt = false;

        for (Map.Entry<String, Double> entry : owes.entrySet()) {
            if (entry.getValue() > 0.01) {
                User creditor = users.get(entry.getKey());
                System.out.println("  " + user.getName() + " owes "
                        + creditor.getName() + ": $" + String.format("%.2f", entry.getValue()));
                hasDebt = true;
            }
        }

        // Also check who owes this user
        for (Map.Entry<String, Map<String, Double>> entry : balanceSheet.entrySet()) {
            if (!entry.getKey().equals(uid)) {
                double owed = entry.getValue().getOrDefault(uid, 0.0);
                if (owed > 0.01) {
                    User debtor = users.get(entry.getKey());
                    System.out.println("  " + debtor.getName() + " owes "
                            + user.getName() + ": $" + String.format("%.2f", owed));
                    hasDebt = true;
                }
            }
        }

        if (!hasDebt) {
            System.out.println("  " + user.getName() + " is all settled up!");
        }
    }

    public void showBalances() {
        boolean any = false;
        for (Map.Entry<String, Map<String, Double>> entry : balanceSheet.entrySet()) {
            for (Map.Entry<String, Double> debt : entry.getValue().entrySet()) {
                if (debt.getValue() > 0.01) {
                    User debtor = users.get(entry.getKey());
                    User creditor = users.get(debt.getKey());
                    System.out.println("  " + debtor.getName() + " owes "
                            + creditor.getName() + ": $" + String.format("%.2f", debt.getValue()));
                    any = true;
                }
            }
        }
        if (!any) System.out.println("  Everyone is settled up!");
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   ExpenseManager (Singleton)
 *     ├── Map<userId, User>
 *     ├── List<Expense>
 *     └── Map<userId, Map<userId, Double>> balanceSheet
 *
 *   Expense
 *     ├── User paidBy
 *     ├── double amount
 *     ├── SplitStrategy (interface)
 *     │     ├── EqualSplit
 *     │     ├── ExactSplit
 *     │     └── PercentageSplit
 *     └── Map<User, Double> shares
 *
 * PATTERNS USED:
 *   ✦ Strategy — split calculation
 *   ✦ Singleton — expense manager
 *
 * EXTENSIONS TO DISCUSS IN INTERVIEW:
 *   - Group expenses
 *   - Simplify debts (minimum transactions via graph algorithm)
 *   - Expense categories and reports
 *   - Currency conversion
 *
 * COMPILE & RUN:
 *   javac Splitwise.java && java Splitwise
 */
