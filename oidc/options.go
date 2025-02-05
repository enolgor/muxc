package oidc

import "time"

type Option func(*OIDCAuthenticator[any])

func WithScopes(scopes ...string) Option {
	return func(auth *OIDCAuthenticator[any]) {
		auth.scopes = scopes
	}
}

func WithStateCookieTTL(ttl time.Duration) Option {
	return func(auth *OIDCAuthenticator[any]) {
		auth.stateCookieTTL = ttl
	}
}

func WithStateCookieHttps(https bool) Option {
	return func(auth *OIDCAuthenticator[any]) {
		auth.stateCookieHttps = https
	}
}

func WithStateCookieName(name string) Option {
	return func(auth *OIDCAuthenticator[any]) {
		auth.stateCookieName = name
	}
}
