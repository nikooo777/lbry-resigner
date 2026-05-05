package main

import (
	"strings"
	"testing"
)

const proxyHost = "https://api.na-backend.odysee.com"

func TestIsProxyEndpoint(t *testing.T) {
	cases := []struct {
		name     string
		endpoint string
		want     bool
	}{
		{"local default", defaultEndpoint, false},
		{"local trailing slash", defaultEndpoint + "/", false},
		{"canonical proxy", proxyHost + proxyPath, true},
		{"proxy trailing slash", proxyHost + proxyPath + "/", true},
		{"proxy double slash", proxyHost + "/api/v1//proxy", true},
		{"proxy with query", proxyHost + proxyPath + "?m=status", true},
		{"proxy in hostname only", "https://proxy.example.com/", false},
		{"proxy in unrelated path", "https://example.com" + proxyPath + "/extra", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isProxyEndpoint(tc.endpoint)
			if got != tc.want {
				t.Errorf("isProxyEndpoint(%q) = %v, want %v", tc.endpoint, got, tc.want)
			}
		})
	}
}

func TestResolveConfig(t *testing.T) {
	canonicalProxy := proxyHost + proxyPath
	cases := []struct {
		name         string
		endpoint     string
		flagToken    string
		wantToken    string
		wantProxy    bool
		wantEndpoint string
		wantErr      bool
	}{
		{
			name:         "local endpoint, no token",
			endpoint:     defaultEndpoint,
			flagToken:    "",
			wantToken:    "",
			wantProxy:    false,
			wantEndpoint: defaultEndpoint,
		},
		{
			name:         "local endpoint, flag token suppressed",
			endpoint:     defaultEndpoint,
			flagToken:    "secret",
			wantToken:    "",
			wantProxy:    false,
			wantEndpoint: defaultEndpoint,
		},
		{
			name:         "proxy canonical with token",
			endpoint:     canonicalProxy,
			flagToken:    "secret",
			wantToken:    "secret",
			wantProxy:    true,
			wantEndpoint: canonicalProxy,
		},
		{
			name:         "proxy trailing slash canonicalized",
			endpoint:     canonicalProxy + "/",
			flagToken:    "secret",
			wantToken:    "secret",
			wantProxy:    true,
			wantEndpoint: canonicalProxy,
		},
		{
			name:         "proxy double slash canonicalized",
			endpoint:     proxyHost + "/api/v1//proxy",
			flagToken:    "secret",
			wantToken:    "secret",
			wantProxy:    true,
			wantEndpoint: canonicalProxy,
		},
		{
			name:      "proxy without token returns error",
			endpoint:  canonicalProxy,
			flagToken: "",
			wantErr:   true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := resolveConfig(tc.endpoint, tc.flagToken)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got cfg=%+v", cfg)
				}
				if !strings.Contains(err.Error(), "--auth-token") {
					t.Errorf("error %q should mention --auth-token", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Endpoint != tc.wantEndpoint {
				t.Errorf("Endpoint=%q want %q", cfg.Endpoint, tc.wantEndpoint)
			}
			if cfg.AuthToken != tc.wantToken {
				t.Errorf("AuthToken=%q want %q", cfg.AuthToken, tc.wantToken)
			}
			if cfg.ProxyMode != tc.wantProxy {
				t.Errorf("ProxyMode=%v want %v", cfg.ProxyMode, tc.wantProxy)
			}
		})
	}
}
