package contact

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/grandcat/zeroconf"
)

const (
	AppServiceType   = "_localitas-app._tcp"
	AppServiceDomain = "local."
)

type AppHealth struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

var DefaultHealth = AppHealth{
	Name:        "contact",
	DisplayName: "Contacts",
	Icon:        "users",
	Version:     "0.1.0",
	Status:      "healthy",
}

// HandleHealth serves /health.json — used by core for both discovery metadata
// and periodic health checks. One endpoint, two purposes.
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DefaultHealth)
}

// BroadcastMDNS registers the app as a _localitas-app._tcp mDNS service.
// TXT records carry only the app name; core fetches /health.json for full metadata.
func BroadcastMDNS(port int, name string) (shutdown func(), err error) {
	txt := []string{fmt.Sprintf("name=%s", name)}

	server, err := zeroconf.Register(
		name,
		AppServiceType,
		AppServiceDomain,
		port,
		txt,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("mDNS register: %w", err)
	}

	log.Printf("Broadcasting mDNS: %s on port %d (%s)", AppServiceType, port, name)
	return server.Shutdown, nil
}
