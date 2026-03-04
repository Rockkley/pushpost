package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rockkley/pushpost/services/auth_service/internal/domain"
	"github.com/stretchr/testify/require"
)

func newSession(sessionID, userID uuid.UUID) *domain.Session {
	return &domain.Session{
		SessionID: sessionID,
		UserID:    userID,
		DeviceID:  uuid.New(),
		Expires:   time.Now().Add(time.Hour).Unix(),
	}
}

// ── Save & Get ────────────────────────────────────────────────────────────────

func TestSessionStore_Save_Get_Success(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	s := newSession(sessionID, userID)

	require.NoError(t, store.Save(ctx, s))

	got, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, sessionID, got.SessionID)
	require.Equal(t, userID, got.UserID)
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	store := NewSessionStore()

	_, err := store.Get(context.Background(), uuid.New())

	require.ErrorIs(t, err, ErrSessionNotFound)
}

func TestSessionStore_Save_OverwritesExistingSession(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()
	sessionID := uuid.New()

	first := newSession(sessionID, uuid.New())
	require.NoError(t, store.Save(ctx, first))

	second := newSession(sessionID, uuid.New()) // same session ID, different user
	require.NoError(t, store.Save(ctx, second))

	got, err := store.Get(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, second.UserID, got.UserID, "second save must overwrite the first")
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestSessionStore_Delete_Success(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()
	sessionID := uuid.New()

	require.NoError(t, store.Save(ctx, newSession(sessionID, uuid.New())))
	require.NoError(t, store.Delete(ctx, sessionID))

	_, err := store.Get(ctx, sessionID)
	require.ErrorIs(t, err, ErrSessionNotFound, "session must not exist after deletion")
}

func TestSessionStore_Delete_NonExistent_NoError(t *testing.T) {
	store := NewSessionStore()

	// Deleting a session that was never saved must not return an error.
	err := store.Delete(context.Background(), uuid.New())
	require.NoError(t, err)
}

func TestSessionStore_Delete_TwiceSameID_NoError(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()
	sessionID := uuid.New()

	require.NoError(t, store.Save(ctx, newSession(sessionID, uuid.New())))
	require.NoError(t, store.Delete(ctx, sessionID))
	require.NoError(t, store.Delete(ctx, sessionID)) // idempotent
}

// ── Multiple Sessions ─────────────────────────────────────────────────────────

func TestSessionStore_MultipleSessionsIndependent(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()

	ids := make([]uuid.UUID, 5)
	for i := range ids {
		ids[i] = uuid.New()
		require.NoError(t, store.Save(ctx, newSession(ids[i], uuid.New())))
	}

	// Delete the middle session; others must survive.
	require.NoError(t, store.Delete(ctx, ids[2]))

	for i, id := range ids {
		_, err := store.Get(ctx, id)
		if i == 2 {
			require.ErrorIs(t, err, ErrSessionNotFound)
		} else {
			require.NoError(t, err)
		}
	}
}

// ── Concurrency ───────────────────────────────────────────────────────────────

func TestSessionStore_ConcurrentSaveGet(t *testing.T) {
	// Run with -race to detect data races.
	store := NewSessionStore()
	ctx := context.Background()
	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			id := uuid.New()
			s := newSession(id, uuid.New())
			_ = store.Save(ctx, s)
			_, _ = store.Get(ctx, id)
			_ = store.Delete(ctx, id)
		}()
	}

	wg.Wait()
}

func TestSessionStore_ConcurrentSaveAndRead(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()
	sessionID := uuid.New()

	require.NoError(t, store.Save(ctx, newSession(sessionID, uuid.New())))

	var wg sync.WaitGroup
	const readers = 30

	wg.Add(readers)
	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			got, err := store.Get(ctx, sessionID)
			require.NoError(t, err)
			require.Equal(t, sessionID, got.SessionID)
		}()
	}

	wg.Wait()
}

// ── Data Integrity ────────────────────────────────────────────────────────────

func TestSessionStore_ReturnedSession_MatchesSaved(t *testing.T) {
	store := NewSessionStore()
	ctx := context.Background()

	original := &domain.Session{
		SessionID: uuid.New(),
		UserID:    uuid.New(),
		DeviceID:  uuid.New(),
		Expires:   time.Now().Add(48 * time.Hour).Unix(),
	}

	require.NoError(t, store.Save(ctx, original))

	got, err := store.Get(ctx, original.SessionID)
	require.NoError(t, err)
	require.Equal(t, original.SessionID, got.SessionID)
	require.Equal(t, original.UserID, got.UserID)
	require.Equal(t, original.DeviceID, got.DeviceID)
	require.Equal(t, original.Expires, got.Expires)
}
