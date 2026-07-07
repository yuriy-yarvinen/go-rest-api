package users

import (
	"errors"
)

// ErrUserNotFound is returned by the repository when an User does not exist.
var ErrUserNotFound = errors.New("User not found")

// User is the domain entity. It knows nothing about HTTP or SQL.
type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserRepository is defined by the domain and implemented by the
// infrastructure layer. The domain depends on this abstraction, never on a
// concrete database.
type UserRepository interface {
	Register(User *User) error
	Login(User *User) error
	GetByID(id int64) (*User, error)
	Update(User *User) error
	Delete(id int64) error
}
