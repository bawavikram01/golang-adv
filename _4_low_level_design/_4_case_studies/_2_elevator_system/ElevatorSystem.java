/*
 * =============================================================
 * LLD CASE STUDY 2: ELEVATOR SYSTEM
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Multiple elevators in a building
 *   - Handle floor requests (external) and destination requests (internal)
 *   - Optimal elevator assignment (nearest elevator strategy)
 *   - State management: IDLE, MOVING_UP, MOVING_DOWN
 *
 * DESIGN PATTERNS USED:
 *   - State (elevator states)
 *   - Strategy (elevator selection algorithm)
 *   - Observer (floor request dispatching)
 *   - Singleton (elevator controller)
 */

import java.util.*;

public class ElevatorSystem {

    public static void main(String[] args) {

        ElevatorController controller = new ElevatorController(3, 10);
        // 3 elevators, 10 floors

        System.out.println("=== ELEVATOR SYSTEM ===");
        controller.displayStatus();

        // External requests (person on a floor presses up/down)
        System.out.println("\n=== EXTERNAL REQUESTS ===");
        controller.requestElevator(3, Direction.UP);
        controller.requestElevator(7, Direction.DOWN);
        controller.requestElevator(1, Direction.UP);

        controller.displayStatus();

        // Process all requests (simulate movement)
        System.out.println("\n=== PROCESSING REQUESTS ===");
        controller.step();  // one step of simulation
        controller.step();
        controller.step();

        controller.displayStatus();
    }
}

// ═══════════════════════════════════════════════════════════════
// ENUMS
// ═══════════════════════════════════════════════════════════════
enum Direction {
    UP, DOWN, IDLE
}

enum DoorState {
    OPEN, CLOSED
}

// ═══════════════════════════════════════════════════════════════
// ELEVATOR
// ═══════════════════════════════════════════════════════════════
class Elevator {
    private final int id;
    private int currentFloor;
    private Direction direction;
    private DoorState doorState;
    private final TreeSet<Integer> upStops;    // sorted ascending
    private final TreeSet<Integer> downStops;  // sorted descending
    private final int maxFloor;

    public Elevator(int id, int maxFloor) {
        this.id = id;
        this.currentFloor = 0;  // ground floor
        this.direction = Direction.IDLE;
        this.doorState = DoorState.CLOSED;
        this.upStops = new TreeSet<>();
        this.downStops = new TreeSet<>(Collections.reverseOrder());
        this.maxFloor = maxFloor;
    }

    public void addStop(int floor) {
        if (floor < 0 || floor > maxFloor) return;

        if (floor > currentFloor) {
            upStops.add(floor);
        } else if (floor < currentFloor) {
            downStops.add(floor);
        }
        // If floor == currentFloor, open doors immediately
        updateDirection();
    }

    public void step() {
        if (direction == Direction.UP && !upStops.isEmpty()) {
            currentFloor++;
            if (upStops.contains(currentFloor)) {
                upStops.remove(currentFloor);
                System.out.println("  Elevator " + id + ": Stopped at floor " + currentFloor + " (going UP)");
                openDoors();
                closeDoors();
            }
        } else if (direction == Direction.DOWN && !downStops.isEmpty()) {
            currentFloor--;
            if (downStops.contains(currentFloor)) {
                downStops.remove(currentFloor);
                System.out.println("  Elevator " + id + ": Stopped at floor " + currentFloor + " (going DOWN)");
                openDoors();
                closeDoors();
            }
        }
        updateDirection();
    }

    private void updateDirection() {
        if (!upStops.isEmpty()) {
            direction = Direction.UP;
        } else if (!downStops.isEmpty()) {
            direction = Direction.DOWN;
        } else {
            direction = Direction.IDLE;
        }
    }

    private void openDoors() {
        doorState = DoorState.OPEN;
    }

    private void closeDoors() {
        doorState = DoorState.CLOSED;
    }

    public int getId() { return id; }
    public int getCurrentFloor() { return currentFloor; }
    public Direction getDirection() { return direction; }
    public boolean isIdle() { return direction == Direction.IDLE; }

    public int distanceTo(int floor) {
        return Math.abs(currentFloor - floor);
    }

    public int getPendingStops() {
        return upStops.size() + downStops.size();
    }

    @Override
    public String toString() {
        return String.format("Elevator %d: Floor=%d, Dir=%s, Stops=%d",
                id, currentFloor, direction, getPendingStops());
    }
}

// ═══════════════════════════════════════════════════════════════
// ELEVATOR SELECTION STRATEGY
// ═══════════════════════════════════════════════════════════════
interface ElevatorSelectionStrategy {
    Elevator selectElevator(List<Elevator> elevators, int requestFloor, Direction direction);
}

// Nearest elevator that is idle or moving in the same direction
class NearestElevatorStrategy implements ElevatorSelectionStrategy {
    @Override
    public Elevator selectElevator(List<Elevator> elevators, int requestFloor, Direction dir) {
        Elevator best = null;
        int bestScore = Integer.MAX_VALUE;

        for (Elevator e : elevators) {
            int distance = e.distanceTo(requestFloor);
            int score;

            if (e.isIdle()) {
                score = distance;  // idle → just distance
            } else if (e.getDirection() == dir) {
                // Moving same direction AND hasn't passed yet
                if ((dir == Direction.UP && e.getCurrentFloor() <= requestFloor)
                        || (dir == Direction.DOWN && e.getCurrentFloor() >= requestFloor)) {
                    score = distance;  // on the way
                } else {
                    score = distance + 100;  // passed already
                }
            } else {
                score = distance + 200;  // moving opposite direction
            }

            if (score < bestScore) {
                bestScore = score;
                best = e;
            }
        }
        return best;
    }
}

// ═══════════════════════════════════════════════════════════════
// ELEVATOR CONTROLLER
// ═══════════════════════════════════════════════════════════════
class ElevatorController {
    private final List<Elevator> elevators;
    private final ElevatorSelectionStrategy strategy;

    public ElevatorController(int numElevators, int maxFloor) {
        this.elevators = new ArrayList<>();
        for (int i = 1; i <= numElevators; i++) {
            elevators.add(new Elevator(i, maxFloor));
        }
        this.strategy = new NearestElevatorStrategy();
    }

    public void requestElevator(int floor, Direction direction) {
        Elevator selected = strategy.selectElevator(elevators, floor, direction);
        if (selected != null) {
            selected.addStop(floor);
            System.out.println("  Request: Floor " + floor + " " + direction
                    + " → Assigned to Elevator " + selected.getId());
        }
    }

    public void step() {
        for (Elevator e : elevators) {
            e.step();
        }
    }

    public void displayStatus() {
        System.out.println("  ┌─── ELEVATOR STATUS ─────────────────┐");
        for (Elevator e : elevators) {
            System.out.println("  │ " + e);
        }
        System.out.println("  └───────────────────────────────────────┘");
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   ElevatorController
 *     ├── List<Elevator>
 *     │     ├── currentFloor, direction, doorState
 *     │     ├── TreeSet<Integer> upStops
 *     │     └── TreeSet<Integer> downStops
 *     └── ElevatorSelectionStrategy (interface)
 *           └── NearestElevatorStrategy
 *
 * PATTERNS USED:
 *   ✦ Strategy — elevator selection algorithm
 *   ✦ State-like — elevator direction management
 *   ✦ TreeSet — sorted stop management (SCAN algorithm)
 *
 * COMPILE & RUN:
 *   javac ElevatorSystem.java && java ElevatorSystem
 */
