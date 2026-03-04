package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	validSecret = "a-valid-secret-that-is-32-chars!!"
	otherSecret = "a-different-secret-32-characters!!"
)

func newManager(secret string) *Manager {
	return NewManager(secret, nil)
}

// ── Generate ──────────────────────────────────────────────────────────────────

func TestJWT_Generate_ReturnsNonEmptyToken(t *testing.T) {
	m := newManager(validSecret)
	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())

	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestJWT_Generate_TokenHasThreeParts(t *testing.T) {
	m := newManager(validSecret)
	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())

	require.NoError(t, err)
	count := 0
	for _, c := range token {
		if c == '.' {
			count++
		}
	}
	require.Equal(t, 2, count, "JWT must consist of header.payload.signature")
}

func TestJWT_Generate_DifferentTokensForDifferentInputs(t *testing.T) {
	m := newManager(validSecret)
	t1, _ := m.Generate(uuid.New(), uuid.New(), uuid.New())
	t2, _ := m.Generate(uuid.New(), uuid.New(), uuid.New())

	require.NotEqual(t, t1, t2)
}

// ── Parse ─────────────────────────────────────────────────────────────────────

func TestJWT_Parse_ValidToken(t *testing.T) {
	m := newManager(validSecret)
	userID := uuid.New()
	deviceID := uuid.New()
	sessionID := uuid.New()

	token, err := m.Generate(userID, deviceID, sessionID)
	require.NoError(t, err)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	require.Equal(t, userID.String(), claims["sub"])
	require.Equal(t, deviceID.String(), claims["did"])
	require.Equal(t, sessionID.String(), claims["sid"])
}

func TestJWT_Parse_ClaimsContainExpiry(t *testing.T) {
	m := newManager(validSecret)
	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	exp, ok := claims["exp"]
	require.True(t, ok, "exp claim must be present")
	require.NotNil(t, exp)
	// exp must be in the future.
	expVal, ok := exp.(float64)
	require.True(t, ok)
	require.Greater(t, int64(expVal), time.Now().Unix())
}

func TestJWT_Parse_EmptyString(t *testing.T) {
	m := newManager(validSecret)
	_, err := m.Parse("")

	require.Error(t, err)
}

func TestJWT_Parse_MalformedToken(t *testing.T) {
	m := newManager(validSecret)

	malformed := []string{
		"not-a-jwt",
		"only.two",
		"a.b.c.d.e", // too many parts
		"   ",
	}

	for _, tok := range malformed {
		_, err := m.Parse(tok)
		require.Errorf(t, err, "expected error for token: %q", tok)
	}
}

func TestJWT_Parse_WrongSigningKey(t *testing.T) {
	signer := newManager(validSecret)
	verifier := newManager(otherSecret)

	token, err := signer.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	_, err = verifier.Parse(token)
	require.Error(t, err)
}

func TestJWT_Parse_TamperedPayload(t *testing.T) {
	m := newManager(validSecret)
	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	// Replace the payload section with a different base64 blob.
	parts := splitToken(token)
	require.Len(t, parts, 3)
	tampered := parts[0] + ".dGFtcGVyZWQ." + parts[2] // "tampered" in base64

	_, err = m.Parse(tampered)
	require.Error(t, err)
}

// ── TTL ───────────────────────────────────────────────────────────────────────

func TestJWT_CustomTTL_UsedInGenerate(t *testing.T) {
	// BUG REGRESSION: jwt.Manager.Generate must use m.ttl, not hardcode 24h.
	ttl := 2 * time.Hour
	m := NewManager(validSecret, &ttl)

	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	exp := int64(claims["exp"].(float64))
	now := time.Now().Unix()
	delta := exp - now

	// Should be approximately 2 hours (±30 seconds).
	require.InDelta(t, int64(2*time.Hour/time.Second), delta, 30,
		"token expiry must reflect the configured TTL, not a hardcoded 24h")
}

func TestJWT_DefaultTTL_Is24Hours(t *testing.T) {
	m := NewManager(validSecret, nil)

	token, err := m.Generate(uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	claims, err := m.Parse(token)
	require.NoError(t, err)

	exp := int64(claims["exp"].(float64))
	now := time.Now().Unix()
	delta := exp - now

	require.InDelta(t, int64(24*time.Hour/time.Second), delta, 30)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func splitToken(token string) []string {
	var parts []string
	start := 0
	for i, c := range token {
		if c == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	return append(parts, token[start:])
}
