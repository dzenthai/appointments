package user

import (
	"appointments/internal/token"
	"appointments/internal/validator"
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicated email")
	ErrUserNotFound   = errors.New("user not found")
	ErrEditConflict   = errors.New("edit conflict")
)

var AnonymousUser = &User{}

type Role string

const (
	RoleClient   = Role("client")
	RoleProvider = Role("provider")
	RoleAdmin    = Role("admin")
)

type User struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"first_name"`
	SecondName string    `json:"second_name"`
	Email      string    `json:"email"`
	Password   password  `json:"-"`
	Role       Role      `json:"role"`
	Verified   bool      `json:"verified"`
	CreatedAt  time.Time `json:"created_at"`
	Version    int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

func (user *User) IsAnonymous() bool {
	return user == AnonymousUser
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}

	p.hash = hash
	p.plaintext = &plaintext

	return nil
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

	v.Check(user.Role != "", "role", "must be provided")
	v.Check(checkUserRole(user.Role), "role", "invalid user role")
}

func checkUserRole(role Role) bool {
	switch role {
	case RoleClient:
		return true
	case RoleProvider:
		return true
	default:
		return false
	}
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetByToken(plaintext string, scope token.Scope) (*User, error) {

	tokenHash := sha256.Sum256([]byte(plaintext))

	query := `
		SELECT users.id, users.first_name, users.second_name, users.email, users.password_hash, users.role, users.verified, users.created_at, users.version
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.token_hash = $1
		AND tokens.scope = $2
		AND tokens.expires_at > $3`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var user User

	err := s.db.QueryRowContext(ctx, query, tokenHash[:], scope, time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.SecondName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *Store) GetByEmail(email string) (*User, error) {
	query :=
		`SELECT id, first_name, second_name, email, password_hash, role, verified, created_at, version FROM users WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	var user User

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.SecondName,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Verified,
		&user.CreatedAt,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (s *Store) Insert(user *User) error {
	query :=
		`INSERT INTO users (first_name, second_name, email, password_hash, role)  
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, verified, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []any{user.FirstName, user.SecondName, user.Email, user.Password.hash, user.Role}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Verified, &user.CreatedAt, &user.Version)
	if err != nil {
		pgErr, exist := errors.AsType[*pgconn.PgError](err)
		if exist && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (s *Store) Update(user *User) error {
	query :=
		`UPDATE users
		SET first_name = $1,
		second_name = $2,
		email = $3,
		password_hash = $4,
		role = $5,
		verified = $6,
		version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []any{user.FirstName, user.SecondName, user.Email, user.Password.hash, user.Role, user.Verified, user.ID, user.Version}

	err := s.db.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		pgErr, exist := errors.AsType[*pgconn.PgError](err)
		if exist && pgErr.Code == "23505" {
			return ErrDuplicateEmail
		}
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
