// Package singleresponsibility demonstrates the Single Responsibility Principle.
//
// PRINCIPLE: A struct/module should have only ONE reason to change.
//
// BAD EXAMPLE:
//
//	A single struct that handles user data, validation, persistence, AND email sending.
//	If any one of those changes, the whole struct must change.
//
// GOOD EXAMPLE (below):
//
//	Separate structs each own exactly one responsibility:
//	- User          → holds data
//	- UserValidator → validates user data
//	- UserRepo      → persists users
//	- EmailService  → sends emails
package singleresponsibility

import (
	"errors"
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────
// Domain entity — only holds data
// ──────────────────────────────────────────────

type User struct {
	ID    string
	Name  string
	Email string
}

// ──────────────────────────────────────────────
// Responsibility 1: Validation
// ──────────────────────────────────────────────

type UserValidator struct{}

func (v *UserValidator) Validate(u User) error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("name is required")
	}
	if !strings.Contains(u.Email, "@") {
		return errors.New("invalid email")
	}
	return nil
}

// ──────────────────────────────────────────────
// Responsibility 2: Persistence
// ──────────────────────────────────────────────

type UserRepository struct {
	store map[string]User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{store: make(map[string]User)}
}

func (r *UserRepository) Save(u User) error {
	r.store[u.ID] = u
	return nil
}

func (r *UserRepository) FindByID(id string) (User, error) {
	u, ok := r.store[id]
	if !ok {
		return User{}, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

// ──────────────────────────────────────────────
// Responsibility 3: Notification
// ──────────────────────────────────────────────

type EmailService struct {
	SentEmails []string // track for testing
}

func (e *EmailService) SendWelcome(u User) error {
	msg := fmt.Sprintf("Welcome %s!", u.Name)
	e.SentEmails = append(e.SentEmails, msg)
	return nil
}

// ──────────────────────────────────────────────
// Orchestrator — composes the single-responsibility pieces
// ──────────────────────────────────────────────

type UserService struct {
	validator *UserValidator
	repo      *UserRepository
	email     *EmailService
}

func NewUserService(v *UserValidator, r *UserRepository, e *EmailService) *UserService {
	return &UserService{validator: v, repo: r, email: e}
}

func (s *UserService) Register(u User) error {
	if err := s.validator.Validate(u); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if err := s.repo.Save(u); err != nil {
		return fmt.Errorf("save failed: %w", err)
	}
	if err := s.email.SendWelcome(u); err != nil {
		return fmt.Errorf("email failed: %w", err)
	}
	return nil
}
