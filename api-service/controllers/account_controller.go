package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"errors"
	"net/http"
	"strconv"
	"strings"

	"igaku/api-service/middleware"
	"igaku/api-service/services"
	"igaku/api-service/utils"
	"igaku/commons/models"
	commonsDtos "igaku/commons/dtos"
	igakuErrors "igaku/commons/errors"
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
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "User ID not found in context",
		})
		return
	}

	id, err := uuid.Parse(idStr.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Invalid user ID format in token",
		})
		return
	}

	details, err := ctrl.service.GetAccountDetails(id)
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

	c.JSON(http.StatusOK, details)
}

// ListAccounts retrieves a paginated list of all users (Admin only).
// @Summary	List All Accounts (Admin)
// @Description	Retrieves a paginated list of all user accounts. Requires Admin privileges.
// @Tags	Accounts
// @Produce	json
// @Param	page query int false "Page number (default: 1)" minimum(1)
// @Param	pageSize query int false "Number of items per page (default: 10)" minimum(1) maximum(100)
// @Param	orderBy query string false "Field name to be ordered by: id (default) or username"
// @Param	orderMethod query string false "Ordering method: asc (default) or desc"
// @Success	200  {object}  dtos.PaginatedResponse{data=[]dtos.AccountDetailsWithID} "Successfully retrieved list of accounts"
// @Failure	400  {object}  dtos.ErrorResponse  "Bad Request - Invalid query parameters (page, pageSize, orderBy, orderMethod)"
// @Failure	401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Admin role"
// @Failure	500  {object}  dtos.ErrorResponse  "Internal Server Error - Failed to retrieve accounts"
// @Security	BearerAuth
// @Router	/accounts/list [get]
func (ctrl *AccountController) ListAccounts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	orderByStr := c.DefaultQuery("orderBy", "id")
	orderMethodStr := c.DefaultQuery("orderMethod", "asc")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid page parameter. Must be a positive integer.",
		})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid pageSize parameter. Must be a positive integer.",
		})
		return
	}

	orderBy, ok := models.UserOrderableFieldsMap[strings.ToLower(orderByStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderBy parameter. Must be `id` or `username`",
		})
		return
	}

	orderMethod, ok := utils.OrderingsMap[strings.ToLower(orderMethodStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderMethod parameter. Must be `asc` or `desc`",
		})
		return
	}

	accList, err := ctrl.service.ListAccounts(page, pageSize, orderBy, orderMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
			Message: "Failed to retrieve accounts list",
		})
		return
	}

	c.JSON(http.StatusOK, accList)
}

func (ctrl *AccountController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/accounts")
	routes.Use(middleware.Authenticate())
	{
		routes.GET("/self", ctrl.GetSelf)
		routes.GET(
			"/list",
			middleware.Authorize(models.Admin),
			ctrl.ListAccounts,
		)
	}
}
