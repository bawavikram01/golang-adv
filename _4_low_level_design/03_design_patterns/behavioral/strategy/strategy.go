// Package strategy demonstrates the Strategy pattern.
//
// INTENT: Define a family of algorithms, encapsulate each one, and make them
// interchangeable. Strategy lets the algorithm vary independently from
// clients that use it.
//
// WHEN TO USE:
//   - Multiple algorithms for the same task (sort, compress, route)
//   - You want to switch behavior at runtime
//   - You want to eliminate large if/else or switch blocks
//
// Go idiom: Strategy = interface. Each algorithm = struct implementing it.
package strategy

import "sort"

// ──────────────────────────────────────────────
// Strategy interface
// ──────────────────────────────────────────────

type SortStrategy interface {
	Sort(data []int) []int
	Name() string
}

// ──────────────────────────────────────────────
// Concrete strategies
// ──────────────────────────────────────────────

// BubbleSort — O(n²) but simple
type BubbleSort struct{}

func (b BubbleSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

func (b BubbleSort) Name() string { return "BubbleSort" }

// QuickSort — uses stdlib sort (which is quicksort-variant)
type QuickSort struct{}

func (q QuickSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	sort.Ints(result)
	return result
}

func (q QuickSort) Name() string { return "QuickSort" }

// ReverseSort — sorts in descending order
type ReverseSort struct{}

func (r ReverseSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

func (r ReverseSort) Name() string { return "ReverseSort" }

// ──────────────────────────────────────────────
// Context — uses the strategy
// ──────────────────────────────────────────────

type Sorter struct {
	strategy SortStrategy
}

func NewSorter(s SortStrategy) *Sorter {
	return &Sorter{strategy: s}
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
	s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
	return s.strategy.Sort(data)
}

func (s *Sorter) StrategyName() string {
	return s.strategy.Name()
}
