package httpd

import (
	"log"
	"net/http"
	"os"
)

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
