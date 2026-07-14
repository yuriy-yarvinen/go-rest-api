package users

import (
	"encoding/json"
	"errors"
)

// ErrUserNotFound is returned by the repository when an User does not exist.
var ErrUserNotFound = errors.New("User not found")
var ErrUserAlreadyExists = errors.New("User already exists")

// User is the domain entity. It knows nothing about HTTP or SQL.
type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// MarshalJSON omits Password from any JSON output. The struct tag alone
// only governs field names, not exclusion, so a handler that forgets to
// clear Password before responding would otherwise leak it (a hash, or
// worse, the plaintext the client just sent on update). Password is still
// read normally on the way in — ShouldBindJSON uses the default decoder,
// which this method doesn't affect.
func (u User) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}{ID: u.ID, Email: u.Email})
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
