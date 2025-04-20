package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"errors"
	"net/http"

	igakuErrors "igaku/errors"
	"igaku/dtos"
	"igaku/middleware"
	"igaku/services"
)

type AccountController struct {
	service services.AccountService
}

func NewAccountController(service services.AccountService) *AccountController {
	return &AccountController{service: service}
}

// GetSelfAccountDetails retrieves details for the currently authenticated user.
// @Summary	Get Own Account Details
// @Description	Retrieves details (username, role, etc.) for the currently logged-in user based on the provided JWT.
// @Tags	Accounts
// @Produce	json
// @Success	200 {object} dtos.AccountDetails "Successfully retrieved account details"
// @Failure	401 {object} dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	404 {object} dtos.ErrorResponse  "Not Found - User associated with token not found"
// @Failure	500 {object} dtos.ErrorResponse  "Internal Server Error - Failed to retrieve account details"
// @Security	BearerAuth
// @Router	/accounts/self [get]
func (ctrl *AccountController) GetSelf(c *gin.Context) {
	idStr, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Message: "User ID not found in context",
		})
		return
	}

	id, err := uuid.Parse(idStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
			Message: "Invalid user ID format in token",
		})
		return
	}

	details, err := ctrl.service.GetAccountDetails(id)
	if err != nil {
		if errors.Is(err, &igakuErrors.UserNotFoundError{}) {
			c.JSON(http.StatusNotFound, dtos.ErrorResponse{
				Message: "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, dtos.ErrorResponse{
				Message: "Failed to retrieve account details",
			})
		}
		return
	}

	c.JSON(http.StatusOK, details)
}

func (ctrl *AccountController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/accounts")
	routes.Use(middleware.Authenticate())
	{
		routes.GET("/self", ctrl.GetSelf)
	}
}
