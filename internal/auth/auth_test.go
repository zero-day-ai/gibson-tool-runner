package auth

import (
	"os"
	"testing"

	"github.com/zero-day-ai/sdk/daemonclient"
)

func TestSetOIDCEnvVars_AllPresent(t *testing.T) {
	t.Setenv(EnvClientID, "tool-runner-client")
	t.Setenv(EnvClientSecret, "tool-runner-secret")
	t.Setenv(EnvZitadelIssuer, "https://auth.example.com")
	// Clear SDK vars so our mapping runs.
	t.Setenv(daemonclient.EnvOIDCClientID, "")
	t.Setenv(daemonclient.EnvOIDCClientSecret, "")
	t.Setenv(daemonclient.EnvOIDCTokenURL, "")

	if err := setOIDCEnvVars(); err != nil {
		t.Fatalf("setOIDCEnvVars: %v", err)
	}

	wantClientID := "tool-runner-client"
	wantTokenURL := "https://auth.example.com/oauth/v2/token"

	if got := os.Getenv(daemonclient.EnvOIDCClientID); got != wantClientID {
		t.Errorf("%s = %q, want %q", daemonclient.EnvOIDCClientID, got, wantClientID)
	}
	if got := os.Getenv(daemonclient.EnvOIDCClientSecret); got != "tool-runner-secret" {
		t.Errorf("%s = %q, want tool-runner-secret", daemonclient.EnvOIDCClientSecret, got)
	}
	if got := os.Getenv(daemonclient.EnvOIDCTokenURL); got != wantTokenURL {
		t.Errorf("%s = %q, want %q", daemonclient.EnvOIDCTokenURL, got, wantTokenURL)
	}
}

func TestSetOIDCEnvVars_TrailingSlash(t *testing.T) {
	t.Setenv(EnvClientID, "id")
	t.Setenv(EnvClientSecret, "secret")
	t.Setenv(EnvZitadelIssuer, "https://auth.example.com/")
	t.Setenv(daemonclient.EnvOIDCClientID, "")

	if err := setOIDCEnvVars(); err != nil {
		t.Fatalf("setOIDCEnvVars: %v", err)
	}

	want := "https://auth.example.com/oauth/v2/token"
	if got := os.Getenv(daemonclient.EnvOIDCTokenURL); got != want {
		t.Errorf("token URL = %q, want %q (trailing slash should be stripped)", got, want)
	}
}

func TestSetOIDCEnvVars_MissingVars(t *testing.T) {
	cases := []struct {
		name   string
		setID  bool
		setSec bool
		setIss bool
	}{
		{"missing_client_id", false, true, true},
		{"missing_client_secret", true, false, true},
		{"missing_issuer", true, true, false},
		{"missing_all", false, false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(daemonclient.EnvOIDCClientID, "")
			if tc.setID {
				t.Setenv(EnvClientID, "x")
			} else {
				t.Setenv(EnvClientID, "")
			}
			if tc.setSec {
				t.Setenv(EnvClientSecret, "x")
			} else {
				t.Setenv(EnvClientSecret, "")
			}
			if tc.setIss {
				t.Setenv(EnvZitadelIssuer, "https://a.example.com")
			} else {
				t.Setenv(EnvZitadelIssuer, "")
			}

			if err := setOIDCEnvVars(); err == nil {
				t.Error("expected error for missing env vars, got nil")
			}
		})
	}
}

func TestSetOIDCEnvVars_SDKVarPreset(t *testing.T) {
	// When OIDC_CLIENT_CREDENTIALS_CLIENT_ID is already set, setOIDCEnvVars must
	// be a no-op (honour operator override / test injection).
	t.Setenv(daemonclient.EnvOIDCClientID, "already-set")
	t.Setenv(EnvClientID, "")
	t.Setenv(EnvClientSecret, "")
	t.Setenv(EnvZitadelIssuer, "")

	if err := setOIDCEnvVars(); err != nil {
		t.Fatalf("setOIDCEnvVars should be no-op when SDK var is set: %v", err)
	}
	// The SDK var should remain untouched.
	if got := os.Getenv(daemonclient.EnvOIDCClientID); got != "already-set" {
		t.Errorf("SDK var changed: got %q, want already-set", got)
	}
}
