package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"webrakshak/config"
	"webrakshak/internal/handlers"
	"webrakshak/internal/proxy"
)

func main() {
	// Load blocklist
	_ = config.LoadBlocked("./config/blocked.json")

	// ensure proxy is removed on shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		proxy.SetSystemProxy(false)
		os.Exit(0)
	}()

	proxy.SetSystemProxy(true)

	// start system proxy server (port 8081)
	proxyServer := &http.Server{
		Addr:    ":8081",
		Handler: http.HandlerFunc(proxy.Handler),
	}
	go func() {
		log.Fatal(proxyServer.ListenAndServe())
	}()

	// HTTP server for UI and API (port 8080) 
	// Mux added to allow API and static file handling on same port
	mux := http.NewServeMux()

	// Serve the provided warning page as the default UI
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./ui/warning.html")
	})
	// Keep serving static files from ./ui under /static/ in case assets are added later
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./ui"))))
	handlers.Register(mux)

	fmt.Println("🚀 GOD-TIER WebRakshak Engine Running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}