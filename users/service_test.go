package users

import (
	"errors"
	"testing"
)

// mockRepository implements UserRepository with per-call function fields,
// so each test only wires up the behavior it cares about.
type mockRepository struct {
	registerFunc func(*User) error
	loginFunc    func(*User) error
	getByIDFunc  func(int64) (*User, error)
	updateFunc   func(*User) error
	deleteFunc   func(int64) error
}

var _ UserRepository = (*mockRepository)(nil)

func (m *mockRepository) Register(user *User) error       { return m.registerFunc(user) }
func (m *mockRepository) Login(user *User) error          { return m.loginFunc(user) }
func (m *mockRepository) GetByID(id int64) (*User, error) { return m.getByIDFunc(id) }
func (m *mockRepository) Update(user *User) error         { return m.updateFunc(user) }
func (m *mockRepository) Delete(id int64) error           { return m.deleteFunc(id) }

func TestServiceRegisterDelegatesToRepository(t *testing.T) {
	user := &User{Email: "alice@gmail.com", Password: "s3cret-pass"}
	var got *User
	repo := &mockRepository{registerFunc: func(u *User) error {
		got = u
		u.ID = 7
		return nil
	}}

	if err := NewService(repo).Register(user); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if got != user {
		t.Error("Register did not pass the same *User through to the repository")
	}
	if user.ID != 7 {
		t.Errorf("user.ID = %d, want 7 (repository-assigned id should be visible to the caller)", user.ID)
	}
}

func TestServiceLoginPropagatesError(t *testing.T) {
	repo := &mockRepository{loginFunc: func(*User) error { return ErrUserNotFound }}

	err := NewService(repo).Login(&User{Email: "alice@gmail.com", Password: "wrong"})
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Login error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestServiceGetByIDNotFound(t *testing.T) {
	repo := &mockRepository{getByIDFunc: func(int64) (*User, error) { return nil, ErrUserNotFound }}

	_, err := NewService(repo).GetByID(99)
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("GetByID error = %v, want %v", err, ErrUserNotFound)
	}
}

func TestServiceUpdateDelegatesToRepository(t *testing.T) {
	user := &User{ID: 5, Email: "alice2@gmail.com"}
	var got *User
	repo := &mockRepository{updateFunc: func(u *User) error {
		got = u
		return nil
	}}

	if err := NewService(repo).Update(user); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got != user {
		t.Error("Update did not pass the same *User through to the repository")
	}
}

func TestServiceDeleteDelegatesToRepository(t *testing.T) {
	var gotID int64
	repo := &mockRepository{deleteFunc: func(id int64) error {
		gotID = id
		return nil
	}}

	if err := NewService(repo).Delete(5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotID != 5 {
		t.Errorf("Delete called repository with id %d, want 5", gotID)
	}
}
