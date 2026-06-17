package user

import (
	"appointments/internal/validator"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicated email")
)

type User struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"first_name"`
	SecondName string    `json:"second_name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Activated  bool      `json:"activated"`
	CreatedAt  time.Time `json:"created_at"`
	Version    int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}

	p.hash = hash
	p.plaintext = &plaintext

	return err
}

func (p *password) Matches(plaintext string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
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

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePassword(v *validator.Validator, password string) {
	key := "password"
	v.Check(password != "", key, "must be provided")
	v.Check(len(password) >= 8, key, "must be at least 8 bytes long")
	v.Check(len(password) <= 72, key, "must not be more than 500 bytes long")
}

func ValidateUser(v *validator.Validator, user User) {
	v.Check(user.FirstName != "", "first_name", "must be provided")
	v.Check(user.SecondName != "", "second_name", "must be provided")

	v.Check(len(user.FirstName) <= 500, "first_name", "must not be more than 500 bytes long")
	v.Check(len(user.SecondName) <= 500, "second_name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Insert(user *User) error {
	query :=
		`INSERT INTO users (first_name, second_name, email, password_hash)  
		VALUES ($1, $2, $3, $4)
		RETURNING id, activated, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []any{user.FirstName, user.SecondName, user.Email, user.Password.hash}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Activated, &user.CreatedAt, &user.Version)
	if err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr):
			if pgErr.Code == "23505" {
				return ErrDuplicateEmail
			}
		default:
			return err
		}
	}
	return nil
}
