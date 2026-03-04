package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func newJSONRequest(t *testing.T, method, path, payload string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeJSONResponse(t *testing.T, resp *httptest.ResponseRecorder, v any) {
	t.Helper()
	require.Equal(t, "application/json", resp.Header().Get("Content-Type"))
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
}
