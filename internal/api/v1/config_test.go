package v1_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ipfs-force-community/threadmirror/internal/testsuit"
	"github.com/stretchr/testify/assert"
)

func TestV1Handler_GetConfigSupabase(t *testing.T) {
	db := testsuit.SetupTestDB(t)
	router := testsuit.SetupTestServer(t, db)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/config/supabase", nil)

	// Create response recorder
	w := httptest.NewRecorder()

	// Execute through router
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Check response format
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")

	// Check data content
	data := response["data"].(map[string]interface{})
	assert.Equal(t, "test-project-ref", data["project_reference"])
	assert.Equal(t, "test-api-key", data["api_anno_key"])

	// Check bucket_names
	assert.Contains(t, data, "bucket_names")
	bucketNames := data["bucket_names"].(map[string]interface{})
	assert.Equal(t, "avatar", bucketNames["avatar"])
	assert.Equal(t, "post-images", bucketNames["post_images"])
}
