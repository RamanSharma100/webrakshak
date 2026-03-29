package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"webrakshak/internal/events"
)

type Telemetry struct {
    CPUUsage float64 `json:"cpu_usage"`
    MemTotal uint64  `json:"mem_total"`
    MemUsed  uint64  `json:"mem_used"`
    Uptime   uint64  `json:"uptime"`
    OSName   string  `json:"os_name"`
}

// Register registers all HTTP endpoints on the provided ServeMux.
func Register(mux *http.ServeMux) {
    mux.HandleFunc("/api/telemetry", telemetryHandler)
    mux.HandleFunc("/api/trace", traceHandler)
    mux.HandleFunc("/api/decisions", getDecisionsHandler)
    mux.HandleFunc("/api/decision", saveDecisionHandler)
    mux.HandleFunc("/api/alert", alertHandler)
    mux.HandleFunc("/api/events", events.SSEHandler)
}

func telemetryHandler(w http.ResponseWriter, r *http.Request) {
    var t Telemetry
    if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Telemetry persistence removed — we just broadcast telemetry to SSE clients.
    events.Broadcast(map[string]interface{}{
        "type":      "telemetry",
        "cpu_usage": t.CPUUsage,
        "mem_total": t.MemTotal,
        "mem_used":  t.MemUsed,
        "uptime":    t.Uptime,
        "os_name":   t.OSName,
        "time":      time.Now().Format(time.RFC3339),
    })
    w.WriteHeader(http.StatusOK)
}

func traceHandler(w http.ResponseWriter, r *http.Request) {
    var t struct{ Trace string `json:"trace"` }
    if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    events.LogEvent("trace", t.Trace)
    w.WriteHeader(http.StatusOK)
}

func getDecisionsHandler(w http.ResponseWriter, r *http.Request) {
    // Decisions persistence removed — return empty map for now
    var decisions = make(map[string]bool)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(decisions)
}

func saveDecisionHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        AppName  string `json:"app_name"`
        IsKilled bool   `json:"is_killed"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Decisions persistence removed — emit an event instead
    events.LogEvent("decision", fmt.Sprintf("App persistence updated: %s -> Kill=%v", req.AppName, req.IsKilled))
    w.WriteHeader(http.StatusOK)
}

func alertHandler(w http.ResponseWriter, r *http.Request) {
    var alert struct{ Message string `json:"message"` }
    if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    events.LogEvent("alert", alert.Message)
    w.WriteHeader(http.StatusOK)
}

