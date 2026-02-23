package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCAuthenticator handles OIDC authentication with PKCE support.
type OIDCAuthenticator struct {
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
	allowedUser  string

	mu     sync.Mutex
	states map[string]stateEntry // state -> PKCE code_verifier
}

type stateEntry struct {
	codeVerifier string
	redirectURI  string
	createdAt    time.Time
}

const stateMaxAge = 10 * time.Minute

func NewOIDC(ctx context.Context, issuer, clientID, clientSecret string) (*OIDCAuthenticator, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc provider discovery: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCAuthenticator{
		provider:     provider,
		verifier:     verifier,
		oauth2Config: oauth2Config,
		states:       make(map[string]stateEntry),
	}, nil
}

// SetAllowedUser restricts login to a specific user identity (email or sub claim).
func (a *OIDCAuthenticator) SetAllowedUser(user string) {
	a.allowedUser = user
}

// AuthURL generates an authorization URL with state and PKCE S256 challenge.
// The redirectURI is stored alongside the state so Callback can use the same URI for token exchange.
func (a *OIDCAuthenticator) AuthURL(redirectURI string) (authURL string, err error) {
	state, err := randomString(32)
	if err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}
	codeVerifier, err := randomString(64)
	if err != nil {
		return "", fmt.Errorf("generate code verifier: %w", err)
	}

	a.mu.Lock()
	a.cleanExpiredStates()
	a.states[state] = stateEntry{codeVerifier: codeVerifier, redirectURI: redirectURI, createdAt: time.Now()}
	a.mu.Unlock()

	cfg := a.oauth2Config
	cfg.RedirectURL = redirectURI
	challenge := s256Challenge(codeVerifier)
	url := cfg.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return url, nil
}

// Callback exchanges the authorization code for tokens and verifies the ID token.
// Returns a user identifier (email or subject claim).
func (a *OIDCAuthenticator) Callback(ctx context.Context, state, code string) (string, error) {
	a.mu.Lock()
	entry, ok := a.states[state]
	if ok {
		delete(a.states, state)
	}
	a.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("invalid or expired state")
	}
	if time.Since(entry.createdAt) > stateMaxAge {
		return "", fmt.Errorf("state expired")
	}

	cfg := a.oauth2Config
	cfg.RedirectURL = entry.redirectURI
	token, err := cfg.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", entry.codeVerifier),
	)
	if err != nil {
		return "", fmt.Errorf("token exchange: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("no id_token in response")
	}

	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return "", fmt.Errorf("verify id_token: %w", err)
	}

	var claims struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", fmt.Errorf("parse claims: %w", err)
	}

	userID := claims.Email
	if userID == "" {
		userID = claims.Sub
	}

	if a.allowedUser != "" && userID != a.allowedUser && claims.Sub != a.allowedUser {
		return "", fmt.Errorf("user %q is not allowed", userID)
	}

	return userID, nil
}

func (a *OIDCAuthenticator) cleanExpiredStates() {
	now := time.Now()
	for k, v := range a.states {
		if now.Sub(v.createdAt) > stateMaxAge {
			delete(a.states, k)
		}
	}
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func s256Challenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
