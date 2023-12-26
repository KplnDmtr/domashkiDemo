package user

import "context"

type User struct {
	Username string
	Login    string
	Password string
	ID       string
}

//go:generate mockgen -source user.go -destination user_mock.go -package user UserRepo
type UserRepo interface {
	AddNewUser(ctx context.Context, user User) (string, error)
	Authenticate(ctx context.Context, user User) (string, error)
	IsUser(ctx context.Context, username string, id string) (bool, error)
}
