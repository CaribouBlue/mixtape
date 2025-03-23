package utils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CaribouBlue/mixtape/internal/config"
	"github.com/CaribouBlue/mixtape/internal/core"
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired error = errors.New("token expired")
	ErrInvalidToken error = errors.New("invalid token")
)

type CookieName = string

const (
	CookieNameAuthorization        CookieName = "authorization"
	CookieNameSessionCorrelationId CookieName = "sessionCorrelationId"
)

func CookieFactory(name CookieName, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}
}

func SetAuthCookie(w http.ResponseWriter, u *core.UserEntity) error {
	expirationDuration := time.Hour * 24
	secretKey := config.GetConfigValue(config.ConfJwtSecret)

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":  u.Id,
		"expires": time.Now().Add(expirationDuration).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return err
	}

	http.SetCookie(w, CookieFactory(CookieNameAuthorization, tokenString, int(expirationDuration.Seconds())))

	return nil
}

func ParseAuthCookie(w http.ResponseWriter, r *http.Request) (*core.UserEntity, error) {
	secretKey := config.GetConfigValue(config.ConfJwtSecret)

	cookie, err := r.Cookie(CookieNameAuthorization)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		expires := int64(claims["expires"].(float64))
		if time.Now().Unix() > expires {
			DeleteCookie(w, r, CookieNameAuthorization)
			return nil, ErrTokenExpired
		}

		userId := int64(claims["userId"].(float64))
		return &core.UserEntity{Id: userId}, nil
	} else {
		return nil, ErrInvalidToken
	}
}

func DeleteCookie(w http.ResponseWriter, r *http.Request, cookieName string) error {
	http.SetCookie(w, CookieFactory(cookieName, "", -1))
	return nil
}

func RefreshCookie(w http.ResponseWriter, r *http.Request, cookieName string) error {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return err
	}

	http.SetCookie(w, CookieFactory(cookieName, cookie.Value, cookie.MaxAge))
	return nil
}
