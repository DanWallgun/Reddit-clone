package user

import (
	"database/sql"
	"errors"
	"redditclone/pkg/utils"
)

type UsersRepo struct {
	DB *sql.DB
}

func NewRepo(db *sql.DB) *UsersRepo {
	return &UsersRepo{
		DB: db,
	}
}

var (
	ErrNoUser            = errors.New("no user found")
	ErrBadPass           = errors.New("invalid password")
	ErrAlreadyRegistered = errors.New("this username is already in use")
)

func (repo *UsersRepo) Register(login, password string) (*User, error) {
	u := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password FROM users WHERE login = ?", login).
		Scan(&u.ID, &u.Login, &u.password)
	if err == nil {
		return nil, ErrAlreadyRegistered
	}

	u.ID = utils.GenerateRandomString(32)
	u.Login = login
	u.password = password

	_, err = repo.DB.Exec(
		"INSERT INTO users VALUES",
		u.ID,
		u.Login,
		u.password,
	)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (repo *UsersRepo) Authorize(login, password string) (*User, error) {
	u := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password FROM users WHERE login = ?", login).
		Scan(&u.ID, &u.Login, &u.password)
	if err != nil {
		return nil, ErrNoUser
	}

	if u.password != password {
		return nil, ErrBadPass
	}

	return u, nil
}

func (repo *UsersRepo) GetByID(id string) (*User, error) {
	u := &User{}
	err := repo.DB.
		QueryRow("SELECT id, login, password FROM users WHERE id = ?", id).
		Scan(&u.ID, &u.Login, &u.password)
	if err != nil {
		return nil, ErrNoUser
	}

	return u, nil
}
