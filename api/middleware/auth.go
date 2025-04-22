package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"os"

	"igaku/dtos"
	"igaku/models"
	"igaku/utils"
)

var jwtSecretKey = []byte(os.Getenv("SECRET_KEY"))

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		claims := utils.Claims{}
		token, err := jwt.ParseWithClaims(
			tokenString,
			&claims,
			func(token *jwt.Token) (interface{}, error) {
				return jwtSecretKey, nil
			},
		)

		if err != nil || !token.Valid {
			if err != nil && errors.Is(err, jwt.ErrTokenExpired) {
				c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
					Message: "Token has expired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, dtos.ErrorResponse{
					Message: "Unauthorized",
				})
			}
			c.Abort()
			return
		}

		c.Set("id", claims.RegisteredClaims.Subject)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func Authorize(allowedRoles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Message: "User role not found in context",
			})
			c.Abort()
			return
		}

		userRole, ok := roleStr.(models.Role)
		if !ok {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Message: "Invalid role type in context",
			})
			c.Abort()
			return
		}

		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, dtos.ErrorResponse{
				Message: "Insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
