package memory

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/rockkley/pushpost/internal/services/auth_service/internal/domain"
	"sync"
)

var ErrSessionNotFound = errors.New("session not found")

type SessionStore struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*domain.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		data: make(map[uuid.UUID]*domain.Session),
	}
}

func (s *SessionStore) Save(ctx context.Context, session *domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[session.SessionID] = session
	return nil
}

func (s *SessionStore) Get(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.data[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, sessionID)
	return nil
}
