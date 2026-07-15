package postgres

import (
	"database/sql"
	"errors"

	"go-rest-api/events"
)

// Repository implements events.EventRepository backed by PostgreSQL.
type Repository struct {
	db *sql.DB
}

// compile-time check that Repository satisfies the domain interface.
var _ events.EventRepository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(event *events.Event) error {
	const query = `
		INSERT INTO events (name, description, location, date_time, user_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	return r.db.QueryRow(query, event.Name, event.Description, event.Location, event.DateTime, event.UserID).
		Scan(&event.ID)
}

func (r *Repository) GetAll() ([]events.Event, error) {
	rows, err := r.db.Query("SELECT id, name, description, location, date_time, user_id FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []events.Event
	for rows.Next() {
		var e events.Event
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Location, &e.DateTime, &e.UserID); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *Repository) GetByID(id int64) (*events.Event, error) {
	row := r.db.QueryRow("SELECT id, name, description, location, date_time, user_id FROM events WHERE id = $1", id)

	var e events.Event
	err := row.Scan(&e.ID, &e.Name, &e.Description, &e.Location, &e.DateTime, &e.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, events.ErrEventNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) Update(event *events.Event) error {
	const query = "UPDATE events SET name = $1, description = $2, location = $3, date_time = $4, user_id = $5 WHERE id = $6"
	result, err := r.db.Exec(query, event.Name, event.Description, event.Location, event.DateTime, event.UserID, event.ID)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM events WHERE id = $1", id)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

func (r *Repository) RegisterUserToEvent(eventID, userID int64) error {
	checkQuery := "SELECT id FROM registrations WHERE event_id = $1 AND user_id = $2"
	var existingID int64
	err := r.db.QueryRow(checkQuery, eventID, userID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if existingID != 0 {
		return errors.New("user is already registered for this event")
	}

	const query = `
		INSERT INTO registrations (event_id, user_id)
		VALUES ($1, $2)`

	_, errCreate := r.db.Exec(query, eventID, userID)
	return errCreate
}

func (r *Repository) UnregisterUserFromEvent(eventID, userID int64) error {
	checkQuery := "SELECT id FROM registrations WHERE event_id = $1 AND user_id = $2"
	var existingID int64
	err := r.db.QueryRow(checkQuery, eventID, userID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if existingID == 0 {
		return errors.New("user is not registered for this event")
	}

	const query = `
		DELETE FROM registrations
		WHERE event_id = $1 AND user_id = $2`

	result, err := r.db.Exec(query, eventID, userID)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

// affectedOrNotFound maps a zero-row UPDATE/DELETE to ErrEventNotFound.
func affectedOrNotFound(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return events.ErrEventNotFound
	}
	return nil
}
