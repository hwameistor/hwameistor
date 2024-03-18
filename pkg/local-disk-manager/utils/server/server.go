package server

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Parse splits the endpoint into protocol and address and returns them
func Parse(ep string) (string, string, error) {
	if strings.HasPrefix(strings.ToLower(ep), "unix://") || strings.HasPrefix(strings.ToLower(ep), "tcp://") {
		s := strings.SplitN(ep, "://", 2)
		if s[1] != "" {
			return s[0], s[1], nil
		}
		return "", "", fmt.Errorf("Invalid endpoint: %v", ep)
	}
	// Assume everything else is a file path for a Unix Domain Socket.
	return "unix", ep, nil
}

// Listen creates a listener for Unix socket or TCP based on the endpoint
// and returns a cleanup function
func Listen(endpoint string) (net.Listener, func(), error) {
	proto, addr, err := Parse(endpoint)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {}
	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) { //nolint: vetshadow
			return nil, nil, fmt.Errorf("%s: %q", addr, err)
		}
		cleanup = func() {
			os.Remove(addr)
		}
	}

	l, err := net.Listen(proto, addr)
	return l, cleanup, err
}
