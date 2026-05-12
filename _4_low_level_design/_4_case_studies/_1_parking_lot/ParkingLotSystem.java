/*
 * =============================================================
 * LLD CASE STUDY 1: PARKING LOT SYSTEM
 * =============================================================
 *
 * REQUIREMENTS:
 *   - Multiple floors, each with parking spots
 *   - Different vehicle types: Bike, Car, Truck
 *   - Different spot sizes: Small, Medium, Large
 *   - Assign nearest available spot
 *   - Track entry/exit, calculate fee
 *
 * DESIGN PATTERNS USED:
 *   - Strategy (pricing)
 *   - Factory (ticket creation)
 *   - Singleton (parking lot instance)
 *   - Enum for types
 *
 * THIS IS THE #1 MOST ASKED LLD INTERVIEW QUESTION.
 */

import java.time.LocalDateTime;
import java.time.Duration;
import java.util.*;

public class ParkingLotSystem {

    public static void main(String[] args) {

        // Initialize parking lot
        ParkingLot lot = ParkingLot.getInstance();
        lot.addFloor(new ParkingFloor("F1", 3, 5, 2));  // 3 small, 5 medium, 2 large
        lot.addFloor(new ParkingFloor("F2", 2, 4, 3));

        System.out.println("=== PARKING LOT STATUS ===");
        lot.displayAvailability();

        // Park vehicles
        System.out.println("\n=== PARKING VEHICLES ===");
        Vehicle bike = new Vehicle("KA-01-1234", VehicleType.BIKE);
        Vehicle car1 = new Vehicle("KA-02-5678", VehicleType.CAR);
        Vehicle car2 = new Vehicle("KA-03-9999", VehicleType.CAR);
        Vehicle truck = new Vehicle("KA-04-0000", VehicleType.TRUCK);

        Ticket t1 = lot.parkVehicle(bike);
        Ticket t2 = lot.parkVehicle(car1);
        Ticket t3 = lot.parkVehicle(car2);
        Ticket t4 = lot.parkVehicle(truck);

        lot.displayAvailability();

        // Unpark and calculate fee
        System.out.println("\n=== UNPARKING VEHICLES ===");
        if (t2 != null) {
            double fee = lot.unparkVehicle(t2);
            System.out.println("  Fee for " + car1.getLicensePlate() + ": $" + String.format("%.2f", fee));
        }

        lot.displayAvailability();
    }
}

// ═══════════════════════════════════════════════════════════════
// ENUMS
// ═══════════════════════════════════════════════════════════════
enum VehicleType {
    BIKE, CAR, TRUCK
}

enum SpotSize {
    SMALL, MEDIUM, LARGE
}

// ═══════════════════════════════════════════════════════════════
// VEHICLE
// ═══════════════════════════════════════════════════════════════
class Vehicle {
    private final String licensePlate;
    private final VehicleType type;

    public Vehicle(String licensePlate, VehicleType type) {
        this.licensePlate = licensePlate;
        this.type = type;
    }

    public String getLicensePlate() { return licensePlate; }
    public VehicleType getType() { return type; }

    public SpotSize getRequiredSpotSize() {
        return switch (type) {
            case BIKE  -> SpotSize.SMALL;
            case CAR   -> SpotSize.MEDIUM;
            case TRUCK -> SpotSize.LARGE;
        };
    }
}

// ═══════════════════════════════════════════════════════════════
// PARKING SPOT
// ═══════════════════════════════════════════════════════════════
class ParkingSpot {
    private final String spotId;
    private final SpotSize size;
    private Vehicle parkedVehicle;

    public ParkingSpot(String spotId, SpotSize size) {
        this.spotId = spotId;
        this.size = size;
    }

    public boolean isAvailable() { return parkedVehicle == null; }

    public boolean canFit(Vehicle vehicle) {
        return isAvailable() && size.ordinal() >= vehicle.getRequiredSpotSize().ordinal();
    }

    public void park(Vehicle vehicle) {
        this.parkedVehicle = vehicle;
    }

    public Vehicle unpark() {
        Vehicle v = parkedVehicle;
        parkedVehicle = null;
        return v;
    }

    public String getSpotId() { return spotId; }
    public SpotSize getSize() { return size; }
    public Vehicle getParkedVehicle() { return parkedVehicle; }
}

// ═══════════════════════════════════════════════════════════════
// PARKING FLOOR
// ═══════════════════════════════════════════════════════════════
class ParkingFloor {
    private final String floorId;
    private final List<ParkingSpot> spots = new ArrayList<>();

    public ParkingFloor(String floorId, int small, int medium, int large) {
        this.floorId = floorId;
        int id = 1;
        for (int i = 0; i < small; i++)  spots.add(new ParkingSpot(floorId + "-S" + id++, SpotSize.SMALL));
        for (int i = 0; i < medium; i++) spots.add(new ParkingSpot(floorId + "-M" + id++, SpotSize.MEDIUM));
        for (int i = 0; i < large; i++)  spots.add(new ParkingSpot(floorId + "-L" + id++, SpotSize.LARGE));
    }

    public ParkingSpot findAvailableSpot(Vehicle vehicle) {
        // Find the SMALLEST available spot that fits (best fit)
        return spots.stream()
                .filter(spot -> spot.canFit(vehicle))
                .min(Comparator.comparingInt(s -> s.getSize().ordinal()))
                .orElse(null);
    }

    public int getAvailableCount(SpotSize size) {
        return (int) spots.stream().filter(s -> s.getSize() == size && s.isAvailable()).count();
    }

    public String getFloorId() { return floorId; }
}

// ═══════════════════════════════════════════════════════════════
// TICKET
// ═══════════════════════════════════════════════════════════════
class Ticket {
    private static int counter = 0;

    private final String ticketId;
    private final Vehicle vehicle;
    private final ParkingSpot spot;
    private final LocalDateTime entryTime;
    private LocalDateTime exitTime;

    public Ticket(Vehicle vehicle, ParkingSpot spot) {
        this.ticketId = "T-" + (++counter);
        this.vehicle = vehicle;
        this.spot = spot;
        this.entryTime = LocalDateTime.now();
    }

    public void setExitTime(LocalDateTime exitTime) { this.exitTime = exitTime; }

    public String getTicketId() { return ticketId; }
    public Vehicle getVehicle() { return vehicle; }
    public ParkingSpot getSpot() { return spot; }
    public LocalDateTime getEntryTime() { return entryTime; }
    public LocalDateTime getExitTime() { return exitTime; }

    public long getDurationMinutes() {
        LocalDateTime exit = exitTime != null ? exitTime : LocalDateTime.now();
        return Math.max(1, Duration.between(entryTime, exit).toMinutes());
    }
}

// ═══════════════════════════════════════════════════════════════
// PRICING STRATEGY
// ═══════════════════════════════════════════════════════════════
interface PricingStrategy {
    double calculateFee(Ticket ticket);
}

class HourlyPricing implements PricingStrategy {
    private static final Map<VehicleType, Double> RATES = Map.of(
            VehicleType.BIKE, 1.0,
            VehicleType.CAR, 2.0,
            VehicleType.TRUCK, 3.0
    );

    @Override
    public double calculateFee(Ticket ticket) {
        double hours = Math.ceil(ticket.getDurationMinutes() / 60.0);
        return hours * RATES.getOrDefault(ticket.getVehicle().getType(), 2.0);
    }
}

// ═══════════════════════════════════════════════════════════════
// PARKING LOT (Singleton)
// ═══════════════════════════════════════════════════════════════
class ParkingLot {
    private static ParkingLot instance;

    private final List<ParkingFloor> floors = new ArrayList<>();
    private final Map<String, Ticket> activeTickets = new HashMap<>();
    private PricingStrategy pricingStrategy = new HourlyPricing();

    private ParkingLot() {}

    public static ParkingLot getInstance() {
        if (instance == null) {
            instance = new ParkingLot();
        }
        return instance;
    }

    public void addFloor(ParkingFloor floor) {
        floors.add(floor);
    }

    public Ticket parkVehicle(Vehicle vehicle) {
        for (ParkingFloor floor : floors) {
            ParkingSpot spot = floor.findAvailableSpot(vehicle);
            if (spot != null) {
                spot.park(vehicle);
                Ticket ticket = new Ticket(vehicle, spot);
                activeTickets.put(ticket.getTicketId(), ticket);
                System.out.println("  ✓ Parked " + vehicle.getLicensePlate()
                        + " (" + vehicle.getType() + ") at spot " + spot.getSpotId()
                        + " | Ticket: " + ticket.getTicketId());
                return ticket;
            }
        }
        System.out.println("  ✗ No available spot for " + vehicle.getLicensePlate());
        return null;
    }

    public double unparkVehicle(Ticket ticket) {
        ticket.setExitTime(LocalDateTime.now());
        ticket.getSpot().unpark();
        activeTickets.remove(ticket.getTicketId());

        double fee = pricingStrategy.calculateFee(ticket);
        System.out.println("  ✓ Unparked " + ticket.getVehicle().getLicensePlate()
                + " from " + ticket.getSpot().getSpotId()
                + " | Duration: " + ticket.getDurationMinutes() + " min");
        return fee;
    }

    public void displayAvailability() {
        System.out.println("  ┌─── AVAILABILITY ───────────────────┐");
        for (ParkingFloor floor : floors) {
            System.out.printf("  │ %s: Small=%d  Medium=%d  Large=%d%n",
                    floor.getFloorId(),
                    floor.getAvailableCount(SpotSize.SMALL),
                    floor.getAvailableCount(SpotSize.MEDIUM),
                    floor.getAvailableCount(SpotSize.LARGE));
        }
        System.out.println("  └─────────────────────────────────────┘");
    }
}

/*
 * CLASS DIAGRAM:
 * ─────────────────────────────────────────────────────────────
 *   ParkingLot (Singleton)
 *     ├── List<ParkingFloor>
 *     │     └── List<ParkingSpot>
 *     │           └── Vehicle (parked)
 *     ├── Map<String, Ticket>
 *     └── PricingStrategy (interface)
 *           ├── HourlyPricing
 *           └── FlatRatePricing (easily addable)
 *
 *   Vehicle → VehicleType (enum)
 *   ParkingSpot → SpotSize (enum)
 *   Ticket → Vehicle + ParkingSpot + timestamps
 *
 * PATTERNS USED:
 *   ✦ Singleton — one parking lot
 *   ✦ Strategy — pricing algorithm
 *   ✦ Enum — type safety for vehicle/spot types
 *
 * COMPILE & RUN:
 *   javac ParkingLotSystem.java && java ParkingLotSystem
 */
