package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func runConnectionProbe(t *testing.T, upstream http.Handler, compat map[string]any) (bool, string) {
	t.Helper()
	t.Setenv("SSRF_WHITELIST", "127.0.0.1")
	t.Setenv("SSRF_WHITELIST_EXTRA", "")
	t.Setenv("SSRF_DNS_WHITELIST_ONLY", "false")
	utils.ResetSSRFWhitelistForTest()
	t.Cleanup(utils.ResetSSRFWhitelistForTest)
	server := httptest.NewServer(upstream)
	t.Cleanup(server.Close)
	payload, err := json.Marshal(ModelTestRequest{
		ModelName: "probe", BaseURL: server.URL, Provider: "generic",
		Spec: &types.ModelSpecOverride{Compat: compat},
	})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/initialization/remote/check", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), types.TenantInfoContextKey, &types.Tenant{ID: 1}))
	(&InitializationHandler{}).CheckRemoteModel(c)
	require.Empty(t, c.Errors)
	require.Equal(t, http.StatusOK, recorder.Code)
	var response struct {
		Data struct {
			Available bool   `json:"available"`
			Message   string `json:"message"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return response.Data.Available, response.Data.Message
}

func TestCheckRemoteModelLeavesRoomForVisibleOutput(t *testing.T) {
	available, message := runConnectionProbe(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			MaxTokens int `json:"max_tokens"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.Header().Set("Content-Type", "application/json")
		// The live Cline endpoint returns this error when the one-token probe
		// is exhausted before any visible content is produced.
		if body.MaxTokens <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"success":false,"error":"empty response content"}`))
			return
		}
		require.LessOrEqual(t, body.MaxTokens, 1024, "connection checks must stay bounded")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}`))
	}), nil)
	require.True(t, available, message)
}

func TestCheckRemoteModelWithExplicitSuccessDataResponse(t *testing.T) {
	available, message := runConnectionProbe(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"choices":[{"message":{"role":"assistant","content":"OK"},"finish_reason":"stop"}]}}`))
	}), map[string]any{"success_data_envelope": true})
	require.True(t, available, message)
}

func TestCheckRemoteModelRejectsEmptyReply(t *testing.T) {
	available, _ := runConnectionProbe(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":""},"finish_reason":"length"}]}`))
	}), nil)
	require.False(t, available, "an empty completion must not be reported as a usable model")
}
