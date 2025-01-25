package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/noonacedia/cinematrique/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(u *User) error {
	query := `
  INSERT INTO users (name, email, password_hash, activated)
  VALUES ($1, $2, $3, $4)
  RETURNING id, created_at, version
  `
	args := []interface{}{u.Name, u.Email, u.Password.hash, u.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&u.ID, &u.CreatedAt, &u.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
  SELECT id, name, email, password_hash, activated, version, created_at
  FROM users WHERE email == $1
  `
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
		&user.CreatedAt,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) Update(u *User) error {
	query := `
  UPDATE users
  SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
  WHERE id = $5 AND version = $6
  RETURNING version
  `
	args := []interface{}{u.Name, u.Email, u.Password.hash, u.Activated, u.ID, u.Version}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&u.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

type MockUserModel struct{}

func (m MockUserModel) Insert(u *User) error {
	return nil
}

func (m MockUserModel) GetByEmail(email string) (*User, error) {
	return &User{}, nil
}

func (m MockUserModel) Update(u *User) error {
	return nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "email cannot be empty")
	v.Check(validator.Matches(email, validator.EmailRx), "email", "email should be in common format")
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(name != "", "name", "name cannot be empty")
	v.Check(len(name) <= 500, "name", "name is too long")
}

func ValidatePasswordPlainText(v *validator.Validator, password string) {
	v.Check(password != "", "password", "password must be presented")
	v.Check(
		len(password) >= 8 && len(password) <= 72,
		"password", "password should have between 8 and 72 bytes",
	)
}

func ValidateUser(v *validator.Validator, user *User) {
	ValidateName(v, user.Name)
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlainText(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
