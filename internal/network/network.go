// Package network provides hostname parsing and loopback detection utilities.
// Adapted from https://github.com/open-policy-agent/conftest/tree/v0.47.0/internal/network
package network

import (
	"net"
	"strings"
)

// Hostname extracts the hostname from an OCI reference string.
func Hostname(ref string) string {
	ref = strings.TrimPrefix(ref, "oci://")

	colon := strings.Index(ref, ":")
	slash := strings.Index(ref, "/")

	cut := colon
	if colon == -1 || (colon > slash && slash != -1) {
		cut = slash
	}

	if cut < 0 {
		return ref
	}

	return ref[0:cut]
}

// IsLoopback reports whether the given host resolves to a loopback address.
func IsLoopback(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0:0:0:0:0:0:0:1" {
		// fast path
		return true
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			return true
		}
	}

	return false
}
