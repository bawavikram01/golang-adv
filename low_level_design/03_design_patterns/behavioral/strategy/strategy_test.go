package strategy

import (
	"reflect"
	"testing"
)

func TestBubbleSort(t *testing.T) {
	sorter := NewSorter(BubbleSort{})
	got := sorter.Sort([]int{5, 3, 8, 1, 2})
	want := []int{1, 2, 3, 5, 8}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BubbleSort = %v, want %v", got, want)
	}
}

func TestQuickSort(t *testing.T) {
	sorter := NewSorter(QuickSort{})
	got := sorter.Sort([]int{9, 1, 4, 7, 3})
	want := []int{1, 3, 4, 7, 9}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("QuickSort = %v, want %v", got, want)
	}
}

func TestReverseSort(t *testing.T) {
	sorter := NewSorter(ReverseSort{})
	got := sorter.Sort([]int{1, 5, 3, 9, 2})
	want := []int{9, 5, 3, 2, 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReverseSort = %v, want %v", got, want)
	}
}

func TestSorter_SwapStrategy(t *testing.T) {
	sorter := NewSorter(BubbleSort{})
	if sorter.StrategyName() != "BubbleSort" {
		t.Errorf("initial strategy = %q", sorter.StrategyName())
	}

	// Switch strategy at runtime
	sorter.SetStrategy(QuickSort{})
	if sorter.StrategyName() != "QuickSort" {
		t.Errorf("after swap = %q", sorter.StrategyName())
	}

	got := sorter.Sort([]int{4, 2, 7})
	want := []int{2, 4, 7}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after swap sort = %v, want %v", got, want)
	}
}

func TestSort_DoesNotMutateOriginal(t *testing.T) {
	original := []int{5, 3, 1}
	sorter := NewSorter(BubbleSort{})
	_ = sorter.Sort(original)
	if !reflect.DeepEqual(original, []int{5, 3, 1}) {
		t.Errorf("original was mutated: %v", original)
	}
}
