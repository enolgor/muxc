package jwtsession

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims[T any] struct {
	Session T `json:"session"`
	jwt.RegisteredClaims
}

type JwtSession[T any] struct {
	key                []byte
	expiration         time.Duration
	sessionCookieHttps bool
	sessionCookieName  string
	parser             *jwt.Parser
	bearerToken        bool
	headerName         string
}

func NewJwtSession[T any](key []byte, expiration time.Duration, opts ...Option) *JwtSession[T] {
	jwts := &JwtSession[T]{
		key:                key,
		expiration:         expiration,
		sessionCookieHttps: true,
		sessionCookieName:  "session",
		parser:             jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})),
		bearerToken:        true,
		headerName:         "Authorization",
	}
	for _, opt := range opts {
		opt((*JwtSession[any])(jwts))
	}
	return jwts
}

func (jwts *JwtSession[T]) ForgeToken(session T) (string, time.Time, error) {
	now := time.Now()
	expiration := now.Add(jwts.expiration)
	claims := Claims[T]{
		session,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	tkn, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwts.key)
	return tkn, expiration, err
}

func (jwts *JwtSession[T]) ForgeCookie(session T) (*http.Cookie, error) {
	tkn, expiration, err := jwts.ForgeToken(session)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     jwts.sessionCookieName,
		Value:    tkn,
		Expires:  expiration,
		Path:     "/",
		Secure:   jwts.sessionCookieHttps,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	return cookie, nil
}

func (jwts *JwtSession[T]) ForgeHeader(session T) (string, error) {
	tkn, _, err := jwts.ForgeToken(session)
	if jwts.bearerToken {
		tkn = "Bearer " + tkn
	}
	return tkn, err
}

func (jwts *JwtSession[T]) GetSessionFromCookie(req *http.Request) (*T, error) {
	if cookie, err := req.Cookie(jwts.sessionCookieName); err != nil {
		return nil, errors.New("session cookie not found")
	} else {
		return jwts.GetSessionFromToken(cookie.Value)
	}
}

func (jwts *JwtSession[T]) GetSessionFromHeader(req *http.Request) (*T, error) {
	header := req.Header.Get(jwts.headerName)
	if header == "" {
		return nil, errors.New("authorization header not found")
	}
	if jwts.bearerToken {
		if !strings.HasSuffix(header, "Bearer ") {
			return nil, errors.New("authorization header not found")
		}
		return jwts.GetSessionFromToken(header[7:])
	}
	return jwts.GetSessionFromToken(header)
}

func (jwts *JwtSession[T]) GetSessionFromToken(tkn string) (*T, error) {
	claims := &Claims[T]{}
	if _, err := jwts.parser.ParseWithClaims(tkn, claims, func(token *jwt.Token) (interface{}, error) {
		return jwts.key, nil
	}); err != nil {
		return nil, err
	}
	return &claims.Session, nil
}

func (jwts *JwtSession[T]) ExpireCookie() *http.Cookie {
	return &http.Cookie{
		Name:     jwts.sessionCookieName,
		Value:    "",
		Expires:  time.Time{},
		MaxAge:   -1,
		Path:     "/",
		Secure:   jwts.sessionCookieHttps,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
