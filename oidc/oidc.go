package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCAuthenticator struct {
	oidcProvider    *oidc.Provider
	oauth2Config    oauth2.Config
	scopes          []string
	stateCookieTTL  time.Duration
	https           bool
	stateCookieName string
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Issuer       string
}

type contextKey int

const (
	profile_context_key contextKey = iota
)

func NewAuthenticator(cfg *Config, opts ...Option) (*OIDCAuthenticator, error) {
	var err error
	auth := &OIDCAuthenticator{
		scopes:         []string{oidc.ScopeOpenID, "profile", "email"},
		stateCookieTTL: 15 * time.Minute,
		https:          true,
	}
	if auth.stateCookieName, err = randomString(10); err != nil {
		return nil, err
	}
	if auth.oidcProvider, err = oidc.NewProvider(context.Background(), cfg.Issuer); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(auth)
	}
	auth.oauth2Config = oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     auth.oidcProvider.Endpoint(),
		Scopes:       auth.scopes,
	}
	return auth, nil
}

func (auth *OIDCAuthenticator) LoginHandler(w http.ResponseWriter, req *http.Request) {
	state, err := randomString(16)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cookie := &http.Cookie{
		Name:     auth.stateCookieName,
		Value:    state,
		Expires:  time.Now().Add(auth.stateCookieTTL),
		Path:     "/",
		Secure:   auth.https,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, req, auth.oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (auth *OIDCAuthenticator) CallbackMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie(auth.stateCookieName)
		if err != nil {
			http.Error(w, "state cookie could not be retrieved", http.StatusInternalServerError)
			return
		}
		if req.URL.Query().Get("state") != cookie.Value {
			http.Error(w, "invalid state cookie", http.StatusBadRequest)
			return
		}
		token, err := auth.oauth2Config.Exchange(req.Context(), req.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange an authorization code for a token.", http.StatusUnauthorized)
			return
		}
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "no id_token field in oauth2 token", http.StatusInternalServerError)
			return
		}
		idToken, err := auth.oidcProvider.Verifier(&oidc.Config{
			ClientID: auth.oauth2Config.ClientID,
		}).Verify(req.Context(), rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token.", http.StatusInternalServerError)
			return
		}
		var claims map[string]any
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cookie = &http.Cookie{
			Name:     auth.stateCookieName,
			Value:    "",
			Expires:  time.Time{},
			MaxAge:   -1,
			Path:     "/",
			Secure:   auth.https,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)
		next(w, req.WithContext(context.WithValue(req.Context(), profile_context_key, claims)))
	}
}

func GetClaims(req *http.Request) (map[string]any, error) {
	if value := req.Context().Value(profile_context_key); value == nil {
		return nil, errors.New("claims not found in request context")
	} else if profile, ok := value.(map[string]any); !ok {
		return nil, errors.New("wrong claims type in request context")
	} else {
		return profile, nil
	}
}

func GetStringValueFromClaims(claims map[string]any, key string) string {
	if value, ok := claims[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func randomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than zero")
	}
	byteLen := (length*6 + 7) / 8
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generating random string: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:length], nil
}
