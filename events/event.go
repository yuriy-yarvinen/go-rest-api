package events

import (
	"errors"
	"time"
)

// ErrEventNotFound is returned by the repository when an event does not exist.
var ErrEventNotFound = errors.New("event not found")

// Event is the domain entity. It knows nothing about HTTP or SQL.
type Event struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Location    string    `json:"location" binding:"required"`
	DateTime    time.Time `json:"date_time"`
	// UserID is the owning user. It's always set server-side from the
	// authenticated token, never trusted from client input — see
	// events/rest.Handler.create/update — so it carries no binding tag.
	UserID int64 `json:"user_id"`
}

// EventRepository is defined by the domain and implemented by the
// infrastructure layer. The domain depends on this abstraction, never on a
// concrete database.
type EventRepository interface {
	Create(event *Event) error
	GetAll() ([]Event, error)
	GetByID(id int64) (*Event, error)
	Update(event *Event) error
	Delete(id int64) error
	RegisterUserToEvent(eventID, userID int64) error
	UnregisterUserFromEvent(eventID, userID int64) error
}
