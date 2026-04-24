package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type payload struct {
	Ts time.Time `json:"ts"`
	ID uuid.UUID `json:"id"`
}

func Encode(secret []byte, ts time.Time, id uuid.UUID) (string, error) {
	b, err := json.Marshal(payload{Ts: ts, ID: id})
	if err != nil {
		return "", fmt.Errorf("cursor encode: %w", err)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(b)
	sig := mac.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(b) +
		"." +
		base64.RawURLEncoding.EncodeToString(sig), nil
}

func Decode(secret []byte, token string) (time.Time, uuid.UUID, error) {
	if token == "" {
		return time.Time{}, uuid.Nil, fmt.Errorf("empty cursor")
	}

	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor format")
	}

	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor payload encoding")
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid cursor signature encoding")
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(b)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return time.Time{}, uuid.Nil, fmt.Errorf("cursor signature invalid")
	}

	var p payload
	if err = json.Unmarshal(b, &p); err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("cursor decode: %w", err)
	}
	return p.Ts, p.ID, nil
}

func Sentinel() (time.Time, uuid.UUID) {
	return time.Now().Add(24 * time.Hour), uuid.Max
}
