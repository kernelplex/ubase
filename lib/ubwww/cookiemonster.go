package ubwww

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kernelplex/ubase/lib/ubsecurity"
)

type CookieContextKey string
type IdentityContextKey string

type UserIdentity struct {
	UserID         int64
	Email          string
	OrganizationID int64
}

func (u *UserIdentity) ToAgent() string {
	return fmt.Sprintf("user:%d", u.UserID)
}

type AuthTokenCookie interface {
	ToUserIdentity() UserIdentity
	IsExpired() bool
	Touch(unixSec int64)
}

type AuthTokenCookieManager[T AuthTokenCookie] interface {
	ClearAuthTokenCookie(w http.ResponseWriter)
	ReadAuthTokenCookie(r *http.Request) (bool, T, error)
	WriteAuthTokenCookie(w http.ResponseWriter, token T) error
	Middleware(handler http.Handler) http.Handler
	MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc
	TokenFromContext(ctx context.Context) (T, bool)
	IdentityFromContext(ctx context.Context) (UserIdentity, bool)
}

type CookieMonster[T AuthTokenCookie] struct {
	encryptionService ubsecurity.EncryptionService
	cookieName        string
	tokenSoftExpiry   int
	secure            bool
	cookieKey         CookieContextKey
	identityKey       IdentityContextKey
}

func NewCookieMonster[T AuthTokenCookie](
	encryptionService ubsecurity.EncryptionService,
	cookieName string,
	secure bool,
	tokenSoftExpiry int64,
	cookieKey CookieContextKey,
	identityKey IdentityContextKey) AuthTokenCookieManager[T] {
	return &CookieMonster[T]{
		encryptionService: encryptionService,
		cookieName:        cookieName,
		secure:            secure,
		tokenSoftExpiry:   int(tokenSoftExpiry),
		cookieKey:         cookieKey,
		identityKey:       identityKey,
	}
}

func (c *CookieMonster[T]) ClearAuthTokenCookie(w http.ResponseWriter) {
	cookie := http.Cookie{
		Name:     c.cookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	}

	slog.Debug("Writing auth token cookie", "cookie", cookie)

	http.SetCookie(w, &cookie)
}

func (c *CookieMonster[T]) ReadAuthTokenCookie(r *http.Request) (bool, T, error) {
	var zero T
	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		slog.Debug("Error reading auth token cookie", "error", err)
		return false, zero, err
	}
	slog.Debug("Auth token cookie found", "cookie", cookie)
	tokenBytes, err := c.encryptionService.Decrypt64(cookie.Value)
	if err != nil {
		slog.Debug("Error decrypting auth token cookie", "error", err)
		return false, zero, fmt.Errorf("failed to decrypt auth token cookie: %w", err)
	}
	slog.Debug("Decrypted auth token cookie", "cookie", cookie)

	var jsonToken T
	err = json.Unmarshal(tokenBytes, &jsonToken)
	if err != nil {
		slog.Debug("Error unmarshaling auth token cookie", "error", err)
		return false, zero, fmt.Errorf("failed to unmarshal auth token cookie: %w", err)
	}
	slog.Debug("Auth token cookie unmarshaled", "token", jsonToken)
	return true, jsonToken, nil
}

func (c *CookieMonster[T]) WriteAuthTokenCookie(w http.ResponseWriter, token T) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	encryptedToken, err := c.encryptionService.Encrypt64(string(tokenJson))
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     c.cookieName,
		Value:    string(encryptedToken),
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   c.tokenSoftExpiry,
		Path:     "/",
	}

	slog.Debug("Writing auth token cookie", "cookie", cookie)

	http.SetCookie(w, &cookie)
	return nil
}

func (c *CookieMonster[T]) TokenFromContext(ctx context.Context) (T, bool) {
	token, ok := ctx.Value(c.cookieKey).(T)
	return token, ok
}

func (c *CookieMonster[T]) IdentityFromContext(ctx context.Context) (UserIdentity, bool) {
	identity, ok := ctx.Value(c.identityKey).(UserIdentity)
	return identity, ok
}

func (c *CookieMonster[T]) middlewareHandler(w http.ResponseWriter, r *http.Request) *http.Request {
	found, token, err := c.ReadAuthTokenCookie(r)

	if err == nil && found {
		if token.IsExpired() {
			slog.Debug("Auth token cookie is expired, clearing cookie")
			c.ClearAuthTokenCookie(w)
		} else {
			updateTime := time.Now().Add(time.Duration(c.tokenSoftExpiry) * time.Second).Unix()
			token.Touch(updateTime)
			err = c.WriteAuthTokenCookie(w, token)
			if err != nil {
				slog.Error("Error updating auth token cookie", "error", err)
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, c.cookieKey, token)
			ctx = context.WithValue(ctx, c.identityKey, token.ToUserIdentity())
			r = r.WithContext(ctx)
		}
	}
	return r
}

func (c *CookieMonster[T]) MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = c.middlewareHandler(w, r)
		handler(w, r)
	}
}

func (c *CookieMonster[T]) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = c.middlewareHandler(w, r)
		handler.ServeHTTP(w, r)
	})
}
