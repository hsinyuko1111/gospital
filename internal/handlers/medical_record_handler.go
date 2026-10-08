package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hsinyuko1111/gospital/internal/models"
	"github.com/hsinyuko1111/gospital/internal/services"
)

// MedicalRecordHandler handles medical record HTTP requests
type MedicalRecordHandler struct {
	medicalRecordService *services.MedicalRecordService
}

// NewMedicalRecordHandler creates a new medical record handler
func NewMedicalRecordHandler(medicalRecordService *services.MedicalRecordService) *MedicalRecordHandler {
	return &MedicalRecordHandler{
		medicalRecordService: medicalRecordService,
	}
}

// CreateMedicalRecord godoc
// @Summary Create a new medical record
// @Description Creates a new medical record for a patient (DynamoDB)
// @Tags Medical Records
// @Accept json
// @Produce json
// @Param record body models.MedicalRecordCreate true "Medical record to create"
// @Success 201 {object} models.MedicalRecord
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/medical-records [post]
func (h *MedicalRecordHandler) CreateMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()

	var record models.MedicalRecordCreate
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid input data",
				Details: []string{err.Error()},
			},
		})
		return
	}

	createdRecord, err := h.medicalRecordService.CreateMedicalRecord(ctx, &record)
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
		if err == services.ErrDoctorNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "DOCTOR_NOT_FOUND",
					Message: "Doctor not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create medical record",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, createdRecord)
}

// GetMedicalRecordByID godoc
// @Summary Get specific medical record
// @Description Retrieves a specific medical record by ID (DynamoDB)
// @Tags Medical Records
// @Accept json
// @Produce json
// @Param recordId path string true "Unique medical record identifier"
// @Success 200 {object} models.MedicalRecord
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/medical-records/{recordId} [get]
func (h *MedicalRecordHandler) GetMedicalRecordByID(c *gin.Context) {
	ctx := c.Request.Context()
	recordID := c.Param("recordId")

	record, err := h.medicalRecordService.GetMedicalRecordByID(ctx, recordID)
	if err != nil {
		if err == services.ErrMedicalRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "MEDICAL_RECORD_NOT_FOUND",
					Message: "Medical record not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to retrieve medical record",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, record)
}

// UpdateMedicalRecord godoc
// @Summary Update medical record
// @Description Updates an existing medical record (DynamoDB)
// @Tags Medical Records
// @Accept json
// @Produce json
// @Param recordId path string true "Unique medical record identifier"
// @Param updates body models.MedicalRecordUpdate true "Medical record updates"
// @Success 200 {object} models.MedicalRecord
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/medical-records/{recordId} [put]
func (h *MedicalRecordHandler) UpdateMedicalRecord(c *gin.Context) {
	ctx := c.Request.Context()
	recordID := c.Param("recordId")

	var updates models.MedicalRecordUpdate
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

	updatedRecord, err := h.medicalRecordService.UpdateMedicalRecord(ctx, recordID, &updates)
	if err != nil {
		if err == services.ErrMedicalRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "MEDICAL_RECORD_NOT_FOUND",
					Message: "Medical record not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to update medical record",
				Details: []string{err.Error()},
			},
		})
		return
	}

	c.JSON(http.StatusOK, updatedRecord)
}
