package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Meta is public: no auth, and it must not fail even before web services init.
func TestHandleMeta_PublicAndSafeWithoutInit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/meta", nil)
	rec := httptest.NewRecorder()

	HandleMeta(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var out metaResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.NotNil(t, out)
}
