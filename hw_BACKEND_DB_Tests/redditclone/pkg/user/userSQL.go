package user

import (
	"context"
	"database/sql"
	"strconv"
)

type UserSQLRepo struct {
	DB *sql.DB
}

func NewUserSQLRepo(db *sql.DB) *UserSQLRepo {
	return &UserSQLRepo{
		DB: db,
	}
}

func (m *UserSQLRepo) AddNewUser(ctx context.Context, user User) (string, error) {
	result, err := m.DB.ExecContext(ctx,
		"INSERT INTO users (`username`, `login`, `password`) VALUES (?, ?, ?)",
		user.Username,
		user.Login,
		user.Password,
	)
	if err != nil {
		return "", err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(lastID)), nil
}

func (m *UserSQLRepo) Authenticate(ctx context.Context, user User) (string, error) {
	row := m.DB.QueryRowContext(ctx, "SELECT id FROM users WHERE username = ? AND password = ? AND login = ?",
		user.Username,
		user.Password,
		user.Login,
	)
	var id string
	err := row.Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil

}

func (m *UserSQLRepo) IsUser(ctx context.Context, username string, id string) (bool, error) {
	row := m.DB.QueryRowContext(ctx, "SELECT login FROM users WHERE username = ? AND id = ?",
		username,
		id,
	)
	var login string
	err := row.Scan(&login)
	if err != nil {
		return false, err
	}
	//
	return true, nil
}
