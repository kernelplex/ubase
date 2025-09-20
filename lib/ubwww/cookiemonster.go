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

type AuthTokenCookieManager interface {
	ClearAuthTokenCookie(w http.ResponseWriter)
	ReadAuthTokenCookie(r *http.Request) (bool, AuthToken, error)
	WriteAuthTokenCookie(w http.ResponseWriter, token AuthToken) error
	Middleware(handler http.Handler) http.Handler
	MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc
	TokenFromContext(ctx context.Context) (AuthToken, bool)
	IdentityFromContext(ctx context.Context) (UserIdentity, bool)
}

type CookieMonster struct {
	encryptionService ubsecurity.EncryptionService
	cookieName        string
	tokenSoftExpiry   int
	secure            bool
	cookieKey         CookieContextKey
	identityKey       IdentityContextKey
}

func NewCookieMonster(
	encryptionService ubsecurity.EncryptionService,
	cookieName string,
	secure bool,
	tokenSoftExpiry int64,
	cookieKey CookieContextKey,
	identityKey IdentityContextKey) AuthTokenCookieManager {
	return &CookieMonster{
		encryptionService: encryptionService,
		cookieName:        cookieName,
		secure:            secure,
		tokenSoftExpiry:   int(tokenSoftExpiry),
		cookieKey:         cookieKey,
		identityKey:       identityKey,
	}
}

func (c *CookieMonster) ClearAuthTokenCookie(w http.ResponseWriter) {
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

func (c *CookieMonster) ReadAuthTokenCookie(r *http.Request) (bool, AuthToken, error) {
	var zero AuthToken
	cookie, err := r.Cookie(c.cookieName)
	if err != nil {
		slog.Debug("Error reading auth token cookie", "error", err)
		return false, zero, err
	}
	slog.Debug("Auth token cookie found", "cookie", cookie.Value)
	tokenBytes, err := c.encryptionService.Decrypt64(cookie.Value)
	if err != nil {
		slog.Debug("Error decrypting auth token cookie", "error", err)
		return false, zero, fmt.Errorf("failed to decrypt auth token cookie: %w", err)
	}
	slog.Debug("************ Decrypted auth token cookie", "cookie", string(tokenBytes))
	slog.Debug("************ Decrypted auth token cookie", "cookie", []byte(string(tokenBytes)))

	var jsonToken AuthToken
	err = json.Unmarshal([]byte(string(tokenBytes)), &jsonToken)
	if err != nil {
		slog.Debug("Error unmarshaling auth token cookie", "error", err)
		return false, zero, fmt.Errorf("failed to unmarshal auth token cookie: %w", err)
	}
	slog.Debug("Auth token cookie unmarshaled", "token", jsonToken)
	return true, jsonToken, nil
}

func (c *CookieMonster) WriteAuthTokenCookie(w http.ResponseWriter, token AuthToken) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	newToken := AuthToken{}
	err = json.Unmarshal(tokenJson, &newToken)
	if err != nil {
		slog.Debug("Error re-unmarshaling auth token cookie", "error", err)
		return fmt.Errorf("failed to re-unmarshal auth token cookie: %w", err)
	}
	slog.Debug("Re-unmarshaled auth token cookie", "token", newToken)

	slog.Debug("Auth token cookie JSON", "token_json", string(tokenJson))
	encryptedToken, err := c.encryptionService.Encrypt64(string(tokenJson))
	slog.Debug("Encrypted auth token cookie", "encrypted_token", encryptedToken)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     c.cookieName,
		Value:    encryptedToken,
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

func (c *CookieMonster) TokenFromContext(ctx context.Context) (AuthToken, bool) {
	token, ok := ctx.Value(c.cookieKey).(AuthToken)
	return token, ok
}

func (c *CookieMonster) IdentityFromContext(ctx context.Context) (UserIdentity, bool) {
	identity, ok := ctx.Value(c.identityKey).(UserIdentity)
	return identity, ok
}

func (c *CookieMonster) middlewareHandler(w http.ResponseWriter, r *http.Request) *http.Request {
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

func (c *CookieMonster) MiddlewareFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = c.middlewareHandler(w, r)
		handler(w, r)
	}
}

func (c *CookieMonster) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = c.middlewareHandler(w, r)
		handler.ServeHTTP(w, r)
	})
}
