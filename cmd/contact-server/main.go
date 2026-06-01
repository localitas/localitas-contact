package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	contact "github.com/localitas/localitas-contact"
	dockerbuild "github.com/localitas/localitas-app-common"
	client "github.com/localitas/localitas-go"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Printf("contact-server %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "docker-build" {
		dockerbuild.Run(dockerbuild.Config{
			AppName: "contact",
			Version: version,
		}, os.Args[2:])
		return
	}

	var (
		listen     = flag.String("listen", ":0", "listen address (default :0 = random port)")
		coreURL = flag.String("core-url", client.DefaultCoreURL(), "base URL of the Localitas core API")
		basePath   = flag.String("base-path", "/", "URL prefix for <base href> (set by proxy mount point)")
		token      = flag.String("token", os.Getenv("LOCALITAS_TOKEN"), "bearer token for install + SQL driver (falls back to $LOCALITAS_TOKEN)")
	)
	flag.Parse()

	ctx := context.Background()
	c := client.New(*coreURL)
	if *token != "" {
		c = c.WithToken(*token)
	}

	app := contact.New(c, *basePath)

	dbID, err := app.Install(ctx)
	if err != nil {
		log.Fatalf("install: %v", err)
	}
	log.Printf("Contacts database ready: %s", dbID)

	if err := app.InitStore(*coreURL, dbID, *token); err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer app.Store.Close()

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)
	mux.HandleFunc("GET /health.json", contact.HandleHealth)

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	fmt.Printf("contact-server listening on http://localhost:%d\n", addr.Port)

	selfURL := fmt.Sprintf("http://localhost:%d", addr.Port)
	if err := c.RegisterService(ctx, "contact", selfURL); err != nil {
		log.Printf("⚠️  service registry failed: %v", err)
	}

	shutdown, err := contact.BroadcastMDNS(addr.Port, contact.DefaultHealth.Name)
	if err != nil {
		log.Printf("⚠️  mDNS broadcast failed: %v", err)
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		if shutdown != nil {
			shutdown()
		}
		os.Exit(0)
	}()

	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
