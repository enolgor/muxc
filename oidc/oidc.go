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

type OIDCAuthenticator[T any] struct {
	oidcProvider            *oidc.Provider
	oauth2Config            oauth2.Config
	scopes                  []string
	stateCookieTTL          time.Duration
	stateCookieHttps        bool
	stateCookieName         string
	stateCookieSameSiteMode http.SameSite
}

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Issuer       string
}

func NewAuthenticator[T any](cfg *Config, opts ...Option) (*OIDCAuthenticator[T], error) {
	var err error
	auth := &OIDCAuthenticator[T]{
		scopes:                  []string{oidc.ScopeOpenID, "profile", "email"},
		stateCookieTTL:          15 * time.Minute,
		stateCookieHttps:        true,
		stateCookieSameSiteMode: http.SameSiteLaxMode,
	}
	if auth.stateCookieName, err = randomString(10); err != nil {
		return nil, err
	}
	if auth.oidcProvider, err = oidc.NewProvider(context.Background(), cfg.Issuer); err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt((*OIDCAuthenticator[any])(auth))
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

func (auth *OIDCAuthenticator[T]) LoginRedirect(w http.ResponseWriter, req *http.Request, state string) {
	cookie := &http.Cookie{
		Name:     auth.stateCookieName,
		Value:    state,
		Expires:  time.Now().Add(auth.stateCookieTTL),
		Path:     "/",
		Secure:   auth.stateCookieHttps,
		HttpOnly: true,
		SameSite: auth.stateCookieSameSiteMode,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, req, auth.oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (auth *OIDCAuthenticator[T]) Callback(w http.ResponseWriter, req *http.Request, claims *T) error {
	cookie, err := req.Cookie(auth.stateCookieName)
	if err != nil {
		return errors.New("state cookie could not be retrieved")
	}
	if req.URL.Query().Get("state") != cookie.Value {
		return errors.New("invalid state cookie")
	}
	token, err := auth.oauth2Config.Exchange(req.Context(), req.URL.Query().Get("code"))
	if err != nil {
		return errors.New("failed to exchange an authorization code for a token")
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return errors.New("no id_token field in oauth2 token")
	}
	idToken, err := auth.oidcProvider.Verifier(&oidc.Config{
		ClientID: auth.oauth2Config.ClientID,
	}).Verify(req.Context(), rawIDToken)
	if err != nil {
		return errors.New("failed to verify ID Token")
	}
	if err := idToken.Claims(claims); err != nil {
		return err
	}
	cookie = &http.Cookie{
		Name:     auth.stateCookieName,
		Value:    "",
		Expires:  time.Time{},
		MaxAge:   -1,
		Path:     "/",
		Secure:   auth.stateCookieHttps,
		HttpOnly: true,
		SameSite: auth.stateCookieSameSiteMode,
	}
	http.SetCookie(w, cookie)
	return nil
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
