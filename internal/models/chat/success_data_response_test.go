package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestModelSuccessDataResponseContract(t *testing.T) {
	t.Setenv("SSRF_WHITELIST", "127.0.0.1")
	t.Setenv("SSRF_WHITELIST_EXTRA", "")
	utils.ResetSSRFWhitelistForTest()
	t.Cleanup(utils.ResetSSRFWhitelistForTest)
	const completion = `{"choices":[{"message":{"role":"assistant","content":"Answer","reasoning_content":"Thought","tool_calls":[{"id":"call-1","type":"function","function":{"name":"lookup","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`
	cases := []struct {
		name, body string
		wrapped    bool
		wantError  bool
	}{
		{"standard stays standard", completion, false, false},
		{"explicit success data", `{"success":true,"data":` + completion + `}`, true, false},
		{"no implicit unwrapping", `{"success":true,"data":` + completion + `}`, false, true},
		{"declared envelope is required", completion, true, true},
		{"failure with data is rejected", `{"success":false,"data":` + completion + `}`, true, true},
		{"missing success is rejected", `{"data":` + completion + `}`, true, true},
		{"null data is rejected", `{"success":true,"data":null}`, true, true},
		{"string data is rejected", `{"success":true,"data":"invalid"}`, true, true},
		{"empty data is rejected", `{"success":true,"data":{}}`, true, true},
		{"conflicting error is rejected", `{"success":true,"error":"provider failed","data":` + completion + `}`, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			model := &types.Model{Name: "probe", Type: types.ModelTypeKnowledgeQA, Source: types.ModelSourceRemote,
				Parameters: types.ModelParameters{BaseURL: server.URL, Provider: "generic",
					Spec: &types.ModelSpecOverride{Compat: map[string]any{"success_data_envelope": tc.wrapped}}}}
			client, err := NewChat(ConfigFromModel(model, "", ""), nil)
			require.NoError(t, err)
			response, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "test"}}, nil)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "Answer", response.Content)
			require.Equal(t, "Thought", response.ReasoningContent)
			require.Equal(t, "tool_calls", response.FinishReason)
			require.Len(t, response.ToolCalls, 1)
			require.Equal(t, "lookup", response.ToolCalls[0].Function.Name)
		})
	}
}

func TestModelSuccessDataResponseKeepsStandardStreaming(t *testing.T) {
	t.Setenv("SSRF_WHITELIST", "127.0.0.1")
	t.Setenv("SSRF_WHITELIST_EXTRA", "")
	utils.ResetSSRFWhitelistForTest()
	t.Cleanup(utils.ResetSSRFWhitelistForTest)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"OK\"}}]}\n\ndata: [DONE]\n\n"))
	}))
	defer server.Close()
	model := &types.Model{Name: "probe", Type: types.ModelTypeKnowledgeQA, Source: types.ModelSourceRemote,
		Parameters: types.ModelParameters{BaseURL: server.URL, Provider: "generic",
			Spec: &types.ModelSpecOverride{Compat: map[string]any{"success_data_envelope": true}}}}
	client, err := NewChat(ConfigFromModel(model, "", ""), nil)
	require.NoError(t, err)
	stream, err := client.ChatStream(context.Background(), []Message{{Role: "user", Content: "test"}}, nil)
	require.NoError(t, err)
	var answer strings.Builder
	for chunk := range stream {
		if chunk.ResponseType == types.ResponseTypeAnswer {
			answer.WriteString(chunk.Content)
		}
	}
	require.Equal(t, "OK", answer.String())
}
