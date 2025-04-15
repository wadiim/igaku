package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"errors"
	"net/http"

	"igaku/dto"
	igaku_errors "igaku/errors"
	"igaku/services"
)

type AuthController struct {
	service services.AuthService
}

func NewAuthController(service services.AuthService) *AuthController {
	return &AuthController{service: service}
}

// Login authenticates the user and returns a JWT token on success.
// @Summary	Login into the system
// @Description	Authenticates a user via username and password. Returns a JWT token as plain text on success.
// @Tags	Authentication
// @Accept	json
// @Produce	plain
// @Param	credentials body dto.LoginCredentials true "User login credentials (username and password)"
// @Success	200 {string} string "Successfully authenticated, returns JWT token"
// @Failure	400 {object} dto.ErrorResponse "Bad Request - Invalid request payload (e.g., missing fields, wrong format)"
// @Failure	401 {object} dto.ErrorResponse "Not Found - Invalid username or password"
// @Failure	500 {object} dto.ErrorResponse "Internal Server Error - Failed to process login (e.g., database error)"
// @Router	/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var creds dto.LoginCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Message: "Invalid request payload",
		})
		return
	}

	token, err := ctrl.service.Login(creds)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, &igaku_errors.InvalidUsernameOrPasswordError{}) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Message: "Invalid login or password",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Message: "Failed to retrieve user",
			})
		}
		return
	}

	c.String(http.StatusOK, "%s", token)
}

func (ctrl *AuthController) RegisterRoutes(router *gin.Engine) {
	router.POST("/login", ctrl.Login)
}
