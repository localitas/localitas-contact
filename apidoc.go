package contact

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type APIEndpoint struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Summary     string          `json:"summary"`
	Description string          `json:"description,omitempty"`
	QueryParams []APIParam      `json:"query_params,omitempty"`
	RequestBody *APIRequestBody `json:"request_body,omitempty"`
	Response    *APIResponse    `json:"response,omitempty"`
}

type APIParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIRequestBody struct {
	ContentType string `json:"content_type"`
	Example     string `json:"example"`
}

type APIResponse struct {
	ContentType string `json:"content_type"`
	Example     string `json:"example"`
}

type APIFieldDoc struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

type APIDoc struct {
	AppName     string        `json:"app_name"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Keywords    []string      `json:"keywords,omitempty"`
	Fields      []APIFieldDoc `json:"fields,omitempty"`
	Endpoints   []APIEndpoint `json:"endpoints"`
}

var ContactAPIDoc = APIDoc{
	AppName:     "Contacts",
	Version:     "0.1.0",
	Description: "Personal contact management with YAML-based data and full-text search",
	Keywords:    []string{"contact", "contacts", "people", "person", "phone", "address", "addressbook", "vcard", "directory", "friend", "colleague"},
	Fields: []APIFieldDoc{
		{Name: "full_name", Description: "Contact's full name", Example: "Alice Johnson"},
		{Name: "emails", Description: "Email addresses (list)", Example: "- alice@example.com\n- alice.work@company.com"},
		{Name: "phones", Description: "Phone numbers (list)", Example: "- +1-555-0123"},
		{Name: "social", Description: "Social media profiles", Example: "linkedin: linkedin.com/in/alice\ngithub: alicej\nx: @alicej\nwebsite: alice.dev"},
		{Name: "work.current", Description: "Current job", Example: "company: Acme Corp\ntitle: Software Engineer\nstarted: 2024-01"},
		{Name: "work.previous", Description: "Previous jobs (list)", Example: "- company: Old Corp\n  title: Junior Dev\n  started: 2022-01\n  ended: 2023-12"},
		{Name: "interactions", Description: "Meeting notes and interactions", Example: "- type: meeting\n  date: 2026-04-15\n  notes: Discussed partnership"},
		{Name: "metadata", Description: "Custom key-value fields", Example: "tags: [prospect, developer]\nlocation: San Francisco, CA\nbirthday: 1990-05-15"},
	},
	Endpoints: []APIEndpoint{
		{
			Method:  "GET",
			Path:    "/api/contacts",
			Summary: "List all contacts",
			QueryParams: []APIParam{
				{Name: "limit", Type: "integer", Description: "Max results (default 50)"},
				{Name: "offset", Type: "integer", Description: "Pagination offset (default 0)"},
			},
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `[{"id":"a1b2c3...","data":"full_name: Alice Johnson\nemails:\n  - alice@example.com","created_at":"2026-04-22T...","updated_at":"2026-04-22T..."}]`,
			},
		},
		{
			Method:  "POST",
			Path:    "/api/contacts",
			Summary: "Create a contact",
			RequestBody: &APIRequestBody{
				ContentType: "application/json",
				Example:     `{"full_name":"Alice Johnson","emails":["alice@example.com"],"phones":["+1-555-0123"]}`,
			},
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `{"id":"a1b2c3...","data":"full_name: Alice Johnson\n...","created_at":"2026-04-22T..."}`,
			},
		},
		{
			Method:  "GET",
			Path:    "/api/contacts/{id}",
			Summary: "Get a contact by ID",
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `{"id":"a1b2c3...","data":"full_name: Alice Johnson\n...","created_at":"2026-04-22T..."}`,
			},
		},
		{
			Method:  "PUT",
			Path:    "/api/contacts/{id}",
			Summary: "Update a contact (raw YAML)",
			RequestBody: &APIRequestBody{
				ContentType: "application/json",
				Example:     `{"data":"full_name: Alice Doe\nemails:\n  - alice@newco.com"}`,
			},
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `{"id":"a1b2c3...","data":"full_name: Alice Doe\n..."}`,
			},
		},
		{
			Method:  "DELETE",
			Path:    "/api/contacts/{id}",
			Summary: "Delete a contact",
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `{"success":true}`,
			},
		},
		{
			Method:  "GET",
			Path:    "/api/search",
			Summary: "Search contacts (FTS5)",
			QueryParams: []APIParam{
				{Name: "q", Type: "string", Required: true, Description: "Search query"},
				{Name: "limit", Type: "integer", Description: "Max results (default 20)"},
			},
			Response: &APIResponse{
				ContentType: "application/json",
				Example:     `[{"id":"a1b2c3...","data":"full_name: Alice Johnson\n..."}]`,
			},
		},
		{
			Method:      "PROPFIND",
			Path:        "/carddav/",
			Summary:     "CardDAV discovery — list address books",
			Description: "Standard CardDAV protocol. Connect Apple Contacts, Thunderbird, or DAVx5 to /carddav/",
		},
		{
			Method:  "REPORT",
			Path:    "/carddav/default/",
			Summary: "CardDAV query — list/filter contacts as vCards",
		},
		{
			Method:      "PUT",
			Path:        "/carddav/default/{id}.vcf",
			Summary:     "CardDAV create/update contact via vCard",
			Description: "Upload a vCard 3.0 to create or update a contact",
		},
		{
			Method:  "DELETE",
			Path:    "/carddav/default/{id}.vcf",
			Summary: "CardDAV delete contact",
		},
	},
}

func HandleSwagger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ContactAPIDoc)
}

// RenderDocsHTML generates the accordion HTML for the empty state docs page.
func RenderDocsHTML(doc APIDoc) template.HTML {
	var sb strings.Builder

	// Fields section
	if len(doc.Fields) > 0 {
		sb.WriteString(`<h3 style="font-size: 0.875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-secondary); margin-bottom: 1rem;">YAML Fields Reference</h3>`)
		sb.WriteString(`<div class="accordion-list">`)
		for _, f := range doc.Fields {
			sb.WriteString(`<details class="glass-panel" style="border-radius: 0.5rem; margin-bottom: 0.5rem;">`)
			sb.WriteString(fmt.Sprintf(`<summary style="padding: 0.75rem 1rem; cursor: pointer; font-weight: 500; color: var(--color-text-primary);">%s</summary>`, template.HTMLEscapeString(f.Name)))
			sb.WriteString(`<div style="padding: 0 1rem 0.75rem; font-size: 0.875rem; color: var(--color-text-secondary);">`)
			sb.WriteString(fmt.Sprintf(`<p style="margin-bottom: 0.5rem;">%s</p>`, template.HTMLEscapeString(f.Description)))
			sb.WriteString(fmt.Sprintf(`<pre style="background: var(--color-bg-base); padding: 0.75rem; border-radius: 0.375rem; overflow-x: auto; font-size: 0.8125rem;">%s</pre>`, template.HTMLEscapeString(f.Example)))
			sb.WriteString(`</div></details>`)
		}
		sb.WriteString(`</div>`)
		sb.WriteString(`<hr style="border-color: var(--color-glass-border); margin: 1.5rem 0;">`)
	}

	// Endpoints section
	sb.WriteString(`<h3 style="font-size: 0.875rem; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; color: var(--color-text-secondary); margin-bottom: 1rem;">API Endpoints</h3>`)
	sb.WriteString(`<div class="accordion-list">`)
	for _, ep := range doc.Endpoints {
		title := fmt.Sprintf("%s %s — %s", ep.Method, ep.Path, ep.Summary)
		sb.WriteString(`<details class="glass-panel" style="border-radius: 0.5rem; margin-bottom: 0.5rem;">`)
		sb.WriteString(fmt.Sprintf(`<summary style="padding: 0.75rem 1rem; cursor: pointer; font-weight: 500; color: var(--color-text-primary);">%s</summary>`, template.HTMLEscapeString(title)))
		sb.WriteString(`<div style="padding: 0 1rem 0.75rem; font-size: 0.875rem; color: var(--color-text-secondary);">`)

		if len(ep.QueryParams) > 0 {
			sb.WriteString(`<p style="margin-bottom: 0.5rem;">Query params: `)
			for i, p := range ep.QueryParams {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("<code>%s</code>", template.HTMLEscapeString(p.Name)))
				if p.Required {
					sb.WriteString(" (required)")
				}
			}
			sb.WriteString(`</p>`)
		}

		var example strings.Builder
		if ep.RequestBody != nil {
			example.WriteString("# Request\n")
			example.WriteString(prettyJSON(ep.RequestBody.Example))
			example.WriteString("\n\n")
		}
		if ep.Response != nil {
			example.WriteString("# Response\n")
			example.WriteString(prettyJSON(ep.Response.Example))
		}
		if example.Len() > 0 {
			sb.WriteString(fmt.Sprintf(`<pre style="background: var(--color-bg-base); padding: 0.75rem; border-radius: 0.375rem; overflow-x: auto; font-size: 0.8125rem;">%s</pre>`, template.HTMLEscapeString(example.String())))
		}

		sb.WriteString(`</div></details>`)
	}
	sb.WriteString(`</div>`)

	return template.HTML(sb.String())
}

func prettyJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return s
	}
	return string(b)
}
