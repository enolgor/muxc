package oidc

import "time"

type Option func(*OIDCAuthenticator)

func WithScopes(scopes ...string) Option {
	return func(auth *OIDCAuthenticator) {
		auth.scopes = scopes
	}
}

func WithStateCookieTTL(ttl time.Duration) Option {
	return func(auth *OIDCAuthenticator) {
		auth.stateCookieTTL = ttl
	}
}

func WithHttps(https bool) Option {
	return func(auth *OIDCAuthenticator) {
		auth.https = https
	}
}

func WithStateCookieName(name string) Option {
	return func(auth *OIDCAuthenticator) {
		auth.stateCookieName = name
	}
}
