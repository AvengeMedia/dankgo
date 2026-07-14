package netutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIPAddress(t *testing.T) {
	ip := "123.45.67.89"

	request, _ := http.NewRequest(http.MethodPost, "https://example.com", nil)
	request.Header.Set("CF-Connecting-IP", ip)
	request.Header.Set("X-Real-Ip", "not-the-ip")
	request.Header.Set("X-Forwarded-For", "not-the-ip")
	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "https://example.com", nil)
	request.Header.Set("X-Real-Ip", ip)
	request.Header.Set("X-Forwarded-For", "not-the-ip")
	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "https://example.com", nil)
	request.Header.Set("X-Forwarded-For", ip)
	assert.Equal(t, ip, GetIPAddress(request))

	request, _ = http.NewRequest(http.MethodPost, "https://example.com", nil)
	request.RemoteAddr = ip
	assert.Equal(t, ip, GetIPAddress(request))
}
