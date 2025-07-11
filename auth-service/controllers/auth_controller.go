package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"fmt"
	"net/http"

	"igaku/auth-service/dtos"
	"igaku/auth-service/services"
	commonsDtos "igaku/commons/dtos"
        commonsErrors "igaku/commons/errors"
	igakuErrors "igaku/auth-service/errors"
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
// @Param	credentials body dtos.LoginCredentials true "User login credentials (username and password)"
// @Success	200 {string} string "Successfully authenticated, returns JWT token"
// @Failure	400 {object} dtos.ErrorResponse "Bad Request - Invalid request payload (e.g., missing fields, wrong format)"
// @Failure	401 {object} dtos.ErrorResponse "Unauthorized - Invalid username or password"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error - Failed to process login (e.g., database error)"
// @Router	/auth/login [post]
func (ctrl *AuthController) Login(c *gin.Context) {
	var creds dtos.LoginCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid request payload",
		})
		return
	}

	token, err := ctrl.service.Login(creds)
	if err != nil {
		if errors.Is(err, &igakuErrors.InvalidUsernameOrPasswordError{}) {
			c.JSON(http.StatusUnauthorized, commonsDtos.ErrorResponse{
				Message: "Invalid login or password",
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve user",
			})
		}
		return
	}

	c.String(http.StatusOK, "%s", token)
}

// Register registers a new user and returns a JWT token on success.
// @Summary	Register in the system
// @Description	Registers a new user via username and password. Returns a JWT token as plain text on success.
// @Tags	Authentication
// @Accept	json
// @Produce	plain
// @Param	fields body dtos.RegistrationFields true "User registration fields (username and password)"
// @Success	200 {string} string "Successfully registered, returns JWT token"
// @Failure	409 {object} dtos.ErrorResponse "Conflict - Username or Email already taken"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error - Failed to process login (e.g., database error)"
// @Router	/auth/register [post]
func (ctrl *AuthController) Register(c *gin.Context) {
	var fields dtos.RegistrationFields
	if err := c.ShouldBindJSON(&fields); err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid request payload",
		})
		return
	}

	token, err := ctrl.service.Register(fields)
	if err != nil {
                var dupEmailErr *commonsErrors.DuplicatedEmailError
		if errors.Is(err, &igakuErrors.UsernameAlreadyTakenError{}) {
			c.JSON(http.StatusConflict, commonsDtos.ErrorResponse{
				Message: fmt.Sprintf("Failed to register: %s",
					err,
				),
			})
		} else if errors.As(err, &dupEmailErr) {
			c.JSON(http.StatusConflict, commonsDtos.ErrorResponse{
				Message: fmt.Sprintf("Failed to register: %s",
					err,
				),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to register",
			})
		}
	}

	c.String(http.StatusOK, "%s", token)
}

func (ctrl *AuthController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/auth")
	{
		routes.POST("/login", ctrl.Login)
		routes.POST("/register", ctrl.Register)
	}
}
