# Parking Lot System — LLD

## Problem Statement

Design a parking lot system that:
- Has multiple floors, each with multiple spots
- Supports different vehicle types (Bike, Car, Truck)
- Different spot sizes for different vehicle types
- Assigns the nearest available spot to a vehicle
- Tracks entry/exit and calculates fees
- Handles concurrent access

## Key Design Decisions

1. **Vehicle ↔ Spot type mapping** — bikes can only park in bike spots, etc.
2. **Strategy for spot assignment** — nearest to entrance
3. **Fee calculation** — strategy pattern (hourly, flat, etc.)
4. **Concurrency** — mutex on spot allocation

## Class Diagram (Mental Model)

```
ParkingLot
├── floors []Floor
├── entryPanels []EntryPanel
├── exitPanels []ExitPanel
└── feeStrategy FeeStrategy

Floor
├── spots []ParkingSpot
└── displayBoard DisplayBoard

ParkingSpot (interface)
├── BikeSpot
├── CarSpot
└── TruckSpot

Vehicle (interface)
├── Bike
├── Car
└── Truck

Ticket
├── vehicle Vehicle
├── spot ParkingSpot
├── entryTime time.Time
└── exitTime time.Time
```

## Patterns Used
- **Strategy** → fee calculation
- **Factory** → vehicle/spot creation
- **Observer** → display board updates when spots change
- **State** → spot (free/occupied)
