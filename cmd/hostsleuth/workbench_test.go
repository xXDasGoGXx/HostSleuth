package main

import (
	"net/http/httptest"
	"testing"
)

func TestWorkbenchRequestAllowedOnlyFromLoopback(t *testing.T) {
	loopback := httptest.NewRequest("GET", "http://127.0.0.1/api/workbench/dns", nil)
	loopback.RemoteAddr = "127.0.0.1:12345"
	if !workbenchRequestAllowed(loopback) {
		t.Fatal("expected IPv4 loopback request to be allowed")
	}

	ipv6 := httptest.NewRequest("GET", "http://[::1]/api/workbench/dns", nil)
	ipv6.RemoteAddr = "[::1]:12345"
	if !workbenchRequestAllowed(ipv6) {
		t.Fatal("expected IPv6 loopback request to be allowed")
	}

	remote := httptest.NewRequest("GET", "http://192.168.1.50/api/workbench/dns", nil)
	remote.RemoteAddr = "192.168.1.99:54321"
	if workbenchRequestAllowed(remote) {
		t.Fatal("non-loopback Workbench request must be rejected")
	}
}
