package utils

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CaribouBlue/mixtape/internal/core"
	jwt "github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired error = errors.New("token expired")
	ErrInvalidToken error = errors.New("invalid token")
)

const (
	CookieAuthorization string = "authorization"
)

type AuthorizationCookie struct {
	UserId  int64 `json:"userId"`
	Expires int64 `json:"expires"`
}

const secretKey = "super secret"

func SetAuthCookie(w http.ResponseWriter, u *core.UserEntity) error {
	expirationDuration := time.Hour * 24

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

	http.SetCookie(w, &http.Cookie{
		Name:     CookieAuthorization,
		Value:    tokenString,
		Path:     "/",
		MaxAge:   int(expirationDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	return nil
}

func ParseAuthCookie(w http.ResponseWriter, r *http.Request) (*core.UserEntity, error) {
	cookie, err := r.Cookie(CookieAuthorization)
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
			ClearAuthCookie(w)
			return nil, ErrTokenExpired
		}

		userId := int64(claims["userId"].(float64))
		return &core.UserEntity{Id: userId}, nil
	} else {
		return nil, ErrInvalidToken
	}
}

func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAuthorization,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})
}
