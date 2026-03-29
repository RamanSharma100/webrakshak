package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"webrakshak/config"
	"webrakshak/internal/events"
)

var lastAlertTimes = make(map[string]time.Time)

// Check if the host matches any blocked substring in the config.
func isWebsiteBlocked(host string) bool {
    cleanHost := strings.ToLower(host)
    for _, blocked := range config.GetBlocked() {
        if strings.Contains(cleanHost, blocked) {
            return true
        }
    }
    return false
}

// Handler is the HTTP handler to be registered as the system proxy.
// It handles CONNECT tunneling for HTTPS (plain TCP tunnel, no MITM) and
// direct TCP proxying for HTTP requests.
func Handler(w http.ResponseWriter, r *http.Request) {
    host := r.URL.Hostname()
    if host == "" {
        host = strings.Split(r.Host, ":")[0]
    }

    if isWebsiteBlocked(host) {
        events.LogEvent("web_block", "Intercepted System Network attempt to restricted domain: "+host)

        lastTime, exists := lastAlertTimes[host]
        if !exists || time.Since(lastTime) > 10*time.Second {
            lastAlertTimes[host] = time.Now()
            // show warning using zenity and open warning page for linux system modal
            go func(blockedHost string) {
                cmd := exec.Command("zenity", "--warning", "--title", "WebRakshak Guardian", "--text", "⚠️ <b>WebRakshak Blocked a Website</b>\n\nYour browser attempted to access <b>"+blockedHost+"</b>.", "--width", "400")
                _ = cmd.Start()
            }(host)
            go func(blockedHost string) {
                _ = exec.Command("xdg-open", fmt.Sprintf("http://127.0.0.1:8080/warning.html?blocked=%s", blockedHost)).Start()
            }(host)
        }

        if r.Method == http.MethodConnect {
            hijacker, ok := w.(http.Hijacker)
            if ok {
                clientConn, _, err := hijacker.Hijack()
                if err == nil {
                    clientConn.Close()
                }
            }
            return
        }

        http.Redirect(w, r, fmt.Sprintf("http://127.0.0.1:8080/warning.html?blocked=%s", host), http.StatusFound)
        return
    }

    if r.Method == http.MethodConnect {
        handleHTTPS(w, r)
    } else {
        destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
        if err != nil {
            http.Error(w, err.Error(), http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        hijacker, ok := w.(http.Hijacker)
        if !ok {
            http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
            return
        }
        clientConn, _, err := hijacker.Hijack()
        if err != nil {
            return
        }
        go transfer(destConn, clientConn)
        go transfer(clientConn, destConn)
    }
}

// TODO: Add certificate handling to manage https requests
func handleHTTPS(w http.ResponseWriter, r *http.Request) {
    destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    hijacker, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }
    clientConn, _, err := hijacker.Hijack()
    if err != nil {
        return
    }
    go transfer(destConn, clientConn)
    go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
    defer destination.Close()
    defer source.Close()
    io.Copy(destination, source)
}

// SetSystemProxy toggles GNOME system proxy settings (manual/none).
func SetSystemProxy(enable bool) {
    if enable {
        exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
        exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "host", "127.0.0.1").Run()
        exec.Command("gsettings", "set", "org.gnome.system.proxy.http", "port", "8081").Run()
        exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "host", "127.0.0.1").Run()
        exec.Command("gsettings", "set", "org.gnome.system.proxy.https", "port", "8081").Run()
    } else {
        exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
    }
}

