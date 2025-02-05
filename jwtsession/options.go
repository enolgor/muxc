package jwtsession

type Option func(*JwtSession[any])

func WithSessionCookieHttps(https bool) Option {
	return func(jwts *JwtSession[any]) {
		jwts.sessionCookieHttps = https
	}
}

func WithSessionCookieName(name string) Option {
	return func(jwts *JwtSession[any]) {
		jwts.sessionCookieName = name
	}
}

func WithBearerToken(bearer bool) Option {
	return func(jwts *JwtSession[any]) {
		jwts.bearerToken = bearer
	}
}

func WithHeaderName(name string) Option {
	return func(jwts *JwtSession[any]) {
		jwts.headerName = name
	}
}
