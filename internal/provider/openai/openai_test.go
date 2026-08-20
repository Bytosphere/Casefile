package openai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"casefile/internal/provider"
)

func newTestProvider(t *testing.T, handler http.HandlerFunc) *Provider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &Provider{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-test",
		Client:  server.Client(),
	}
}

func TestComplete_ContentOnly(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"  final answer  "}}]}`))
	})

	res, err := p.Complete(context.Background(), provider.Request{
		Messages: []provider.Message{{Role: provider.RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if res.Content != "final answer" {
		t.Errorf("expected trimmed content, got %q", res.Content)
	}
	if len(res.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(res.ToolCalls))
	}
}

func TestComplete_ToolCallsPresent(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"","tool_calls":[
			{"id":"call_1","type":"function","function":{"name":"casefile_search","arguments":"{\"pattern\":\"foo\"}"}}
		]}}]}`))
	})

	res, err := p.Complete(context.Background(), provider.Request{
		Messages: []provider.Message{{Role: provider.RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(res.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(res.ToolCalls))
	}
	tc := res.ToolCalls[0]
	if tc.ID != "call_1" || tc.Name != "casefile_search" {
		t.Errorf("unexpected tool call: %+v", tc)
	}
	if tc.Arguments["pattern"] != "foo" {
		t.Errorf("expected pattern=foo, got %v", tc.Arguments)
	}
}

func TestComplete_NonSuccessStatus(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
	})

	_, err := p.Complete(context.Background(), provider.Request{
		Messages: []provider.Message{{Role: provider.RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected an error for non-2xx status")
	}
}

func TestComplete_EmptyChoices(t *testing.T) {
	p := newTestProvider(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	})

	_, err := p.Complete(context.Background(), provider.Request{
		Messages: []provider.Message{{Role: provider.RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected an error for empty choices")
	}
}
