package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"

	"igaku/api-service/services"
	"igaku/commons/models"
	igakuErrors "igaku/commons/errors"
	commonsDtos "igaku/commons/dtos"
)

type InternalAccountController struct {
	service services.AccountService
}

func NewInternalAccountController(service services.AccountService) *InternalAccountController {
	return &InternalAccountController{service: service}
}

func (ctrl *InternalAccountController) FindByUsername(c *gin.Context) {
	username := c.Param("username")

	user, err := ctrl.service.GetAccountByUsername(username)

	if err != nil {
		if errors.Is(err, &igakuErrors.UserNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve account details",
			})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ctrl *InternalAccountController) Persist(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid request body",
		})
		return
	}

	if err := ctrl.service.Persist(&user); err != nil {
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to persist user",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (ctrl *InternalAccountController) RegisterRoutes(router *gin.Engine) {
	// TODO: Change the path prefix to `accounts` here, and to
	// `/api/accounts` everywhere else.
	routes := router.Group("/internal-accounts")
	{
		routes.GET("/find-by-username/:username", ctrl.FindByUsername)
		routes.POST("/persist", ctrl.Persist)
	}
}

// TODO: Unit test these endpoints
