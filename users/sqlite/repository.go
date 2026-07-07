package sqlite

import (
	"database/sql"
	"errors"

	"go-rest-api/users"
	"go-rest-api/utils"
)

// Repository implements users.UserRepository backed by SQLite.
type Repository struct {
	db *sql.DB
}

// compile-time check that Repository satisfies the domain interface.
var _ users.UserRepository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Register(user *users.User) error {
	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = passwordHash

	const query = "INSERT INTO users (email, password) VALUES (?, ?)"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Email, user.Password)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func (r *Repository) Login(user *users.User) error {
	var stored users.User
	row := r.db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", user.Email)
	err := row.Scan(&stored.ID, &stored.Email, &stored.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return users.ErrUserNotFound
	}
	if err != nil {
		return err
	}

	if !utils.CheckPasswordHash(user.Password, stored.Password) {
		return users.ErrUserNotFound
	}

	user.ID = stored.ID
	user.Password = ""
	return nil
}

func (r *Repository) GetByID(id int64) (*users.User, error) {
	row := r.db.QueryRow("SELECT id, email FROM users WHERE id = ?", id)

	var u users.User
	err := row.Scan(&u.ID, &u.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, users.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) Update(user *users.User) error {
	const query = "UPDATE users SET email = ? WHERE id = ?"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(user.Email, user.ID)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

func (r *Repository) Delete(id int64) error {
	stmt, err := r.db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

// affectedOrNotFound maps a zero-row UPDATE/DELETE to ErrUserNotFound.
func affectedOrNotFound(result sql.Result) error {
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return users.ErrUserNotFound
	}
	return nil
}
