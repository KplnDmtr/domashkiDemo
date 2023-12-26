package user

import (
	"context"
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

func TestAddNewUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserSQLRepo{
		DB: db,
	}

	username := "user1"
	login := "log"
	paswword := "pass"
	user := User{
		Username: username,
		Login:    login,
		Password: paswword,
	}

	// ok query
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(username, login, paswword).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.AddNewUser(context.Background(), user)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if id != "1" {
		t.Errorf("bad id: want %v, have %v", id, 1)
		return
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// bad query
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(username, login, paswword).
		WillReturnError(fmt.Errorf("db error"))

	_, err = repo.AddNewUser(context.Background(), user)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// last ID error
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(username, login, paswword).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("something wrong")))
	_, err = repo.AddNewUser(context.Background(), user)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAuthenticate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := &UserSQLRepo{
		DB: db,
	}

	username := "user"
	login := "log"
	paswword := "pass"
	user := User{
		Username: username,
		Login:    login,
		Password: paswword,
	}

	rows := sqlmock.NewRows([]string{"id"})

	mock.
		ExpectQuery("SELECT id FROM users WHERE").
		WithArgs(username, user.Password, login).
		WillReturnRows(rows)
	_, err = repo.Authenticate(context.Background(), user)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	rows = sqlmock.NewRows([]string{"id"}).AddRow("1")
	mock.
		ExpectQuery("SELECT id FROM users WHERE").
		WithArgs(username, user.Password, login).
		WillReturnRows(rows)
	res, err := repo.Authenticate(context.Background(), user)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if res != "1" {
		t.Errorf("bad id: want %v, have %v", 1, res)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestIsUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewUserSQLRepo(db)

	username := "user"
	id := "1"

	rows := sqlmock.NewRows([]string{"login"})

	mock.
		ExpectQuery("SELECT login FROM users WHERE").
		WithArgs(username, id).
		WillReturnRows(rows)
	_, err = repo.IsUser(context.Background(), username, id)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	rows = sqlmock.NewRows([]string{"login"}).AddRow("1")
	mock.
		ExpectQuery("SELECT login FROM users WHERE").
		WithArgs(username, id).
		WillReturnRows(rows)
	res, err := repo.IsUser(context.Background(), username, id)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if res != true {
		t.Errorf("bad res: want %v, have %v", true, res)
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}
