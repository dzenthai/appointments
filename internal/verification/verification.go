package verification

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"
)

type Verification struct {
	UserID int64
	Code
	TTL time.Time
}

type Code struct {
	plaintext string
	hash      []byte
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func NewCode(userID int64, ttl time.Duration) (*Verification, error) {
	vry := &Verification{
		UserID: userID,
		TTL:    time.Now().Add(ttl),
	}

	randBytes := make([]byte, 16)

	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}

	vry.Code.plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)

	hash := sha256.Sum256([]byte(vry.Code.plaintext))

	vry.Code.hash = hash[:]

	return vry, nil
}

func (c Code) Plaintext() string {
	return c.plaintext
}

func (s *Store) Create(v *Verification) error {
	query :=
		`INSERT INTO verifications (user_id, code_hash, ttl)
		VALUES ($1, $2, $3) 
		ON CONFLICT (user_id) 
      	DO UPDATE SET code_hash = EXCLUDED.code_hash, ttl = EXCLUDED.ttl`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []any{v.UserID, v.Code.hash, v.TTL}

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) DeleteAllByUserID(userID int64) error {
	query := `DELETE FROM verifications WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
