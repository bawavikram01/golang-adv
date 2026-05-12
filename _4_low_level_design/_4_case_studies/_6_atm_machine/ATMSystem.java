/*
 * =============================================================
 * LLD CASE STUDY 6: ATM MACHINE
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Card insertion and PIN verification
 *   - Check balance, withdraw, deposit
 *   - Cash dispensed using chain of responsibility (denominations)
 *   - State management (idle, card inserted, authenticated, etc.)
 *   - Transaction logging
 *
 * DESIGN PATTERNS USED:
 *   - State (ATM states)
 *   - Chain of Responsibility (cash dispensing)
 *   - Strategy (transaction types)
 */

import java.util.*;

public class ATMSystem {

    public static void main(String[] args) {

        // Setup ATM with cash
        ATM atm = new ATM("ATM-001");
        atm.loadCash(2000, 10);  // 10 × ₹2000
        atm.loadCash(500, 20);   // 20 × ₹500
        atm.loadCash(200, 30);   // 30 × ₹200
        atm.loadCash(100, 50);   // 50 × ₹100

        // Setup bank accounts
        BankAccount aliceAccount = new BankAccount("ACC001", "Alice", 50000, "1234");
        BankAccount bobAccount = new BankAccount("ACC002", "Bob", 10000, "5678");

        Card aliceCard = new Card("4111-1111-1111-1111", aliceAccount);
        Card bobCard = new Card("4222-2222-2222-2222", bobAccount);

        // ═══════════════════════════════════════════════════════
        // Scenario 1: Successful withdrawal
        // ═══════════════════════════════════════════════════════
        System.out.println("=== SCENARIO 1: Successful Withdrawal ===");
        atm.insertCard(aliceCard);
        atm.enterPin("1234");
        atm.checkBalance();
        atm.withdraw(4700);
        atm.ejectCard();

        // ═══════════════════════════════════════════════════════
        // Scenario 2: Wrong PIN
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== SCENARIO 2: Wrong PIN ===");
        atm.insertCard(bobCard);
        atm.enterPin("0000");  // wrong!
        atm.ejectCard();

        // ═══════════════════════════════════════════════════════
        // Scenario 3: Insufficient funds
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== SCENARIO 3: Insufficient Funds ===");
        atm.insertCard(bobCard);
        atm.enterPin("5678");
        atm.withdraw(99999);
        atm.ejectCard();
    }
}

// ═══════════════════════════════════════════════════════════════
// BANK ACCOUNT
// ═══════════════════════════════════════════════════════════════
class BankAccount {
    private final String accountId;
    private final String holderName;
    private double balance;
    private final String pin;

    public BankAccount(String id, String name, double balance, String pin) {
        this.accountId = id;
        this.holderName = name;
        this.balance = balance;
        this.pin = pin;
    }

    public boolean validatePin(String input) { return pin.equals(input); }
    public double getBalance() { return balance; }

    public boolean debit(double amount) {
        if (amount > balance) return false;
        balance -= amount;
        return true;
    }

    public void credit(double amount) { balance += amount; }
    public String getHolderName() { return holderName; }
}

// ═══════════════════════════════════════════════════════════════
// CARD
// ═══════════════════════════════════════════════════════════════
class Card {
    private final String cardNumber;
    private final BankAccount linkedAccount;

    public Card(String number, BankAccount account) {
        this.cardNumber = number;
        this.linkedAccount = account;
    }

    public String getCardNumber() { return cardNumber; }
    public BankAccount getAccount() { return linkedAccount; }
    public String getMaskedNumber() {
        return "****-****-****-" + cardNumber.substring(cardNumber.length() - 4);
    }
}

// ═══════════════════════════════════════════════════════════════
// CASH DISPENSER (Chain of Responsibility)
// ═══════════════════════════════════════════════════════════════
class CashDispenser {
    private final TreeMap<Integer, Integer> cassettes = new TreeMap<>(Collections.reverseOrder());
    // denomination → count, reverse order so largest first

    public void loadCash(int denomination, int count) {
        cassettes.merge(denomination, count, Integer::sum);
    }

    public boolean canDispense(int amount) {
        int remaining = amount;
        for (Map.Entry<Integer, Integer> entry : cassettes.entrySet()) {
            int denom = entry.getKey();
            int available = entry.getValue();
            int needed = Math.min(remaining / denom, available);
            remaining -= needed * denom;
        }
        return remaining == 0;
    }

    public Map<Integer, Integer> dispense(int amount) {
        if (!canDispense(amount)) return null;

        Map<Integer, Integer> dispensed = new LinkedHashMap<>();
        int remaining = amount;

        for (Map.Entry<Integer, Integer> entry : cassettes.entrySet()) {
            int denom = entry.getKey();
            int available = entry.getValue();
            int needed = Math.min(remaining / denom, available);

            if (needed > 0) {
                dispensed.put(denom, needed);
                cassettes.put(denom, available - needed);
                remaining -= needed * denom;
            }
        }

        return dispensed;
    }

    public int getTotalCash() {
        return cassettes.entrySet().stream()
                .mapToInt(e -> e.getKey() * e.getValue())
                .sum();
    }
}

// ═══════════════════════════════════════════════════════════════
// ATM STATES
// ═══════════════════════════════════════════════════════════════
interface ATMState {
    void insertCard(ATM atm, Card card);
    void enterPin(ATM atm, String pin);
    void checkBalance(ATM atm);
    void withdraw(ATM atm, int amount);
    void ejectCard(ATM atm);
}

class IdleState implements ATMState {
    @Override
    public void insertCard(ATM atm, Card card) {
        System.out.println("  💳 Card inserted: " + card.getMaskedNumber());
        atm.setCurrentCard(card);
        atm.setState(new CardInsertedState());
    }
    @Override public void enterPin(ATM atm, String pin)  { System.out.println("  ✗ Insert card first."); }
    @Override public void checkBalance(ATM atm)          { System.out.println("  ✗ Insert card first."); }
    @Override public void withdraw(ATM atm, int amount)  { System.out.println("  ✗ Insert card first."); }
    @Override public void ejectCard(ATM atm)             { System.out.println("  ✗ No card inserted."); }
}

class CardInsertedState implements ATMState {
    @Override public void insertCard(ATM atm, Card card) { System.out.println("  ✗ Card already inserted."); }

    @Override
    public void enterPin(ATM atm, String pin) {
        if (atm.getCurrentCard().getAccount().validatePin(pin)) {
            System.out.println("  ✓ PIN verified. Welcome, "
                    + atm.getCurrentCard().getAccount().getHolderName() + "!");
            atm.setState(new AuthenticatedState());
        } else {
            System.out.println("  ✗ Wrong PIN!");
        }
    }

    @Override public void checkBalance(ATM atm)         { System.out.println("  ✗ Enter PIN first."); }
    @Override public void withdraw(ATM atm, int amount) { System.out.println("  ✗ Enter PIN first."); }

    @Override
    public void ejectCard(ATM atm) {
        System.out.println("  💳 Card ejected.");
        atm.setCurrentCard(null);
        atm.setState(new IdleState());
    }
}

class AuthenticatedState implements ATMState {
    @Override public void insertCard(ATM atm, Card card) { System.out.println("  ✗ Card already inserted."); }
    @Override public void enterPin(ATM atm, String pin)  { System.out.println("  ✗ Already authenticated."); }

    @Override
    public void checkBalance(ATM atm) {
        double balance = atm.getCurrentCard().getAccount().getBalance();
        System.out.println("  💰 Balance: ₹" + String.format("%.2f", balance));
    }

    @Override
    public void withdraw(ATM atm, int amount) {
        BankAccount account = atm.getCurrentCard().getAccount();

        if (amount <= 0 || amount % 100 != 0) {
            System.out.println("  ✗ Amount must be positive and multiple of 100.");
            return;
        }

        if (amount > account.getBalance()) {
            System.out.println("  ✗ Insufficient balance! Available: ₹" + account.getBalance());
            return;
        }

        Map<Integer, Integer> cash = atm.getDispenser().dispense(amount);
        if (cash == null) {
            System.out.println("  ✗ ATM cannot dispense this amount.");
            return;
        }

        account.debit(amount);
        System.out.println("  ✓ Dispensing ₹" + amount + ":");
        for (Map.Entry<Integer, Integer> entry : cash.entrySet()) {
            System.out.println("    💵 " + entry.getValue() + " × ₹" + entry.getKey());
        }
        System.out.println("  Remaining balance: ₹" + String.format("%.2f", account.getBalance()));
    }

    @Override
    public void ejectCard(ATM atm) {
        System.out.println("  💳 Card ejected. Thank you!");
        atm.setCurrentCard(null);
        atm.setState(new IdleState());
    }
}

// ═══════════════════════════════════════════════════════════════
// ATM — Context
// ═══════════════════════════════════════════════════════════════
class ATM {
    private final String atmId;
    private ATMState currentState;
    private Card currentCard;
    private final CashDispenser dispenser;

    public ATM(String atmId) {
        this.atmId = atmId;
        this.currentState = new IdleState();
        this.dispenser = new CashDispenser();
        System.out.println("ATM " + atmId + " initialized.");
    }

    public void loadCash(int denomination, int count) {
        dispenser.loadCash(denomination, count);
    }

    // Delegate all actions to current state
    public void insertCard(Card card)   { currentState.insertCard(this, card); }
    public void enterPin(String pin)    { currentState.enterPin(this, pin); }
    public void checkBalance()          { currentState.checkBalance(this); }
    public void withdraw(int amount)    { currentState.withdraw(this, amount); }
    public void ejectCard()             { currentState.ejectCard(this); }

    void setState(ATMState state)       { this.currentState = state; }
    void setCurrentCard(Card card)      { this.currentCard = card; }
    Card getCurrentCard()               { return currentCard; }
    CashDispenser getDispenser()        { return dispenser; }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   ATM
 *     ├── ATMState (interface)
 *     │     ├── IdleState
 *     │     ├── CardInsertedState
 *     │     └── AuthenticatedState
 *     ├── CashDispenser (Chain of Responsibility)
 *     │     └── TreeMap<denomination, count>
 *     └── Card → BankAccount
 *
 * PATTERNS USED:
 *   ✦ State — ATM behavior changes per state
 *   ✦ Chain of Responsibility — cash dispensing by denomination
 *
 * COMPILE & RUN:
 *   javac ATMSystem.java && java ATMSystem
 */
