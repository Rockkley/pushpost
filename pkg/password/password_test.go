package password

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword_Hash_ReturnsNonEmptyString(t *testing.T) {
	hash, err := Hash("SomePassword1")

	require.NoError(t, err)
	require.NotEmpty(t, hash)
}

func TestPassword_Hash_Is60Chars(t *testing.T) {
	hash, err := Hash("SomePassword1")

	require.NoError(t, err)
	require.Len(t, hash, 60, "bcrypt hash at DefaultCost must be exactly 60 chars")
}

func TestPassword_Hash_IsValidBcrypt(t *testing.T) {
	const plain = "ValidPass99"
	hash, err := Hash(plain)
	require.NoError(t, err)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	require.NoError(t, err, "produced hash must be verifiable by bcrypt directly")
}

func TestPassword_Hash_DifferentHashesForSameInput(t *testing.T) {
	// bcrypt includes a random salt, so two hashes of the same input differ.
	h1, err1 := Hash("SamePassword1")
	h2, err2 := Hash("SamePassword1")

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotEqual(t, h1, h2, "bcrypt must produce different hashes due to random salt")
}

func TestPassword_Hash_DoesNotStoreOrReturnPlaintext(t *testing.T) {
	const plain = "Secret123"
	hash, err := Hash(plain)

	require.NoError(t, err)
	require.NotContains(t, hash, plain, "hash must not contain the plaintext password")
}

func TestPassword_Compare_CorrectPassword(t *testing.T) {
	const plain = "CorrectPass1"
	hash, err := Hash(plain)
	require.NoError(t, err)

	err = Compare(plain, hash)
	require.NoError(t, err)
}

func TestPassword_Compare_WrongPassword(t *testing.T) {
	hash, err := Hash("CorrectPass1")
	require.NoError(t, err)

	err = Compare("WrongPass1", hash)
	require.Error(t, err)
}

func TestPassword_Compare_EmptyPassword(t *testing.T) {
	hash, err := Hash("ActualPass1")
	require.NoError(t, err)

	err = Compare("", hash)
	require.Error(t, err)
}

func TestPassword_Compare_EmptyHash(t *testing.T) {
	err := Compare("SomePass1", "")
	require.Error(t, err)
}

func TestPassword_Compare_HashFromExternalBcrypt(t *testing.T) {
	// Ensure Compare works with a hash produced outside of our package.
	const plain = "ExtHash99"
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	require.NoError(t, err)

	err = Compare(plain, string(b))
	require.NoError(t, err)
}

func TestPassword_Compare_LongPassword(t *testing.T) {
	// bcrypt silently truncates passwords > 72 bytes. Verify no panic occurs.
	longPass := string(make([]byte, 100))
	for i := range longPass {
		longPass = longPass[:i] + "A" + longPass[i+1:]
	}
	hash, err := Hash(longPass)
	require.NoError(t, err)

	err = Compare(longPass, hash)
	require.NoError(t, err)
}
