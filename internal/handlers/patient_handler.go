package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// PatientHandler handles patient HTTP requests
type PatientHandler struct {
	patientService *services.PatientService
}

// NewPatientHandler creates a new patient handler
func NewPatientHandler(patientService *services.PatientService) *PatientHandler {
	return &PatientHandler{
		patientService: patientService,
	}
}

// CreatePatient godoc
// @Summary Register a new patient
// @Description Creates a new patient record in the system (MySQL)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patient body models.PatientCreate true "Patient to register"
// @Success 201 {object} models.Patient
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /api/v1/patients [post]
func (h *PatientHandler) CreatePatient(c *gin.Context) {
	ctx := c.Request.Context()

	var patient models.PatientCreate
	if err := c.ShouldBindJSON(&patient); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	createdPatient, err := h.patientService.CreatePatient(ctx, &patient)
	if err != nil {
		if err == services.ErrPatientAlreadyExists {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_ALREADY_EXISTS",
					Message: "Patient already exists",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create patient",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, createdPatient)
}

// GetPatientByID godoc
// @Summary Get patient details
// @Description Retrieves detailed information about a specific patient (MySQL)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patientId path string true "Unique patient identifier"
// @Success 200 {object} models.Patient
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/patients/{patientId} [get]
func (h *PatientHandler) GetPatientByID(c *gin.Context) {
	ctx := c.Request.Context()
	patientID := c.Param("patientId")

	patient, err := h.patientService.GetPatientByID(ctx, patientID)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve patient",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// ListPatients godoc
// @Summary List or search patients
// @Description Retrieves a list of patients with optional filters (MySQL)
// @Tags Patients
// @Accept json
// @Produce json
// @Param name query string false "Filter by patient name"
// @Param email query string false "Filter by email"
// @Param phone query string false "Filter by phone number"
// @Param status query string false "Filter by patient status" Enums(active, inactive)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(20)
// @Success 200 {object} object{data=[]models.Patient,pagination=models.Pagination}
// @Failure 400 {object} models.ErrorResponse
// @Router /api/v1/patients [get]
func (h *PatientHandler) ListPatients(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	filters := make(map[string]interface{})
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if phone := c.Query("phone"); phone != "" {
		filters["phone"] = phone
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	patients, pagination, err := h.patientService.ListPatients(ctx, filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve patients",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       patients,
		"pagination": pagination,
	})
}

// UpdatePatient godoc
// @Summary Update patient information
// @Description Updates existing patient information (MySQL)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patientId path string true "Unique patient identifier"
// @Param updates body models.PatientUpdate true "Patient updates"
// @Success 200 {object} models.Patient
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/patients/{patientId} [put]
func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	ctx := c.Request.Context()
	patientID := c.Param("patientId")

	var updates models.PatientUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	updatedPatient, err := h.patientService.UpdatePatient(ctx, patientID, &updates)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update patient",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedPatient)
}

// DeactivatePatient godoc
// @Summary Deactivate a patient
// @Description Marks a patient as inactive in the system (MySQL)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patientId path string true "Unique patient identifier"
// @Success 200 {object} object{message=string,patientId=string}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/patients/{patientId} [delete]
func (h *PatientHandler) DeactivatePatient(c *gin.Context) {
	ctx := c.Request.Context()
	patientID := c.Param("patientId")

	err := h.patientService.DeactivatePatient(ctx, patientID)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to deactivate patient",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Patient deactivated successfully",
		"patientId": patientID,
	})
}

// GetPatientMedicalRecords godoc
// @Summary Get patient's medical history
// @Description Retrieves all medical records for a specific patient (DynamoDB)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patientId path string true "Unique patient identifier"
// @Param startDate query string false "Filter records from this date" Format(date)
// @Param endDate query string false "Filter records until this date" Format(date)
// @Success 200 {object} object{patientId=string,records=[]models.MedicalRecord}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/patients/{patientId}/medical-records [get]
func (h *PatientHandler) GetPatientMedicalRecords(c *gin.Context) {
	ctx := c.Request.Context()
	patientID := c.Param("patientId")

	var startDate, endDate *string
	if sd := c.Query("startDate"); sd != "" {
		startDate = &sd
	}
	if ed := c.Query("endDate"); ed != "" {
		endDate = &ed
	}

	records, err := h.patientService.GetPatientMedicalRecords(ctx, patientID, startDate, endDate)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve medical records",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patientId": patientID,
		"records":   records,
	})
}

// GetPatientInvoices godoc
// @Summary Get patient's invoices
// @Description Retrieves all billing invoices for a specific patient (DynamoDB)
// @Tags Patients
// @Accept json
// @Produce json
// @Param patientId path string true "Unique patient identifier"
// @Param status query string false "Filter by payment status" Enums(pending, paid, overdue, cancelled)
// @Success 200 {object} object{patientId=string,invoices=[]models.Invoice}
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/patients/{patientId}/invoices [get]
func (h *PatientHandler) GetPatientInvoices(c *gin.Context) {
	ctx := c.Request.Context()
	patientID := c.Param("patientId")

	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}

	invoices, err := h.patientService.GetPatientInvoices(ctx, patientID, status)
	if err != nil {
		if err == services.ErrPatientNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PATIENT_NOT_FOUND",
					Message: "Patient not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve invoices",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"patientId": patientID,
		"invoices":  invoices,
	})
}
