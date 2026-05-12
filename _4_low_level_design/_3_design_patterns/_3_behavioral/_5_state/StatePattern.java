/*
 * =============================================================
 * BEHAVIORAL PATTERN 5: STATE
 * =============================================================
 *
 * INTENT: Allow an object to alter its behavior when its internal
 *         state changes. The object will appear to change its class.
 *
 * ANALOGY: A traffic light — same light, different behavior
 *          based on current state (red, yellow, green).
 *
 * USE WHEN:
 *   - Object behavior depends on state, and changes at runtime
 *   - You see large if/else or switch on state variables
 *   - State transitions have complex rules
 *
 * STATE vs STRATEGY:
 *   - Strategy: client chooses the algorithm
 *   - State: object changes behavior automatically based on context
 */

public class StatePattern {

    public static void main(String[] args) {

        // ═══════════════════════════════════════════════════════
        // Vending Machine — classic state pattern example
        // ═══════════════════════════════════════════════════════
        System.out.println("=== VENDING MACHINE ===");

        VendingMachine machine = new VendingMachine(3);  // 3 items in stock

        // Scenario 1: Normal purchase
        machine.insertCoin();
        machine.selectItem();
        machine.dispense();

        System.out.println();

        // Scenario 2: Try to select without inserting coin
        machine.selectItem();

        System.out.println();

        // Scenario 3: Insert coin, then eject
        machine.insertCoin();
        machine.ejectCoin();

        System.out.println();

        // Scenario 4: Buy remaining items until sold out
        machine.insertCoin();
        machine.selectItem();
        machine.dispense();

        System.out.println();

        machine.insertCoin();
        machine.selectItem();
        machine.dispense();  // last item!

        System.out.println();

        // Scenario 5: Try to buy when sold out
        machine.insertCoin();

        // ═══════════════════════════════════════════════════════
        // Order Status — real-world state machine
        // ═══════════════════════════════════════════════════════
        System.out.println("\n=== ORDER STATE MACHINE ===");

        Order order = new Order("ORD-001");
        order.next();  // pending → confirmed
        order.next();  // confirmed → shipped
        order.next();  // shipped → delivered
        order.next();  // delivered → no further transition

        System.out.println();

        Order order2 = new Order("ORD-002");
        order2.next();  // pending → confirmed
        order2.cancel(); // confirmed → cancelled
        order2.next();   // cancelled → can't proceed
    }
}

// ═══════════════════════════════════════════════════════════════
// VENDING MACHINE STATE PATTERN
// ═══════════════════════════════════════════════════════════════

// State interface
interface VendingState {
    void insertCoin(VendingMachine machine);
    void ejectCoin(VendingMachine machine);
    void selectItem(VendingMachine machine);
    void dispense(VendingMachine machine);
}

class VendingMachine {
    // All possible states
    private VendingState idleState;
    private VendingState hasCoinState;
    private VendingState dispensingState;
    private VendingState soldOutState;

    private VendingState currentState;
    private int itemCount;

    public VendingMachine(int itemCount) {
        idleState = new IdleState();
        hasCoinState = new HasCoinState();
        dispensingState = new DispensingState();
        soldOutState = new SoldOutState();

        this.itemCount = itemCount;
        this.currentState = itemCount > 0 ? idleState : soldOutState;
        System.out.println("  Machine ready with " + itemCount + " items.");
    }

    // Delegate ALL actions to current state
    public void insertCoin() { currentState.insertCoin(this); }
    public void ejectCoin()  { currentState.ejectCoin(this); }
    public void selectItem() { currentState.selectItem(this); }
    public void dispense()   { currentState.dispense(this); }

    // State transitions
    void setState(VendingState state) { this.currentState = state; }
    VendingState getIdleState()       { return idleState; }
    VendingState getHasCoinState()    { return hasCoinState; }
    VendingState getDispensingState() { return dispensingState; }
    VendingState getSoldOutState()    { return soldOutState; }

    void releaseItem() { itemCount--; }
    int getItemCount() { return itemCount; }
}

// Concrete States
class IdleState implements VendingState {
    @Override
    public void insertCoin(VendingMachine m) {
        System.out.println("  💰 Coin inserted.");
        m.setState(m.getHasCoinState());
    }

    @Override
    public void ejectCoin(VendingMachine m) {
        System.out.println("  ✗ No coin to eject.");
    }

    @Override
    public void selectItem(VendingMachine m) {
        System.out.println("  ✗ Please insert a coin first.");
    }

    @Override
    public void dispense(VendingMachine m) {
        System.out.println("  ✗ No item selected.");
    }
}

class HasCoinState implements VendingState {
    @Override
    public void insertCoin(VendingMachine m) {
        System.out.println("  ✗ Coin already inserted.");
    }

    @Override
    public void ejectCoin(VendingMachine m) {
        System.out.println("  💰 Coin ejected.");
        m.setState(m.getIdleState());
    }

    @Override
    public void selectItem(VendingMachine m) {
        System.out.println("  ✓ Item selected.");
        m.setState(m.getDispensingState());
    }

    @Override
    public void dispense(VendingMachine m) {
        System.out.println("  ✗ Select an item first.");
    }
}

class DispensingState implements VendingState {
    @Override
    public void insertCoin(VendingMachine m) {
        System.out.println("  ✗ Please wait, dispensing...");
    }

    @Override
    public void ejectCoin(VendingMachine m) {
        System.out.println("  ✗ Already dispensing, cannot eject.");
    }

    @Override
    public void selectItem(VendingMachine m) {
        System.out.println("  ✗ Already dispensing...");
    }

    @Override
    public void dispense(VendingMachine m) {
        m.releaseItem();
        System.out.println("  📦 Item dispensed! (" + m.getItemCount() + " remaining)");
        if (m.getItemCount() > 0) {
            m.setState(m.getIdleState());
        } else {
            System.out.println("  ⚠️ Machine sold out!");
            m.setState(m.getSoldOutState());
        }
    }
}

class SoldOutState implements VendingState {
    @Override public void insertCoin(VendingMachine m)  { System.out.println("  ✗ SOLD OUT. Cannot accept coins."); }
    @Override public void ejectCoin(VendingMachine m)   { System.out.println("  ✗ No coin inserted."); }
    @Override public void selectItem(VendingMachine m)  { System.out.println("  ✗ SOLD OUT."); }
    @Override public void dispense(VendingMachine m)    { System.out.println("  ✗ SOLD OUT."); }
}

// ═══════════════════════════════════════════════════════════════
// ORDER STATE MACHINE
// ═══════════════════════════════════════════════════════════════
interface OrderState {
    void next(Order order);
    void cancel(Order order);
    String getStatus();
}

class Order {
    private String orderId;
    private OrderState state;

    public Order(String orderId) {
        this.orderId = orderId;
        this.state = new PendingState();
        System.out.println("  Order " + orderId + " created → " + state.getStatus());
    }

    public void next() {
        state.next(this);
    }

    public void cancel() {
        state.cancel(this);
    }

    void setState(OrderState state) {
        this.state = state;
        System.out.println("  Order " + orderId + " → " + state.getStatus());
    }
}

class PendingState implements OrderState {
    @Override public void next(Order o)   { o.setState(new ConfirmedState()); }
    @Override public void cancel(Order o) { o.setState(new CancelledState()); }
    @Override public String getStatus()   { return "PENDING"; }
}

class ConfirmedState implements OrderState {
    @Override public void next(Order o)   { o.setState(new ShippedState()); }
    @Override public void cancel(Order o) { o.setState(new CancelledState()); }
    @Override public String getStatus()   { return "CONFIRMED"; }
}

class ShippedState implements OrderState {
    @Override public void next(Order o)   { o.setState(new DeliveredState()); }
    @Override public void cancel(Order o) { System.out.println("  ✗ Cannot cancel — already shipped!"); }
    @Override public String getStatus()   { return "SHIPPED"; }
}

class DeliveredState implements OrderState {
    @Override public void next(Order o)   { System.out.println("  ✗ Already delivered. No further transitions."); }
    @Override public void cancel(Order o) { System.out.println("  ✗ Cannot cancel — already delivered!"); }
    @Override public String getStatus()   { return "DELIVERED"; }
}

class CancelledState implements OrderState {
    @Override public void next(Order o)   { System.out.println("  ✗ Order is cancelled. Cannot proceed."); }
    @Override public void cancel(Order o) { System.out.println("  ✗ Already cancelled."); }
    @Override public String getStatus()   { return "CANCELLED"; }
}

/*
 * KEY TAKEAWAYS:
 * ─────────────────────────────────────────────────────────────
 * ✦ State pattern replaces large if/else or switch on state variables.
 * ✦ Each state is a class implementing the State interface.
 * ✦ Context (VendingMachine) delegates behavior to current state.
 * ✦ State transitions are handled BY the states themselves.
 * ✦ Adding new states doesn't modify existing state classes (OCP).
 *
 * ✦ Very common in interviews:
 *   - Vending Machine
 *   - ATM Machine
 *   - Order/Workflow status
 *   - Traffic Light
 *   - Connection state (TCP)
 *
 * COMPILE & RUN:
 *   javac StatePattern.java && java StatePattern
 */
