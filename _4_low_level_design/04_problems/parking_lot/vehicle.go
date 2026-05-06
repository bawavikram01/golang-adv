package parkinglot

import "fmt"

// ──────────────────────────────────────────────
// Vehicle types
// ──────────────────────────────────────────────

type VehicleType string

const (
	VehicleBike  VehicleType = "Bike"
	VehicleCar   VehicleType = "Car"
	VehicleTruck VehicleType = "Truck"
)

type Vehicle struct {
	LicensePlate string
	Type         VehicleType
	Owner        string
}

func NewVehicle(plate string, vType VehicleType, owner string) *Vehicle {
	return &Vehicle{
		LicensePlate: plate,
		Type:         vType,
		Owner:        owner,
	}
}

func (v *Vehicle) String() string {
	return fmt.Sprintf("%s [%s] - %s", v.LicensePlate, v.Type, v.Owner)
}

// ──────────────────────────────────────────────
// Parking Spot
// ──────────────────────────────────────────────

type SpotType string

const (
	SpotBike  SpotType = "Bike"
	SpotCar   SpotType = "Car"
	SpotTruck SpotType = "Truck"
)

type ParkingSpot struct {
	ID       string
	Floor    int
	Number   int
	Type     SpotType
	Occupied bool
	Vehicle  *Vehicle
}

func (ps *ParkingSpot) CanFit(v *Vehicle) bool {
	switch v.Type {
	case VehicleBike:
		return ps.Type == SpotBike
	case VehicleCar:
		return ps.Type == SpotCar
	case VehicleTruck:
		return ps.Type == SpotTruck
	default:
		return false
	}
}

func (ps *ParkingSpot) Park(v *Vehicle) error {
	if ps.Occupied {
		return fmt.Errorf("spot %s is already occupied", ps.ID)
	}
	if !ps.CanFit(v) {
		return fmt.Errorf("vehicle %s cannot fit in spot %s", v.Type, ps.Type)
	}
	ps.Occupied = true
	ps.Vehicle = v
	return nil
}

func (ps *ParkingSpot) Vacate() {
	ps.Occupied = false
	ps.Vehicle = nil
}
