package lowleveldesign

import (
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// LLD PROBLEM: PARKING LOT SYSTEM
// =============================================================================
// This is the #1 most asked LLD interview question.
//
// Requirements:
// 1. Multi-level parking lot
// 2. Different vehicle sizes (motorcycle, car, bus)
// 3. Different spot sizes (small, medium, large)
// 4. Track available spots per level
// 5. Assign nearest available spot
// 6. Calculate parking fee based on duration
// 7. Thread-safe (concurrent entry/exit)
//
// Design Principles Used:
// - Strategy Pattern: Different pricing strategies
// - Factory Method: Creating spots and tickets
// - Observer: Notify when lot is full/available
// - SRP: Each struct has one responsibility

// =============================================================================
// DOMAIN MODELS
// =============================================================================

type VehicleSize int

const (
	SizeMotorcycle VehicleSize = iota
	SizeCar
	SizeBus
)

func (s VehicleSize) String() string {
	return [...]string{"Motorcycle", "Car", "Bus"}[s]
}

type SpotSize int

const (
	SpotSmall SpotSize = iota
	SpotMedium
	SpotLarge
)

func (s SpotSize) String() string {
	return [...]string{"Small", "Medium", "Large"}[s]
}

// Vehicle represents any vehicle entering the lot
type Vehicle struct {
	LicensePlate string
	Size         VehicleSize
	EntryTime    time.Time
}

// ParkingSpot represents a single parking spot
type ParkingSpot struct {
	ID      string
	Level   int
	SpotNum int
	Size    SpotSize
	Vehicle *Vehicle // nil if available
	mu      sync.Mutex
}

func NewParkingSpot(level, num int, size SpotSize) *ParkingSpot {
	return &ParkingSpot{
		ID:      fmt.Sprintf("L%d-S%d", level, num),
		Level:   level,
		SpotNum: num,
		Size:    size,
	}
}

func (s *ParkingSpot) IsAvailable() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Vehicle == nil
}

func (s *ParkingSpot) CanFit(v *Vehicle) bool {
	// Small spot: only motorcycles
	// Medium spot: motorcycles + cars
	// Large spot: anything
	switch s.Size {
	case SpotSmall:
		return v.Size == SizeMotorcycle
	case SpotMedium:
		return v.Size <= SizeCar
	case SpotLarge:
		return true
	}
	return false
}

func (s *ParkingSpot) Park(v *Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Vehicle != nil {
		return fmt.Errorf("spot %s is occupied", s.ID)
	}
	if !s.CanFit(v) {
		return fmt.Errorf("vehicle %s (size: %v) cannot fit in spot %s (size: %v)",
			v.LicensePlate, v.Size, s.ID, s.Size)
	}
	s.Vehicle = v
	return nil
}

func (s *ParkingSpot) Unpark() *Vehicle {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.Vehicle
	s.Vehicle = nil
	return v
}

// =============================================================================
// PARKING TICKET
// =============================================================================

type ParkingTicket struct {
	ID        string
	Vehicle   *Vehicle
	Spot      *ParkingSpot
	EntryTime time.Time
	ExitTime  time.Time
	Fee       float64
	IsPaid    bool
}

// =============================================================================
// PRICING STRATEGY (Strategy Pattern)
// =============================================================================

type ParkingPricingStrategy interface {
	Calculate(duration time.Duration, vehicleSize VehicleSize) float64
}

// Hourly pricing: different rates per vehicle size
type HourlyPricing struct {
	rates map[VehicleSize]float64 // per hour
}

func NewHourlyPricing() *HourlyPricing {
	return &HourlyPricing{
		rates: map[VehicleSize]float64{
			SizeMotorcycle: 10.0,
			SizeCar:        20.0,
			SizeBus:        50.0,
		},
	}
}

func (p *HourlyPricing) Calculate(duration time.Duration, size VehicleSize) float64 {
	hours := duration.Hours()
	if hours < 1 {
		hours = 1 // minimum 1 hour
	}
	return hours * p.rates[size]
}

// Flat rate pricing
type FlatRateParkingPricing struct {
	rates map[VehicleSize]float64
}

func NewFlatRatePricing() *FlatRateParkingPricing {
	return &FlatRateParkingPricing{
		rates: map[VehicleSize]float64{
			SizeMotorcycle: 50.0,
			SizeCar:        100.0,
			SizeBus:        200.0,
		},
	}
}

func (p *FlatRateParkingPricing) Calculate(_ time.Duration, size VehicleSize) float64 {
	return p.rates[size]
}

// =============================================================================
// PARKING LEVEL
// =============================================================================

type ParkingLevel struct {
	level int
	spots []*ParkingSpot
	mu    sync.RWMutex
}

func NewParkingLevel(level int, small, medium, large int) *ParkingLevel {
	pl := &ParkingLevel{level: level}
	spotNum := 1
	for i := 0; i < small; i++ {
		pl.spots = append(pl.spots, NewParkingSpot(level, spotNum, SpotSmall))
		spotNum++
	}
	for i := 0; i < medium; i++ {
		pl.spots = append(pl.spots, NewParkingSpot(level, spotNum, SpotMedium))
		spotNum++
	}
	for i := 0; i < large; i++ {
		pl.spots = append(pl.spots, NewParkingSpot(level, spotNum, SpotLarge))
		spotNum++
	}
	return pl
}

func (l *ParkingLevel) FindAvailableSpot(v *Vehicle) *ParkingSpot {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, spot := range l.spots {
		if spot.IsAvailable() && spot.CanFit(v) {
			return spot
		}
	}
	return nil
}

func (l *ParkingLevel) AvailableCount() map[SpotSize]int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	count := map[SpotSize]int{SpotSmall: 0, SpotMedium: 0, SpotLarge: 0}
	for _, spot := range l.spots {
		if spot.IsAvailable() {
			count[spot.Size]++
		}
	}
	return count
}

// =============================================================================
// PARKING LOT (Facade + Singleton-ish)
// =============================================================================

type ParkingLot struct {
	name    string
	levels  []*ParkingLevel
	tickets map[string]*ParkingTicket // licensePlate -> ticket
	pricing ParkingPricingStrategy
	mu      sync.RWMutex
	nextID  int
}

func NewParkingLot(name string, pricing ParkingPricingStrategy) *ParkingLot {
	return &ParkingLot{
		name:    name,
		tickets: make(map[string]*ParkingTicket),
		pricing: pricing,
	}
}

func (pl *ParkingLot) AddLevel(small, medium, large int) {
	level := len(pl.levels) + 1
	pl.levels = append(pl.levels, NewParkingLevel(level, small, medium, large))
}

// Entry: Vehicle enters, gets a ticket
func (pl *ParkingLot) Entry(licensePlate string, size VehicleSize) (*ParkingTicket, error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	// Check if vehicle is already parked
	if _, exists := pl.tickets[licensePlate]; exists {
		return nil, fmt.Errorf("vehicle %s is already parked", licensePlate)
	}

	vehicle := &Vehicle{
		LicensePlate: licensePlate,
		Size:         size,
		EntryTime:    time.Now(),
	}

	// Find available spot (nearest = lowest level first)
	var spot *ParkingSpot
	for _, level := range pl.levels {
		spot = level.FindAvailableSpot(vehicle)
		if spot != nil {
			break
		}
	}

	if spot == nil {
		return nil, fmt.Errorf("parking lot full — no spot available for %v", size)
	}

	// Park the vehicle
	if err := spot.Park(vehicle); err != nil {
		return nil, err
	}

	// Issue ticket
	pl.nextID++
	ticket := &ParkingTicket{
		ID:        fmt.Sprintf("TKT-%04d", pl.nextID),
		Vehicle:   vehicle,
		Spot:      spot,
		EntryTime: vehicle.EntryTime,
	}
	pl.tickets[licensePlate] = ticket

	fmt.Printf("✅ %s parked at spot %s (Level %d)\n",
		licensePlate, spot.ID, spot.Level)
	return ticket, nil
}

// Exit: Vehicle leaves, fee calculated
func (pl *ParkingLot) Exit(licensePlate string) (*ParkingTicket, error) {
	pl.mu.Lock()
	defer pl.mu.Unlock()

	ticket, exists := pl.tickets[licensePlate]
	if !exists {
		return nil, fmt.Errorf("vehicle %s not found in parking lot", licensePlate)
	}

	// Calculate fee
	ticket.ExitTime = time.Now()
	duration := ticket.ExitTime.Sub(ticket.EntryTime)
	ticket.Fee = pl.pricing.Calculate(duration, ticket.Vehicle.Size)
	ticket.IsPaid = true

	// Free the spot
	ticket.Spot.Unpark()
	delete(pl.tickets, licensePlate)

	fmt.Printf("🚗 %s exited from spot %s. Duration: %v, Fee: ₹%.2f\n",
		licensePlate, ticket.Spot.ID, duration.Round(time.Minute), ticket.Fee)
	return ticket, nil
}

// Status: Display parking lot status
func (pl *ParkingLot) Status() {
	pl.mu.RLock()
	defer pl.mu.RUnlock()

	fmt.Printf("\n=== %s Status ===\n", pl.name)
	for i, level := range pl.levels {
		avail := level.AvailableCount()
		fmt.Printf("Level %d: Small=%d, Medium=%d, Large=%d available\n",
			i+1, avail[SpotSmall], avail[SpotMedium], avail[SpotLarge])
	}
	fmt.Printf("Vehicles parked: %d\n\n", len(pl.tickets))
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleParkingLot() {
	// Create parking lot with hourly pricing
	lot := NewParkingLot("CityCenter Mall Parking", NewHourlyPricing())
	lot.AddLevel(5, 10, 3) // Level 1: 5 small, 10 medium, 3 large
	lot.AddLevel(5, 10, 3) // Level 2: same

	lot.Status()

	// Vehicles enter
	lot.Entry("KA-01-1234", SizeCar)
	lot.Entry("KA-02-5678", SizeMotorcycle)
	lot.Entry("KA-03-9999", SizeBus)
	lot.Entry("KA-04-1111", SizeCar)

	lot.Status()

	// Vehicles exit
	lot.Exit("KA-01-1234")
	lot.Exit("KA-02-5678")

	lot.Status()
}
