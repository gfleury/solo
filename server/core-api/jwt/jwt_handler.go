package jwt

import (
	"context"
	"net/http"
	"strings"

	"github.com/gfleury/solo/server/oauth2/session"
	"github.com/golang-jwt/jwt"
)

type ClaimsCtx string

var Claims ClaimsCtx = "Claims"

func VerifyJWT(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt_token := ""
		jwt_cookie, err := r.Cookie("JWT")
		if err != nil {
			auth_header := r.Header.Get("Authorization")
			if jwt_header := strings.Split(auth_header, "Bearer "); len(jwt_header) > 1 {
				jwt_token = jwt_header[1]
			}
		} else {
			jwt_token = jwt_cookie.Value
		}

		if jwt_token != "" {
			token, err := jwt.Parse(jwt_token, func(token *jwt.Token) (interface{}, error) {
				_, ok := token.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					_, err := w.Write([]byte("You're Unauthorized"))
					if err != nil {
						return nil, err
					}
				}
				return session.JWT_SECRET, nil
			})
			// parsing errors result
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("You're Unauthorized due to error parsing the JWT"))
				return
			}
			// if there's a token
			if token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					ctx := context.WithValue(r.Context(), Claims, claims)
					h.ServeHTTP(w, r.WithContext(ctx))
					return
				}

				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid JWT Claims"))
				return

			} else {
				w.WriteHeader(http.StatusUnauthorized)
				_, err := w.Write([]byte("You're Unauthorized due to invalid token"))
				if err != nil {
					return
				}
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("You're Unauthorized due to No token in the header"))
			if err != nil {
				return
			}
		}
		// response for if there's no token header
	})
}

func getFromClaim(c context.Context, key string) string {
	if claims, ok := c.Value(Claims).(jwt.MapClaims); ok {
		return claims[key].(string)
	}
	return ""
}

func GetUserFromClaim(c context.Context) string {
	return getFromClaim(c, "user")
}

func GetNameFromClaim(c context.Context) string {
	return getFromClaim(c, "name")
}

func GetEmailFromClaim(c context.Context) string {
	return getFromClaim(c, "email")
}

func GetProviderFromClaim(c context.Context) string {
	return getFromClaim(c, "provider")
}

func GetAvatarFromClaim(c context.Context) string {
	return getFromClaim(c, "avatar")
}
