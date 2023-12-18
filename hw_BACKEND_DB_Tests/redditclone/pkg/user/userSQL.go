package user

import (
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

func (m *UserSQLRepo) AddNewUser(user User) (string, error) {
	result, err := m.DB.Exec(
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

func (m *UserSQLRepo) Authenticate(user User) (string, error) {
	row := m.DB.QueryRow("SELECT id FROM users WHERE username = ? AND password = ? AND login = ?",
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

func (m *UserSQLRepo) IsUser(username, id string) (bool, error) {
	row := m.DB.QueryRow("SELECT login FROM users WHERE username = ? AND id = ?",
		username,
		id,
	)
	var login string
	err := row.Scan(&login)
	if err != nil {
		return false, err
	}
	return true, nil
}
