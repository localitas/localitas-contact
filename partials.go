package contact

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/localitas/localitas-go"
	"gopkg.in/yaml.v3"
)

type ContactListItem struct {
	ID      string
	Name    string
	Company string
}

type SidebarData struct {
	Contacts          []ContactListItem
	SelectedContactID string
}

type EditorData struct {
	Contact     *Contact
	ContactName string
}

type partialHandler struct {
	app *App
}

func (p *partialHandler) tmpl() (*template.Template, error) {
	tmpl := template.New("")
	partials := []string{
		"templates/partials/_sidebar_list.html",
		"templates/partials/_editor.html",
		"templates/partials/_empty.html",
		"templates/partials/_header.html",
	}
	for _, file := range partials {
		content, err := TemplatesFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := tmpl.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
	}
	return tmpl, nil
}

// tmplOOB returns templates with the sidebar list having hx-swap-oob="true"
// for use when the sidebar is returned alongside a primary response (save/create/delete).
func (p *partialHandler) tmplOOB() (*template.Template, error) {
	tmpl := template.New("")
	partials := []string{
		"templates/partials/_editor.html",
		"templates/partials/_empty.html",
		"templates/partials/_header.html",
	}
	for _, file := range partials {
		content, err := TemplatesFS.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}
		if _, err := tmpl.Parse(string(content)); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}
	}
	sidebarContent, err := TemplatesFS.ReadFile("templates/partials/_sidebar_list.html")
	if err != nil {
		return nil, err
	}
	oobSidebar := strings.Replace(string(sidebarContent),
		`id="contacts-sidebar-list"`,
		`id="contacts-sidebar-list" hx-swap-oob="true"`, 1)
	if _, err := tmpl.Parse(oobSidebar); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (p *partialHandler) sidebarItems(r *http.Request) []ContactListItem {
	userID := client.UserIDFromRequest(r)
	contacts, _ := p.app.Store.List(r.Context(), userID, 100, 0)
	items := make([]ContactListItem, 0, len(contacts))
	for _, c := range contacts {
		if c.ID == "" {
			continue
		}
		var data ContactData
		if err := yaml.Unmarshal([]byte(c.Data), &data); err == nil {
			company := ""
			if data.Work.Current != nil {
				company = data.Work.Current.Company
			}
			items = append(items, ContactListItem{
				ID:      c.ID,
				Name:    data.FullName,
				Company: company,
			})
		}
	}
	return items
}

func (p *partialHandler) handleSidebar(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	selectedID := r.URL.Query().Get("selected")
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Contacts:          p.sidebarItems(r),
		SelectedContactID: selectedID,
	})
}

func (p *partialHandler) handleEditor(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	contact, err := p.app.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}
	var data ContactData
	contactName := "Unnamed"
	if err := yaml.Unmarshal([]byte(contact.Data), &data); err == nil && data.FullName != "" {
		contactName = data.FullName
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "editor", EditorData{Contact: contact, ContactName: contactName})
}

func (p *partialHandler) handleEmpty(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "empty", struct {
		DocsHTML template.HTML
	}{
		DocsHTML: RenderDocsHTML(ContactAPIDoc),
	})
}

func (p *partialHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID := client.UserIDFromRequest(r)
	contact, err := p.app.Store.Create(r.Context(), userID, ContactData{FullName: "New Contact"})
	if err != nil {
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}
	templateYAML := NewContactYAML()
	if err := p.app.Store.UpdateYAML(r.Context(), contact.ID, templateYAML); err == nil {
		contact.Data = templateYAML
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Contacts:          p.sidebarItems(r),
		SelectedContactID: contact.ID,
	})
	tmpl.ExecuteTemplate(w, "editor", EditorData{
		Contact:     contact,
		ContactName: "New Contact",
	})
}

func (p *partialHandler) handleSave(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	if err := p.app.Store.UpdateYAML(r.Context(), id, r.FormValue("data")); err != nil {
		http.Error(w, "Failed to save: "+err.Error(), http.StatusBadRequest)
		return
	}
	contact, err := p.app.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to reload contact after save", http.StatusInternalServerError)
		return
	}
	var data ContactData
	contactName := "Unnamed"
	if err := yaml.Unmarshal([]byte(contact.Data), &data); err == nil && data.FullName != "" {
		contactName = data.FullName
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Contacts:          p.sidebarItems(r),
		SelectedContactID: id,
	})
	tmpl.ExecuteTemplate(w, "editor", EditorData{
		Contact:     contact,
		ContactName: contactName,
	})
}

func (p *partialHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmplOOB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Contact ID required", http.StatusBadRequest)
		return
	}
	if err := p.app.Store.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "empty", nil)
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{
		Contacts: p.sidebarItems(r),
	})
}

func (p *partialHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	tmpl, err := p.tmpl()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	query := r.FormValue("query")
	if query == "" {
		p.handleSidebar(w, r)
		return
	}
	userID := client.UserIDFromRequest(r)
	contacts, err := p.app.Store.Search(r.Context(), userID, query, 100)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	items := make([]ContactListItem, 0, len(contacts))
	for _, c := range contacts {
		var data ContactData
		if err := yaml.Unmarshal([]byte(c.Data), &data); err == nil {
			company := ""
			if data.Work.Current != nil {
				company = data.Work.Current.Company
			}
			items = append(items, ContactListItem{
				ID:      c.ID,
				Name:    data.FullName,
				Company: company,
			})
		}
	}
	w.Header().Set("Content-Type", "text/html")
	tmpl.ExecuteTemplate(w, "sidebar_list", SidebarData{Contacts: items})
}
