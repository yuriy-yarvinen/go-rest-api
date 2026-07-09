package postgres

import (
	"database/sql"
	"errors"
	"go-rest-api/users"
	"go-rest-api/utils"
)

// Repository implements Users.UserRepository backed by PostgreSQL.
type Repository struct {
	db *sql.DB
}

// compile-time check that Repository satisfies the domain interface.
var _ users.UserRepository = (*Repository)(nil)

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Register(user *users.User) error {
	const query = `
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id`
	passwordHash, err := utils.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = passwordHash

	return r.db.QueryRow(query, user.Email, user.Password).Scan(&user.ID)
}

func (r *Repository) Login(user *users.User) error {
	var stored users.User

	row := r.db.QueryRow("SELECT password FROM users WHERE email = $1", user.Email)
	err := row.Scan(&stored.Password)
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
	row := r.db.QueryRow("SELECT id, email FROM users WHERE id = $1", id)

	var e users.User
	err := row.Scan(&e.ID, &e.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, users.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) Update(User *users.User) error {
	const query = "UPDATE users SET email = $1 WHERE id = $2"
	result, err := r.db.Exec(query, User.Email, User.ID)
	if err != nil {
		return err
	}
	return affectedOrNotFound(result)
}

func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
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
