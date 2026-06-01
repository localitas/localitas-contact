package contact

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSwagger_ReturnsValidJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/swagger.json", nil)
	w := httptest.NewRecorder()

	HandleSwagger(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %q", ct)
	}

	var spec APIDoc
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if spec.AppName != "Contacts" {
		t.Errorf("expected app_name Contacts, got %q", spec.AppName)
	}

	expected := []string{"/api/contacts", "/api/search"}
	for _, p := range expected {
		found := false
		for _, ep := range spec.Endpoints {
			if strings.Contains(ep.Path, p) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected endpoint containing %s", p)
		}
	}
}

func TestRenderDocsHTML_ContainsFieldsAndEndpoints(t *testing.T) {
	html := string(RenderDocsHTML(ContactAPIDoc))

	if !strings.Contains(html, "YAML Fields Reference") {
		t.Error("expected YAML Fields Reference heading")
	}
	if !strings.Contains(html, "API Endpoints") {
		t.Error("expected API Endpoints heading")
	}
	if !strings.Contains(html, "full_name") {
		t.Error("expected full_name field")
	}
	if !strings.Contains(html, "GET /api/contacts") {
		t.Error("expected GET /api/contacts endpoint")
	}
	if !strings.Contains(html, "POST /api/contacts") {
		t.Error("expected POST /api/contacts endpoint")
	}
	if !strings.Contains(html, "DELETE /api/contacts/{id}") {
		t.Error("expected DELETE endpoint")
	}
}

func TestSwaggerSpec_HasAllEndpoints(t *testing.T) {
	expected := []string{
		"/api/contacts",
		"/api/contacts/{id}",
		"/api/search",
	}
	for _, p := range expected {
		found := false
		for _, ep := range ContactAPIDoc.Endpoints {
			if strings.Contains(ep.Path, p) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected endpoint containing %s", p)
		}
	}
}
