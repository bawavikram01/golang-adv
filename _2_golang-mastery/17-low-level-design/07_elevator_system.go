package lowleveldesign

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// =============================================================================
// LLD PROBLEM: ELEVATOR SYSTEM
// =============================================================================
// Design an elevator system for a building.
//
// Requirements:
// 1. Multiple elevators
// 2. Handle requests from floors (up/down buttons)
// 3. Handle requests from inside elevator (destination floor)
// 4. Efficient scheduling (minimize wait time)
// 5. Support for VIP/express elevators
// 6. Display current floor and direction
// 7. Thread-safe (concurrent requests)
//
// Design Patterns:
// - State Pattern: Elevator states (idle, moving up, moving down, doors open)
// - Strategy Pattern: Scheduling algorithms (SCAN, LOOK, FCFS)
// - Observer Pattern: Notify displays of floor changes
// - Command Pattern: Floor requests as commands

// =============================================================================
// DOMAIN MODELS
// =============================================================================

type Direction int

const (
	DirIdle Direction = iota
	DirUp
	DirDown
)

func (d Direction) String() string {
	return [...]string{"IDLE", "UP", "DOWN"}[d]
}

type ElevatorState int

const (
	ElevatorIdle ElevatorState = iota
	ElevatorMovingUp
	ElevatorMovingDown
	ElevatorDoorsOpen
)

func (s ElevatorState) String() string {
	return [...]string{"IDLE", "MOVING_UP", "MOVING_DOWN", "DOORS_OPEN"}[s]
}

// Request represents a floor request
type ElevatorRequest struct {
	FromFloor   int
	ToFloor     int
	Direction   Direction
	RequestedAt time.Time
}

// =============================================================================
// ELEVATOR
// =============================================================================

type Elevator struct {
	ID           int
	CurrentFloor int
	Direction    Direction
	State        ElevatorState
	Capacity     int
	Passengers   int
	Stops        map[int]bool // floors to stop at
	mu           sync.Mutex
	MinFloor     int
	MaxFloor     int
}

func NewElevator(id, minFloor, maxFloor, capacity int) *Elevator {
	return &Elevator{
		ID:           id,
		CurrentFloor: 1,
		Direction:    DirIdle,
		State:        ElevatorIdle,
		Capacity:     capacity,
		Stops:        make(map[int]bool),
		MinFloor:     minFloor,
		MaxFloor:     maxFloor,
	}
}

func (e *Elevator) AddStop(floor int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if floor >= e.MinFloor && floor <= e.MaxFloor {
		e.Stops[floor] = true
	}
}

func (e *Elevator) RemoveStop(floor int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.Stops, floor)
}

func (e *Elevator) HasStops() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.Stops) > 0
}

func (e *Elevator) ShouldStop() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Stops[e.CurrentFloor]
}

func (e *Elevator) IsAvailable() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.Passengers < e.Capacity
}

// DistanceTo calculates how far this elevator is from a floor
// considering current direction (SCAN algorithm distance)
func (e *Elevator) DistanceTo(floor int, requestDir Direction) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.State == ElevatorIdle {
		return abs(e.CurrentFloor - floor)
	}

	// If moving towards the floor in the same direction, just distance
	if e.Direction == DirUp && floor >= e.CurrentFloor && requestDir == DirUp {
		return floor - e.CurrentFloor
	}
	if e.Direction == DirDown && floor <= e.CurrentFloor && requestDir == DirDown {
		return e.CurrentFloor - floor
	}

	// Otherwise, need to reverse — penalize
	if e.Direction == DirUp {
		return (e.MaxFloor - e.CurrentFloor) + (e.MaxFloor - floor)
	}
	return (e.CurrentFloor - e.MinFloor) + (floor - e.MinFloor)
}

func (e *Elevator) MoveOne() {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch e.Direction {
	case DirUp:
		if e.CurrentFloor < e.MaxFloor {
			e.CurrentFloor++
			e.State = ElevatorMovingUp
		}
	case DirDown:
		if e.CurrentFloor > e.MinFloor {
			e.CurrentFloor--
			e.State = ElevatorMovingDown
		}
	}
}

func (e *Elevator) OpenDoors() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.State = ElevatorDoorsOpen
	delete(e.Stops, e.CurrentFloor)
	fmt.Printf("  🛗 Elevator %d: Doors OPEN at floor %d\n", e.ID, e.CurrentFloor)
}

func (e *Elevator) CloseDoors() {
	e.mu.Lock()
	defer e.mu.Unlock()
	fmt.Printf("  🛗 Elevator %d: Doors CLOSED at floor %d\n", e.ID, e.CurrentFloor)
}

func (e *Elevator) UpdateDirection() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.Stops) == 0 {
		e.Direction = DirIdle
		e.State = ElevatorIdle
		return
	}

	// Check if there are stops above or below
	hasAbove, hasBelow := false, false
	for floor := range e.Stops {
		if floor > e.CurrentFloor {
			hasAbove = true
		}
		if floor < e.CurrentFloor {
			hasBelow = true
		}
	}

	switch e.Direction {
	case DirUp:
		if !hasAbove {
			e.Direction = DirDown
		}
	case DirDown:
		if !hasBelow {
			e.Direction = DirUp
		}
	case DirIdle:
		if hasAbove {
			e.Direction = DirUp
		} else if hasBelow {
			e.Direction = DirDown
		}
	}
}

func (e *Elevator) Status() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return fmt.Sprintf("Elevator %d: Floor=%d, Dir=%s, State=%s, Stops=%v",
		e.ID, e.CurrentFloor, e.Direction, e.State, e.Stops)
}

// =============================================================================
// SCHEDULING STRATEGY (Strategy Pattern)
// =============================================================================

type SchedulingStrategy interface {
	SelectElevator(elevators []*Elevator, request *ElevatorRequest) *Elevator
}

// LOOK Algorithm: Select the nearest elevator moving in the same direction
type LOOKStrategy struct{}

func (s *LOOKStrategy) SelectElevator(elevators []*Elevator, req *ElevatorRequest) *Elevator {
	var best *Elevator
	bestDist := int(^uint(0) >> 1) // max int

	for _, e := range elevators {
		if !e.IsAvailable() {
			continue
		}
		dist := e.DistanceTo(req.FromFloor, req.Direction)
		if dist < bestDist {
			bestDist = dist
			best = e
		}
	}
	return best
}

// Simple FCFS (First Come First Served) — round robin
type FCFSStrategy struct {
	next int
}

func (s *FCFSStrategy) SelectElevator(elevators []*Elevator, _ *ElevatorRequest) *Elevator {
	if len(elevators) == 0 {
		return nil
	}
	e := elevators[s.next%len(elevators)]
	s.next++
	return e
}

// =============================================================================
// ELEVATOR CONTROLLER (Facade)
// =============================================================================

type ElevatorController struct {
	elevators []*Elevator
	strategy  SchedulingStrategy
	requests  chan *ElevatorRequest
	mu        sync.Mutex
	running   bool
}

func NewElevatorController(numElevators, floors, capacity int, strategy SchedulingStrategy) *ElevatorController {
	ec := &ElevatorController{
		elevators: make([]*Elevator, numElevators),
		strategy:  strategy,
		requests:  make(chan *ElevatorRequest, 100),
	}
	for i := 0; i < numElevators; i++ {
		ec.elevators[i] = NewElevator(i+1, 1, floors, capacity)
	}
	return ec
}

// RequestElevator: Called when someone presses up/down button on a floor
func (ec *ElevatorController) RequestElevator(fromFloor int, dir Direction) {
	req := &ElevatorRequest{
		FromFloor:   fromFloor,
		Direction:   dir,
		RequestedAt: time.Now(),
	}

	elevator := ec.strategy.SelectElevator(ec.elevators, req)
	if elevator == nil {
		fmt.Printf("⚠️  No available elevator for floor %d\n", fromFloor)
		return
	}

	elevator.AddStop(fromFloor)
	elevator.UpdateDirection()
	fmt.Printf("📍 Floor %d (%s) → Assigned to Elevator %d\n",
		fromFloor, dir, elevator.ID)
}

// SelectFloor: Called when passenger inside elevator selects a floor
func (ec *ElevatorController) SelectFloor(elevatorID, floor int) {
	if elevatorID < 1 || elevatorID > len(ec.elevators) {
		return
	}
	e := ec.elevators[elevatorID-1]
	e.AddStop(floor)
	e.UpdateDirection()
	fmt.Printf("🔘 Elevator %d: Passenger selected floor %d\n", elevatorID, floor)
}

// Step: Simulate one time step (move all elevators by one floor)
func (ec *ElevatorController) Step() {
	for _, e := range ec.elevators {
		if !e.HasStops() {
			continue
		}

		if e.ShouldStop() {
			e.OpenDoors()
			// Simulate passenger exchange
			time.Sleep(50 * time.Millisecond)
			e.CloseDoors()
			e.UpdateDirection()
		} else {
			e.MoveOne()
		}
	}
}

// Status: Display all elevator statuses
func (ec *ElevatorController) Status() {
	fmt.Println("\n=== Elevator System Status ===")
	for _, e := range ec.elevators {
		fmt.Println(e.Status())
	}
	fmt.Println()
}

// Run: Start the simulation loop
func (ec *ElevatorController) Run(steps int) {
	for i := 0; i < steps; i++ {
		ec.Step()
		time.Sleep(100 * time.Millisecond)
	}
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleElevator() {
	// Building: 10 floors, 3 elevators, 8 person capacity
	controller := NewElevatorController(3, 10, 8, &LOOKStrategy{})

	controller.Status()

	// Simulate requests
	controller.RequestElevator(5, DirUp)   // Someone on floor 5 wants to go up
	controller.RequestElevator(3, DirDown) // Someone on floor 3 wants to go down
	controller.RequestElevator(8, DirDown) // Someone on floor 8 wants to go down

	// Simulate a few steps
	fmt.Println("--- Running simulation ---")
	controller.Run(10)

	// Passenger inside elevator 1 selects floor 7
	controller.SelectFloor(1, 7)

	controller.Run(5)
	controller.Status()
}

// =============================================================================
// HELPER
// =============================================================================

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// init seeds the random number generator (used in examples)
func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
