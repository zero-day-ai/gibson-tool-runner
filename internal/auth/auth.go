// Package auth wires up daemon-connection credentials for the gibson-tool-runner.
//
// The tool runner is a service (one deployment, one Zitadel service account) that
// must present both a Zitadel service-account JWT and (when running in Kubernetes)
// a SPIFFE X.509 SVID for mTLS.  This package translates the tool-runner-specific
// environment variables into the SDK's generic OIDC credential env vars, then
// delegates connection establishment to [github.com/zero-day-ai/sdk/daemonclient].
//
// Environment variables consumed by this package:
//
//	ZITADEL_TOOL_RUNNER_CLIENT_ID     — Zitadel service-account client_id  (required)
//	ZITADEL_TOOL_RUNNER_CLIENT_SECRET — Zitadel service-account secret      (required)
//	ZITADEL_ISSUER                    — Zitadel base URL, e.g.
//	                                    https://auth.example.com             (required)
//	GIBSON_DAEMON_ADDRESS             — gRPC address of Envoy front-door
//	                                    (default: localhost:50002)
//
// The token URL is derived as <ZITADEL_ISSUER>/oauth/v2/token.
//
// SPIFFE is auto-detected: if the SPIRE Workload API socket exists at
// /run/spire/sockets/agent.sock the SDK will use mTLS; otherwise it falls
// back to server-side TLS only (for external / customer-network deployments).
//
// Spec: unified-identity-and-authorization Requirements 12.1–12.3.
//
// TODO(unified-identity-and-authorization 7.1): consolidate to agent.Connect
// once SDK >= v0.84.0 ships that entry point.
package auth

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/zero-day-ai/sdk/daemonclient"
)

// Env var names specific to the tool runner.
const (
	// EnvClientID is the Zitadel service-account client_id.
	EnvClientID = "ZITADEL_TOOL_RUNNER_CLIENT_ID"

	// EnvClientSecret is the Zitadel service-account client_secret.
	EnvClientSecret = "ZITADEL_TOOL_RUNNER_CLIENT_SECRET"

	// EnvZitadelIssuer is the Zitadel base URL; used to derive the token URL.
	EnvZitadelIssuer = "ZITADEL_ISSUER"
)

// ConnectDaemon returns a connected [daemonclient.Client] using Zitadel
// service-account client-credentials + (when available) SPIFFE mTLS.
//
// Call [daemonclient.Client.Close] when done.
//
// Returns an error when required environment variables are missing or when
// the daemon address cannot be dialled.
func ConnectDaemon(ctx context.Context) (*daemonclient.Client, error) {
	if err := setOIDCEnvVars(); err != nil {
		return nil, err
	}

	addr := daemonclient.GetDaemonAddress()
	client, err := daemonclient.New(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("auth: connect daemon at %s: %w", addr, err)
	}
	return client, nil
}

// setOIDCEnvVars translates the tool-runner-specific env vars into the SDK's
// generic OIDC client-credentials env vars so daemonclient.New can pick them up
// via its credential auto-detection logic.
//
// It is idempotent: if OIDC_CLIENT_CREDENTIALS_CLIENT_ID is already set (e.g.
// in tests that set the SDK vars directly) it is left alone.
func setOIDCEnvVars() error {
	// Honour pre-set SDK vars (allows overriding in tests and operator scripts).
	if os.Getenv(daemonclient.EnvOIDCClientID) != "" {
		return nil
	}

	clientID := os.Getenv(EnvClientID)
	clientSecret := os.Getenv(EnvClientSecret)
	issuer := strings.TrimRight(os.Getenv(EnvZitadelIssuer), "/")

	var missing []string
	if clientID == "" {
		missing = append(missing, EnvClientID)
	}
	if clientSecret == "" {
		missing = append(missing, EnvClientSecret)
	}
	if issuer == "" {
		missing = append(missing, EnvZitadelIssuer)
	}
	if len(missing) > 0 {
		return fmt.Errorf("auth: required env vars not set: %s", strings.Join(missing, ", "))
	}

	tokenURL := issuer + "/oauth/v2/token"

	os.Setenv(daemonclient.EnvOIDCClientID, clientID)
	os.Setenv(daemonclient.EnvOIDCClientSecret, clientSecret)
	os.Setenv(daemonclient.EnvOIDCTokenURL, tokenURL)

	return nil
}
