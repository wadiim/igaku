package controllers

import (
	"github.com/google/uuid"
	"github.com/gin-gonic/gin"

	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	commonsDtos "igaku/commons/dtos"
	commonsErrors "igaku/commons/errors"
	commonsModels "igaku/commons/models"
	commonsUtils "igaku/commons/utils"
	"igaku/med-service/dtos"
	medErrors "igaku/med-service/errors"
	"igaku/med-service/middleware"
	"igaku/med-service/models"
	"igaku/med-service/services"
)

type MedController struct {
	service services.MedService
}

func NewMedController(service services.MedService) *MedController {
	return &MedController{service: service}
}

// GetBySubstring retrieves the list of diseases that match provided substring
// @Summary	List diseases that match provided substring (Doctor)
// @Description	Retrieves a paginated list of diseases that match provided substring. Requires Doctor privileges.
// @Tags	Diseases
// @Produce	json
// @Param	name path string false "Partial disease name to search for (case‑insensitive)"
// @Param	page query int false "Page number (default: 1)" minimum(1)
// @Param	pageSize query int false "Number of items per page (default: 10)" minimum(1) maximum(100)
// @Success	200  {object}  commonsDtos.PaginatedResponse{data=[]dtos.DiseaseDetails} "Successfully retrieved list of diseases"
// @Failure	400  {object}  dtos.ErrorResponse  "Bad Request - Invalid query parameters (name, page, pageSize)"
// @Failure	401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure	500  {object}  dtos.ErrorResponse  "Internal Server Error - Failed to retrieve diseases"
// @Security	BearerAuth
// @Router	/med/disease/{name} [get]
func (ctrl *MedController) GetBySubstring(c *gin.Context) {
	name := c.Param("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "5")

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

	diseases, err := ctrl.service.GetBySubstring(name, page, pageSize)
	if err != nil {
		if errors.Is(err, &medErrors.DiseaseNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve disease details",
			})
		}
		return
	}

	c.JSON(http.StatusOK, diseases)
}

// GetRecommendedDrugs returns a paginated, sortable list of drugs that are recommended for a given disease.
// @Summary	List recommended drugs for a disease (Doctor)
// @Description	Retrieves drugs that are indicated for the specified disease. Supports pagination and sorting. Requires Doctor privileges.
// @Tags	Drugs
// @Produce	json
// @Param	disease path string true "Disease RxNormID"
// @Param	page query int false "Page number (default: 1)" minimum(1)
// @Param	pageSize query int false "Number of items per page (default: 5)" minimum(1) maximum(100)
// @Param	orderBy query string false "Field to sort by: `id`, `name` or `substance` (default: `id`)" enums(id,name,substance)
// @Param	orderMethod query string false "Sort direction: `asc` or `desc` (default: `asc`)" enums(asc,desc)
// @Success	200 {object} commonsDtos.PaginatedResponse{data=[]dtos.DrugDetails} "Successfully retrieved list of recommended drugs"
// @Failure	400 {object} dtos.ErrorResponse "Bad Request – Invalid query parameters"
// @Failure	401 {object} dtos.ErrorResponse "Unauthorized – Invalid or missing token"
// @Failure	403 {object} dtos.ErrorResponse "Forbidden – User does not have Doctor role"
// @Failure	404 {object} dtos.ErrorResponse "Not Found – Drug not found"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error – Failed to retrieve drugs"
// @Security	BearerAuth
// @Router    /med/drug/recommend/{disease} [get]
func (ctrl *MedController) GetRecommendedDrugs(c *gin.Context) {
	disease := c.Param("disease")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "5")
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

	orderBy, ok := models.DrugOrderableFieldsMap[strings.ToLower(orderByStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderBy parameter. Must be `id`, `name` or `substance`",
		})
		return
	}

	orderMethod, ok := commonsUtils.OrderingsMap[strings.ToLower(orderMethodStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderMethod parameter. Must be `asc` or `desc`",
		})
		return
	}

	drugs, err := ctrl.service.GetRecommendedDrugs(disease, page, pageSize, orderBy, orderMethod)
	if err != nil {
		if errors.Is(err, &medErrors.RxClassUnavailableError{}) {
			c.JSON(http.StatusServiceUnavailable, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else if errors.Is(err, &medErrors.SubstanceNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else if errors.Is(err, &medErrors.DrugNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve list of drugs",
			})
			return
		}
	}

	c.JSON(http.StatusOK, drugs)
}

// GetDrugsByName returns a paginated, sortable list of drugs that match a given name substring.
// @Summary	List drugs by name substring (Doctor)
// @Description	Retrieves drugs whose names contain the supplied substring. Supports pagination and sorting. Requires Doctor privileges.
// @Tags	Drugs
// @Produce	json
// @Param	name path string true "Partial drug name to search for (case‑insensitive)"
// @Param	page query int false "Page number (default: 1)" minimum(1)
// @Param	pageSize query int false "Number of items per page (default: 5)" minimum(1) maximum(100)
// @Param	orderBy query string false "Field to sort by: `id`, `name` or `substance` (default: `id`)" enums(id,name,substance)
// @Param	orderMethod query string false "Sort direction: `asc` or `desc` (default: `asc`)" enums(asc,desc)
// @Success	200 {object} commonsDtos.PaginatedResponse{data=[]dtos.DrugDetails} "Successfully retrieved list of drugs"
// @Failure	400 {object} dtos.ErrorResponse "Bad Request – Invalid query parameters"
// @Failure	401 {object} dtos.ErrorResponse "Unauthorized – Invalid or missing token"
// @Failure	403 {object} dtos.ErrorResponse "Forbidden – User does not have Doctor role"
// @Failure	404 {object} dtos.ErrorResponse "Not Found – No matching drugs"
// @Failure	500 {object} dtos.ErrorResponse "Internal Server Error – Failed to retrieve drugs"
// @Security	BearerAuth
// @Router    /med/drug/{name} [get]
func (ctrl *MedController) GetDrugsByName(c *gin.Context) {
	name := c.Param("name")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "5")
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

	orderBy, ok := models.DrugOrderableFieldsMap[strings.ToLower(orderByStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderBy parameter. Must be `id`, `name` or `substance`",
		})
		return
	}

	orderMethod, ok := commonsUtils.OrderingsMap[strings.ToLower(orderMethodStr)]
	if !ok {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid orderMethod parameter. Must be `asc` or `desc`",
		})
		return
	}

	drugs, err := ctrl.service.GetDrugsByName(name, page, pageSize, orderBy, orderMethod)
	if err != nil {
		if errors.Is(err, &medErrors.RxClassUnavailableError{}) {
			c.JSON(http.StatusServiceUnavailable, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else if errors.Is(err, &medErrors.DrugNotFoundError{}) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
			return
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to retrieve list of drugs",
			})
			return
		}
	}

	c.JSON(http.StatusOK, drugs)
}

// GetByNationalID retrieves data of patient with given national ID
// @Summary	Retrieve data of patient with given national ID (Doctor)
// @Description	Retrieves username, email and national ID of patient. Requires Doctor privileges.
// @Tags	Patient
// @Produce	json
// @Param	national_id path string true "11‑digit national identification number of the patient"
// @Success	200  {object}  dtos.PatientDetails "Successfully retrieved patient data"
// @Failure	400  {object}  dtos.ErrorResponse  "Bad Request - Invalid path parameter (national_id)"
// @Failure	401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure	403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure	500  {object}  dtos.ErrorResponse  "Internal Server Error - Failed to retrieve patient data"
// @Security	BearerAuth
// @Router	/med/patient/{national_id} [get]
func (ctrl *MedController) GetByNationalID(c *gin.Context) {
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

// CreatePrescription creates a new prescription and adds a medical‑history entry for the patient.
// @Summary      Create a prescription (Doctor)
// @Description  Stores a prescription (patient, disease, drugs) and creates a medical‑history record linking the patient with the prescribing doctor.
// @Tags         Prescription
// @Accept       json
// @Produce      json
// @Param        prescription body dtos.PrescriptionDetails true "Prescription payload"
// @Success      201 {object} nil "Prescription created successfully"
// @Failure      400 {object} commonsDtos.ErrorResponse "Invalid doctor ID"
// @Failure      401  {object}  dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure      403  {object}  dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure      500 {object} commonsDtos.ErrorResponse "Server error"
// @Security     BearerAuth
// @Router       /med/prescribe [post]
func (ctrl *MedController) CreatePrescription(c *gin.Context) {
	val, exists := c.Get("id")
	if !exists {
		c.JSON(http.StatusUnauthorized, commonsDtos.ErrorResponse{
			Message: "Doctor ID not found in token",
		})
		return
	}

	doctorID, err := uuid.Parse(val.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid doctor ID",
		})
	}

	var prescription dtos.PrescriptionDetails
	if err := c.ShouldBindJSON(&prescription); err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid request payload",
		})
		return
	}

	patientID := prescription.Patient.ID
	drugs := prescription.Drugs
	err = ctrl.service.AddMedicalHistoryItem(patientID, doctorID, drugs)
	if err != nil {
		var medHistItemInsertError *medErrors.MedicalHistoryItemInsertError

		if errors.As(err, &medHistItemInsertError) {
			c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, nil)
}

// GetMedicalHistoryItemByPatientID returns the medical history records for a given patient.
// @Summary      Get medical history items by patient ID (Doctor)
// @Description  Retrieves a list of medical history entries for the specified patient UUID.
// @Tags         MedicalHistory
// @Produce      json
// @Param        patient_id path string true "Patient UUID"
// @Success      200 {object} []models.MedicalHistoryItem
// @Failure      400 {object} commonsDtos.ErrorResponse "Bad Request - Invalid patient ID"
// @Failure      401 {object} dtos.ErrorResponse  "Unauthorized - Invalid or missing token"
// @Failure      403 {object} dtos.ErrorResponse  "Forbidden - User does not have Doctor role"
// @Failure      404 {object} commonsDtos.ErrorResponse "Not Found - Medical history item not found"
// @Failure      500 {object} commonsDtos.ErrorResponse "Internal Server Error – Failed to retrieve item"
// @Security     BearerAuth
// @Router       /med/history/{patient_id} [get]
func (ctrl *MedController) GetMedicalHistoryItemByPatientID(c *gin.Context) {
	patientIDStr := c.Param("patient_id")
	patientID, err := uuid.Parse(patientIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, commonsDtos.ErrorResponse{
			Message: "Invalid patient ID",
		})
		return
	}

	patient, err := ctrl.service.GetMedicalHistoryItemByPatientID(patientID)
	if err != nil {
		var medHistItemNotFoundError *medErrors.MedicalHistoryItemNotFoundError

		if errors.As(err, &medHistItemNotFoundError) {
			c.JSON(http.StatusNotFound, commonsDtos.ErrorResponse{
				Message: err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, commonsDtos.ErrorResponse{
				Message: "Failed to find medical history item",
			})
		}
		return
	}

	c.JSON(http.StatusOK, patient)
}

func (ctrl *MedController) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/med")
	routes.Use(middleware.Authenticate())
	{
		routes.GET(
			"/disease/:name",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetBySubstring,
		)
		routes.GET(
			"/patient/:national_id",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetByNationalID,
		)
		routes.GET(
			"/drug/recommend/:disease",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetRecommendedDrugs,
		)
		routes.GET(
			"/drug/:name",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.GetDrugsByName,
		)
		routes.POST(
			"/prescribe",
			middleware.Authorize(commonsModels.Doctor),
			ctrl.CreatePrescription,
		)
		routes.GET(
			"/history/:patient_id",
			middleware.Authorize(commonsModels.Patient),
			ctrl.GetMedicalHistoryItemByPatientID,
		)
	}
}
