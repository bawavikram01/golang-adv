package interfacesegregation

import "testing"

func TestHuman_ImplementsAllInterfaces(t *testing.T) {
	h := Human{Name: "Vikram"}

	// Human satisfies Worker
	var w Worker = h
	if got := w.Work(); got != "Vikram is working" {
		t.Errorf("Work() = %q", got)
	}

	// Human satisfies Eater
	var e Eater = h
	if got := e.Eat(); got != "Vikram is eating" {
		t.Errorf("Eat() = %q", got)
	}

	// Human satisfies Sleeper
	var s Sleeper = h
	if got := s.Sleep(); got != "Vikram is sleeping" {
		t.Errorf("Sleep() = %q", got)
	}

	// Human satisfies composed LivingWorker
	var lw LivingWorker = h
	_ = lw
}

func TestRobot_OnlyImplementsWorker(t *testing.T) {
	r := Robot{Model: "T-800"}

	// Robot satisfies Worker
	var w Worker = r
	if got := w.Work(); got != "Robot T-800 is working" {
		t.Errorf("Work() = %q", got)
	}

	// Robot does NOT satisfy Eater or Sleeper — and that's CORRECT.
	// The following would NOT compile:
	// var e Eater = r   // compile error
	// var s Sleeper = r  // compile error
}

func TestAssign_WorksWithAnyWorker(t *testing.T) {
	human := Human{Name: "Vikram"}
	robot := Robot{Model: "R2D2"}

	// Both can be assigned work
	r1 := Assign(human, "code review")
	if r1 != "Assigned 'code review': Vikram is working" {
		t.Errorf("unexpected: %s", r1)
	}

	r2 := Assign(robot, "deploy")
	if r2 != "Assigned 'deploy': Robot R2D2 is working" {
		t.Errorf("unexpected: %s", r2)
	}
}

func TestFeedAll_OnlyAcceptsEaters(t *testing.T) {
	humans := []Eater{
		Human{Name: "Alice"},
		Human{Name: "Bob"},
	}

	results := FeedAll(humans)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "Alice is eating" {
		t.Errorf("got %q", results[0])
	}
}
