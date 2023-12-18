package session

import (
	"fmt"
	"os"
)

type Session struct {
	secretKey string
}

func (s *Session) GetKey() string {
	return s.secretKey
}

func (s *Session) SetKey(key string) {
	s.secretKey = key
}

func (s *Session) DownloadKey() error {
	key := os.Getenv("SecretKey")
	if key == "" {
		return fmt.Errorf("secret key not found in environment variables")
	}
	s.secretKey = key
	return nil
}
