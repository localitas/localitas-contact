---
title: Contacts
description: Contact management with CardDAV support
---

# Contacts

Store and manage contacts with full vCard support and CardDAV synchronization.

## Contact Management

Create, read, update, and delete contacts. Each contact supports standard vCard fields: name, email addresses, phone numbers, addresses, organization, and notes.

**GET /api/contacts** - List all contacts
**POST /api/contacts** - Create a new contact
**GET /api/contacts/{id}** - Get a contact by ID
**PUT /api/contacts/{id}** - Update a contact
**DELETE /api/contacts/{id}** - Delete a contact

## Search

Search contacts by name, email, phone number, or organization.

**GET /api/search** - Search contacts with a query string

## CardDAV Server

The app exposes a CardDAV server at `/carddav/` for external address book clients (Apple Contacts, Thunderbird, etc.). The well-known endpoint `/.well-known/carddav` redirects to it.

## HTMX Partials

The UI uses server-rendered HTMX partials for sidebar navigation, contact editing, creating, saving, deleting, and searching. These endpoints power the interactive web interface without client-side JavaScript frameworks.

**GET /partials/sidebar** - Render the contact list sidebar
**GET /partials/editor/{id}** - Render the contact editor for a specific contact
**POST /partials/create** - Create a contact and return updated UI
**POST /partials/save/{id}** - Save edits and return updated UI
**DELETE /partials/delete/{id}** - Delete and return updated sidebar

## Data Format

Contacts are stored in a normalized SQLite schema and can be exported as vCard (VCF) format through the CardDAV interface.

## Build & Deploy

### Version

```bash
./contact-server --version
```

### Build from source

```bash
# Development (native)
cd apps/contact && go build -o bin/contact-server ./cmd/contact-server

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -trimpath -o bin/contact-server-linux-amd64 ./cmd/contact-server
```

### Docker

Build a Docker image directly from the binary:

```bash
# Default base image (debian:12-slim)
./contact-server docker-build

# Custom base image
./contact-server docker-build --base ubuntu:24.04

# Custom Dockerfile
./contact-server docker-build --dockerfile ./my.Dockerfile

# Tag and push to registry
./contact-server docker-build --tag ghcr.io/localitas/contact:latest --push
```

The `docker-build` command requires a Linux amd64 binary in the same directory. Run `make deploy-build` from the project root first.

### Download

Pre-built binaries are available on the [GitHub releases page](https://github.com/localitas/localitas/releases).

Each release includes three builds per app:
- `contact-server-darwin-arm64` (macOS Apple Silicon)
- `contact-server-linux-amd64` (Linux x86_64)
- `contact-server-linux-arm64` (Linux ARM64)

Download with the GitHub CLI:

    gh release download --repo localitas/localitas --pattern 'contact-server-*'

### Release

All app binaries are published to GitHub releases as part of `make deploy-upload-image`.
