package events

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

var clients = make(map[chan string]bool)

// Broadcast sends a JSON event to all connected SSE clients.
func Broadcast(data map[string]interface{}) {
    eventJSON, _ := json.Marshal(data)
    for client := range clients {
        client <- string(eventJSON)
    }
}

// LogEvent writes a row to the events table (if DB is available) and broadcasts
// the event over SSE.
func LogEvent(eventType, message string) {
    // No persistent DB configured; just broadcast and log locally.
    log.Printf("event: %s - %s", eventType, message)
    Broadcast(map[string]interface{}{
        "type":    eventType,
        "message": message,
        "time":    time.Now().Format(time.RFC3339),
    })
}

// SSEHandler is an HTTP handler that upgrades a connection to Server-Sent Events
// and forwards messages pushed via Broadcast.
func SSEHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    clientChan := make(chan string)
    clients[clientChan] = true
    defer func() { delete(clients, clientChan) }()

    for msg := range clientChan {
        fmt.Fprintf(w, "data: %s\n\n", msg)
        if flusher, ok := w.(http.Flusher); ok {
            flusher.Flush()
        }
    }
}

