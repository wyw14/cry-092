package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const principalKey = "principal"

type Principal struct {
	UserID string
	Roles  map[string]bool
}

func Authenticate(signingKey []byte, issuer string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorized(c)
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return signingKey, nil
		}, jwt.WithIssuer(issuer), jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			unauthorized(c)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			unauthorized(c)
			return
		}
		sub, err := claims.GetSubject()
		if err != nil || sub == "" {
			unauthorized(c)
			return
		}
		roles := map[string]bool{}
		if values, ok := claims["roles"].([]any); ok {
			for _, v := range values {
				if role, ok := v.(string); ok {
					roles[role] = true
				}
			}
		}
		c.Set(principalKey, Principal{UserID: sub, Roles: roles})
		c.Next()
	}
}
func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHENTICATED", "message": "访问令牌无效或已过期", "field_errors": []any{}, "request_id": CurrentRequestID(c)})
}
func Current(c *gin.Context) (Principal, bool) {
	value, ok := c.Get(principalKey)
	if !ok {
		return Principal{}, false
	}
	p, ok := value.(Principal)
	return p, ok
}
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, ok := Current(c)
		if !ok {
			unauthorized(c)
			return
		}
		for _, role := range roles {
			if p.Roles[role] {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "当前身份没有操作权限", "field_errors": []any{}, "request_id": CurrentRequestID(c)})
	}
}
