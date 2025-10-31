package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type AuthService struct {
	users    map[string]string
	sessions map[string]string
	mu       sync.RWMutex
}

func NewAuthService() *AuthService {
	return &AuthService{
		users:    map[string]string{"admin": "admin"},
		sessions: make(map[string]string),
	}
}

func (a *AuthService) Login(username, password string) (string, error) {
	a.mu.RLock()
	pwd, ok := a.users[username]
	a.mu.RUnlock()
	if !ok || pwd != password {
		return "", ErrInvalidCreds
	}
	sid := newSessionID()
	a.mu.Lock()
	a.sessions[sid] = username
	a.mu.Unlock()
	return sid, nil
}

func (a *AuthService) Logout(sessionID string) {
	a.mu.Lock()
	delete(a.sessions, sessionID)
	a.mu.Unlock()
}

func (a *AuthService) Who(sessionID string) (string, error) {
	a.mu.RLock()
	u := a.sessions[sessionID]
	a.mu.RUnlock()
	if u == "" {
		return "", ErrUnauthorized
	}
	return u, nil
}

func newSessionID() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
