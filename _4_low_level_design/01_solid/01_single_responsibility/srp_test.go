package singleresponsibility

import "testing"

func TestUserValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool
	}{
		{"valid user", User{ID: "1", Name: "Vikram", Email: "v@x.com"}, false},
		{"empty name", User{ID: "2", Name: "", Email: "v@x.com"}, true},
		{"invalid email", User{ID: "3", Name: "Vikram", Email: "nope"}, true},
	}

	v := &UserValidator{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate(tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserRepository_SaveAndFind(t *testing.T) {
	repo := NewUserRepository()
	u := User{ID: "42", Name: "Alice", Email: "alice@x.com"}

	if err := repo.Save(u); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := repo.FindByID("42")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got.Name != "Alice" {
		t.Errorf("got Name = %q, want Alice", got.Name)
	}
}

func TestUserRepository_NotFound(t *testing.T) {
	repo := NewUserRepository()
	_, err := repo.FindByID("ghost")
	if err == nil {
		t.Error("expected error for missing user")
	}
}

func TestEmailService_SendWelcome(t *testing.T) {
	e := &EmailService{}
	u := User{ID: "1", Name: "Bob", Email: "bob@x.com"}

	if err := e.SendWelcome(u); err != nil {
		t.Fatalf("SendWelcome() error = %v", err)
	}
	if len(e.SentEmails) != 1 {
		t.Fatalf("expected 1 email, got %d", len(e.SentEmails))
	}
	if e.SentEmails[0] != "Welcome Bob!" {
		t.Errorf("got %q", e.SentEmails[0])
	}
}

func TestUserService_Register(t *testing.T) {
	v := &UserValidator{}
	repo := NewUserRepository()
	email := &EmailService{}
	svc := NewUserService(v, repo, email)

	u := User{ID: "1", Name: "Vikram", Email: "v@x.com"}
	if err := svc.Register(u); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Verify each responsibility did its job
	got, _ := repo.FindByID("1")
	if got.Name != "Vikram" {
		t.Errorf("user not saved correctly")
	}
	if len(email.SentEmails) != 1 {
		t.Errorf("welcome email not sent")
	}
}

func TestUserService_Register_ValidationFails(t *testing.T) {
	svc := NewUserService(&UserValidator{}, NewUserRepository(), &EmailService{})
	err := svc.Register(User{ID: "1", Name: "", Email: "v@x.com"})
	if err == nil {
		t.Error("expected validation error")
	}
}
