package jwtsession

type Option[T any] func(*JwtSession[T])

func WithSessionCookieHttps[T any](https bool) Option[T] {
	return func(jwts *JwtSession[T]) {
		jwts.sessionCookieHttps = https
	}
}

func WithSessionCookieName[T any](name string) Option[T] {
	return func(jwts *JwtSession[T]) {
		jwts.sessionCookieName = name
	}
}

func WithBearerToken[T any](bearer bool) Option[T] {
	return func(jwts *JwtSession[T]) {
		jwts.bearerToken = bearer
	}
}

func WithHeaderName[T any](name string) Option[T] {
	return func(jwts *JwtSession[T]) {
		jwts.headerName = name
	}
}
