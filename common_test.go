package httpd

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
}

func TestClientIP(t *testing.T) {
	req := &http.Request{RemoteAddr: "1.2.3.4:5678", Header: http.Header{}}
	assert.Equal(t, "1.2.3.4:5678", ClientIP(req))
	req.Header.Add("X-Real-Ip", "1.1.1.1")
	assert.Equal(t, "1.1.1.1", ClientIP(req))
	req.Header.Add("X-Forwarded-For", "2.2.2.2, 3.3.3.3, 4.4.4.4")
	assert.Equal(t, "2.2.2.2", ClientIP(req))
}
