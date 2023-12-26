package session

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

type SessionDB struct {
	ID         int
	IAT        int64
	Expiration int64
	UserID     string
	UserAgent  string
}

type SessionSQL struct {
	secretKey interface{}
	Sessions  *sql.DB
}

func NewSessionSQLRepo(db *sql.DB) *SessionSQL {
	return &SessionSQL{
		Sessions: db,
	}
}

func (s *SessionSQL) GetExp(ctx context.Context, id string, iat int64) int64 {
	row := s.Sessions.QueryRowContext(ctx, "SELECT expiration FROM sessions WHERE userid = ? AND iat = ?",
		id,
		iat,
	)
	var exp int64
	err := row.Scan(&exp)
	if err == sql.ErrNoRows {
		return 0
	}
	return exp
}

func (s *SessionSQL) AddNewSess(ctx context.Context, id string, exp int64, iat int64) error {
	_, err := s.Sessions.ExecContext(ctx,
		"INSERT INTO sessions (`userid`, `expiration`,`iat`) VALUES (?, ?, ?)",
		id,
		exp,
		iat,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *SessionSQL) GetKey() interface{} {
	return s.secretKey
}

func (s *SessionSQL) SetKey(key interface{}) {
	s.secretKey = []byte(key.(string))
}

func (s *SessionSQL) DownloadKey() error {
	key := os.Getenv("SecretKey")
	if key == "" {
		return fmt.Errorf("secret key not found in environment variables")
	}
	s.secretKey = []byte(key)
	return nil
}

func (s *SessionSQL) DeleteSess(ctx context.Context, userID string, iat int64) error {
	_, err := s.Sessions.ExecContext(ctx,
		"DELETE FROM sessions WHERE userid = ? AND iat = ?",
		userID,
		iat,
	)
	if err != nil {
		return fmt.Errorf("error in session deletekey %w", err)
	}
	return nil
}
