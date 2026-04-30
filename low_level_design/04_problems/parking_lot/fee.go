package parkinglot

import "time"

// ──────────────────────────────────────────────
// Fee Strategy — Strategy Pattern
// ──────────────────────────────────────────────

type FeeStrategy interface {
	Calculate(vehicleType VehicleType, duration time.Duration) float64
	Name() string
}

// HourlyFee charges per hour with different rates per vehicle type.
type HourlyFee struct {
	rates map[VehicleType]float64
}

func NewHourlyFee() *HourlyFee {
	return &HourlyFee{
		rates: map[VehicleType]float64{
			VehicleBike:  10,
			VehicleCar:   20,
			VehicleTruck: 40,
		},
	}
}

func (h *HourlyFee) Calculate(vType VehicleType, duration time.Duration) float64 {
	hours := duration.Hours()
	if hours < 1 {
		hours = 1 // minimum 1 hour charge
	}
	return hours * h.rates[vType]
}

func (h *HourlyFee) Name() string { return "Hourly" }

// FlatFee charges a flat rate regardless of duration.
type FlatFee struct {
	rates map[VehicleType]float64
}

func NewFlatFee() *FlatFee {
	return &FlatFee{
		rates: map[VehicleType]float64{
			VehicleBike:  50,
			VehicleCar:   100,
			VehicleTruck: 200,
		},
	}
}

func (f *FlatFee) Calculate(vType VehicleType, _ time.Duration) float64 {
	return f.rates[vType]
}

func (f *FlatFee) Name() string { return "Flat" }
