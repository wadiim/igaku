package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"os"

	"igaku/dtos"
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
