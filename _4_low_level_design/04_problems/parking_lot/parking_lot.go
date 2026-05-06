package parkinglot

import (
	"fmt"
	"time"
)

// ──────────────────────────────────────────────
// Ticket
// ──────────────────────────────────────────────

type Ticket struct {
	ID        string
	Vehicle   *Vehicle
	Spot      *ParkingSpot
	EntryTime time.Time
	ExitTime  time.Time
	Fee       float64
	Paid      bool
}

// ──────────────────────────────────────────────
// Floor
// ──────────────────────────────────────────────

type Floor struct {
	Number int
	Spots  []*ParkingSpot
}

func NewFloor(number int, bikeSpots, carSpots, truckSpots int) *Floor {
	f := &Floor{Number: number}
	spotNum := 1

	for i := 0; i < bikeSpots; i++ {
		f.Spots = append(f.Spots, &ParkingSpot{
			ID:     fmt.Sprintf("F%d-B%d", number, spotNum),
			Floor:  number,
			Number: spotNum,
			Type:   SpotBike,
		})
		spotNum++
	}
	for i := 0; i < carSpots; i++ {
		f.Spots = append(f.Spots, &ParkingSpot{
			ID:     fmt.Sprintf("F%d-C%d", number, spotNum),
			Floor:  number,
			Number: spotNum,
			Type:   SpotCar,
		})
		spotNum++
	}
	for i := 0; i < truckSpots; i++ {
		f.Spots = append(f.Spots, &ParkingSpot{
			ID:     fmt.Sprintf("F%d-T%d", number, spotNum),
			Floor:  number,
			Number: spotNum,
			Type:   SpotTruck,
		})
		spotNum++
	}
	return f
}

func (f *Floor) FindAvailableSpot(vType VehicleType) *ParkingSpot {
	for _, spot := range f.Spots {
		if !spot.Occupied && spot.CanFit(&Vehicle{Type: vType}) {
			return spot
		}
	}
	return nil
}

func (f *Floor) AvailableCount(spotType SpotType) int {
	count := 0
	for _, spot := range f.Spots {
		if !spot.Occupied && spot.Type == spotType {
			count++
		}
	}
	return count
}

// ──────────────────────────────────────────────
// Parking Lot — the main orchestrator
// ──────────────────────────────────────────────

type ParkingLot struct {
	Name        string
	Floors      []*Floor
	FeeStrategy FeeStrategy
	tickets     map[string]*Ticket // ticketID -> ticket
	vehicleMap  map[string]*Ticket // licensePlate -> ticket
	nextID      int
}

func NewParkingLot(name string, floors []*Floor, feeStrategy FeeStrategy) *ParkingLot {
	return &ParkingLot{
		Name:        name,
		Floors:      floors,
		FeeStrategy: feeStrategy,
		tickets:     make(map[string]*Ticket),
		vehicleMap:  make(map[string]*Ticket),
	}
}

func (pl *ParkingLot) ParkVehicle(v *Vehicle) (*Ticket, error) {
	// Check if already parked
	if _, exists := pl.vehicleMap[v.LicensePlate]; exists {
		return nil, fmt.Errorf("vehicle %s is already parked", v.LicensePlate)
	}

	// Find spot across floors
	for _, floor := range pl.Floors {
		spot := floor.FindAvailableSpot(v.Type)
		if spot != nil {
			if err := spot.Park(v); err != nil {
				return nil, err
			}

			pl.nextID++
			ticket := &Ticket{
				ID:        fmt.Sprintf("TKT-%d", pl.nextID),
				Vehicle:   v,
				Spot:      spot,
				EntryTime: time.Now(),
			}
			pl.tickets[ticket.ID] = ticket
			pl.vehicleMap[v.LicensePlate] = ticket
			return ticket, nil
		}
	}

	return nil, fmt.Errorf("no available spot for %s", v.Type)
}

func (pl *ParkingLot) UnparkVehicle(ticketID string) (*Ticket, error) {
	ticket, exists := pl.tickets[ticketID]
	if !exists {
		return nil, fmt.Errorf("ticket %s not found", ticketID)
	}

	ticket.ExitTime = time.Now()
	duration := ticket.ExitTime.Sub(ticket.EntryTime)
	ticket.Fee = pl.FeeStrategy.Calculate(ticket.Vehicle.Type, duration)
	ticket.Paid = true

	ticket.Spot.Vacate()
	delete(pl.vehicleMap, ticket.Vehicle.LicensePlate)

	return ticket, nil
}

func (pl *ParkingLot) AvailableSpots() map[string]int {
	result := make(map[string]int)
	for _, floor := range pl.Floors {
		result[fmt.Sprintf("Floor %d - Bike", floor.Number)] = floor.AvailableCount(SpotBike)
		result[fmt.Sprintf("Floor %d - Car", floor.Number)] = floor.AvailableCount(SpotCar)
		result[fmt.Sprintf("Floor %d - Truck", floor.Number)] = floor.AvailableCount(SpotTruck)
	}
	return result
}

func (pl *ParkingLot) IsFull(vType VehicleType) bool {
	for _, floor := range pl.Floors {
		if floor.FindAvailableSpot(vType) != nil {
			return false
		}
	}
	return true
}
