package oop

import "testing"

func TestBankAccount_Encapsulation(t *testing.T) {
	acc := NewBankAccount("Vikram", 1000)

	if acc.Balance() != 1000 {
		t.Errorf("initial balance = %v, want 1000", acc.Balance())
	}

	if err := acc.Deposit(500); err != nil {
		t.Fatalf("Deposit() error: %v", err)
	}
	if acc.Balance() != 1500 {
		t.Errorf("after deposit balance = %v, want 1500", acc.Balance())
	}

	if err := acc.Withdraw(200); err != nil {
		t.Fatalf("Withdraw() error: %v", err)
	}
	if acc.Balance() != 1300 {
		t.Errorf("after withdrawal balance = %v, want 1300", acc.Balance())
	}
}

func TestBankAccount_Errors(t *testing.T) {
	acc := NewBankAccount("Test", 100)

	if err := acc.Deposit(-50); err == nil {
		t.Error("expected error for negative deposit")
	}
	if err := acc.Withdraw(0); err == nil {
		t.Error("expected error for zero withdrawal")
	}
	if err := acc.Withdraw(999); err == nil {
		t.Error("expected error for overdraft")
	}
}

func TestPolymorphism_RenderAll(t *testing.T) {
	items := []Drawable{
		CircleShape{Radius: 5},
		RectShape{W: 10, H: 20},
		TextBox{Text: "hello"},
	}

	results := RenderAll(items)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0] != "Drawing circle r=5.0" {
		t.Errorf("got %q", results[0])
	}
}

func TestComposition_DogEmbedding(t *testing.T) {
	dog := Dog{
		Animal: Animal{Name: "Rex", Species: "Dog"},
		Breed:  "Labrador",
	}

	// Dog overrides Speak
	if got := dog.Speak(); got != "Rex barks! Woof!" {
		t.Errorf("Speak() = %q", got)
	}

	// Dog gets String() from embedded Animal
	if got := dog.String(); got != "Rex the Dog" {
		t.Errorf("String() = %q", got)
	}

	// Dog has its own method
	if got := dog.Fetch(); got != "Rex fetches the ball" {
		t.Errorf("Fetch() = %q", got)
	}
}

func TestMakeThemSpeak(t *testing.T) {
	speakers := []Speaker{
		Dog{Animal: Animal{Name: "Buddy", Species: "Dog"}},
		Cat{Animal: Animal{Name: "Whiskers", Species: "Cat"}},
	}

	results := MakeThemSpeak(speakers)
	if results[0] != "Buddy barks! Woof!" {
		t.Errorf("dog speak = %q", results[0])
	}
	if results[1] != "Whiskers purrs... meow" {
		t.Errorf("cat speak = %q", results[1])
	}
}
