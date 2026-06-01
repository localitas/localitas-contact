package contact

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/localitas/localitas-go"
	"gopkg.in/yaml.v3"
)

type handler struct {
	app *App
}

func (h *handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	limit := intParam(r, "limit", 50)
	offset := intParam(r, "offset", 0)

	contacts, err := h.app.Store.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to list contacts: %v", err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		if len(contacts) == 0 {
			fmt.Fprintf(w, `<div class="text-center py-8 text-secondary">
				<i data-lucide="users" class="w-8 h-8 mx-auto mb-2 opacity-50"></i>
				<p>No contacts yet. Create your first contact!</p>
			</div>`)
			return
		}
		for _, c := range contacts {
			renderContactCard(w, c)
		}
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	c, err := h.app.Store.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, "contact not found: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID := client.UserIDFromRequest(r)
	var data ContactData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	c, err := h.app.Store.Create(r.Context(), userID, data)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to create contact: %v", err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body: %v", err)
		return
	}
	if err := h.app.Store.UpdateYAML(r.Context(), id, req.Data); err != nil {
		writeErr(w, http.StatusBadRequest, "%v", err)
		return
	}
	c, err := h.app.Store.Get(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "saved but failed to reload: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.app.Store.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to delete contact: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *handler) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeErr(w, http.StatusBadRequest, "query parameter 'q' is required")
		return
	}
	limit := intParam(r, "limit", 20)

	userID := client.UserIDFromRequest(r)
	contacts, err := h.app.Store.Search(r.Context(), userID, query, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "search failed: %v", err)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		if len(contacts) == 0 {
			fmt.Fprintf(w, `<div class="text-center py-8 text-secondary">
				<i data-lucide="search-x" class="w-8 h-8 mx-auto mb-2 opacity-50"></i>
				<p>No contacts found for "%s"</p>
			</div>`, template.HTMLEscapeString(query))
			return
		}
		for _, c := range contacts {
			renderContactCard(w, c)
		}
		return
	}

	writeJSON(w, http.StatusOK, contacts)
}

func renderContactCard(w http.ResponseWriter, c *Contact) {
	var data ContactData
	if err := yaml.Unmarshal([]byte(c.Data), &data); err != nil {
		return
	}

	company := ""
	if data.Work.Current != nil {
		company = data.Work.Current.Company
	}
	email := ""
	if len(data.Emails) > 0 {
		email = data.Emails[0]
	}

	fmt.Fprintf(w, `<div class="glass-panel p-4 rounded-lg hover:bg-glass-hover transition-all cursor-pointer">
		<div class="flex items-start justify-between">
			<div class="flex-1">
				<h3 class="font-semibold text-primary">%s</h3>
				<p class="text-sm text-secondary">%s</p>
				<p class="text-xs text-secondary mt-1">%s</p>
			</div>
		</div>
	</div>`,
		template.HTMLEscapeString(data.FullName),
		template.HTMLEscapeString(company),
		template.HTMLEscapeString(email))
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, format string, args ...interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf(format, args...)})
}

func intParam(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
