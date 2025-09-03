package middleware

import (
	"fmt"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/th1enq/ViettelSMS_ServerService/internal/delivery/http/presenter"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
)

type JWTMiddleware interface {
	RequireAuth() gin.HandlerFunc
	RequireScope(requireScope string) gin.HandlerFunc
}

type jwtMiddleware struct {
	presenter presenter.Presenter
	jwtSecret []byte
}

func NewJWTMiddleware(
	presenter presenter.Presenter,
	jwtSecret []byte,
) JWTMiddleware {
	return &jwtMiddleware{
		presenter: presenter,
		jwtSecret: jwtSecret,
	}
}

func (s *jwtMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			if apiKey == "<API_KEY>" {
				c.Next()
				return
			}
		}

		token := s.extractTokenFromHeader(c)
		if token == "" {
			s.presenter.Unauthorized(c, "Missing or malformed token", fmt.Errorf("missing or malformed token"))
			c.Abort()
			return
		}

		claims, err := s.validateToken(token)
		if err != nil {
			s.presenter.Unauthorized(c, "Invalid token", err)
			c.Abort()
			return
		}

		if claims.Blocked {
			s.presenter.Forbidden(c, "User is blocked", fmt.Errorf("user is blocked"))
			c.Abort()
			return
		}

		c.Set("userID", claims.Sub)
		c.Set("scopes", claims.Scopes)

		c.Next()
	}
}

func (s *jwtMiddleware) RequireScope(requireScope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("scopes")
		if !exists {
			s.presenter.Forbidden(c, "Insufficient scope", fmt.Errorf("insufficient scope"))
			c.Abort()
			return
		}

		if slices.Contains(scopes.([]string), requireScope) {
			c.Next()
			return
		}

		s.presenter.Forbidden(c, "Insufficient scope", fmt.Errorf("insufficient scope"))
		c.Abort()
	}
}

func (s *jwtMiddleware) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(authHeader, "Bearer ")
}

func (s *jwtMiddleware) validateToken(token string) (*dto.Claims, error) {
	accessToken, err := jwt.ParseWithClaims(token, &dto.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected token signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := accessToken.Claims.(*dto.Claims); ok && accessToken.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims: %w", err)
}
