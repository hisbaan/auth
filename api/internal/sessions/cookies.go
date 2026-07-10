package sessions

import (
	"net/http"
	"time"
)

const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
)

func SetCookies(w http.ResponseWriter, cookieDomain string, tokens SessionTokens) {
	setCookie(w, cookieDomain, AccessTokenCookieName, tokens.AccessToken, time.Duration(tokens.ExpiresIn)*time.Second)
	setCookie(w, cookieDomain, RefreshTokenCookieName, tokens.RefreshToken, time.Duration(tokens.RefreshExpiresIn)*time.Second)
}

func ClearCookies(w http.ResponseWriter, cookieDomain string) {
	setCookie(w, cookieDomain, AccessTokenCookieName, "", -time.Hour)
	setCookie(w, cookieDomain, RefreshTokenCookieName, "", -time.Hour)
}

func setCookie(w http.ResponseWriter, cookieDomain string, name string, value string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
		Secure:   cookieDomain != "localhost",
		SameSite: http.SameSiteLaxMode,
		Domain:   cookieDomain,
		Path:     "/",
		Expires:  time.Now().Add(ttl),
	})
}
