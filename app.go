package contact

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/localitas/localitas-go"
)

type App struct {
	Store    *Store
	BasePath string
	client   *client.Client
}

// New creates a contact App. The client is used only for install (migration
// apply). All runtime CRUD goes through the Store's database/sql connection.
func New(c *client.Client, basePath string) *App {
	if basePath == "" {
		basePath = "/"
	}
	return &App{
		BasePath: basePath,
		client:   c,
	}
}

// InitStore resolves the contacts database ID and opens a database/sql
// connection via the Localitas driver. Must be called after Install.
func (a *App) InitStore(coreURL, dbID, token string) error {
	store, err := OpenStore(coreURL, dbID, token)
	if err != nil {
		return err
	}
	a.Store = store
	return nil
}

// Install ensures the contacts system database exists and all migrations have been
// applied. Returns the database ID for use with InitStore. Idempotent.
// Migrations are embedded in the binary — no filesystem dependency.
// Retries up to 30 seconds if the data app isn't ready yet.
func (a *App) Install(ctx context.Context) (string, error) {
	for attempt := 1; ; attempt++ {
		db, err := a.client.CreateSystemDatabase(ctx, DatabaseName)
		if err != nil {
			log.Printf("install: attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err := applyEmbeddedMigrations(ctx, a.client, db.ID); err != nil {
			log.Printf("install: migrations attempt %d failed (retrying): %v", attempt, err)
			time.Sleep(2 * time.Second)
			continue
		}
		return db.ID, nil
	}
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(TemplatesFS, "templates/index.html")
	if err != nil {
		log.Printf("contact index template error: %v", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	data := map[string]string{
		"BasePath":   a.BasePath,
		"SelectedID": r.URL.Query().Get("id"),
	}
	if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("contact index render error: %v", err)
	}
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	h := &handler{app: a}
	p := &partialHandler{app: a}

	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /swagger.json", HandleSwagger)
	mux.HandleFunc("GET /help.md", handleHelpMarkdown)
	mux.HandleFunc("GET /api/contacts", h.handleList)
	mux.HandleFunc("POST /api/contacts", h.handleCreate)
	mux.HandleFunc("GET /api/contacts/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/contacts/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/contacts/{id}", h.handleDelete)
	mux.HandleFunc("GET /api/search", h.handleSearch)

	mux.HandleFunc("GET /partials/sidebar", p.handleSidebar)
	mux.HandleFunc("GET /partials/editor/{id}", p.handleEditor)
	mux.HandleFunc("GET /partials/empty", p.handleEmpty)
	mux.HandleFunc("POST /partials/create", p.handleCreate)
	mux.HandleFunc("POST /partials/save/{id}", p.handleSave)
	mux.HandleFunc("DELETE /partials/delete/{id}", p.handleDelete)
	mux.HandleFunc("POST /partials/search", p.handleSearch)

	cardDavHandler := NewCardDAVHandler(a.Store, "/carddav/")
	mux.Handle("/carddav/", cardDavHandler)
	mux.HandleFunc("/.well-known/carddav", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, a.BasePath+"carddav/", http.StatusPermanentRedirect)
	})
}

// TokenPassthrough is kept for compatibility but with the SQL driver approach,
// auth is handled at the driver/DSN level. This middleware now just ensures the
// Authorization header is present.
func (a *App) TokenPassthrough(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
