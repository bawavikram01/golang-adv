// Package interfacesegregation demonstrates the Interface Segregation Principle.
//
// PRINCIPLE: No client should be forced to depend on methods it does not use.
//
// Split fat interfaces into small, focused ones.
//
// Go's stdlib is the BEST example of ISP:
//
//	io.Reader (1 method), io.Writer (1 method), io.Closer (1 method)
//	Then composed: io.ReadWriter, io.ReadCloser, etc.
//
// VIOLATION: A single giant interface that forces all implementors to provide
// methods they don't need.
package interfacesegregation

import "fmt"

// ──────────────────────────────────────────────
// BAD: Fat interface forces unnecessary implementations
// ──────────────────────────────────────────────

// BadWorker forces ALL workers to implement Eat() — but robots don't eat!
type BadWorker interface {
	Work()
	Eat()
	Sleep()
}

// ──────────────────────────────────────────────
// GOOD: Segregated interfaces — small and focused
// ──────────────────────────────────────────────

// Worker — anyone who can do work
type Worker interface {
	Work() string
}

// Eater — biological entities that need food
type Eater interface {
	Eat() string
}

// Sleeper — entities that sleep
type Sleeper interface {
	Sleep() string
}

// ──────────────────────────────────────────────
// Concrete types — implement only what they need
// ──────────────────────────────────────────────

// Human implements Worker, Eater, and Sleeper.
type Human struct {
	Name string
}

func (h Human) Work() string  { return fmt.Sprintf("%s is working", h.Name) }
func (h Human) Eat() string   { return fmt.Sprintf("%s is eating", h.Name) }
func (h Human) Sleep() string { return fmt.Sprintf("%s is sleeping", h.Name) }

// Robot implements only Worker — no Eat or Sleep needed!
type Robot struct {
	Model string
}

func (r Robot) Work() string { return fmt.Sprintf("Robot %s is working", r.Model) }

// ──────────────────────────────────────────────
// Composed interfaces — build bigger contracts from small ones
// ──────────────────────────────────────────────

// LivingWorker is a composed interface for entities that work AND eat AND sleep.
type LivingWorker interface {
	Worker
	Eater
	Sleeper
}

// ──────────────────────────────────────────────
// Functions accept ONLY what they need
// ──────────────────────────────────────────────

// Assign accepts anything that can Work — human, robot, whatever.
func Assign(w Worker, task string) string {
	return fmt.Sprintf("Assigned '%s': %s", task, w.Work())
}

// FeedAll only accepts Eaters — won't accidentally try to feed a robot.
func FeedAll(eaters []Eater) []string {
	results := make([]string, len(eaters))
	for i, e := range eaters {
		results[i] = e.Eat()
	}
	return results
}
