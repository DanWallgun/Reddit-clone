package user

import (
	"fmt"
	"redditclone/pkg/utils"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	elemID := "qwertyID"

	// good query
	rows := sqlmock.
		NewRows([]string{"id", "login", "password"})
	expect := []*User{
		{"CoolGuy", elemID, "CoolPass"},
	}
	for _, item := range expect {
		rows = rows.AddRow(item.ID, item.Login, item.password)
	}

	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemID).
		WillReturnRows(rows)

	repo := &UsersRepo{
		DB: db,
	}
	item, err := repo.GetByID(elemID)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemID).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.GetByID(elemID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// row scan error
	rows = sqlmock.NewRows([]string{"id", "login"}).
		AddRow(elemID, "CoolGuy")

	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemID).
		WillReturnRows(rows)

	_, err = repo.GetByID(elemID)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	elemID := "qwertyID"
	elemLogin := "CoolGuy"
	elemPassword := "CoolPass"

	// good query
	rows := sqlmock.
		NewRows([]string{"id", "login", "password"})
	expect := []*User{
		{elemLogin, elemID, elemPassword},
	}
	for _, item := range expect {
		rows = rows.AddRow(item.ID, item.Login, item.password)
	}

	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemLogin).
		WillReturnRows(rows)

	repo := &UsersRepo{
		DB: db,
	}
	item, err := repo.Authorize(elemLogin, elemPassword)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(item, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], item)
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemLogin).
		WillReturnError(fmt.Errorf("db_error"))

	_, err = repo.Authorize(elemLogin, elemPassword)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// bad pass error
	rows = sqlmock.NewRows([]string{"id", "login", "password"}).
		AddRow(elemID, elemLogin, elemPassword)

	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemLogin).
		WillReturnRows(rows)

	_, err = repo.Authorize(elemLogin, "12345678")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}

func TestRegister(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	utils.IsRandom = false
	defer func() { utils.IsRandom = true }()

	elemID := "Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk="
	elemLogin := "CoolGuy"
	elemPassword := "CoolPass"

	// already registered
	rows := sqlmock.
		NewRows([]string{"id", "login", "password"})
	expect := []*User{
		{elemLogin, elemID, elemPassword},
	}
	for _, item := range expect {
		rows = rows.AddRow(item.ID, item.Login, item.password)
	}

	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemLogin).
		WillReturnRows(rows)

	repo := NewRepo(db)
	_, err = repo.Register(elemLogin, elemPassword)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// good query
	mock.
		ExpectExec("INSERT INTO users VALUES").
		WithArgs(elemID, elemLogin, elemPassword).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = repo.Register(elemLogin, elemPassword)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	// err query
	mock.
		ExpectQuery("SELECT id, login, password FROM users WHERE").
		WithArgs(elemLogin).
		WillReturnError(fmt.Errorf("No user found"))

	_, err = repo.Register(elemLogin, elemPassword)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
}
