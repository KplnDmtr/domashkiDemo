package session

//go:generate mockgen -source manager.go -destination manager_mock.go -package session SessionManager
type SessionManager interface {
	GetKey() interface{}
	SetKey(interface{})
	AddNewSess(id string, exp int64, iat int64) error
	GetExp(id string, iat int64) int64
	DeleteSess(userID string, iat int64) error
}
