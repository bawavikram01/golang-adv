/*
 * =============================================================
 * BEHAVIORAL PATTERN 1: STRATEGY
 * =============================================================
 *
 * INTENT: Define a family of algorithms, encapsulate each one,
 *         and make them interchangeable at RUNTIME.
 *
 * ANALOGY: Google Maps — same destination, different strategies:
 *          driving, walking, cycling, public transit.
 *
 * USE WHEN:
 *   - Multiple algorithms for the same task
 *   - You want to switch algorithms at runtime
 *   - You see if/else or switch for selecting behavior
 *
 * THIS IS THE MOST IMPORTANT DESIGN PATTERN. Master it.
 */

import java.util.List;
import java.util.ArrayList;
import java.util.Collections;

public class StrategyPattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Sorting Strategy
        // ═══════════════════════════════════════════════════════
        System.out.println("=== SORTING STRATEGY ===");

        int[] data = {5, 2, 8, 1, 9, 3};

        Sorter sorter = new Sorter(new BubbleSortStrategy());
        sorter.sort(data.clone());

        sorter.setStrategy(new QuickSortStrategy());
        sorter.sort(data.clone());

        sorter.setStrategy(new MergeSortStrategy());
        sorter.sort(data.clone());

        // ═══════════════════════════════════════════════════════
        // Payment Strategy
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== PAYMENT STRATEGY ===");

        ShoppingCart cart = new ShoppingCart();
        cart.addItem("Laptop", 999.99);
        cart.addItem("Mouse", 29.99);

        // Pay with credit card
        cart.pay(new CreditCardPayment("4111-1111-1111-1111", "Alice"));
        System.out.println();

        // Same cart, different payment — just swap strategy!
        cart.pay(new PayPalPayment("alice@email.com"));
        System.out.println();

        cart.pay(new CryptoPayment("0xABC123DEF456"));

        // ═══════════════════════════════════════════════════════
        // Compression Strategy
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== COMPRESSION STRATEGY ===");

        FileCompressor compressor = new FileCompressor();

        compressor.setStrategy(new ZipCompression());
        compressor.compress("document.pdf");

        compressor.setStrategy(new RarCompression());
        compressor.compress("images.folder");

        compressor.setStrategy(new GzipCompression());
        compressor.compress("logs.txt");
    }
}

// ═══════════════════════════════════════════════════════════════
// SORTING STRATEGY
// ═══════════════════════════════════════════════════════════════
interface SortStrategy {
    void sort(int[] array);
    String getName();
}

class BubbleSortStrategy implements SortStrategy {
    @Override
    public void sort(int[] array) {
        for (int i = 0; i < array.length; i++)
            for (int j = 0; j < array.length - i - 1; j++)
                if (array[j] > array[j + 1]) {
                    int tmp = array[j];
                    array[j] = array[j + 1];
                    array[j + 1] = tmp;
                }
    }
    @Override public String getName() { return "Bubble Sort"; }
}

class QuickSortStrategy implements SortStrategy {
    @Override
    public void sort(int[] array) {
        quickSort(array, 0, array.length - 1);
    }
    private void quickSort(int[] a, int lo, int hi) {
        if (lo < hi) {
            int p = partition(a, lo, hi);
            quickSort(a, lo, p - 1);
            quickSort(a, p + 1, hi);
        }
    }
    private int partition(int[] a, int lo, int hi) {
        int pivot = a[hi], i = lo;
        for (int j = lo; j < hi; j++)
            if (a[j] < pivot) { int t = a[i]; a[i] = a[j]; a[j] = t; i++; }
        int t = a[i]; a[i] = a[hi]; a[hi] = t;
        return i;
    }
    @Override public String getName() { return "Quick Sort"; }
}

class MergeSortStrategy implements SortStrategy {
    @Override
    public void sort(int[] array) {
        java.util.Arrays.sort(array);  // simplified — uses mergesort internally
    }
    @Override public String getName() { return "Merge Sort"; }
}

// Context — uses a strategy
class Sorter {
    private SortStrategy strategy;

    public Sorter(SortStrategy strategy) {
        this.strategy = strategy;
    }

    public void setStrategy(SortStrategy strategy) {
        this.strategy = strategy;
    }

    public void sort(int[] array) {
        System.out.print("  " + strategy.getName() + ": ");
        strategy.sort(array);
        System.out.println(java.util.Arrays.toString(array));
    }
}

// ═══════════════════════════════════════════════════════════════
// PAYMENT STRATEGY
// ═══════════════════════════════════════════════════════════════
interface PaymentStrategy {
    void pay(double amount);
}

class CreditCardPayment implements PaymentStrategy {
    private String cardNumber;
    private String name;

    public CreditCardPayment(String cardNumber, String name) {
        this.cardNumber = cardNumber;
        this.name = name;
    }

    @Override
    public void pay(double amount) {
        String masked = "****" + cardNumber.substring(cardNumber.length() - 4);
        System.out.println("  💳 Paid $" + String.format("%.2f", amount) + " with card " + masked + " (" + name + ")");
    }
}

class PayPalPayment implements PaymentStrategy {
    private String email;

    public PayPalPayment(String email) { this.email = email; }

    @Override
    public void pay(double amount) {
        System.out.println("  🅿️ Paid $" + String.format("%.2f", amount) + " via PayPal (" + email + ")");
    }
}

class CryptoPayment implements PaymentStrategy {
    private String walletAddress;

    public CryptoPayment(String walletAddress) { this.walletAddress = walletAddress; }

    @Override
    public void pay(double amount) {
        System.out.println("  ₿ Paid $" + String.format("%.2f", amount) + " via crypto (" + walletAddress + ")");
    }
}

class ShoppingCart {
    private List<String> items = new ArrayList<>();
    private double total = 0;

    public void addItem(String name, double price) {
        items.add(name);
        total += price;
    }

    public void pay(PaymentStrategy strategy) {
        System.out.println("  Cart: " + items + " = $" + String.format("%.2f", total));
        strategy.pay(total);
    }
}

// ═══════════════════════════════════════════════════════════════
// COMPRESSION STRATEGY
// ═══════════════════════════════════════════════════════════════
interface CompressionStrategy {
    void compress(String file);
}

class ZipCompression implements CompressionStrategy {
    @Override public void compress(String file) {
        System.out.println("  📦 Compressed " + file + " → " + file + ".zip");
    }
}

class RarCompression implements CompressionStrategy {
    @Override public void compress(String file) {
        System.out.println("  📦 Compressed " + file + " → " + file + ".rar");
    }
}

class GzipCompression implements CompressionStrategy {
    @Override public void compress(String file) {
        System.out.println("  📦 Compressed " + file + " → " + file + ".gz");
    }
}

class FileCompressor {
    private CompressionStrategy strategy;

    public void setStrategy(CompressionStrategy strategy) {
        this.strategy = strategy;
    }

    public void compress(String file) {
        strategy.compress(file);
    }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ Strategy = interface + multiple implementations + context.
 * ✦ The CONTEXT holds a reference to a strategy and delegates.
 * ✦ Strategies are SWAPPABLE at runtime via setter.
 *
 * ✦ Eliminates if/else chains:
 *   BAD:  if (type == "card") ... else if (type == "paypal") ...
 *   GOOD: strategy.pay(amount);
 *
 * ✦ Perfect example of:
 *   - OCP: add new strategies without changing existing code
 *   - DIP: context depends on interface, not implementations
 *   - Composition over inheritance
 *
 * COMPILE & RUN:
 *   javac StrategyPattern.java && java StrategyPattern
 */
