package parkinglot

import "testing"

func TestParkingLot_ParkAndUnpark(t *testing.T) {
	floors := []*Floor{
		NewFloor(1, 2, 3, 1), // 2 bike, 3 car, 1 truck
	}
	pl := NewParkingLot("Test Lot", floors, NewHourlyFee())

	car := NewVehicle("KA-01-1234", VehicleCar, "Vikram")
	ticket, err := pl.ParkVehicle(car)
	if err != nil {
		t.Fatalf("ParkVehicle() error: %v", err)
	}
	if ticket.Vehicle.LicensePlate != "KA-01-1234" {
		t.Errorf("ticket vehicle = %q", ticket.Vehicle.LicensePlate)
	}

	// Unpark
	result, err := pl.UnparkVehicle(ticket.ID)
	if err != nil {
		t.Fatalf("UnparkVehicle() error: %v", err)
	}
	if !result.Paid {
		t.Error("ticket should be marked as paid")
	}
	if result.Fee <= 0 {
		t.Errorf("fee should be > 0, got %v", result.Fee)
	}
}

func TestParkingLot_DuplicateVehicle(t *testing.T) {
	floors := []*Floor{NewFloor(1, 0, 2, 0)}
	pl := NewParkingLot("Lot", floors, NewFlatFee())

	car := NewVehicle("KA-01-9999", VehicleCar, "Alice")
	_, _ = pl.ParkVehicle(car)

	_, err := pl.ParkVehicle(car)
	if err == nil {
		t.Error("expected error for duplicate vehicle")
	}
}

func TestParkingLot_Full(t *testing.T) {
	floors := []*Floor{NewFloor(1, 0, 1, 0)} // only 1 car spot
	pl := NewParkingLot("Tiny", floors, NewFlatFee())

	c1 := NewVehicle("KA-01-0001", VehicleCar, "A")
	_, _ = pl.ParkVehicle(c1)

	c2 := NewVehicle("KA-01-0002", VehicleCar, "B")
	_, err := pl.ParkVehicle(c2)
	if err == nil {
		t.Error("expected error when lot is full")
	}

	if !pl.IsFull(VehicleCar) {
		t.Error("IsFull should return true")
	}
}

func TestParkingLot_MultipleFloors(t *testing.T) {
	floors := []*Floor{
		NewFloor(1, 1, 1, 0),
		NewFloor(2, 1, 1, 0),
	}
	pl := NewParkingLot("Multi", floors, NewHourlyFee())

	b1 := NewVehicle("KA-01-B1", VehicleBike, "X")
	b2 := NewVehicle("KA-01-B2", VehicleBike, "Y")

	_, err1 := pl.ParkVehicle(b1)
	_, err2 := pl.ParkVehicle(b2)

	if err1 != nil || err2 != nil {
		t.Errorf("errors: %v, %v", err1, err2)
	}
}

func TestParkingLot_AvailableSpots(t *testing.T) {
	floors := []*Floor{NewFloor(1, 2, 3, 1)}
	pl := NewParkingLot("Lot", floors, NewFlatFee())

	spots := pl.AvailableSpots()
	if spots["Floor 1 - Car"] != 3 {
		t.Errorf("car spots = %d, want 3", spots["Floor 1 - Car"])
	}
	if spots["Floor 1 - Bike"] != 2 {
		t.Errorf("bike spots = %d, want 2", spots["Floor 1 - Bike"])
	}
}

func TestParkingSpot_CanFit(t *testing.T) {
	spot := &ParkingSpot{ID: "S1", Type: SpotCar}
	car := &Vehicle{Type: VehicleCar}
	bike := &Vehicle{Type: VehicleBike}

	if !spot.CanFit(car) {
		t.Error("car spot should fit car")
	}
	if spot.CanFit(bike) {
		t.Error("car spot should NOT fit bike")
	}
}

func TestFeeStrategy_Hourly(t *testing.T) {
	fee := NewHourlyFee()
	if fee.Name() != "Hourly" {
		t.Errorf("Name() = %q", fee.Name())
	}
}

func TestFeeStrategy_Flat(t *testing.T) {
	fee := NewFlatFee()
	if fee.Name() != "Flat" {
		t.Errorf("Name() = %q", fee.Name())
	}
}

func TestUnpark_InvalidTicket(t *testing.T) {
	floors := []*Floor{NewFloor(1, 1, 1, 1)}
	pl := NewParkingLot("Lot", floors, NewFlatFee())

	_, err := pl.UnparkVehicle("FAKE-TICKET")
	if err == nil {
		t.Error("expected error for invalid ticket")
	}
}
