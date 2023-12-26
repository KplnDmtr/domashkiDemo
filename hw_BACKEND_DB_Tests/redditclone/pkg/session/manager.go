package session

import "context"

//go:generate mockgen -source manager.go -destination manager_mock.go -package session SessionManager
type SessionManager interface {
	GetKey() interface{}
	SetKey(interface{})
	AddNewSess(ctx context.Context, id string, exp int64, iat int64) error
	GetExp(ctx context.Context, id string, iat int64) int64
	DeleteSess(ctx context.Context, userID string, iat int64) error
}
