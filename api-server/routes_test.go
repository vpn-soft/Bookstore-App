package api_server

import (
	"testing"
)

func TestRoutes(t *testing.T) {
	StartAPIServer("888", "test")
	server := getApiServer("test")
	recoveryMid(handler(server))
}
