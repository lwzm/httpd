package httpd

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// ClientIP Guess client real ip
func ClientIP(r *http.Request) string {
	if ip := strings.TrimSpace(strings.Split(
		r.Header.Get("X-Forwarded-For"), ",")[0]); ip != "" {
		return ip
	}

	if ip := strings.TrimSpace(r.Header.Get("X-Real-Ip")); ip != "" {
		return ip
	}

	// return strings.Split(r.RemoteAddr, ":")[0]
	return r.RemoteAddr
}

// Start the common http server
func Start() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}
	port = ":" + port
	log.Println(port)
	log.Fatal(http.ListenAndServe(port, nil))
}
