package user

type User struct {
	Username string
	Login    string
	Password string
	ID       string
}

//go:generate mockgen -source user.go -destination user_mock.go -package user UserRepo
type UserRepo interface {
	AddNewUser(user User) (string, error)
	Authenticate(user User) (string, error)
	IsUser(username, id string) (bool, error)
}
