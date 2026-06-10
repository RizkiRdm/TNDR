package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBaseClient_DoRequest(t *testing.T) {
	type testRequest struct {
		Input string `json:"input"`
	}
	type testResponse struct {
		Output string `json:"output"`
	}

	tests := []struct {
		name       string
		statusCode int
		respBody   interface{}
		wantErr    error
	}{
		{
			name:       "Success",
			statusCode: http.StatusOK,
			respBody:   testResponse{Output: "hello"},
			wantErr:    nil,
		},
		{
			name:       "Invalid Key",
			statusCode: http.StatusUnauthorized,
			respBody:   nil,
			wantErr:    ErrInvalidKey,
		},
		{
			name:       "Rate Limit",
			statusCode: http.StatusTooManyRequests,
			respBody:   nil,
			wantErr:    ErrRateLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.respBody != nil {
					json.NewEncoder(w).Encode(tt.respBody)
				}
			}))
			defer server.Close()

			client := NewBaseClient()
			var resp testResponse
			err := client.DoRequest(context.Background(), server.URL, nil, testRequest{Input: "test"}, &resp)

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("DoRequest() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("DoRequest() unexpected error: %v", err)
			}

			if resp.Output != "hello" {
				t.Errorf("DoRequest() got = %v, want hello", resp.Output)
			}
		})
	}
}

func TestBaseClient_StreamSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data: chunk1\n\n"))
		w.Write([]byte("data: chunk2\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client := NewBaseClient()
	var received []string
	err := client.StreamSSE(context.Background(), server.URL, nil, nil, func(data []byte) error {
		received = append(received, string(data))
		return nil
	})

	if err != nil {
		t.Fatalf("StreamSSE() unexpected error: %v", err)
	}

	if len(received) != 2 {
		t.Errorf("StreamSSE() received %d chunks, want 2", len(received))
	}

	if received[0] != "chunk1" || received[1] != "chunk2" {
		t.Errorf("StreamSSE() unexpected data: %v", received)
	}
}
