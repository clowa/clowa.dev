package main

import (
	"reflect"
	"testing"
)

// Unset (or blank) API_TRUSTED_PROXIES must fall back to the loopback default so
// the api trusts only Caddy out of the box.
func TestTrustedProxiesDefaultsToLoopback(t *testing.T) {
	t.Setenv("API_TRUSTED_PROXIES", "")
	if got := trustedProxies(); !reflect.DeepEqual(got, defaultTrustedProxies) {
		t.Errorf("trustedProxies() = %v, want %v", got, defaultTrustedProxies)
	}
}

// A comma-separated list is split, trimmed, and empty entries dropped.
func TestTrustedProxiesParsesList(t *testing.T) {
	t.Setenv("API_TRUSTED_PROXIES", " 127.0.0.1 , 10.0.0.0/8 ,")
	want := []string{"127.0.0.1", "10.0.0.0/8"}
	if got := trustedProxies(); !reflect.DeepEqual(got, want) {
		t.Errorf("trustedProxies() = %v, want %v", got, want)
	}
}

// A value that is only separators/whitespace has no real entries, so it must fall
// back to the default rather than trusting an empty (all) set.
func TestTrustedProxiesBlankFallsBackToDefault(t *testing.T) {
	t.Setenv("API_TRUSTED_PROXIES", " , ")
	if got := trustedProxies(); !reflect.DeepEqual(got, defaultTrustedProxies) {
		t.Errorf("trustedProxies() = %v, want default %v", got, defaultTrustedProxies)
	}
}
