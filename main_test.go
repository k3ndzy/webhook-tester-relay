package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStore_ListAndAdd(t *testing.T) {
	s := NewStore()
	items := s.List()
	if len(items) == 0 {
		t.Fatal("Expected default items to be populated")
	}

	initialCount := len(items)
	newItem := Item{
		Name:     "Test Automated Entry",
		Category: "Unit Testing",
		Status:   "Verified",
		Detail:   "Verification test payload",
	}
	added := s.Add(newItem)
	if added.ID == "" {
		t.Errorf("Expected generated ID, got empty string")
	}

	afterCount := len(s.List())
	if afterCount != initialCount+1 {
		t.Errorf("Expected count %d, got %d", initialCount+1, afterCount)
	}
}

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse json: %v", err)
	}

	if body["status"] != "UP" {
		t.Errorf("Expected status UP, got %v", body["status"])
	}
}

func TestItemsEndpoint_Get(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	rr := httptest.NewRecorder()

	handleItems(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var items []Item
	if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
		t.Fatalf("Failed to parse items json: %v", err)
	}
	if len(items) == 0 {
		t.Errorf("Expected non-empty items slice")
	}
}

func TestItemsEndpoint_Post(t *testing.T) {
	payload := `{
		"name": "Integration Test Model",
		"category": "Testing",
		"status": "Ready",
		"detail": "Automated payload"
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/items", bytes.NewBufferString(payload))
	rr := httptest.NewRecorder()

	handleItems(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var created Item
	if err := json.Unmarshal(rr.Body.Bytes(), &created); err != nil {
		t.Fatalf("Failed to unmarshal created item: %v", err)
	}
	if created.Name != "Integration Test Model" {
		t.Errorf("Expected name match, got %s", created.Name)
	}
}

func TestStatsEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rr := httptest.NewRecorder()

	handleStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}
}
