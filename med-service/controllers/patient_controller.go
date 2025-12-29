package controllers

import (
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"regexp"

	medErrors "igaku/med-service/errors"
	"igaku/med-service/middleware"
	"igaku/med-service/services"
	commonsDtos "igaku/commons/dtos"
	commonsModels "igaku/commons/models"
	commonsErrors "igaku/commons/errors"
)

type PatientController struct {
	service services.PatientService
}

func NewPatientController(service services.PatientService) *PatientController {
	return &PatientController{service: service}
}

// GetByNationalID retrieves data of patient with given national ID
// @Summary	Retrieve data of patient with given national ID (Doctor)
// @Description	Retrieves username, email and national ID of patient. Requires Doctor privileges.
// @Tags	Patient
// @Produce	json
// @Param	national_id path string true "National ID"
// @Success	200  {object}  dtos.PatientDetails "Successfully retrieved patient data"
// @Failure	400  {object}  dtos.ErrorResponse  "Bad Request - Invalid path parameter (national_id)"
// @Failure	401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure	500  {object}  dtos.ErrorResponse  "Internal Server Error - Failed to retrieve patient data"
// @Security	BearerAuth
// @Router	/med/patient/{national_id} [get]
func (ctrl *PatientController) GetByNationalID(c *gin.Context) {
	nationalID := c.Param("national_id")
	re := regexp.MustCompile(`^\d{11}$`)
	if !re.MatchString(nationalID) {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid national_id parameter. Must be a 11-digit number",
		})
		return
	}

	patient, err := ctrl.service.GetPatientByNationalID(nationalID)
	if err != nil {
		var patientNotFoundErr *medErrors.PatientNotFoundError
		var userNotFoundErr *commonsErrors.UserNotFoundError

		if errors.As(err, &patientNotFoundErr) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else if errors.As(err, &userNotFoundErr) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to find patient",
			})
		}
		return
	}

	c.JSON(http.StatusOK, patient)
}

func (ctrl *PatientController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/med/patient")
	routes.Use(middleware.Authenticate())
	{
		routes.GET(
			"/:national_id",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetByNationalID,
		)
	}
}
