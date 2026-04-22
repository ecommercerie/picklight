package poller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPoll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"stats": map[string]interface{}{"orders_pending": 7},
		})
	}))
	defer server.Close()

	val, err := Poll(server.URL, "stats.orders_pending", false)
	if err != nil {
		t.Fatal(err)
	}
	if val != 7 {
		t.Errorf("expected 7, got %d", val)
	}
}

func TestPollNestedPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"a": map[string]interface{}{"b": map[string]interface{}{"c": 42}},
		})
	}))
	defer server.Close()

	val, err := Poll(server.URL, "a.b.c", false)
	if err != nil {
		t.Fatal(err)
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestPollKeyNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"foo": 1})
	}))
	defer server.Close()

	_, err := Poll(server.URL, "bar", false)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestPollHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	_, err := Poll(server.URL, "x", false)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestExtractPath(t *testing.T) {
	data := map[string]interface{}{
		"stats": map[string]interface{}{"orders_pending": float64(3)},
	}
	val, err := extractPath(data, "stats.orders_pending")
	if err != nil {
		t.Fatal(err)
	}
	if val != float64(3) {
		t.Errorf("expected 3, got %v", val)
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected int
		wantErr  bool
	}{
		{float64(5), 5, false},
		{int(3), 3, false},
		{"10", 10, false},
		{"abc", 0, true},
		{true, 0, true},
	}
	for _, tt := range tests {
		got, err := toInt(tt.input)
		if tt.wantErr && err == nil {
			t.Errorf("expected error for %v", tt.input)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("unexpected error for %v: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("toInt(%v) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
