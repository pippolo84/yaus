package yaus

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

type MockBackend struct{}

func (b *MockBackend) Get(key string) (string, error) {
	if key == "test-hash" {
		return "http://www.example.com", nil
	}

	return "", errors.New("not found")
}

func (b *MockBackend) Put(key, _ string) error {
	if key == "test-hash" {
		return nil
	}

	return errors.New("wrong key")
}

type MockShortener struct{}

func (s *MockShortener) Hash(text string) string {
	if text == "http://www.example.com" {
		return "test-hash"
	}

	return ""
}

func TestRedirect(t *testing.T) {
	srv := Server{
		cache:     &MockBackend{},
		shortener: &MockShortener{},
	}

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = mux.SetURLVars(req, map[string]string{
		"hash": "test-hash",
	})

	rr := httptest.NewRecorder()

	handler := srv.redirect()
	handler(rr, req)

	if rr.Result().StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusTemporaryRedirect,
			rr.Result().StatusCode,
		)
	}
	location := rr.Result().Header.Get("Location")
	if location != "http://www.example.com" {
		t.Fatalf("expected Location header %q, got %q", "http://www.example.com", location)
	}
}

func TestShorten(t *testing.T) {
	srv := Server{
		cache:     &MockBackend{},
		shortener: &MockShortener{},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		"/shorten",
		strings.NewReader(`{"url":"http://www.example.com"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := srv.shorten()
	handler(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, rr.Result().StatusCode)
	}

	var resp struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Hash != "test-hash" {
		t.Fatalf("expected hash %q, got %q", "test-hash", resp.Hash)
	}
}
